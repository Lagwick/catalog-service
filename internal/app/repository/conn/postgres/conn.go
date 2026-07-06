package rcpostgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strconv"
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

func NewClient(ctx context.Context, cfg section.RepositoryPostgres) (*Client, error) {
	var u url.URL
	u.Scheme = "postgres"
	u.Host = cfg.Address
	u.User = url.UserPassword(cfg.Username, cfg.Password)
	u.Path = cfg.Name

	args := make(url.Values)
	args.Set("sslmode", "disable")
	u.RawQuery = args.Encode()

	dsn := u.String()

	log.Printf("Initializing PostgreSQL connection read_timeout=%s write_timeout=%s", cfg.ReadTimeout, cfg.WriteTimeout)

	connector := pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		pgdriver.WithReadTimeout(cfg.ReadTimeout.Duration),
		pgdriver.WithWriteTimeout(cfg.WriteTimeout.Duration),
	)

	sqlDB := sql.OpenDB(connector)
	sqlDB.SetMaxOpenConns(10)

	bunDB := bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns())

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	log.Printf("PostgreSQL connection established")

	return &Client{
		_bunDB:   newTxInjector(bunDB),
		rawBunDB: bunDB,
		cfg:      cfg,
	}, nil
}

func (c *Client) Migrate(ctx context.Context) (oldVer, newVer int64, err error) {
	migrations := migrate.NewMigrations()

	if err = migrations.Discover(migration.Postgres); err != nil {
		return 0, 0, fmt.Errorf("failed to discover migrations: %w", err)
	}

	opts := []migrate.MigratorOption{
		migrate.WithTableName(c.cfg.MigrationTable),
		migrate.WithLocksTableName(c.cfg.MigrationTable + "_lock"),
		migrate.WithMarkAppliedOnSuccess(true),
	}

	m := migrate.NewMigrator(c.rawBunDB, migrations, opts...)

	if err = m.Init(ctx); err != nil {
		return 0, 0, fmt.Errorf("failed to init migrator: %w", err)
	}

	applied, err := m.AppliedMigrations(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) > 0 {
		oldVer, _ = strconv.ParseInt(applied[0].Name, 10, 64)
	}

	newVer = oldVer

	mgg, err := m.Migrate(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to run migrations: %w", err)
	}

	for _, mg := range mgg.Migrations {
		v, _ := strconv.ParseInt(mg.Name, 10, 64)
		if v > newVer {
			newVer = v
		}
	}

	return oldVer, newVer, nil
}

func (c *Client) InsideTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return c.rawBunDB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(setTxToContext(ctx, tx))
	})
}
