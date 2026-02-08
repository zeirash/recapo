package store

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/zeirash/recapo/arion/model"
)

func Test_product_GetProductByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		productID  int
		shopID     []int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Product
		wantErr    bool
	}{
		{
			name:      "get product by ID without shop filter",
			productID: 1,
			shopID:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "original_price", "created_at", "updated_at"}).
					AddRow(1, 10, "Product A", "A great product", 1000, 800, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, name, description, price, original_price, created_at, updated_at\s+FROM products\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Product{
				ID:            1,
				ShopID:        10,
				Name:          "Product A",
				Description:   "A great product",
				Price:         1000,
				OriginalPrice: 800,
				CreatedAt:     fixedTime,
			},
			wantErr: false,
		},
		{
			name:      "get product by ID with shop filter",
			productID: 1,
			shopID:    []int{10},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "original_price", "created_at", "updated_at"}).
					AddRow(1, 10, "Product A", "A great product", 1000, 800, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, name, description, price, original_price, created_at, updated_at\s+FROM products\s+WHERE id = \$1\s+AND shop_id = \$2`).
					WithArgs(1, 10).
					WillReturnRows(rows)
			},
			wantResult: &model.Product{
				ID:            1,
				ShopID:        10,
				Name:          "Product A",
				Description:   "A great product",
				Price:         1000,
				OriginalPrice: 800,
				CreatedAt:     fixedTime,
			},
			wantErr: false,
		},
		{
			name:      "get non-existent product returns nil",
			productID: 9999,
			shopID:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, description, price, original_price, created_at, updated_at\s+FROM products\s+WHERE id = \$1`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:      "get product returns error on database failure",
			productID: 1,
			shopID:    nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, description, price, original_price, created_at, updated_at\s+FROM products\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewProductStoreWithDB(db)

			got, gotErr := store.GetProductByID(tt.productID, tt.shopID...)

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

func Test_product_GetProductsByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name        string
		shopID      int
		searchQuery *string
		mockSetup   func(mock sqlmock.Sqlmock)
		wantResult  []model.Product
		wantErr     bool
	}{
		{
			name:        "get products by shop ID returns multiple products",
			shopID:      10,
			searchQuery: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "original_price", "created_at", "updated_at"}).
					AddRow(1, 10, "Product A", "Description A", 1000, 800, fixedTime, nil).
					AddRow(2, 10, "Product B", "Description B", 2000, 1500, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, name, description, price, original_price, created_at, updated_at\s+FROM products\s+WHERE shop_id = \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			wantResult: []model.Product{
				{ID: 1, ShopID: 10, Name: "Product A", Description: "Description A", Price: 1000, OriginalPrice: 800, CreatedAt: fixedTime},
				{ID: 2, ShopID: 10, Name: "Product B", Description: "Description B", Price: 2000, OriginalPrice: 1500, CreatedAt: fixedTime},
			},
			wantErr: false,
		},
		{
			name:        "get products by shop ID returns empty slice when no products exist",
			shopID:      9999,
			searchQuery: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "original_price", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT id, shop_id, name, description, price, original_price, created_at, updated_at\s+FROM products\s+WHERE shop_id = \$1`).
					WithArgs(9999).
					WillReturnRows(rows)
			},
			wantResult: []model.Product{},
			wantErr:    false,
		},
		{
			name:        "get products by shop ID returns nil when db returns no rows",
			shopID:      9999,
			searchQuery: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, description, price, original_price, created_at, updated_at\s+FROM products\s+WHERE shop_id = \$1`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:        "get products returns error on database failure",
			shopID:      10,
			searchQuery: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, description, price, original_price, created_at, updated_at\s+FROM products\s+WHERE shop_id = \$1`).
					WithArgs(10).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:        "get products by shop ID with search query returns matching products",
			shopID:      10,
			searchQuery: strPtr("widget"),
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "original_price", "created_at", "updated_at"}).
					AddRow(1, 10, "Widget A", "A useful widget", 1000, 800, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, name, description, price, original_price, created_at, updated_at\s+FROM products\s+WHERE shop_id = \$1\s+AND name ILIKE \$2`).
					WithArgs(10, "%widget%").
					WillReturnRows(rows)
			},
			wantResult: []model.Product{
				{ID: 1, ShopID: 10, Name: "Widget A", Description: "A useful widget", Price: 1000, OriginalPrice: 800, CreatedAt: fixedTime},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewProductStoreWithDB(db)

			got, gotErr := store.GetProductsByShopID(tt.shopID, tt.searchQuery)

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

func Test_product_CreateProduct(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	type input struct {
		name          string
		description   *string
		price         int
		shopID        int
		originalPrice *int
	}

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		input      input
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Product
		wantErr    bool
	}{
		{
			name: "successfully create product",
			input: input{
				name:          "New Product",
				description:   strPtr("Product description"),
				price:         1500,
				shopID:        10,
				originalPrice: nil,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "description"}).AddRow(1, "Product description")
				mock.ExpectQuery(`INSERT INTO products \(name, description, price, original_price, shop_id, created_at\)\s+VALUES \(\$1, COALESCE\(\$2, ''\), \$3, \$4, \$5, \$6\)\s+RETURNING id, description`).
					WithArgs("New Product", strPtr("Product description"), 1500, 1500, 10, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.Product{
				ID:            1,
				Name:          "New Product",
				Description:   "Product description",
				Price:         1500,
				OriginalPrice: 1500,
				ShopID:        10,
			},
			wantErr: false,
		},
		{
			name: "successfully create product with original_price",
			input: input{
				name:          "New Product",
				description:   nil,
				price:         1500,
				shopID:        10,
				originalPrice: intPtr(1200),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "description"}).AddRow(1, "")
				mock.ExpectQuery(`INSERT INTO products \(name, description, price, original_price, shop_id, created_at\)\s+VALUES \(\$1, COALESCE\(\$2, ''\), \$3, \$4, \$5, \$6\)\s+RETURNING id, description`).
					WithArgs("New Product", nil, 1500, 1200, 10, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.Product{
				ID:            1,
				Name:          "New Product",
				Description:   "",
				Price:         1500,
				OriginalPrice: 1200,
				ShopID:        10,
			},
			wantErr: false,
		},
		{
			name: "create product with duplicate name returns ErrDuplicateProductName",
			input: input{
				name:          "Existing Product",
				description:   nil,
				price:         1000,
				shopID:        10,
				originalPrice: nil,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO products \(name, description, price, original_price, shop_id, created_at\)\s+VALUES \(\$1, COALESCE\(\$2, ''\), \$3, \$4, \$5, \$6\)\s+RETURNING id, description`).
					WithArgs("Existing Product", nil, 1000, 1000, 10, sqlmock.AnyArg()).
					WillReturnError(&pq.Error{Code: "23505"})
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "create product returns error on database failure",
			input: input{
				name:          "New Product",
				description:   nil,
				price:         1000,
				shopID:        10,
				originalPrice: nil,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO products \(name, description, price, original_price, shop_id, created_at\)\s+VALUES \(\$1, COALESCE\(\$2, ''\), \$3, \$4, \$5, \$6\)\s+RETURNING id, description`).
					WithArgs("New Product", nil, 1000, 1000, 10, sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewProductStoreWithDB(db)

			got, gotErr := store.CreateProduct(tt.input.name, tt.input.description, tt.input.price, tt.input.shopID, tt.input.originalPrice)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateProduct() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateProduct() succeeded unexpectedly")
			}

			// Set expected CreatedAt to match for DeepEqual comparison
			tt.wantResult.CreatedAt = got.CreatedAt

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateProduct() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_product_UpdateProduct(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name       string
		productID  int
		input      UpdateProductInput
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Product
		wantErr    bool
	}{
		{
			name:      "update product with name",
			productID: 1,
			input: UpdateProductInput{
				Name: strPtr("Updated Product"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "original_price", "created_at", "updated_at"}).
					AddRow(1, 10, "Updated Product", "Description", 1000, 800, fixedTime, updatedTime)
				mock.ExpectQuery(`UPDATE products\s+SET name = 'Updated Product',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, name, description, price, original_price, created_at, updated_at`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Product{
				ID:            1,
				ShopID:        10,
				Name:          "Updated Product",
				Description:   "Description",
				Price:         1000,
				OriginalPrice: 800,
				CreatedAt:     fixedTime,
				UpdatedAt:     sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name:      "update product with price",
			productID: 1,
			input: UpdateProductInput{
				Price: intPtr(2000),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "original_price", "created_at", "updated_at"}).
					AddRow(1, 10, "Product A", "Description", 2000, 800, fixedTime, updatedTime)
				mock.ExpectQuery(`UPDATE products\s+SET price = 2000,updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, name, description, price, original_price, created_at, updated_at`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Product{
				ID:            1,
				ShopID:        10,
				Name:          "Product A",
				Description:   "Description",
				Price:         2000,
				OriginalPrice: 800,
				CreatedAt:     fixedTime,
				UpdatedAt:     sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name:      "update product with description",
			productID: 1,
			input: UpdateProductInput{
				Description: strPtr("Updated description"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "original_price", "created_at", "updated_at"}).
					AddRow(1, 10, "Product A", "Updated description", 1000, 800, fixedTime, updatedTime)
				mock.ExpectQuery(`UPDATE products\s+SET description = 'Updated description',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, name, description, price, original_price, created_at, updated_at`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Product{
				ID:            1,
				ShopID:        10,
				Name:          "Product A",
				Description:   "Updated description",
				Price:         1000,
				OriginalPrice: 800,
				CreatedAt:     fixedTime,
				UpdatedAt:     sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name:      "update product with originalPrice",
			productID: 1,
			input: UpdateProductInput{
				OriginalPrice: intPtr(1200),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "description", "price", "original_price", "created_at", "updated_at"}).
					AddRow(1, 10, "Product A", "Description", 1000, 1200, fixedTime, updatedTime)
				mock.ExpectQuery(`UPDATE products\s+SET original_price = 1200,updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, name, description, price, original_price, created_at, updated_at`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Product{
				ID:            1,
				ShopID:        10,
				Name:          "Product A",
				Description:   "Description",
				Price:         1000,
				OriginalPrice: 1200,
				CreatedAt:     fixedTime,
				UpdatedAt:     sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name:      "update product with duplicate name returns ErrDuplicateProductName",
			productID: 1,
			input: UpdateProductInput{
				Name: strPtr("Existing Name"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE products\s+SET name = 'Existing Name',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, name, description, price, original_price, created_at, updated_at`).
					WithArgs(1).
					WillReturnError(&pq.Error{Code: "23505"})
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:      "update non-existent product returns error",
			productID: 9999,
			input: UpdateProductInput{
				Name: strPtr("New Name"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE products\s+SET name = 'New Name',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, name, description, price, original_price, created_at, updated_at`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewProductStoreWithDB(db)

			got, gotErr := store.UpdateProduct(tt.productID, tt.input)

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

func Test_product_DeleteProductByID(t *testing.T) {
	tests := []struct {
		name      string
		productID int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "successfully delete product",
			productID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM products\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:      "delete non-existent product succeeds",
			productID: 9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM products\s+WHERE id = \$1`).
					WithArgs(9999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name:      "delete product returns error on database failure",
			productID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM products\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewProductStoreWithDB(db)

			gotErr := store.DeleteProductByID(tt.productID)

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
