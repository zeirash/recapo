package store

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeirash/recapo/arion/common/constant"
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
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "in_progress", "", fixedTime, nil)
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   5000,
				Status:       "in_progress",
				CreatedAt:    fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "get order by ID with shop filter",
			id:     1,
			shopID: []int{10},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "in_progress", "", fixedTime, nil)
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.id = \$1\s+AND o.shop_id = \$2`).
					WithArgs(1, 10).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   5000,
				Status:       "in_progress",
				CreatedAt:    fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "get non-existent order returns nil",
			id:     9999,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.id = \$1`).
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
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.id = \$1`).
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
	strPtr := func(s string) *string { return &s }
	ptrTime := func(t time.Time) *time.Time { return &t }

	tests := []struct {
		name       string
		shopID     int
		opts       model.OrderFilterOptions
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult []model.Order
		wantErr    bool
	}{
		{
			name:   "get orders by shop ID returns multiple orders",
			shopID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "in_progress", "", fixedTime, nil).
					AddRow(2, 10, "Jane Doe", 3000, "done", "", fixedTime, nil)
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.shop_id = \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			wantResult: []model.Order{
				{
					ID:           1,
					ShopID:       10,
					CustomerName: "John Doe",
					TotalPrice:   5000,
					Status:       "in_progress",
					CreatedAt:    fixedTime,
				},
				{
					ID:           2,
					ShopID:       10,
					CustomerName: "Jane Doe",
					TotalPrice:   3000,
					Status:       "done",
					CreatedAt:    fixedTime,
				},
			},
			wantErr: false,
		},
		{
			name:   "get orders by shop ID returns empty slice when no orders exist",
			shopID: 9999,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.shop_id = \$1`).
					WithArgs(9999).
					WillReturnRows(rows)
			},
			wantResult: []model.Order{},
			wantErr:    false,
		},
		{
			name:   "get orders returns error on database failure",
			shopID: 10,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.shop_id = \$1`).
					WithArgs(10).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get orders by shop ID with search query filters by customer name or phone",
			shopID: 10,
			opts:   model.OrderFilterOptions{SearchQuery: strPtr("john")},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "in_progress", "", fixedTime, nil)
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.shop_id = \$1\s+AND \(c.name ILIKE \$2 OR c.phone ILIKE \$2\)`).
					WithArgs(10, "%john%").
					WillReturnRows(rows)
			},
			wantResult: []model.Order{
				{
					ID:           1,
					ShopID:       10,
					CustomerName: "John Doe",
					TotalPrice:   5000,
					Status:       "in_progress",
					CreatedAt:    fixedTime,
				},
			},
			wantErr: false,
		},
		{
			name:   "get orders by shop ID with date range filters by created_at",
			shopID: 10,
			opts: model.OrderFilterOptions{
				DateFrom: ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				DateTo:   ptrTime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)), // exclusive: < Feb 1 = through Jan 31
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "in_progress", "", fixedTime, nil)
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.shop_id = \$1\s+AND o.created_at >= \$2\s+AND o.created_at < \$3`).
					WithArgs(10, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)).
					WillReturnRows(rows)
			},
			wantResult: []model.Order{
				{
					ID:           1,
					ShopID:       10,
					CustomerName: "John Doe",
					TotalPrice:   5000,
					Status:       "in_progress",
					CreatedAt:    fixedTime,
				},
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
			store := NewOrderStoreWithDB(db)

			got, gotErr := store.GetOrdersByShopID(tt.shopID, tt.opts)

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

	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }

	type input struct {
		customerID  int
		shopID      int
		notes       *string
		totalPrice  *int
	}

	tests := []struct {
		name       string
		input      input
		useTx      bool // when true, pass tx from db.Begin() (mock must ExpectBegin)
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Order
		wantErr    bool
	}{
		{
			name:  "successfully create order",
			useTx: false,
			input: input{
				customerID: 1,
				shopID:     10,
				notes:      strPtr("test notes"),
				totalPrice: nil,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "total_price", "status", "customer_name", "shop_id", "notes", "created_at"}).
					AddRow(1, 0, constant.OrderStatusCreated, "John Doe", 10, "test notes", fixedTime)
				mock.ExpectQuery(`WITH inserted AS \(\s+INSERT INTO orders \(total_price, status, customer_id, shop_id, notes, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, COALESCE\(\$5, ''\), \$6\)\s+RETURNING id, total_price, status, customer_id, shop_id, notes, created_at\s+\)\s+SELECT i.id, i.total_price, i.status, c.name as customer_name, i.shop_id, i.notes, i.created_at\s+FROM inserted i\s+INNER JOIN customers c ON i.customer_id = c.id`).
					WithArgs(0, constant.OrderStatusCreated, 1, 10, strPtr("test notes"), sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   0,
				Status:       constant.OrderStatusCreated,
				Notes:        "test notes",
				CreatedAt:    fixedTime,
			},
			wantErr: false,
		},
		{
			name:  "create order returns error on database failure",
			useTx: false,
			input: input{
				customerID: 1,
				shopID:     10,
				totalPrice: nil,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH inserted AS \(\s+INSERT INTO orders \(total_price, status, customer_id, shop_id, notes, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, COALESCE\(\$5, ''\), \$6\)\s+RETURNING id, total_price, status, customer_id, shop_id, notes, created_at\s+\)\s+SELECT i.id, i.total_price, i.status, c.name as customer_name, i.shop_id, i.notes, i.created_at\s+FROM inserted i\s+INNER JOIN customers c ON i.customer_id = c.id`).
					WithArgs(0, constant.OrderStatusCreated, 1, 10, (*string)(nil), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:  "successfully create order with totalPrice",
			useTx: false,
			input: input{
				customerID: 1,
				shopID:     10,
				notes:      nil,
				totalPrice: intPtr(5000),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "total_price", "status", "customer_name", "shop_id", "notes", "created_at"}).
					AddRow(1, 5000, constant.OrderStatusCreated, "John Doe", 10, "", fixedTime)
				mock.ExpectQuery(`WITH inserted AS \(\s+INSERT INTO orders \(total_price, status, customer_id, shop_id, notes, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, COALESCE\(\$5, ''\), \$6\)\s+RETURNING id, total_price, status, customer_id, shop_id, notes, created_at\s+\)\s+SELECT i.id, i.total_price, i.status, c.name as customer_name, i.shop_id, i.notes, i.created_at\s+FROM inserted i\s+INNER JOIN customers c ON i.customer_id = c.id`).
					WithArgs(5000, constant.OrderStatusCreated, 1, 10, (*string)(nil), sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   5000,
				Status:       constant.OrderStatusCreated,
				Notes:        "",
				CreatedAt:    fixedTime,
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
			store := NewOrderStoreWithDB(db)

			var got *model.Order
			var gotErr error
			if tt.useTx {
				tx, err := db.Begin()
				if err != nil {
					t.Fatalf("failed to begin tx: %v", err)
				}
				defer tx.Rollback()
				got, gotErr = store.CreateOrder(tx, tt.input.customerID, tt.input.shopID, tt.input.notes, tt.input.totalPrice)
			} else {
				got, gotErr = store.CreateOrder(nil, tt.input.customerID, tt.input.shopID, tt.input.notes, tt.input.totalPrice)
			}

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
				Status: strPtr("done"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "done", "", fixedTime, updatedTime)
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE orders\s+SET status = 'done',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, customer_id, total_price, status, notes, created_at, updated_at\s+\)\s+SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.notes, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN customers c ON u.customer_id = c.id`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   5000,
				Status:       "done",
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
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 10000, "in_progress", "", fixedTime, updatedTime)
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE orders\s+SET total_price = 10000,updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, customer_id, total_price, status, notes, created_at, updated_at\s+\)\s+SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.notes, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN customers c ON u.customer_id = c.id`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   10000,
				Status:       "in_progress",
				CreatedAt:    fixedTime,
				UpdatedAt:    sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name: "update order with notes",
			id:   1,
			input: UpdateOrderInput{
				Notes: strPtr("updated notes"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", 5000, "in_progress", "updated notes", fixedTime, updatedTime)
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE orders\s+SET notes = 'updated notes',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, customer_id, total_price, status, notes, created_at, updated_at\s+\)\s+SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.notes, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN customers c ON u.customer_id = c.id`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           1,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   5000,
				Status:       "in_progress",
				Notes:        "updated notes",
				CreatedAt:    fixedTime,
				UpdatedAt:    sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name: "update non-existent order returns error",
			id:   9999,
			input: UpdateOrderInput{
				Status: strPtr("done"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE orders\s+SET status = 'completed',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, customer_id, total_price, status, notes, created_at, updated_at\s+\)\s+SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.notes, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN customers c ON u.customer_id = c.id`).
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
				Status: strPtr("done"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE orders\s+SET status = 'completed',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, shop_id, customer_id, total_price, status, notes, created_at, updated_at\s+\)\s+SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.notes, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN customers c ON u.customer_id = c.id`).
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

			got, gotErr := store.UpdateOrder(nil, tt.id, tt.input)

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

func Test_order_CreateTempOrder(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name          string
		customerName  string
		customerPhone string
		shopID        int
		mockSetup     func(mock sqlmock.Sqlmock)
		want          *model.TempOrder
		wantErr       bool
	}{
		{
			name:          "successfully create temp order",
			customerName:  "Jane Doe",
			customerPhone: "+62812345678",
			shopID:        5,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "customer_name", "customer_phone", "shop_id", "total_price", "status", "created_at"}).
					AddRow(1, "Jane Doe", "+62812345678", 5, 0, "pending", fixedTime)
				mock.ExpectQuery(`INSERT INTO temp_orders \(customer_name, customer_phone, status, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, \$5\)\s+RETURNING id, customer_name, customer_phone, shop_id, total_price, status, created_at`).
					WithArgs("Jane Doe", "+62812345678", "pending", 5, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			want: &model.TempOrder{
				ID:            1,
				CustomerName:  "Jane Doe",
				CustomerPhone: "+62812345678",
				ShopID:        5,
				TotalPrice:    0,
				Status:        "pending",
				CreatedAt:     fixedTime,
			},
			wantErr: false,
		},
		{
			name:          "create temp order returns error on database failure",
			customerName:  "Jane Doe",
			customerPhone: "+62812345678",
			shopID:        5,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO temp_orders \(customer_name, customer_phone, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4\)\s+RETURNING id, customer_name, customer_phone, shop_id, total_price, status, created_at`).
					WithArgs("Jane Doe", "+62812345678", 5, sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
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

			mock.ExpectBegin()
			tt.mockSetup(mock)
			store := NewOrderStoreWithDB(db)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin tx: %v", err)
			}
			defer tx.Rollback()

			got, gotErr := store.CreateTempOrder(tx, tt.customerName, tt.customerPhone, tt.shopID)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTempOrder() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTempOrder() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateTempOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_order_UpdateTempOrderTotalPrice(t *testing.T) {
	tests := []struct {
		name        string
		tempOrderID int
		totalPrice  int
		mockSetup   func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:        "successfully update temp order total price",
			tempOrderID: 1,
			totalPrice:  2500,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE temp_orders\s+SET total_price = \$1, updated_at = now\(\)\s+WHERE id = \$2`).
					WithArgs(2500, 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:        "update temp order total price returns error on database failure",
			tempOrderID: 1,
			totalPrice:  2500,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE temp_orders\s+SET total_price = \$1, updated_at = now\(\)\s+WHERE id = \$2`).
					WithArgs(2500, 1).
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
			store := NewOrderStoreWithDB(db)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin tx: %v", err)
			}
			defer tx.Rollback()

			gotErr := store.UpdateTempOrderTotalPrice(tx, tt.tempOrderID, tt.totalPrice)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTempOrderTotalPrice() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTempOrderTotalPrice() succeeded unexpectedly")
			}
		})
	}
}

