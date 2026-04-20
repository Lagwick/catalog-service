package service

import (
	"context"

	"github.com/Lagwick/catalog-service/internal/app/entity"
	"github.com/google/uuid"
)

type (
	Category interface {
		Create(ctx context.Context, req entity.RequestCategoryCreate) (entity.Category, error)
		GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Category, error)
		Update(ctx context.Context, guid uuid.UUID, req entity.RequestCategoryUpdate) (entity.Category, error)
		Delete(ctx context.Context, guid uuid.UUID) error
		List(ctx context.Context) ([]entity.Category, error)
	}

	Product interface {
		Create(ctx context.Context, req entity.RequestProductCreate) (entity.Product, error)
		GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Product, error)
		Update(ctx context.Context, guid uuid.UUID, req entity.RequestProductUpdate) (entity.Product, error)
		Delete(ctx context.Context, guid uuid.UUID) error
		List(ctx context.Context) ([]entity.Product, error)
	}
)
