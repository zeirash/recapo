package store

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeirash/recapo/arion/model"
)

func Test_order_GetOrderByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		id         int
		shopID     []int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Order
		wantErr    bool
	}{
		{
			name:   "get order by ID without shop filter",
			id:     1,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "pending", fixedTime, nil)
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   5000,
				Status:       "pending",
				CreatedAt:    fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "get order by ID with shop filter",
			id:     1,
			shopID: []int{10},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "pending", fixedTime, nil)
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.id = \$1\s+AND o.shop_id = \$2`).
					WithArgs(1, 10).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   5000,
				Status:       "pending",
				CreatedAt:    fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "get non-existent order returns nil",
			id:     9999,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.id = \$1`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:   "get order returns error on database failure",
			id:     1,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.id = \$1`).
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
			store := NewOrderStoreWithDB(db)

			got, gotErr := store.GetOrderByID(tt.id, tt.shopID...)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetOrderByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetOrderByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetOrderByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_order_GetOrdersByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		shopID     int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult []model.Order
		wantErr    bool
	}{
		{
			name:   "get orders by shop ID returns multiple orders",
			shopID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "pending", fixedTime, nil).
					AddRow(2, 10, "Jane Doe", 3000, "completed", fixedTime, nil)
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.shop_id = \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			wantResult: []model.Order{
				{
					ID:           1,
					ShopID:       10,
					CustomerName: "John Doe",
					TotalPrice:   5000,
					Status:       "pending",
					CreatedAt:    fixedTime,
				},
				{
					ID:           2,
					ShopID:       10,
					CustomerName: "Jane Doe",
					TotalPrice:   3000,
					Status:       "completed",
					CreatedAt:    fixedTime,
				},
			},
			wantErr: false,
		},
		{
			name:   "get orders by shop ID returns empty slice when no orders exist",
			shopID: 9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.shop_id = \$1`).
					WithArgs(9999).
					WillReturnRows(rows)
			},
			wantResult: []model.Order{},
			wantErr:    false,
		},
		{
			name:   "get orders returns error on database failure",
			shopID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.shop_id = \$1`).
					WithArgs(10).
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
			store := NewOrderStoreWithDB(db)

			got, gotErr := store.GetOrdersByShopID(tt.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetOrdersByShopID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetOrdersByShopID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetOrdersByShopID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_order_CreateOrder(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	type input struct {
		customerID int
		shopID     int
		status     string
	}

	tests := []struct {
		name       string
		input      input
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Order
		wantErr    bool
	}{
		{
			name: "successfully create order",
			input: input{
				customerID: 1,
				shopID:     10,
				status:     "pending",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "total_price", "status", "customer_name", "shop_id", "created_at"}).
					AddRow(1, 0, "pending", "John Doe", 10, fixedTime)
				mock.ExpectQuery(`WITH inserted AS \(\s+INSERT INTO orders \(total_price, status, customer_id, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, \$5\)\s+RETURNING id, total_price, status, customer_id, shop_id, created_at\s+\)\s+SELECT i.id, i.total_price, i.status, c.name as customer_name, i.shop_id, i.created_at\s+FROM inserted i\s+INNER JOIN customers c ON i.customer_id = c.id`).
					WithArgs(0, "pending", 1, 10, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   0,
				Status:       "pending",
				CreatedAt:    fixedTime,
			},
			wantErr: false,
		},
		{
			name: "create order returns error on database failure",
			input: input{
				customerID: 1,
				shopID:     10,
				status:     "pending",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH inserted AS \(\s+INSERT INTO orders \(total_price, status, customer_id, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, \$5\)\s+RETURNING id, total_price, status, customer_id, shop_id, created_at\s+\)\s+SELECT i.id, i.total_price, i.status, c.name as customer_name, i.shop_id, i.created_at\s+FROM inserted i\s+INNER JOIN customers c ON i.customer_id = c.id`).
					WithArgs(0, "pending", 1, 10, sqlmock.AnyArg()).
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
			store := NewOrderStoreWithDB(db)

			got, gotErr := store.CreateOrder(tt.input.customerID, tt.input.shopID, tt.input.status)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateOrder() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateOrder() succeeded unexpectedly")
			}

			// Set expected CreatedAt to match for DeepEqual comparison
			tt.wantResult.CreatedAt = got.CreatedAt

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateOrder() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_order_UpdateOrder(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	intPtr := func(i int) *int { return &i }
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		id         int
		input      UpdateOrderInput
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Order
		wantErr    bool
	}{
		{
			name: "update order with status",
			id:   1,
			input: UpdateOrderInput{
				Status: strPtr("completed"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "completed", fixedTime, updatedTime)
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE orders\s+SET status = 'completed',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, customer_id, total_price, status, created_at, updated_at\s+\)\s+SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN customers c ON u.customer_id = c.id`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   5000,
				Status:       "completed",
				CreatedAt:    fixedTime,
				UpdatedAt:    sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name: "update order with total price",
			id:   1,
			input: UpdateOrderInput{
				TotalPrice: intPtr(10000),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 10000, "pending", fixedTime, updatedTime)
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE orders\s+SET total_price = 10000,updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, customer_id, total_price, status, created_at, updated_at\s+\)\s+SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN customers c ON u.customer_id = c.id`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   10000,
				Status:       "pending",
				CreatedAt:    fixedTime,
				UpdatedAt:    sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name: "update non-existent order returns error",
			id:   9999,
			input: UpdateOrderInput{
				Status: strPtr("completed"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE orders\s+SET status = 'completed',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, customer_id, total_price, status, created_at, updated_at\s+\)\s+SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN customers c ON u.customer_id = c.id`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "update order returns error on database failure",
			id:   1,
			input: UpdateOrderInput{
				Status: strPtr("completed"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE orders\s+SET status = 'completed',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, customer_id, total_price, status, created_at, updated_at\s+\)\s+SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN customers c ON u.customer_id = c.id`).
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
			store := NewOrderStoreWithDB(db)

			got, gotErr := store.UpdateOrder(tt.id, tt.input)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateOrder() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateOrder() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateOrder() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_order_DeleteOrderByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "successfully delete order",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM orders\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "delete order returns error on database failure",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM orders\s+WHERE id = \$1`).
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

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			var o order
			gotErr := o.DeleteOrderByID(tx, tt.id)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteOrderByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteOrderByID() succeeded unexpectedly")
			}
		})
	}
}
