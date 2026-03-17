package store

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeirash/recapo/arion/model"
)

func Test_orderpayment_CreateOrderPayment(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		orderID   int
		amount    int
		useTx     bool
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:    "successfully create order payment without tx",
			orderID: 10,
			amount:  50000,
			useTx:   false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`INSERT INTO order_payments \(order_id, amount, created_at\)\s+VALUES \(\$1, \$2, \$3\)\s+RETURNING id`).
					WithArgs(10, 50000, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:    "successfully create order payment with tx",
			orderID: 5,
			amount:  100000,
			useTx:   true,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`INSERT INTO order_payments \(order_id, amount, created_at\)\s+VALUES \(\$1, \$2, \$3\)\s+RETURNING id`).
					WithArgs(5, 100000, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:    "returns error on database failure",
			orderID: 10,
			amount:  50000,
			useTx:   false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO order_payments \(order_id, amount, created_at\)\s+VALUES \(\$1, \$2, \$3\)\s+RETURNING id`).
					WithArgs(10, 50000, sqlmock.AnyArg()).
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
			store := NewOrderPaymentStoreWithDB(db)

			var got *model.OrderPayment
			var gotErr error
			if tt.useTx {
				tx, err := db.Begin()
				if err != nil {
					t.Fatalf("failed to begin tx: %v", err)
				}
				defer tx.Rollback()
				got, gotErr = store.CreateOrderPayment(context.Background(), tx, tt.orderID, tt.amount)
			} else {
				got, gotErr = store.CreateOrderPayment(context.Background(), nil, tt.orderID, tt.amount)
			}

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateOrderPayment() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateOrderPayment() succeeded unexpectedly")
			}

			if got.OrderID != tt.orderID {
				t.Errorf("CreateOrderPayment() OrderID = %v, want %v", got.OrderID, tt.orderID)
			}
			if got.Amount != tt.amount {
				t.Errorf("CreateOrderPayment() Amount = %v, want %v", got.Amount, tt.amount)
			}
			_ = fixedTime // CreatedAt is set to time.Now() internally; just verify it's populated
			if got.CreatedAt.IsZero() {
				t.Error("CreateOrderPayment() CreatedAt should not be zero")
			}
		})
	}
}

func Test_orderpayment_GetOrderPaymentsByOrderID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		orderID    int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult []model.OrderPayment
		wantErr    bool
	}{
		{
			name:    "returns multiple payments for order",
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "order_id", "amount", "created_at", "updated_at"}).
					AddRow(1, 10, 50000, fixedTime, nil).
					AddRow(2, 10, 25000, fixedTime, sql.NullTime{Time: updatedTime, Valid: true})
				mock.ExpectQuery(`SELECT id, order_id, amount, created_at, updated_at\s+FROM order_payments\s+WHERE order_id = \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			wantResult: []model.OrderPayment{
				{ID: 1, OrderID: 10, Amount: 50000, CreatedAt: fixedTime},
				{ID: 2, OrderID: 10, Amount: 25000, CreatedAt: fixedTime, UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true}},
			},
			wantErr: false,
		},
		{
			name:    "returns empty slice when no payments exist for order",
			orderID: 9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "order_id", "amount", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT id, order_id, amount, created_at, updated_at\s+FROM order_payments\s+WHERE order_id = \$1`).
					WithArgs(9999).
					WillReturnRows(rows)
			},
			wantResult: []model.OrderPayment{},
			wantErr:    false,
		},
		{
			name:    "returns error on database failure",
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, order_id, amount, created_at, updated_at\s+FROM order_payments\s+WHERE order_id = \$1`).
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
			store := NewOrderPaymentStoreWithDB(db)

			got, gotErr := store.GetOrderPaymentsByOrderID(context.Background(), tt.orderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetOrderPaymentsByOrderID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetOrderPaymentsByOrderID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetOrderPaymentsByOrderID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_orderpayment_UpdateOrderPaymentAmountByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		id         int
		orderID    int
		amount     int
		useTx      bool
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.OrderPayment
		wantErr    bool
	}{
		{
			name:    "successfully update payment amount without tx",
			id:      1,
			orderID: 10,
			amount:  75000,
			useTx:   false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "order_id", "amount", "created_at", "updated_at"}).
					AddRow(1, 10, 75000, fixedTime, sql.NullTime{Time: updatedTime, Valid: true})
				mock.ExpectQuery(`UPDATE order_payments\s+SET amount = \$1, updated_at = \$2\s+WHERE id = \$3 AND order_id = \$4\s+RETURNING id, order_id, amount, created_at, updated_at`).
					WithArgs(75000, sqlmock.AnyArg(), 1, 10).
					WillReturnRows(rows)
			},
			wantResult: &model.OrderPayment{
				ID:        1,
				OrderID:   10,
				Amount:    75000,
				CreatedAt: fixedTime,
				UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name:    "successfully update payment amount with tx",
			id:      2,
			orderID: 5,
			amount:  30000,
			useTx:   true,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "order_id", "amount", "created_at", "updated_at"}).
					AddRow(2, 5, 30000, fixedTime, sql.NullTime{Time: updatedTime, Valid: true})
				mock.ExpectQuery(`UPDATE order_payments\s+SET amount = \$1, updated_at = \$2\s+WHERE id = \$3 AND order_id = \$4\s+RETURNING id, order_id, amount, created_at, updated_at`).
					WithArgs(30000, sqlmock.AnyArg(), 2, 5).
					WillReturnRows(rows)
			},
			wantResult: &model.OrderPayment{
				ID:        2,
				OrderID:   5,
				Amount:    30000,
				CreatedAt: fixedTime,
				UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name:    "returns error when payment not found",
			id:      9999,
			orderID: 10,
			amount:  50000,
			useTx:   false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE order_payments\s+SET amount = \$1, updated_at = \$2\s+WHERE id = \$3 AND order_id = \$4\s+RETURNING id, order_id, amount, created_at, updated_at`).
					WithArgs(50000, sqlmock.AnyArg(), 9999, 10).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:    "returns error on database failure",
			id:      1,
			orderID: 10,
			amount:  50000,
			useTx:   false,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE order_payments\s+SET amount = \$1, updated_at = \$2\s+WHERE id = \$3 AND order_id = \$4\s+RETURNING id, order_id, amount, created_at, updated_at`).
					WithArgs(50000, sqlmock.AnyArg(), 1, 10).
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
			store := NewOrderPaymentStoreWithDB(db)

			var got *model.OrderPayment
			var gotErr error
			if tt.useTx {
				tx, err := db.Begin()
				if err != nil {
					t.Fatalf("failed to begin tx: %v", err)
				}
				defer tx.Rollback()
				got, gotErr = store.UpdateOrderPaymentAmountByID(context.Background(), tx, tt.id, tt.orderID, tt.amount)
			} else {
				got, gotErr = store.UpdateOrderPaymentAmountByID(context.Background(), nil, tt.id, tt.orderID, tt.amount)
			}

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateOrderPaymentAmountByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateOrderPaymentAmountByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateOrderPaymentAmountByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_orderpayment_DeleteOrderPaymentByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		orderID   int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:    "successfully delete order payment",
			id:      1,
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM order_payments\s+WHERE id = \$1 AND order_id = \$2`).
					WithArgs(1, 10).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "delete non-existent payment succeeds",
			id:      9999,
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM order_payments\s+WHERE id = \$1 AND order_id = \$2`).
					WithArgs(9999, 10).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name:    "returns error on database failure",
			id:      1,
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM order_payments\s+WHERE id = \$1 AND order_id = \$2`).
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
			store := NewOrderPaymentStoreWithDB(db)

			gotErr := store.DeleteOrderPaymentByID(context.Background(), tt.id, tt.orderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteOrderPaymentByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteOrderPaymentByID() succeeded unexpectedly")
			}
		})
	}
}

func Test_orderpayment_DeleteOrderPaymentsByOrderID(t *testing.T) {
	tests := []struct {
		name      string
		orderID   int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:    "successfully delete all payments for order",
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM order_payments\s+WHERE order_id = \$1`).
					WithArgs(10).
					WillReturnResult(sqlmock.NewResult(0, 3))
			},
			wantErr: false,
		},
		{
			name:    "succeeds when no payments exist for order",
			orderID: 9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM order_payments\s+WHERE order_id = \$1`).
					WithArgs(9999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name:    "returns error on database failure",
			orderID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`DELETE FROM order_payments\s+WHERE order_id = \$1`).
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

			var o orderpayment
			gotErr := o.DeleteOrderPaymentsByOrderID(context.Background(), tx, tt.orderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteOrderPaymentsByOrderID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteOrderPaymentsByOrderID() succeeded unexpectedly")
			}
		})
	}
}
