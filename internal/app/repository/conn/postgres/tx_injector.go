package rcpostgres

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type txInjector struct {
	fallback bun.IDB
	sqlDB    *sql.DB
}

type sqlConn interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func newTxInjector(db bun.IDB, sqlDB *sql.DB) bun.IDB {
	return &txInjector{fallback: db, sqlDB: sqlDB}
}

func (x *txInjector) getIDB(ctx context.Context) bun.IDB {
	tx := getTxFromContext(ctx)
	if tx.Tx != nil {
		return tx
	}
	return x.fallback
}

func (x *txInjector) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return x.raw(ctx).ExecContext(ctx, query, args...)
}

func (x *txInjector) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return x.raw(ctx).QueryContext(ctx, query, args...)
}

func (x *txInjector) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return x.raw(ctx).QueryRowContext(ctx, query, args...)
}

func (x *txInjector) NewSelect() *bun.SelectQuery {
	q := x.fallback.NewSelect().Conn(x)
	return q
}

func (x *txInjector) NewInsert() *bun.InsertQuery {
	return x.fallback.NewInsert().Conn(x)
}

func (x *txInjector) NewUpdate() *bun.UpdateQuery {
	return x.fallback.NewUpdate().Conn(x)
}

func (x *txInjector) NewDelete() *bun.DeleteQuery {
	return x.fallback.NewDelete().Conn(x)
}

func (x *txInjector) NewMerge() *bun.MergeQuery {
	return x.fallback.NewMerge().Conn(x)
}

func (x *txInjector) NewRaw(query string, args ...any) *bun.RawQuery {
	return x.fallback.NewRaw(query, args...).Conn(x)
}

func (x *txInjector) NewValues(model any) *bun.ValuesQuery {
	return x.fallback.NewValues(model).Conn(x)
}

func (x *txInjector) NewCreateTable() *bun.CreateTableQuery {
	return x.fallback.NewCreateTable().Conn(x)
}

func (x *txInjector) NewDropTable() *bun.DropTableQuery {
	return x.fallback.NewDropTable().Conn(x)
}

func (x *txInjector) NewCreateIndex() *bun.CreateIndexQuery {
	return x.fallback.NewCreateIndex().Conn(x)
}

func (x *txInjector) NewDropIndex() *bun.DropIndexQuery {
	return x.fallback.NewDropIndex().Conn(x)
}

func (x *txInjector) NewTruncateTable() *bun.TruncateTableQuery {
	return x.fallback.NewTruncateTable().Conn(x)
}

func (x *txInjector) NewAddColumn() *bun.AddColumnQuery {
	return x.fallback.NewAddColumn().Conn(x)
}

func (x *txInjector) NewDropColumn() *bun.DropColumnQuery {
	return x.fallback.NewDropColumn().Conn(x)
}

func (x *txInjector) Dialect() schema.Dialect {
	return x.fallback.Dialect()
}

func (x *txInjector) BeginTx(ctx context.Context, opts *sql.TxOptions) (bun.Tx, error) {
	return x.getIDB(ctx).BeginTx(ctx, opts)
}

func (x *txInjector) RunInTx(ctx context.Context, opts *sql.TxOptions, f func(ctx context.Context, tx bun.Tx) error) error {
	return x.getIDB(ctx).RunInTx(ctx, opts, f)
}

func (x *txInjector) raw(ctx context.Context) sqlConn {
	tx := getTxFromContext(ctx)
	if tx.Tx != nil {
		return tx.Tx
	}

	return x.sqlDB
}
