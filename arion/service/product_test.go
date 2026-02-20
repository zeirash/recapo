package service

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/response"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/store"
)

func Test_pservice_CreateProduct(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	strPtr := func(s string) *string { return &s }

	type input struct {
		shopID        int
		name          string
		description   *string
		price         int
		originalPrice *int
	}

	tests := []struct {
		name       string
		input      input
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockProductStore
		wantResult response.ProductData
		wantErr    bool
	}{
		{
			name: "successfully create product",
			input: input{
				shopID:        10,
				name:          "Product A",
				description:   strPtr("A great product"),
				price:         1000,
				originalPrice: nil,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					CreateProduct("Product A", strPtr("A great product"), 1000, 10, nil).
					Return(&model.Product{
						ID:            1,
						Name:          "Product A",
						Description:   "A great product",
						Price:         1000,
						OriginalPrice: 1000,
						CreatedAt:     fixedTime,
					}, nil)
				return mock
			},
			wantResult: response.ProductData{
				ID:            1,
				Name:          "Product A",
				Description:   "A great product",
				Price:         1000,
				OriginalPrice: 1000,
				CreatedAt:     fixedTime,
			},
			wantErr: false,
		},
		{
			name: "create product without description",
			input: input{
				shopID:        10,
				name:          "Product B",
				description:   nil,
				price:         500,
				originalPrice: nil,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					CreateProduct("Product B", nil, 500, 10, nil).
					Return(&model.Product{
						ID:            2,
						Name:          "Product B",
						Price:         500,
						OriginalPrice: 500,
						CreatedAt:     fixedTime,
					}, nil)
				return mock
			},
			wantResult: response.ProductData{
				ID:            2,
				Name:          "Product B",
				Price:         500,
				OriginalPrice: 500,
				CreatedAt:     fixedTime,
			},
			wantErr: false,
		},
		{
			name: "create product returns error on database failure",
			input: input{
				shopID:        10,
				name:          "Product A",
				description:   nil,
				price:         1000,
				originalPrice: nil,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					CreateProduct("Product A", nil, 1000, 10, nil).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.ProductData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := productStore
			defer func() { productStore = oldStore }()
			productStore = tt.mockSetup(ctrl)

			var p pservice
			got, gotErr := p.CreateProduct(tt.input.shopID, tt.input.name, tt.input.description, tt.input.price, tt.input.originalPrice)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateProduct() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateProduct() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateProduct() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_pservice_GetProductByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	type input struct {
		productID int
		shopID    []int
	}

	tests := []struct {
		name       string
		input      input
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockProductStore
		wantResult *response.ProductData
		wantErr    bool
	}{
		{
			name: "get product by ID without shop filter",
			input: input{
				productID: 1,
				shopID:    nil,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductByID(1).
					Return(&model.Product{
						ID:            1,
						Name:          "Product A",
						Description:   "A great product",
						Price:         1000,
						OriginalPrice: 1000,
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{Time: fixedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: &response.ProductData{
				ID:            1,
				Name:          "Product A",
				Description:   "A great product",
				Price:         1000,
				OriginalPrice: 1000,
				CreatedAt:     fixedTime,
				UpdatedAt:     &fixedTime,
			},
			wantErr: false,
		},
		{
			name: "get product by ID with shop filter",
			input: input{
				productID: 1,
				shopID:    []int{10},
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductByID(1, 10).
					Return(&model.Product{
						ID:            1,
						Name:          "Product A",
						Description:   "A great product",
						Price:         1000,
						OriginalPrice: 1000,
						CreatedAt:     fixedTime,
					}, nil)
				return mock
			},
			wantResult: &response.ProductData{
				ID:            1,
				Name:          "Product A",
				Description:   "A great product",
				Price:         1000,
				OriginalPrice: 1000,
				CreatedAt:     fixedTime,
			},
			wantErr: false,
		},
		{
			name: "get product not found returns error",
			input: input{
				productID: 9999,
				shopID:    nil,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductByID(9999).
					Return(nil, nil)
				return mock
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "get product returns error on database failure",
			input: input{
				productID: 1,
				shopID:    nil,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductByID(1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := productStore
			defer func() { productStore = oldStore }()
			productStore = tt.mockSetup(ctrl)

			var p pservice
			got, gotErr := p.GetProductByID(tt.input.productID, tt.input.shopID...)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetProductByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetProductByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetProductByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_pservice_GetProductsByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name        string
		shopID      int
		searchQuery *string
		mockSetup   func(ctrl *gomock.Controller) *mock_store.MockProductStore
		wantResult  []response.ProductData
		wantErr     bool
	}{
		{
			name:        "get products by shop ID returns multiple products",
			shopID:      10,
			searchQuery: nil,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductsByShopID(10, nil).
					Return([]model.Product{
						{ID: 1, Name: "Product A", Description: "Desc A", Price: 1000, OriginalPrice: 800, CreatedAt: fixedTime, UpdatedAt: sql.NullTime{Time: fixedTime, Valid: true}},
						{ID: 2, Name: "Product B", Price: 500, OriginalPrice: 500, CreatedAt: fixedTime},
					}, nil)
				return mock
			},
			wantResult: []response.ProductData{
				{ID: 1, Name: "Product A", Description: "Desc A", Price: 1000, OriginalPrice: 800, CreatedAt: fixedTime, UpdatedAt: &fixedTime},
				{ID: 2, Name: "Product B", Price: 500, OriginalPrice: 500, CreatedAt: fixedTime},
			},
			wantErr: false,
		},
		{
			name:        "get products by shop ID returns empty slice",
			shopID:      10,
			searchQuery: nil,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductsByShopID(10, nil).
					Return([]model.Product{}, nil)
				return mock
			},
			wantResult: []response.ProductData{},
			wantErr:    false,
		},
		{
			name:        "get products by shop ID returns error on database failure",
			shopID:      10,
			searchQuery: nil,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductsByShopID(10, nil).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: []response.ProductData{},
			wantErr:    true,
		},
		{
			name:        "get products by shop ID with search query returns matching products",
			shopID:      10,
			searchQuery: strPtr("widget"),
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductsByShopID(10, strPtr("widget")).
					Return([]model.Product{
						{ID: 1, Name: "Widget A", Description: "A useful widget", Price: 1000, OriginalPrice: 800, CreatedAt: fixedTime},
					}, nil)
				return mock
			},
			wantResult: []response.ProductData{
				{ID: 1, Name: "Widget A", Description: "A useful widget", Price: 1000, OriginalPrice: 800, CreatedAt: fixedTime},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := productStore
			defer func() { productStore = oldStore }()
			productStore = tt.mockSetup(ctrl)

			var p pservice
			got, gotErr := p.GetProductsByShopID(tt.shopID, tt.searchQuery)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetProductsByShopID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetProductsByShopID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetProductsByShopID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_pservice_UpdateProduct(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name       string
		input      UpdateProductInput
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockProductStore
		wantResult response.ProductData
		wantErr    bool
	}{
		{
			name: "successfully update product name",
			input: UpdateProductInput{
				ID:   1,
				Name: strPtr("Updated Product"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					UpdateProduct(1, store.UpdateProductInput{Name: strPtr("Updated Product")}).
					Return(&model.Product{
						ID:            1,
						Name:          "Updated Product",
						Description:   "Desc",
						Price:         1000,
						OriginalPrice: 800,
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.ProductData{
				ID:            1,
				Name:          "Updated Product",
				Description:   "Desc",
				Price:         1000,
				OriginalPrice: 800,
				CreatedAt:     fixedTime,
				UpdatedAt:     &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "successfully update product price",
			input: UpdateProductInput{
				ID:    1,
				Price: intPtr(2000),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					UpdateProduct(1, store.UpdateProductInput{Price: intPtr(2000)}).
					Return(&model.Product{
						ID:            1,
						Name:          "Product A",
						Description:   "Desc",
						Price:         2000,
						OriginalPrice: 800,
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.ProductData{
				ID:            1,
				Name:          "Product A",
				Description:   "Desc",
				Price:         2000,
				OriginalPrice: 800,
				CreatedAt:     fixedTime,
				UpdatedAt:     &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "update product not found returns error",
			input: UpdateProductInput{
				ID:   9999,
				Name: strPtr("Updated"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					UpdateProduct(9999, store.UpdateProductInput{Name: strPtr("Updated")}).
					Return(nil, nil)
				return mock
			},
			wantResult: response.ProductData{},
			wantErr:    true,
		},
		{
			name: "update product returns error on database failure",
			input: UpdateProductInput{
				ID:   1,
				Name: strPtr("Updated"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					UpdateProduct(1, store.UpdateProductInput{Name: strPtr("Updated")}).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.ProductData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := productStore
			defer func() { productStore = oldStore }()
			productStore = tt.mockSetup(ctrl)

			var p pservice
			got, gotErr := p.UpdateProduct(tt.input)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateProduct() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateProduct() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateProduct() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_pservice_DeleteProductByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockProductStore
		wantErr   bool
	}{
		{
			name: "successfully delete product",
			id:   1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					DeleteProductByID(1).
					Return(nil)
				return mock
			},
			wantErr: false,
		},
		{
			name: "delete product returns error on database failure",
			id:   1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					DeleteProductByID(1).
					Return(errors.New("database error"))
				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := productStore
			defer func() { productStore = oldStore }()
			productStore = tt.mockSetup(ctrl)

			var p pservice
			gotErr := p.DeleteProductByID(tt.id)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteProductByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteProductByID() succeeded unexpectedly")
			}
		})
	}
}

func Test_pservice_GetPurchaseListProducts(t *testing.T) {
	tests := []struct {
		name      string
		shopID    int
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockProductStore
		want      []response.PurchaseListProductData
		wantErr   bool
	}{
		{
			name:   "returns purchase list products from store",
			shopID: 10,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductsListByActiveOrders(10).
					Return([]model.PurchaseProduct{
						{ProductName: "Product A", Price: 1000, Qty: 5},
						{ProductName: "Product B", Price: 2000, Qty: 3},
					}, nil)
				return mock
			},
			want: []response.PurchaseListProductData{
				{ProductName: "Product A", Price: 1000, Qty: 5},
				{ProductName: "Product B", Price: 2000, Qty: 3},
			},
			wantErr: false,
		},
		{
			name:   "returns empty list when store returns no products",
			shopID: 20,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductsListByActiveOrders(20).
					Return([]model.PurchaseProduct{}, nil)
				return mock
			},
			want:    []response.PurchaseListProductData{},
			wantErr: false,
		},
		{
			name:   "returns error when store fails",
			shopID: 10,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockProductStore {
				mock := mock_store.NewMockProductStore(ctrl)
				mock.EXPECT().
					GetProductsListByActiveOrders(10).
					Return(nil, errors.New("database error"))
				return mock
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := productStore
			defer func() { productStore = oldStore }()
			productStore = tt.mockSetup(ctrl)

			var p pservice
			got, gotErr := p.GetPurchaseListProducts(tt.shopID)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetPurchaseListProducts() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetPurchaseListProducts() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPurchaseListProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}
