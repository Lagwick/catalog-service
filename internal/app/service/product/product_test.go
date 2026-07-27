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

func matchGUIDs(guids ...uuid.UUID) interface{} {
	return mock.MatchedBy(func(actual []uuid.UUID) bool {
		if len(actual) != len(guids) {
			return false
		}

		for i := range actual {
			if actual[i] != guids[i] {
				return false
			}
		}

		return true
	})
}

func expectInsideTx(transactionalRepo *mocks.MockTransactional, ctx context.Context) {
	transactionalRepo.EXPECT().
		InsideTx(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		}).
		Once()
}

type createProductSuite struct {
	suite.Suite
	srv               *srv
	productRepo       *mocks.MockProduct
	categoryRepo      *mocks.MockCategory
	transactionalRepo *mocks.MockTransactional
	ctx               context.Context
}

func (s *createProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.transactionalRepo = mocks.NewMockTransactional(s.T())
	s.srv = &srv{
		repoProduct:       s.productRepo,
		repoCategory:      s.categoryRepo,
		repoTransactional: s.transactionalRepo,
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
	categoryErr := errors.New("category error")
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
			want: want{},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.req.CategoryGUID)).
					Return([]entity.Category{{GUID: args.req.CategoryGUID}}, nil).
					Once()

				s.productRepo.EXPECT().
					Create(s.ctx, mock.MatchedBy(func(p entity.Product) bool {
						return p.GUID != uuid.Nil &&
							p.Name == args.req.Name &&
							p.Description == args.req.Description &&
							p.Price == args.req.Price &&
							p.CategoryGUID == args.req.CategoryGUID &&
							!p.CreatedAt.IsZero() &&
							!p.UpdatedAt.IsZero()
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
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{{GUID: uuid.New(), Name: args.req.Name}}, nil).
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
			want: want{err: listErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return(nil, listErr).
					Once()
			},
		},
		{
			name: "category error",
			args: args{
				req: entity.RequestProductCreate{
					Name:         "Test Product",
					Price:        100,
					CategoryGUID: categoryGUID,
				},
			},
			want: want{err: categoryErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.req.CategoryGUID)).
					Return(nil, categoryErr).
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
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.req.CategoryGUID)).
					Return([]entity.Category{}, nil).
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
			want: want{err: createErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.categoryRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.req.CategoryGUID)).
					Return([]entity.Category{{GUID: args.req.CategoryGUID}}, nil).
					Once()

				s.productRepo.EXPECT().
					Create(s.ctx, mock.Anything).
					Return(createErr).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			expectInsideTx(s.transactionalRepo, s.ctx)

			tc.prepare(tc.args)

			result, err := s.srv.Create(s.ctx, tc.args.req)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Equal(entity.Product{}, result)

				return
			}

			s.NoError(err)
			s.NotEqual(uuid.Nil, result.GUID)
			s.Equal(tc.args.req.Name, result.Name)
			s.Equal(tc.args.req.Description, result.Description)
			s.Equal(tc.args.req.Price, result.Price)
			s.Equal(tc.args.req.CategoryGUID, result.CategoryGUID)
			s.False(result.CreatedAt.IsZero())
			s.False(result.UpdatedAt.IsZero())
		})
	}
}

type getByGUIDsProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *getByGUIDsProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestGetByGUIDsProductSuite(t *testing.T) {
	suite.Run(t, new(getByGUIDsProductSuite))
}

func (s *getByGUIDsProductSuite) TestGetByGUIDs() {
	type args struct {
		guids []uuid.UUID
	}
	type want struct {
		products []entity.Product
		err      error
	}

	productGUID := uuid.New()
	getErr := errors.New("get error")

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				guids: []uuid.UUID{productGUID},
			},
			want: want{
				products: []entity.Product{
					{
						GUID: productGUID,
						Name: "Test Product",
					},
				},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, args.guids).
					Return([]entity.Product{
						{
							GUID: productGUID,
							Name: "Test Product",
						},
					}, nil).
					Once()
			},
		},
		{
			name: "error",
			args: args{
				guids: []uuid.UUID{productGUID},
			},
			want: want{err: getErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, args.guids).
					Return(nil, getErr).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			result, err := s.srv.GetByGUIDs(s.ctx, tc.args.guids)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Nil(result)

				return
			}

			s.NoError(err)
			s.Equal(tc.want.products, result)
		})
	}
}

type deleteProductSuite struct {
	suite.Suite
	srv               *srv
	productRepo       *mocks.MockProduct
	categoryRepo      *mocks.MockCategory
	transactionalRepo *mocks.MockTransactional
	ctx               context.Context
}

