package store

import (
	"database/sql"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	OrderPaymentStore interface {
		CreateOrderPayment(tx database.Tx, orderID int, amount int) (*model.OrderPayment, error)
		GetOrderPaymentsByOrderID(orderID int) ([]model.OrderPayment, error)
		UpdateOrderPaymentAmountByID(tx database.Tx, id, orderID, amount int) (*model.OrderPayment, error)
		DeleteOrderPaymentByID(id, orderID int) error
		DeleteOrderPaymentsByOrderID(tx database.Tx, orderID int) error
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

func (o *orderpayment) CreateOrderPayment(tx database.Tx, orderID int, amount int) (*model.OrderPayment, error) {
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
		err = tx.QueryRow(q, args...).Scan(&id)
	} else {
		err = o.db.QueryRow(q, args...).Scan(&id)
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

func (o *orderpayment) GetOrderPaymentsByOrderID(orderID int) ([]model.OrderPayment, error) {
	q := `
		SELECT id, order_id, amount, created_at, updated_at
		FROM order_payments
		WHERE order_id = $1
	`
	rows, err := o.db.Query(q, orderID)
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

func (o *orderpayment) UpdateOrderPaymentAmountByID(tx database.Tx, id, orderID, amount int) (*model.OrderPayment, error) {
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
		err = tx.QueryRow(q, amount, now, id, orderID).Scan(&orderPayment.ID, &orderPayment.OrderID, &orderPayment.Amount, &orderPayment.CreatedAt, &orderPayment.UpdatedAt)
	} else {
		err = o.db.QueryRow(q, amount, now, id, orderID).Scan(&orderPayment.ID, &orderPayment.OrderID, &orderPayment.Amount, &orderPayment.CreatedAt, &orderPayment.UpdatedAt)
	}
	if err != nil {
		return nil, err
	}
	return &orderPayment, nil
}

func (o *orderpayment) DeleteOrderPaymentByID(id, orderID int) error {
	q := `
		DELETE FROM order_payments
		WHERE id = $1 AND order_id = $2
	`

	_, err := o.db.Exec(q, id, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (o *orderpayment) DeleteOrderPaymentsByOrderID(tx database.Tx, orderID int) error {
	q := `
		DELETE FROM order_payments
		WHERE order_id = $1
	`

	_, err := tx.Exec(q, orderID)
	if err != nil {
		return err
	}
	return nil
}
