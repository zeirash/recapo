package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	OrderPaymentStore interface {
		CreateOrderPayment(ctx context.Context, tx database.Tx, orderID int, amount int) (*model.OrderPayment, error)
		GetOrderPaymentsByOrderID(ctx context.Context, orderID int) ([]model.OrderPayment, error)
		GetPaymentsSumByShopID(ctx context.Context, shopID int, opts model.OrderFilterOptions) (int, error)
		UpdateOrderPaymentAmountByID(ctx context.Context, tx database.Tx, id, orderID, amount int) (*model.OrderPayment, error)
		DeleteOrderPaymentByID(ctx context.Context, id, orderID int) error
		DeleteOrderPaymentsByOrderID(ctx context.Context, tx database.Tx, orderID int) error
	}

	orderpayment struct {
		db *sql.DB
	}

	CreateOrderPaymentInput struct {
		OrderID int
		Amount  int
	}
)

func NewOrderPaymentStore() OrderPaymentStore {
	return &orderpayment{db: database.GetDB()}
}

// NewOrderPaymentStoreWithDB creates an OrderPaymentStore with a custom db connection (for testing)
func NewOrderPaymentStoreWithDB(db *sql.DB) OrderPaymentStore {
	return &orderpayment{db: db}
}

func (o *orderpayment) CreateOrderPayment(ctx context.Context, tx database.Tx, orderID int, amount int) (*model.OrderPayment, error) {
	now := time.Now()
	q := `
		INSERT INTO order_payments (order_id, amount, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int
	var err error
	args := []interface{}{orderID, amount, now}
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, args...).Scan(&id)
	} else {
		err = o.db.QueryRowContext(ctx, q, args...).Scan(&id)
	}
	if err != nil {
		return nil, err
	}

	return &model.OrderPayment{
		ID:        id,
		OrderID:   orderID,
		Amount:    amount,
		CreatedAt: now,
	}, nil
}

func (o *orderpayment) GetOrderPaymentsByOrderID(ctx context.Context, orderID int) ([]model.OrderPayment, error) {
	q := `
		SELECT id, order_id, amount, created_at, updated_at
		FROM order_payments
		WHERE order_id = $1
	`
	rows, err := o.db.QueryContext(ctx, q, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderPayments := []model.OrderPayment{}
	for rows.Next() {
		var orderPayment model.OrderPayment
		err := rows.Scan(&orderPayment.ID, &orderPayment.OrderID, &orderPayment.Amount, &orderPayment.CreatedAt, &orderPayment.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orderPayments = append(orderPayments, orderPayment)
	}

	return orderPayments, nil
}

func (o *orderpayment) UpdateOrderPaymentAmountByID(ctx context.Context, tx database.Tx, id, orderID, amount int) (*model.OrderPayment, error) {
	now := time.Now()
	var orderPayment model.OrderPayment

	q := `
		UPDATE order_payments
		SET amount = $1, updated_at = $2
		WHERE id = $3 AND order_id = $4
		RETURNING id, order_id, amount, created_at, updated_at
	`

	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, amount, now, id, orderID).Scan(&orderPayment.ID, &orderPayment.OrderID, &orderPayment.Amount, &orderPayment.CreatedAt, &orderPayment.UpdatedAt)
	} else {
		err = o.db.QueryRowContext(ctx, q, amount, now, id, orderID).Scan(&orderPayment.ID, &orderPayment.OrderID, &orderPayment.Amount, &orderPayment.CreatedAt, &orderPayment.UpdatedAt)
	}
	if err != nil {
		return nil, err
	}
	return &orderPayment, nil
}

func (o *orderpayment) DeleteOrderPaymentByID(ctx context.Context, id, orderID int) error {
	q := `
		DELETE FROM order_payments
		WHERE id = $1 AND order_id = $2
	`

	_, err := o.db.ExecContext(ctx, q, id, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (o *orderpayment) DeleteOrderPaymentsByOrderID(ctx context.Context, tx database.Tx, orderID int) error {
	q := `
		DELETE FROM order_payments
		WHERE order_id = $1
	`

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, q, orderID)
	} else {
		_, err = o.db.ExecContext(ctx, q, orderID)
	}
	if err != nil {
		return err
	}
	return nil
}

func (o *orderpayment) GetPaymentsSumByShopID(ctx context.Context, shopID int, opts model.OrderFilterOptions) (int, error) {
	args := []interface{}{shopID}
	q := `
		SELECT COALESCE(SUM(op.amount), 0)
		FROM order_payments op
		JOIN orders ord ON ord.id = op.order_id
		WHERE ord.shop_id = $1
	`

	argIdx := 2
	if opts.DateFrom != nil {
		q += fmt.Sprintf(" AND op.created_at::date >= $%d", argIdx)
		args = append(args, *opts.DateFrom)
		argIdx++
	}
	if opts.DateTo != nil {
		q += fmt.Sprintf(" AND op.created_at::date <= $%d", argIdx)
		args = append(args, *opts.DateTo)
		argIdx++
	}

	var total int
	err := o.db.QueryRowContext(ctx, q, args...).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}