func (s *deleteProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.transactionalRepo = mocks.NewMockTransactional(s.T())
	s.srv = &srv{
		repoProduct:       s.productRepo,
		repoCategory:      s.categoryRepo,
		repoTransactional: s.transactionalRepo,
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
	getErr := errors.New("get error")
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
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{{GUID: args.guid}}, nil).
					Once()

				s.productRepo.EXPECT().
					Delete(s.ctx, args.guid).
					Return(nil).
					Once()
			},
		},
		{
			name: "get error",
			args: args{
				guid: productGUID,
			},
			want: want{err: getErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return(nil, getErr).
					Once()
			},
		},
		{
			name: "not found",
			args: args{
				guid: productGUID,
			},
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
		{
			name: "delete error",
			args: args{
				guid: productGUID,
			},
			want: want{err: deleteErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{{GUID: args.guid}}, nil).
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
			expectInsideTx(s.transactionalRepo, s.ctx)

			tc.prepare(tc.args)

			err := s.srv.Delete(s.ctx, tc.args.guid)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)

				return
			}

			s.NoError(err)
		})
	}
}

type listProductSuite struct {
	suite.Suite
	srv          *srv
	productRepo  *mocks.MockProduct
	categoryRepo *mocks.MockCategory
	ctx          context.Context
}

func (s *listProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.srv = &srv{
		repoProduct:  s.productRepo,
		repoCategory: s.categoryRepo,
	}
}

func TestListProductSuite(t *testing.T) {
	suite.Run(t, new(listProductSuite))
}

