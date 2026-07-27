package scategory

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Lagwick/catalog-service/internal/app/entity"
	"github.com/Lagwick/catalog-service/internal/app/repository"
	"github.com/Lagwick/catalog-service/internal/app/service"
)

type srv struct {
	repoCategory      repository.Category
	repoProduct       repository.Product
	repoTransactional repository.Transactional
}

func NewService(
	repoCategory repository.Category,
	repoProduct repository.Product,
	repoTransactional repository.Transactional,
) service.Category {
	return &srv{
		repoCategory:      repoCategory,
		repoProduct:       repoProduct,
		repoTransactional: repoTransactional,
	}
}

func (s *srv) Create(ctx context.Context, req entity.RequestCategoryCreate) (entity.Category, error) {
	var category entity.Category

	err := s.repoTransactional.InsideTx(ctx, func(ctx context.Context) error {
		existing, err := s.repoCategory.List(ctx, &req.Name)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return entity.ErrAlreadyExists
		}

		now := time.Now()
		category = entity.Category{
			GUID:      uuid.New(),
			Name:      req.Name,
			CreatedAt: now,
			UpdatedAt: now,
		}

		return s.repoCategory.Create(ctx, category)
	})
	if err != nil {
		return entity.Category{}, err
	}

	return category, nil
}

func (s *srv) GetByGUIDs(ctx context.Context, guids []uuid.UUID) ([]entity.Category, error) {
	return s.repoCategory.GetByGUIDs(ctx, guids)
}

func (s *srv) Update(ctx context.Context, guid uuid.UUID, req entity.RequestCategoryUpdate) (entity.Category, error) {
	var category entity.Category

	err := s.repoTransactional.InsideTx(ctx, func(ctx context.Context) error {
		categories, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{guid})
		if err != nil {
			return err
		}
		if len(categories) == 0 {
			return entity.ErrNotFound
		}
		category = categories[0]

		if req.Name != "" {
			existing, err := s.repoCategory.List(ctx, &req.Name)
			if err != nil {
				return err
			}

			for _, e := range existing {
				if e.GUID != guid {
					return entity.ErrAlreadyExists
				}
			}

			category.Name = req.Name
		}

		category.UpdatedAt = time.Now()

		return s.repoCategory.Update(ctx, category)
	})
	if err != nil {
		return entity.Category{}, err
	}

	return category, nil
}

func (s *srv) Delete(ctx context.Context, guid uuid.UUID) error {
	return s.repoTransactional.InsideTx(ctx, func(ctx context.Context) error {
		categories, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{guid})
		if err != nil {
			return err
		}
		if len(categories) == 0 {
			return entity.ErrNotFound
		}

		products, err := s.repoProduct.List(ctx, nil, &guid, nil, nil)
		if err != nil {
			return err
		}
		if len(products) > 0 {
			return entity.ErrCategoryHasProducts
		}

		return s.repoCategory.Delete(ctx, guid)
	})
}

func (s *srv) List(ctx context.Context) ([]entity.Category, error) {
	return s.repoCategory.List(ctx, nil)
}
