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

func Test_orderitem_GetOrderItemByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		id         int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.OrderItem
		wantErr    bool
	}{
		{
			name: "get order item by ID",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "order_id", "product_name", "price", "qty", "created_at", "updated_at"}).
					AddRow(1, 10, "Product A", 1000, 2, fixedTime, nil)
				mock.ExpectQuery(`SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at\s+FROM order_items oi\s+INNER JOIN products p ON oi.product_id = p.id\s+WHERE oi.id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.OrderItem{
				ID:          1,
				OrderID:     10,
				ProductName: "Product A",
				Price:       1000,
				Qty:         2,
				CreatedAt:   fixedTime,
			},
			wantErr: false,
		},
		{
			name: "get non-existent order item returns nil",
			id:   9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at\s+FROM order_items oi\s+INNER JOIN products p ON oi.product_id = p.id\s+WHERE oi.id = \$1`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name: "get order item returns error on database failure",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at\s+FROM order_items oi\s+INNER JOIN products p ON oi.product_id = p.id\s+WHERE oi.id = \$1`).
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
			store := NewOrderItemStoreWithDB(db)

			got, gotErr := store.GetOrderItemByID(tt.id)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetOrderItemByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetOrderItemByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetOrderItemByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_orderitem_GetOrderItemsByOrderID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		orderID    int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult []model.OrderItem
		wantErr    bool
	}{
		{
			name:    "get order items by order ID returns multiple items",
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "order_id", "product_name", "price", "qty", "created_at", "updated_at"}).
					AddRow(1, 10, "Product A", 1000, 2, fixedTime, nil).
					AddRow(2, 10, "Product B", 2000, 1, fixedTime, nil)
				mock.ExpectQuery(`SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at\s+FROM order_items oi\s+INNER JOIN products p ON oi.product_id = p.id\s+WHERE oi.order_id = \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			wantResult: []model.OrderItem{
				{
					ID:          1,
					OrderID:     10,
					ProductName: "Product A",
					Price:       1000,
					Qty:         2,
					CreatedAt:   fixedTime,
				},
				{
					ID:          2,
					OrderID:     10,
					ProductName: "Product B",
					Price:       2000,
					Qty:         1,
					CreatedAt:   fixedTime,
				},
			},
			wantErr: false,
		},
		{
			name:    "get order items by order ID returns empty slice when no items exist",
			orderID: 9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "order_id", "product_name", "price", "qty", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at\s+FROM order_items oi\s+INNER JOIN products p ON oi.product_id = p.id\s+WHERE oi.order_id = \$1`).
					WithArgs(9999).
					WillReturnRows(rows)
			},
			wantResult: []model.OrderItem{},
			wantErr:    false,
		},
		{
			name:    "get order items returns error on database failure",
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at\s+FROM order_items oi\s+INNER JOIN products p ON oi.product_id = p.id\s+WHERE oi.order_id = \$1`).
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
			store := NewOrderItemStoreWithDB(db)

			got, gotErr := store.GetOrderItemsByOrderID(tt.orderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetOrderItemsByOrderID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetOrderItemsByOrderID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetOrderItemsByOrderID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_orderitem_CreateOrderItem(t *testing.T) {
	type input struct {
		orderID   int
		productID int
		qty       int
	}

	tests := []struct {
		name       string
		input      input
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.OrderItem
		wantErr    bool
	}{
		{
			name: "successfully create order item",
			input: input{
				orderID:   10,
				productID: 5,
				qty:       2,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
				rows := sqlmock.NewRows([]string{"id", "order_id", "product_name", "price", "qty", "created_at"}).
					AddRow(1, 10, "Product A", 1000, 2, fixedTime)
				mock.ExpectQuery(`WITH inserted AS \(\s+INSERT INTO order_items \(order_id, product_id, qty, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4\)\s+RETURNING id, order_id, product_id, qty, created_at\s+\)\s+SELECT i.id, i.order_id, p.name as product_name, p.price as price, i.qty, i.created_at\s+FROM inserted i\s+INNER JOIN products p ON i.product_id = p.id`).
					WithArgs(10, 5, 2, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.OrderItem{
				ID:          1,
				OrderID:     10,
				ProductName: "Product A",
				Price:       1000,
				Qty:         2,
			},
			wantErr: false,
		},
		{
			name: "create order item returns error on database failure",
			input: input{
				orderID:   10,
				productID: 5,
				qty:       2,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH inserted AS \(\s+INSERT INTO order_items \(order_id, product_id, qty, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4\)\s+RETURNING id, order_id, product_id, qty, created_at\s+\)\s+SELECT i.id, i.order_id, p.name as product_name, p.price as price, i.qty, i.created_at\s+FROM inserted i\s+INNER JOIN products p ON i.product_id = p.id`).
					WithArgs(10, 5, 2, sqlmock.AnyArg()).
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
			store := NewOrderItemStoreWithDB(db)

			got, gotErr := store.CreateOrderItem(tt.input.orderID, tt.input.productID, tt.input.qty)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateOrderItem() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateOrderItem() succeeded unexpectedly")
			}

			// Set expected CreatedAt to match for DeepEqual comparison
			tt.wantResult.CreatedAt = got.CreatedAt
			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateOrderItem() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_orderitem_UpdateOrderItemByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name       string
		id         int
		orderID    int
		input      UpdateOrderItemInput
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.OrderItem
		wantErr    bool
	}{
		{
			name:    "update order item with qty",
			id:      1,
			orderID: 10,
			input: UpdateOrderItemInput{
				Qty: intPtr(5),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "order_id", "product_name", "price", "qty", "created_at", "updated_at"}).
					AddRow(1, 10, "Product A", 1000, 5, fixedTime, updatedTime)
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE order_items\s+SET qty = 5,updated_at = now\(\)\s+WHERE id = \$1 AND order_id = \$2\s+RETURNING id, order_id, product_id, qty, created_at, updated_at\s+\)\s+SELECT u.id, u.order_id, p.name as product_name, p.price as price, u.qty, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN products p ON u.product_id = p.id`).
					WithArgs(1, 10).
					WillReturnRows(rows)
			},
			wantResult: &model.OrderItem{
				ID:          1,
				OrderID:     10,
				ProductName: "Product A",
				Price:       1000,
				Qty:         5,
				CreatedAt:   fixedTime,
				UpdatedAt:   sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name:    "update order item with product ID",
			id:      1,
			orderID: 10,
			input: UpdateOrderItemInput{
				ProductID: intPtr(3),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "order_id", "product_name", "price", "qty", "created_at", "updated_at"}).
					AddRow(1, 10, "Product B", 2000, 2, fixedTime, updatedTime)
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE order_items\s+SET product_id = 3,updated_at = now\(\)\s+WHERE id = \$1 AND order_id = \$2\s+RETURNING id, order_id, product_id, qty, created_at, updated_at\s+\)\s+SELECT u.id, u.order_id, p.name as product_name, p.price as price, u.qty, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN products p ON u.product_id = p.id`).
					WithArgs(1, 10).
					WillReturnRows(rows)
			},
			wantResult: &model.OrderItem{
				ID:          1,
				OrderID:     10,
				ProductName: "Product B",
				Price:       2000,
				Qty:         2,
				CreatedAt:   fixedTime,
				UpdatedAt:   sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name:    "update non-existent order item returns error",
			id:      9999,
			orderID: 10,
			input: UpdateOrderItemInput{
				Qty: intPtr(5),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE order_items\s+SET qty = 5,updated_at = now\(\)\s+WHERE id = \$1 AND order_id = \$2\s+RETURNING id, order_id, product_id, qty, created_at, updated_at\s+\)\s+SELECT u.id, u.order_id, p.name as product_name, p.price as price, u.qty, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN products p ON u.product_id = p.id`).
					WithArgs(9999, 10).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:    "update order item returns error on database failure",
			id:      1,
			orderID: 10,
			input: UpdateOrderItemInput{
				Qty: intPtr(5),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`WITH updated AS \(\s+UPDATE order_items\s+SET qty = 5,updated_at = now\(\)\s+WHERE id = \$1 AND order_id = \$2\s+RETURNING id, order_id, product_id, qty, created_at, updated_at\s+\)\s+SELECT u.id, u.order_id, p.name as product_name, p.price as price, u.qty, u.created_at, u.updated_at\s+FROM updated u\s+INNER JOIN products p ON u.product_id = p.id`).
					WithArgs(1, 10).
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
			store := NewOrderItemStoreWithDB(db)

			got, gotErr := store.UpdateOrderItemByID(tt.id, tt.orderID, tt.input)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateOrderItemByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateOrderItemByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateOrderItemByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_orderitem_DeleteOrderItemByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		orderID   int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:    "successfully delete order item",
			id:      1,
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM order_items\s+WHERE id = \$1 AND order_id = \$2`).
					WithArgs(1, 10).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "delete non-existent order item succeeds",
			id:      9999,
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM order_items\s+WHERE id = \$1 AND order_id = \$2`).
					WithArgs(9999, 10).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name:    "delete order item returns error on database failure",
			id:      1,
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM order_items\s+WHERE id = \$1 AND order_id = \$2`).
					WithArgs(1, 10).
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
			store := NewOrderItemStoreWithDB(db)

			gotErr := store.DeleteOrderItemByID(tt.id, tt.orderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteOrderItemByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteOrderItemByID() succeeded unexpectedly")
			}
		})
	}
}

func Test_orderitem_DeleteOrderItemsByOrderID(t *testing.T) {
	tests := []struct {
		name      string
		orderID   int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:    "successfully delete order items by order ID",
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM order_items\s+WHERE order_id = \$1`).
					WithArgs(10).
					WillReturnResult(sqlmock.NewResult(0, 3))
			},
			wantErr: false,
		},
		{
			name:    "delete order items returns error on database failure",
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM order_items\s+WHERE order_id = \$1`).
					WithArgs(10).
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

			var o orderitem
			gotErr := o.DeleteOrderItemsByOrderID(tx, tt.orderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteOrderItemsByOrderID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteOrderItemsByOrderID() succeeded unexpectedly")
			}
		})
	}
}

func Test_orderitem_CreateTempOrderItem(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	queryRegex := `WITH inserted AS \(\s*INSERT INTO temp_order_items \(temp_order_id, product_id, qty, created_at\)\s*VALUES \(\$1, \$2, \$3, \$4\)\s*RETURNING id, temp_order_id, product_id, qty, created_at\s*\)\s*SELECT i\.id, i\.temp_order_id, p\.name as product_name, p\.price as price, i\.qty, i\.created_at\s*FROM inserted i\s*INNER JOIN products p ON i\.product_id = p\.id`

	tests := []struct {
		name        string
		tempOrderID int
		productID   int
		qty         int
		mockSetup   func(mock sqlmock.Sqlmock)
		want        *model.TempOrderItem
		wantErr     bool
	}{
		{
			name:        "successfully create temp order item",
			tempOrderID: 1,
			productID:   10,
			qty:         2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "temp_order_id", "product_name", "price", "qty", "created_at"}).
					AddRow(1, 1, "Product A", 1000, 2, fixedTime)
				mock.ExpectQuery(queryRegex).
					WithArgs(1, 10, 2, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			want: &model.TempOrderItem{
				ID:          1,
				TempOrderID: 1,
				ProductName: "Product A",
				Price:       1000,
				Qty:         2,
				CreatedAt:   fixedTime,
			},
			wantErr: false,
		},
		{
			name:        "create temp order item returns error on database failure",
			tempOrderID: 1,
			productID:   10,
			qty:         2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRegex).
					WithArgs(1, 10, 2, sqlmock.AnyArg()).
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
			store := NewOrderItemStoreWithDB(db)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin tx: %v", err)
			}
			defer tx.Rollback()

			got, gotErr := store.CreateTempOrderItem(tx, tt.tempOrderID, tt.productID, tt.qty)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTempOrderItem() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTempOrderItem() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateTempOrderItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_orderitem_GetTempOrderItemsByTempOrderID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		tempOrderID int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult []model.TempOrderItem
		wantErr    bool
	}{
		{
			name:        "get temp order items by temp order ID returns multiple items",
			tempOrderID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "temp_order_id", "product_name", "price", "qty", "created_at"}).
					AddRow(1, 1, "Product A", 1000, 2, fixedTime).
					AddRow(2, 1, "Product B", 500, 1, fixedTime)
				mock.ExpectQuery(`SELECT ti.id, ti.temp_order_id, p.name as product_name, p.price as price, ti.qty, ti.created_at\s+FROM temp_order_items ti\s+INNER JOIN products p ON ti.product_id = p.id\s+WHERE ti.temp_order_id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: []model.TempOrderItem{
				{ID: 1, TempOrderID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
				{ID: 2, TempOrderID: 1, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
			},
			wantErr: false,
		},
		{
			name:        "get temp order items returns empty slice when none exist",
			tempOrderID: 99,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "temp_order_id", "product_name", "price", "qty", "created_at"})
				mock.ExpectQuery(`SELECT ti.id, ti.temp_order_id, p.name as product_name, p.price as price, ti.qty, ti.created_at\s+FROM temp_order_items ti\s+INNER JOIN products p ON ti.product_id = p.id\s+WHERE ti.temp_order_id = \$1`).
					WithArgs(99).
					WillReturnRows(rows)
			},
			wantResult: []model.TempOrderItem{},
			wantErr:    false,
		},
		{
			name:        "get temp order items returns error on database failure",
			tempOrderID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT ti.id, ti.temp_order_id, p.name as product_name, p.price as price, ti.qty, ti.created_at\s+FROM temp_order_items ti\s+INNER JOIN products p ON ti.product_id = p.id\s+WHERE ti.temp_order_id = \$1`).
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
			store := NewOrderItemStoreWithDB(db)

			got, gotErr := store.GetTempOrderItemsByTempOrderID(tt.tempOrderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTempOrderItemsByTempOrderID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTempOrderItemsByTempOrderID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetTempOrderItemsByTempOrderID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
