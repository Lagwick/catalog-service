package rcpostgres

import (
	"context"

	"github.com/uptrace/bun"
)

type contextKeyTx struct{}

func getTxFromContext(ctx context.Context) bun.Tx {
	v, ok := ctx.Value(contextKeyTx{}).(bun.Tx)
	if ok {
		return v
	}
	return bun.Tx{}
}

func setTxToContext(ctx context.Context, tx bun.Tx) context.Context {
	return context.WithValue(ctx, contextKeyTx{}, tx)
}
