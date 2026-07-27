package sproduct

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/Lagwick/catalog-service/internal/app/entity"
	"github.com/Lagwick/catalog-service/internal/app/repository"
	"github.com/Lagwick/catalog-service/internal/app/service"
)

type srv struct {
	repoProduct       repository.Product
	repoCategory      repository.Category
	repoTransactional repository.Transactional
}

func NewService(
	repoProduct repository.Product,
	repoCategory repository.Category,
	repoTransactional repository.Transactional,
) service.Product {
	return &srv{
		repoProduct:       repoProduct,
		repoCategory:      repoCategory,
		repoTransactional: repoTransactional,
	}
}

func (s *srv) Create(ctx context.Context, req entity.RequestProductCreate) (entity.Product, error) {
	var product entity.Product

	err := s.repoTransactional.InsideTx(ctx, func(ctx context.Context) error {
		existing, err := s.repoProduct.List(ctx, &req.Name, nil, nil, nil)
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return entity.ErrAlreadyExists
		}

		categories, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{req.CategoryGUID})
		if err != nil {
			return err
		}
		if len(categories) == 0 {
			return entity.ErrNotFound
		}

		now := time.Now()
		product = entity.Product{
			GUID:         uuid.New(),
			Name:         req.Name,
			Description:  req.Description,
			Price:        req.Price,
			CategoryGUID: req.CategoryGUID,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		return s.repoProduct.Create(ctx, product)
	})
	if err != nil {
		return entity.Product{}, err
	}

	return product, nil
}

func (s *srv) GetByGUIDs(ctx context.Context, guids []uuid.UUID) ([]entity.Product, error) {
	return s.repoProduct.GetByGUIDs(ctx, guids)
}

func (s *srv) Update(ctx context.Context, guid uuid.UUID, req entity.RequestProductUpdate) (entity.Product, error) {
	var product entity.Product

	err := s.repoTransactional.InsideTx(ctx, func(ctx context.Context) error {
		products, err := s.repoProduct.GetByGUIDs(ctx, []uuid.UUID{guid})
		if err != nil {
			return err
		}
		if len(products) == 0 {
			return entity.ErrNotFound
		}
		product = products[0]

		if req.Name != "" {
			existing, err := s.repoProduct.List(ctx, &req.Name, nil, nil, nil)
			if err != nil {
				return err
			}
			for _, e := range existing {
				if e.GUID != guid {
					return entity.ErrAlreadyExists
				}
			}
			product.Name = req.Name
		}

		if req.Description != nil {
			product.Description = req.Description
		}

		if req.Price > 0 {
			product.Price = req.Price
		}

		if req.CategoryGUID != uuid.Nil {
			categories, err := s.repoCategory.GetByGUIDs(ctx, []uuid.UUID{req.CategoryGUID})
			if err != nil {
				return err
			}
			if len(categories) == 0 {
				return entity.ErrNotFound
			}

			product.CategoryGUID = req.CategoryGUID
		}

		product.UpdatedAt = time.Now()

		return s.repoProduct.Update(ctx, product)
	})
	if err != nil {
		return entity.Product{}, err
	}

	return product, nil
}

func (s *srv) Delete(ctx context.Context, guid uuid.UUID) error {
	return s.repoTransactional.InsideTx(ctx, func(ctx context.Context) error {
		products, err := s.repoProduct.GetByGUIDs(ctx, []uuid.UUID{guid})
		if err != nil {
			return err
		}
		if len(products) == 0 {
			return entity.ErrNotFound
		}

		return s.repoProduct.Delete(ctx, guid)
	})
}

func (s *srv) List(ctx context.Context, req entity.RequestProductList) ([]entity.Product, error) {
	return s.repoProduct.List(ctx, nil, req.CategoryGUID, req.MinPrice, req.MaxPrice)
}
