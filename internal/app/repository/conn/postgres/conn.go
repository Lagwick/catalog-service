package rcpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"

	"github.com/Lagwick/catalog-service/internal/app/config/section"
	"github.com/Lagwick/catalog-service/migration"
)

type (
	Client struct {
		_bunDB
		rawBunDB *bun.DB

		cfg section.RepositoryPostgres
	}

	_bunDB = bun.IDB
)

func (c *Client) GetRawBunDB() *bun.DB {
	return c.rawBunDB
}

func NewConn(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.Username, cfg.Password),
		Host:   cfg.Address,
		Path:   cfg.Name,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	dsn := u.String()

	connector := pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		pgdriver.WithDialTimeout(cfg.ReadTimeout.Duration),
		pgdriver.WithReadTimeout(cfg.ReadTimeout.Duration),
		pgdriver.WithWriteTimeout(cfg.WriteTimeout.Duration),
	)

	sqlDB := sql.OpenDB(connector)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(10)

	bunDB := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := bunDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}

	return &Client{
		_bunDB:   bunDB,
		rawBunDB: bunDB,
		cfg:      cfg,
	}, nil
}

func (c *Client) Migrate(ctx context.Context) (oldVer, newVer int64, err error) {
	migrations := migrate.NewMigrations()

	if err = migrations.Discover(migration.Postgres); err != nil {
		return 0, 0, fmt.Errorf("discover migrations: %w", err)
	}

	opts := []migrate.MigratorOption{
		migrate.WithTableName(c.cfg.MigrationTable),
		migrate.WithLocksTableName(c.cfg.MigrationTable + "_lock"),
		migrate.WithMarkAppliedOnSuccess(true),
	}

	migrator := migrate.NewMigrator(
		c.rawBunDB,
		migrations,
		opts...,
	)

	if err = migrator.Init(ctx); err != nil {
		return 0, 0, fmt.Errorf("init migrator: %w", err)
	}

	applied, err := migrator.AppliedMigrations(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("get applied migrations: %w", err)
	}

	if len(applied) > 0 {
		parts := strings.SplitN(applied[0].Name, "_", 2)
		oldVer, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("parse old migration version %q: %w", applied[0].Name, err)
		}
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return oldVer, oldVer, fmt.Errorf("apply migrations: %w", err)
	}

	newVer = oldVer
	if group != nil {
		for _, m := range group.Migrations {
			parts := strings.SplitN(m.Name, "_", 2)
			ver, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return oldVer, oldVer, fmt.Errorf("parse migration version %q: %w", m.Name, err)
			}
			if ver > newVer {
				newVer = ver
			}
		}
	}

	return oldVer, newVer, nil
}
