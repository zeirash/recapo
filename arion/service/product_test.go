package service

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
		imageURL      *string
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
					CreateProduct("Product A", strPtr("A great product"), 1000, 10, nil, nil).
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
					CreateProduct("Product B", nil, 500, 10, nil, nil).
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
					CreateProduct("Product A", nil, 1000, 10, nil, nil).
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
			got, gotErr := p.CreateProduct(tt.input.shopID, tt.input.name, tt.input.description, tt.input.price, tt.input.originalPrice, tt.input.imageURL)

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

func Test_pservice_UploadProductImage(t *testing.T) {
	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

	tmpDir := t.TempDir()
	oldCfg := cfg
	cfg.UploadDir = tmpDir
	defer func() { cfg = oldCfg }()

	tests := []struct {
		name         string
		fileContent  []byte
		r2BucketName string // if non-empty, enables R2 path
		r2UploadErr  error  // error r2UploadFunc should return (nil = success with mock URL)
		wantURLPfx   string
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:        "successfully upload jpeg image",
			fileContent: jpegBytes,
			wantURLPfx:  "/uploads/products/",
			wantErr:     false,
		},
		{
			name:        "returns error for unsupported file type",
			fileContent: []byte("hello plain text"),
			wantErr:     true,
			wantErrMsg:  "unsupported image type",
		},
		{
			name:         "with R2 configured, delegates to R2 upload",
			fileContent:  jpegBytes,
			r2BucketName: "test-bucket",
			r2UploadErr:  errors.New("mock R2 unavailable"),
			wantErr:      true,
			wantErrMsg:   "mock R2 unavailable",
		},
		{
			name:         "with R2 configured, returns R2 URL on success",
			fileContent:  jpegBytes,
			r2BucketName: "test-bucket",
			r2UploadErr:  nil,
			wantURLPfx:   "https://pub-test.r2.dev/products/",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.r2BucketName != "" {
				cfg.R2BucketName = tt.r2BucketName
				cfg.R2PublicURL = "https://pub-test.r2.dev"
				defer func() {
					cfg.R2BucketName = ""
					cfg.R2PublicURL = ""
				}()

				uploadErr := tt.r2UploadErr
				old := r2UploadFunc
				r2UploadFunc = func(key string, body io.Reader, contentType string) error {
					return uploadErr
				}
				defer func() { r2UploadFunc = old }()
			}

			var p pservice
			got, gotErr := p.UploadProductImage(bytes.NewReader(tt.fileContent))

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UploadProductImage() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				if tt.wantErrMsg != "" && !strings.Contains(gotErr.Error(), tt.wantErrMsg) {
					t.Errorf("UploadProductImage() error = %v, want containing %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UploadProductImage() succeeded unexpectedly")
			}
			if !strings.HasPrefix(got, tt.wantURLPfx) {
				t.Errorf("UploadProductImage() url = %v, want prefix %v", got, tt.wantURLPfx)
			}
		})
	}
}

func Test_pservice_DeleteProductImage(t *testing.T) {
	tmpDir := t.TempDir()
	oldCfg := cfg
	cfg.UploadDir = tmpDir
	defer func() { cfg = oldCfg }()

	// create a real temp file to test successful deletion
	createTempImage := func(t *testing.T) string {
		t.Helper()
		productDir := filepath.Join(tmpDir, "products")
		if err := os.MkdirAll(productDir, 0755); err != nil {
			t.Fatalf("failed to create upload dir: %v", err)
		}
		f, err := os.CreateTemp(productDir, "test-*.jpg")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		f.Close()
		return "/uploads/products/" + filepath.Base(f.Name())
	}

	tests := []struct {
		name         string
		imageURL     func(t *testing.T) string
		r2BucketName string // if non-empty, enables R2 path
		r2DeleteErr  error  // error r2DeleteFunc should return
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:     "successfully delete image",
			imageURL: func(t *testing.T) string { return createTempImage(t) },
			wantErr:  false,
		},
		{
			name:       "returns error for invalid URL prefix",
			imageURL:   func(t *testing.T) string { return "/some/other/path/image.jpg" },
			wantErr:    true,
			wantErrMsg: "invalid image URL",
		},
		{
			name:       "returns error when file does not exist",
			imageURL:   func(t *testing.T) string { return "/uploads/products/nonexistent.jpg" },
			wantErr:    true,
			wantErrMsg: "image not found",
		},
		{
			name:         "with R2 configured, delegates to R2 delete for R2 URL",
			imageURL:     func(t *testing.T) string { return "https://pub-test.r2.dev/products/abc.jpg" },
			r2BucketName: "test-bucket",
			r2DeleteErr:  errors.New("mock R2 unavailable"),
			wantErr:      true,
			wantErrMsg:   "mock R2 unavailable",
		},
		{
			name:         "with R2 configured, falls back to local path for non-R2 URL",
			imageURL:     func(t *testing.T) string { return "/uploads/products/nonexistent.jpg" },
			r2BucketName: "test-bucket",
			wantErr:      true,
			wantErrMsg:   "image not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.r2BucketName != "" {
				cfg.R2BucketName = tt.r2BucketName
				cfg.R2PublicURL = "https://pub-test.r2.dev"
				defer func() {
					cfg.R2BucketName = ""
					cfg.R2PublicURL = ""
				}()

				deleteErr := tt.r2DeleteErr
				old := r2DeleteFunc
				r2DeleteFunc = func(key string) error {
					return deleteErr
				}
				defer func() { r2DeleteFunc = old }()
			}

			var p pservice
			gotErr := p.DeleteProductImage(tt.imageURL(t))

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteProductImage() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				if tt.wantErrMsg != "" && !strings.Contains(gotErr.Error(), tt.wantErrMsg) {
					t.Errorf("DeleteProductImage() error = %v, want containing %v", gotErr.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteProductImage() succeeded unexpectedly")
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
