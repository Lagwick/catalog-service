package sproduct

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Lagwick/catalog-service/internal/app/entity"
	"github.com/Lagwick/catalog-service/internal/app/repository/mocks"
	"github.com/Lagwick/catalog-service/internal/pkg/testutil"
)

type createProductSuite struct {
	suite.Suite
	svc          *svc
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *createProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.svc = &svc{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestCreateProductSuite(t *testing.T) {
	suite.Run(t, new(createProductSuite))
}

func (s *createProductSuite) TestCreate() {
	type args struct {
		req entity.RequestProductCreate
	}
	type want struct {
		err error
	}

	categoryGUID := uuid.New()

	listErr := errors.New("list error")
	createErr := errors.New("create error")

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Test Product",
					Description:  testutil.PtrString("A test product"),
					Price:        1000,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: nil},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
					).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUID(s.ctx, args.req.CategoryGUID).
					Return(entity.Category{GUID: categoryGUID}, nil).
					Once()

				s.productRepo.EXPECT().
					Create(s.ctx, mock.MatchedBy(func(p entity.Product) bool {
						return p.Name == args.req.Name &&
							p.Description == args.req.Description &&
							p.Price == args.req.Price &&
							p.CategoryGUID == args.req.CategoryGUID
					})).
					Return(nil).
					Once()
			},
		},
		{
			name: "already exists",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Existing Product",
					Price:        500,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: entity.ErrAlreadyExists},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
					).
					Return([]entity.Product{{Name: "Existing Product"}}, nil).
					Once()
			},
		},
		{
			name: "list error",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Test Product",
					Price:        100,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{
				err: listErr,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
					).
					Return(nil, listErr).
					Once()
			},
		},
		{
			name: "create error",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Test Product",
					Price:        100,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{
				err: createErr,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
					).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUID(s.ctx, args.req.CategoryGUID).
					Return(entity.Category{GUID: args.req.CategoryGUID}, nil).
					Once()

				s.productRepo.EXPECT().
					Create(s.ctx, mock.Anything).
					Return(createErr).
					Once()
			},
		},
		{
			name: "category not found",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "New Product",
					Price:        1000,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
					).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUID(s.ctx, args.req.CategoryGUID).
					Return(entity.Category{}, entity.ErrNotFound).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			result, err := s.svc.Create(s.ctx, tc.args.req)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				return
			} else {
				s.NoError(err)
				s.NotEmpty(result.GUID)
				s.Equal(tc.args.req.Name, result.Name)
				s.Equal(tc.args.req.Description, result.Description)
				s.Equal(tc.args.req.Price, result.Price)
				s.Equal(tc.args.req.CategoryGUID, result.CategoryGUID)
			}
		})
	}
}

type getByGUIDProductSuite struct {
	suite.Suite
	svc          *svc
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *getByGUIDProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.svc = &svc{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestGetByGUIDProductSuite(t *testing.T) {
	suite.Run(t, new(getByGUIDProductSuite))
}

func (s *getByGUIDProductSuite) TestGetByGUID() {
	type args struct {
		guid uuid.UUID
	}

	type want struct {
		product entity.Product
		err     error
	}

	productGUID := uuid.New()

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				guid: productGUID,
			},
			want: want{
				product: entity.Product{
					GUID: productGUID,
					Name: "Test Product",
				},
				err: nil,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(entity.Product{
						GUID: productGUID,
						Name: "Test Product",
					}, nil).
					Once()
			},
		},
		{
			name: "not found",
			args: args{
				guid: productGUID,
			},
			want: want{
				err: entity.ErrNotFound,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(entity.Product{}, entity.ErrNotFound).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			result, err := s.svc.GetByGUID(s.ctx, tc.args.guid)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Equal(entity.Product{}, result)
			} else {
				s.NoError(err)
				s.Equal(tc.want.product, result)
			}
		})
	}
}

type deleteProductSuite struct {
	suite.Suite
	svc          *svc
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *deleteProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())

	s.svc = &svc{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestDeleteProductSuite(t *testing.T) {
	suite.Run(t, new(deleteProductSuite))
}

func (s *deleteProductSuite) TestDelete() {
	type args struct {
		guid uuid.UUID
	}

	type want struct {
		err error
	}

	productGUID := uuid.New()
	deleteErr := errors.New("delete failed")

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				guid: productGUID,
			},
			want: want{},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(entity.Product{GUID: args.guid}, nil).
					Once()

				s.productRepo.EXPECT().
					Delete(s.ctx, args.guid).
					Return(nil).
					Once()
			},
		},
		{
			name: "not found",
			args: args{
				guid: productGUID,
			},
			want: want{
				err: entity.ErrNotFound,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(entity.Product{}, entity.ErrNotFound).
					Once()
			},
		},
		{
			name: "delete error",
			args: args{
				guid: productGUID,
			},
			want: want{
				err: deleteErr,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(entity.Product{GUID: args.guid}, nil).
					Once()

				s.productRepo.EXPECT().
					Delete(s.ctx, args.guid).
					Return(deleteErr).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			err := s.svc.Delete(s.ctx, tc.args.guid)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
			} else {
				s.NoError(err)
			}
		})
	}
}

