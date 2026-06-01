package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/Lagwick/catalog-service/internal/app/entity"
)

type (
	Category interface {
		Transactional
		Create(ctx context.Context, category entity.Category) error
		GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Category, error)
		Update(ctx context.Context, category entity.Category) error
		Delete(ctx context.Context, guid uuid.UUID) error
		List(ctx context.Context, name *string) ([]entity.Category, error)
	}

	Product interface {
		Transactional
		Create(ctx context.Context, product entity.Product) error
		GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Product, error)
		Update(ctx context.Context, product entity.Product) error
		Delete(ctx context.Context, guid uuid.UUID) error
		List(ctx context.Context, name *string, categoryGUID *uuid.UUID) ([]entity.Product, error)
	}

	Transactional interface {
		InsideTx(ctx context.Context, fn func(ctx context.Context) error) error
	}

	Migrate interface {
		Migrate(ctx context.Context) (oldVer, newVer int64, err error)
	}
)