func (s *listProductSuite) TestList() {
	type args struct {
		req entity.RequestProductList
	}
	type want struct {
		products []entity.Product
		err      error
	}

	categoryGUID := uuid.New()
	minPrice := int64(100)
	maxPrice := int64(1000)
	listErr := errors.New("list error")

	testCases := []struct {
		name    string
		args    args
		want    want
		prepare func(args args)
	}{
		{
			name: "success",
			args: args{
				req: entity.RequestProductList{
					CategoryGUID: &categoryGUID,
					MinPrice:     &minPrice,
					MaxPrice:     &maxPrice,
				},
			},
			want: want{
				products: []entity.Product{
					{Name: "iPhone", Price: 1000},
					{Name: "MacBook", Price: 2000},
				},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(s.ctx, (*string)(nil), args.req.CategoryGUID, args.req.MinPrice, args.req.MaxPrice).
					Return([]entity.Product{
						{Name: "iPhone", Price: 1000},
						{Name: "MacBook", Price: 2000},
					}, nil).
					Once()
			},
		},
		{
			name: "empty result",
			args: args{
				req: entity.RequestProductList{},
			},
			want: want{
				products: []entity.Product{},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(s.ctx, (*string)(nil), (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
		{
			name: "list error",
			args: args{
				req: entity.RequestProductList{},
			},
			want: want{err: listErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					List(s.ctx, (*string)(nil), (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return(nil, listErr).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.prepare(tc.args)

			result, err := s.srv.List(s.ctx, tc.args.req)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Nil(result)

				return
			}

			s.NoError(err)
			s.Equal(tc.want.products, result)
		})
	}
}

type updateProductSuite struct {
	suite.Suite
	srv               *srv
	productRepo       *mocks.MockProduct
	categoryRepo      *mocks.MockCategory
	transactionalRepo *mocks.MockTransactional
	ctx               context.Context
}

func (s *updateProductSuite) SetupTest() {
	s.ctx = context.Background()
	s.productRepo = mocks.NewMockProduct(s.T())
	s.categoryRepo = mocks.NewMockCategory(s.T())
	s.transactionalRepo = mocks.NewMockTransactional(s.T())
	s.srv = &srv{
		repoProduct:       s.productRepo,
		repoCategory:      s.categoryRepo,
		repoTransactional: s.transactionalRepo,
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
		product entity.Product
		err     error
	}

	productGUID := uuid.New()
	otherProductGUID := uuid.New()
	categoryGUID := uuid.New()

	existingProduct := entity.Product{
		GUID:         productGUID,
		Name:         "Old Product",
		Description:  testutil.PtrString("old description"),
		Price:        10,
		CategoryGUID: categoryGUID,
	}

	updateListErr := errors.New("list error")
	getErr := errors.New("get error")
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
					Name:        "New Product",
					Description: testutil.PtrString("new description"),
					Price:       100,
				},
			},
			want: want{
				product: entity.Product{
					GUID:         productGUID,
					Name:         "New Product",
					Description:  testutil.PtrString("new description"),
					Price:        100,
					CategoryGUID: categoryGUID,
				},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{existingProduct}, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.productRepo.EXPECT().
					Update(s.ctx, mock.MatchedBy(func(p entity.Product) bool {
						return p.GUID == args.guid &&
							p.Name == args.req.Name &&
							p.Description == args.req.Description &&
							p.Price == args.req.Price &&
							p.CategoryGUID == existingProduct.CategoryGUID &&
							!p.UpdatedAt.IsZero()
					})).
					Return(nil).
					Once()
			},
		},
		{
			name: "same name",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name:  existingProduct.Name,
					Price: 999,
				},
			},
			want: want{
				product: entity.Product{
					GUID:         productGUID,
					Name:         existingProduct.Name,
					Description:  existingProduct.Description,
					Price:        999,
					CategoryGUID: categoryGUID,
				},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{existingProduct}, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{{GUID: args.guid, Name: args.req.Name}}, nil).
					Once()

				s.productRepo.EXPECT().
					Update(s.ctx, mock.MatchedBy(func(p entity.Product) bool {
						return p.GUID == args.guid &&
							p.Name == args.req.Name &&
							p.Price == args.req.Price
					})).
					Return(nil).
					Once()
			},
		},
		{
			name: "partial update without name",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Description: testutil.PtrString("changed description"),
				},
			},
			want: want{
				product: entity.Product{
					GUID:         productGUID,
					Name:         existingProduct.Name,
					Description:  testutil.PtrString("changed description"),
					Price:        existingProduct.Price,
					CategoryGUID: categoryGUID,
				},
			},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{existingProduct}, nil).
					Once()

				s.productRepo.EXPECT().
					Update(s.ctx, mock.MatchedBy(func(p entity.Product) bool {
						return p.GUID == args.guid &&
							p.Name == existingProduct.Name &&
							p.Description == args.req.Description &&
							p.Price == existingProduct.Price
					})).
					Return(nil).
					Once()
			},
		},
		{
			name: "get error",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name: "New Product",
				},
			},
			want: want{err: getErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return(nil, getErr).
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
			want: want{err: entity.ErrNotFound},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{}, nil).
					Once()
			},
		},
		{
			name: "list error",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name: "New Product",
				},
			},
			want: want{err: updateListErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{existingProduct}, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return(nil, updateListErr).
					Once()
			},
		},
		{
			name: "duplicate name",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name: "Duplicate Product",
				},
			},
			want: want{err: entity.ErrAlreadyExists},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{existingProduct}, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{{GUID: otherProductGUID, Name: args.req.Name}}, nil).
					Once()
			},
		},
		{
			name: "update error",
			args: args{
				guid: productGUID,
				req: entity.RequestProductUpdate{
					Name:  "New Product",
					Price: 100,
				},
			},
			want: want{err: updateErr},
			prepare: func(args args) {
				s.productRepo.EXPECT().
					GetByGUIDs(s.ctx, matchGUIDs(args.guid)).
					Return([]entity.Product{existingProduct}, nil).
					Once()

				s.productRepo.EXPECT().
					List(s.ctx, &args.req.Name, (*uuid.UUID)(nil), (*int64)(nil), (*int64)(nil)).
					Return([]entity.Product{}, nil).
					Once()

				s.productRepo.EXPECT().
					Update(s.ctx, mock.Anything).
					Return(updateErr).
					Once()
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			expectInsideTx(s.transactionalRepo, s.ctx)

			tc.prepare(tc.args)

			result, err := s.srv.Update(s.ctx, tc.args.guid, tc.args.req)

			if tc.want.err != nil {
				s.ErrorIs(err, tc.want.err)
				s.Equal(entity.Product{}, result)

				return
			}

			s.NoError(err)
			s.Equal(tc.want.product.GUID, result.GUID)
			s.Equal(tc.want.product.Name, result.Name)
			s.Equal(tc.want.product.Description, result.Description)
			s.Equal(tc.want.product.Price, result.Price)
			s.Equal(tc.want.product.CategoryGUID, result.CategoryGUID)
			s.False(result.UpdatedAt.IsZero())
		})
	}
}

func TestNewService(t *testing.T) {
	productRepo := mocks.NewMockProduct(t)
	categoryRepo := mocks.NewMockCategory(t)
	transactionalRepo := mocks.NewMockTransactional(t)

	s := NewService(productRepo, categoryRepo, transactionalRepo)

	require.NotNil(t, s)
}