type listProductSuite struct {
	suite.Suite
	svc          *svc
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *listProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())

	s.svc = &svc{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestListProductSuite(t *testing.T) {
	suite.Run(t, new(listProductSuite))
}

func (s *listProductSuite) TestList() {
	type want struct {
		products []entity.Product
		err      error
	}

	testCases := []struct {
		name    string
		want    want
		prepare func()
	}{
		{
			name: "success",
			want: want{
				products: []entity.Product{
					{
						Name:  "iPhone",
						Price: 1000,
					},
					{
						Name:  "MacBook",
						Price: 2000,
					},
				},
			},
			prepare: func() {
				s.productRepo.EXPECT().
					List(s.ctx, (*string)(nil), (*uuid.UUID)(nil)).
					Return([]entity.Product{
						{
							Name:  "iPhone",
							Price: 1000,
						},
						{
							Name:  "MacBook",
							Price: 2000,
						},
					}, nil).
					Once()
			},
		},
		{
			name: "empty result",
			want: want{
				products: []entity.Product{},
			},
			prepare: func() {
				s.productRepo.EXPECT().
					List(s.ctx, (*string)(nil), (*uuid.UUID)(nil)).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare()

			result, err := s.svc.List(s.ctx)

			s.NoError(err)
			s.Len(result, len(tc.want.products))

			for i := range tc.want.products {
				s.Equal(tc.want.products[i].Name, result[i].Name)
				s.Equal(tc.want.products[i].Price, result[i].Price)
			}
		})
	}
}

type updateProductSuite struct {
	suite.Suite
	svc          *svc
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *updateProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())

	s.svc = &svc{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestUpdateProductSuite(t *testing.T) {
	suite.Run(t, new(updateProductSuite))
}

func (s *updateProductSuite) TestUpdate() {
	type args struct {
		guid uuid.UUID
		req  entity.RequestProductUpdate
	}

	type want struct {
		err error
	}

	productGUID := uuid.New()
	oldCategoryGUID := uuid.New()
	newCategoryGUID := uuid.New()

	existingProduct := entity.Product{
		GUID:         productGUID,
		Name:         "Old Product",
		Description:  testutil.PtrString("old description"),
		Price:        10,
		CategoryGUID: oldCategoryGUID,
	}

	updateListErr := errors.New("list error")
	updateErr := errors.New("update error")

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "full update",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name:         "New Product",
					Description:  testutil.PtrString("new description"),
					Price:        100,
					CategoryGUID: newCategoryGUID,
				},
			},
			want: want{},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(existingProduct, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUID(s.ctx, newCategoryGUID).
					Return(entity.Category{GUID: newCategoryGUID}, nil).
					Once()

				s.productRepo.EXPECT().
					Update(
						s.ctx,
						mock.MatchedBy(func(p entity.Product) bool {
							return p.Name == args.req.Name &&
								p.Description == args.req.Description &&
								p.Price == args.req.Price &&
								p.CategoryGUID == args.req.CategoryGUID
						}),
					).
					Return(nil).
					Once()
			},
		},
		{
			name: "same name same category",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name:         existingProduct.Name,
					Description:  existingProduct.Description,
					Price:        999,
					CategoryGUID: existingProduct.CategoryGUID,
				},
			},
			want: want{},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(existingProduct, nil).
					Once()

				s.productRepo.EXPECT().
					Update(
						s.ctx,
						mock.MatchedBy(func(p entity.Product) bool {
							return p.Name == args.req.Name &&
								p.Price == args.req.Price &&
								p.CategoryGUID == args.req.CategoryGUID
						}),
					).
					Return(nil).
					Once()
			},
		},
		{
			name: "list error",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name:         "New Product",
					CategoryGUID: oldCategoryGUID,
				},
			},
			want: want{
				err: updateListErr,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(existingProduct, nil).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
					).
					Return(nil, updateListErr).
					Once()
			},
		},
		{
			name: "update error",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name:         "New Product",
					Description:  testutil.PtrString("new description"),
					Price:        100,
					CategoryGUID: newCategoryGUID,
				},
			},
			want: want{
				err: updateErr,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(existingProduct, nil).
					Once()

				s.productRepo.EXPECT().
					List(
						s.ctx,
						&args.req.Name,
						(*uuid.UUID)(nil),
					).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUID(s.ctx, newCategoryGUID).
					Return(entity.Category{GUID: newCategoryGUID}, nil).
					Once()

				s.productRepo.EXPECT().
					Update(
						s.ctx,
						mock.Anything,
					).
					Return(updateErr).
					Once()
			},
		},
		{
			name: "not found",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name: "New Product",
				},
			},
			want: want{
				err: entity.ErrNotFound,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(entity.Product{}, entity.ErrNotFound).
					Once()
			},
		},
		{
			name: "duplicate name",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name:         "Duplicate Product",
					CategoryGUID: oldCategoryGUID,
				},
			},
			want: want{
				err: entity.ErrAlreadyExists,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(existingProduct, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil)).
					Return([]entity.Product{
						{Name: "Duplicate Product"},
					}, nil).
					Once()
			},
		},
		{
			name: "category not found",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name:         "New Product",
					CategoryGUID: newCategoryGUID,
				},
			},
			want: want{
				err: entity.ErrNotFound,
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					InsideTx(s.ctx, mock.AnythingOfType("func(context.Context) error")).
					RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					}).
					Once()

				s.productRepo.EXPECT().
					GetByGUID(s.ctx, args.guid).
					Return(existingProduct, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUID(s.ctx, newCategoryGUID).
					Return(entity.Category{}, entity.ErrNotFound).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			_, err := s.svc.Update(
				s.ctx,
				tc.args.guid,
				tc.args.req,
			)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	productRepo := mocks.NewMockProduct(t)
	categoryRepo := mocks.NewMockCategory(t)

	s := NewService(productRepo, categoryRepo)

	require.NotNil(t, s)
}