func Test_order_GetTempOrdersByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	strPtr := func(s string) *string { return &s }
	ptrTime := func(t time.Time) *time.Time { return &t }

	tests := []struct {
		name       string
		shopID     int
		opts       model.OrderFilterOptions
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult []model.TempOrder
		wantErr    bool
	}{
		{
			name:   "get temp orders by shop ID returns multiple temp orders",
			shopID: 5,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "customer_phone", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 5, "Jane Doe", "+62812345678", 2500, "pending", fixedTime, nil).
					AddRow(2, 5, "John Doe", "+62887654321", 1000, "pending", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at\s+FROM temp_orders\s+WHERE shop_id = \$1`).
					WithArgs(5).
					WillReturnRows(rows)
			},
			wantResult: []model.TempOrder{
				{ID: 1, ShopID: 5, CustomerName: "Jane Doe", CustomerPhone: "+62812345678", TotalPrice: 2500, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{}},
				{ID: 2, ShopID: 5, CustomerName: "John Doe", CustomerPhone: "+62887654321", TotalPrice: 1000, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{}},
			},
			wantErr: false,
		},
		{
			name:   "get temp orders by shop ID returns empty slice when none exist",
			shopID: 99,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "customer_phone", "total_price", "status", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at\s+FROM temp_orders\s+WHERE shop_id = \$1`).
					WithArgs(99).
					WillReturnRows(rows)
			},
			wantResult: []model.TempOrder{},
			wantErr:    false,
		},
		{
			name:   "get temp orders returns error on database failure",
			shopID: 5,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at\s+FROM temp_orders\s+WHERE shop_id = \$1`).
					WithArgs(5).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get temp orders with search query filters by customer name or phone",
			shopID: 5,
			opts:   model.OrderFilterOptions{SearchQuery: strPtr("62812")},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "customer_phone", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 5, "Jane Doe", "+62812345678", 2500, "pending", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at\s+FROM temp_orders\s+WHERE shop_id = \$1\s+AND \(customer_name ILIKE \$2 OR customer_phone ILIKE \$2\)`).
					WithArgs(5, "%62812%").
					WillReturnRows(rows)
			},
			wantResult: []model.TempOrder{
				{ID: 1, ShopID: 5, CustomerName: "Jane Doe", CustomerPhone: "+62812345678", TotalPrice: 2500, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{}},
			},
			wantErr: false,
		},
		{
			name:   "get temp orders with date range filters by created_at",
			shopID: 5,
			opts: model.OrderFilterOptions{
				DateFrom: ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				DateTo:   ptrTime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "customer_phone", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 5, "Jane Doe", "+62812345678", 2500, "pending", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at\s+FROM temp_orders\s+WHERE shop_id = \$1\s+AND created_at >= \$2\s+AND created_at < \$3`).
					WithArgs(5, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)).
					WillReturnRows(rows)
			},
			wantResult: []model.TempOrder{
				{ID: 1, ShopID: 5, CustomerName: "Jane Doe", CustomerPhone: "+62812345678", TotalPrice: 2500, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{}},
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
			store := NewOrderStoreWithDB(db)

			got, gotErr := store.GetTempOrdersByShopID(tt.shopID, tt.opts)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTempOrdersByShopID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTempOrdersByShopID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetTempOrdersByShopID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_order_GetTempOrderByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		id         int
		shopID     []int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.TempOrder
		wantErr    bool
	}{
		{
			name:   "get temp order by ID without shop filter",
			id:     1,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "customer_phone", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 5, "Jane Doe", "+62812345678", 2500, "pending", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at\s+FROM temp_orders\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.TempOrder{
				ID:            1,
				ShopID:        5,
				CustomerName:  "Jane Doe",
				CustomerPhone: "+62812345678",
				TotalPrice:    2500,
				Status:        "pending",
				CreatedAt:     fixedTime,
				UpdatedAt:     sql.NullTime{},
			},
			wantErr: false,
		},
		{
			name:   "get temp order by ID with shop ID filter",
			id:     1,
			shopID: []int{5},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "customer_phone", "total_price", "status", "created_at", "updated_at"}).
					AddRow(1, 5, "Jane Doe", "+62812345678", 2500, "pending", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at\s+FROM temp_orders\s+WHERE id = \$1\s+AND shop_id = \$2`).
					WithArgs(1, 5).
					WillReturnRows(rows)
			},
			wantResult: &model.TempOrder{
				ID:            1,
				ShopID:        5,
				CustomerName:  "Jane Doe",
				CustomerPhone: "+62812345678",
				TotalPrice:    2500,
				Status:        "pending",
				CreatedAt:     fixedTime,
				UpdatedAt:     sql.NullTime{},
			},
			wantErr: false,
		},
		{
			name:   "get temp order by ID not found returns nil",
			id:     999,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at\s+FROM temp_orders\s+WHERE id = \$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:   "get temp order by ID returns error on database failure",
			id:     1,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at\s+FROM temp_orders\s+WHERE id = \$1`).
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

			got, gotErr := store.GetTempOrderByID(tt.id, tt.shopID...)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTempOrderByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTempOrderByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetTempOrderByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_order_GetActiveOrderByCustomerID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		customerID int
		shopID     int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Order
		wantErr    bool
	}{
		{
			name:       "returns active order when customer has one",
			customerID: 1,
			shopID:     10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "customer_name", "total_price", "status", "notes", "created_at", "updated_at"}).
					AddRow(5, 10, "John Doe", 3000, constant.OrderStatusInProgress, "notes", fixedTime, nil)
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.customer_id = \$1 AND o.shop_id = \$2 AND o.status IN \(\$3, \$4\)\s+ORDER BY o.created_at DESC\s+LIMIT 1`).
					WithArgs(1, 10, constant.OrderStatusCreated, constant.OrderStatusInProgress).
					WillReturnRows(rows)
			},
			wantResult: &model.Order{
				ID:           5,
				ShopID:       10,
				CustomerName: "John Doe",
				TotalPrice:   3000,
				Status:       constant.OrderStatusInProgress,
				Notes:        "notes",
				CreatedAt:    fixedTime,
			},
			wantErr: false,
		},
		{
			name:       "returns nil nil when customer has no active order",
			customerID: 99,
			shopID:     10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.customer_id = \$1 AND o.shop_id = \$2 AND o.status IN \(\$3, \$4\)\s+ORDER BY o.created_at DESC\s+LIMIT 1`).
					WithArgs(99, 10, constant.OrderStatusCreated, constant.OrderStatusInProgress).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:       "returns error on database failure",
			customerID: 1,
			shopID:     10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at\s+FROM orders o\s+INNER JOIN customers c ON o.customer_id = c.id\s+WHERE o.customer_id = \$1 AND o.shop_id = \$2 AND o.status IN \(\$3, \$4\)\s+ORDER BY o.created_at DESC\s+LIMIT 1`).
					WithArgs(1, 10, constant.OrderStatusCreated, constant.OrderStatusInProgress).
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

			got, gotErr := store.GetActiveOrderByCustomerID(tt.customerID, tt.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetActiveOrderByCustomerID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetActiveOrderByCustomerID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetActiveOrderByCustomerID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_order_UpdateTempOrderStatus(t *testing.T) {
	tests := []struct {
		name        string
		useTx       bool
		tempOrderID int
		status      string
		mockSetup   func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:        "successfully update temp order status with tx",
			useTx:       true,
			tempOrderID: 1,
			status:      "created",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE temp_orders\s+SET status = \$1, updated_at = now\(\)\s+WHERE id = \$2`).
					WithArgs("created", 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:        "successfully update temp order status with nil tx (uses db)",
			useTx:       false,
			tempOrderID: 10,
			status:      "rejected",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE temp_orders\s+SET status = \$1, updated_at = now\(\)\s+WHERE id = \$2`).
					WithArgs("rejected", 10).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:        "update temp order status returns error on database failure",
			useTx:       true,
			tempOrderID: 1,
			status:      "created",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE temp_orders\s+SET status = \$1, updated_at = now\(\)\s+WHERE id = \$2`).
					WithArgs("created", 1).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name:        "update temp order status with nil tx returns error on database failure",
			useTx:       false,
			tempOrderID: 5,
			status:      "accepted",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE temp_orders\s+SET status = \$1, updated_at = now\(\)\s+WHERE id = \$2`).
					WithArgs("accepted", 5).
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
			store := NewOrderStoreWithDB(db)

			var gotErr error
			if tt.useTx {
				tx, err := db.Begin()
				if err != nil {
					t.Fatalf("failed to begin tx: %v", err)
				}
				defer tx.Rollback()
				gotErr = store.UpdateTempOrderStatus(tx, tt.tempOrderID, tt.status)
			} else {
				gotErr = store.UpdateTempOrderStatus(nil, tt.tempOrderID, tt.status)
			}

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateTempOrderStatus() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateTempOrderStatus() succeeded unexpectedly")
			}
		})
	}
}
