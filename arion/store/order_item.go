package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	OrderItemStore interface {
		GetOrderItemByID(ctx context.Context, id int) (*model.OrderItem, error)
		GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]model.OrderItem, error)
		CreateOrderItem(ctx context.Context, tx database.Tx, orderID, productID, qty int) (*model.OrderItem, error)
		UpdateOrderItemByID(ctx context.Context, tx database.Tx, id, orderID int, input UpdateOrderItemInput) (*model.OrderItem, error)
		DeleteOrderItemByID(ctx context.Context, id, orderID int) error
		DeleteOrderItemsByOrderID(ctx context.Context, tx database.Tx, orderID int) error
		GetOrderItemByProductID(ctx context.Context, productID int, orderID int) (*model.OrderItem, error)

		CreateTempOrderItem(ctx context.Context, tx database.Tx, tempOrderID, productID, qty int) (*model.TempOrderItem, error)
		GetTempOrderItemsByTempOrderID(ctx context.Context, tempOrderID int) ([]model.TempOrderItem, error)
	}

	orderitem struct {
		db *sql.DB
	}

	UpdateOrderItemInput struct {
		ProductID *int
		Qty       *int
	}
)

func NewOrderItemStore() OrderItemStore {
	return &orderitem{db: database.GetDB()}
}

// NewOrderItemStoreWithDB creates an OrderItemStore with a custom db connection (for testing)
func NewOrderItemStoreWithDB(db *sql.DB) OrderItemStore {
	return &orderitem{db: db}
}

func (o *orderitem) GetOrderItemByID(ctx context.Context, id int) (*model.OrderItem, error) {
	q := `
		SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at
		FROM order_items oi
		INNER JOIN products p ON oi.product_id = p.id
		WHERE oi.id = $1
	`

	var orderItem model.OrderItem
	err := o.db.QueryRowContext(ctx, q, id).Scan(&orderItem.ID, &orderItem.OrderID, &orderItem.ProductName, &orderItem.Price, &orderItem.Qty, &orderItem.CreatedAt, &orderItem.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &orderItem, nil
}

func (o *orderitem) GetOrderItemsByOrderID(ctx context.Context, orderID int) ([]model.OrderItem, error) {
	q := `
		SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at
		FROM order_items oi
		INNER JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
		ORDER BY oi.created_at ASC
	`

	rows, err := o.db.QueryContext(ctx, q, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderItems := []model.OrderItem{}
	for rows.Next() {
		var orderItem model.OrderItem
		err := rows.Scan(&orderItem.ID, &orderItem.OrderID, &orderItem.ProductName, &orderItem.Price, &orderItem.Qty, &orderItem.CreatedAt, &orderItem.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, orderItem)
	}

	return orderItems, nil
}

func (o *orderitem) CreateOrderItem(ctx context.Context, tx database.Tx, orderID, productID, qty int) (*model.OrderItem, error) {
	now := time.Now()
	var orderItem model.OrderItem

	q := `
		WITH inserted AS (
			INSERT INTO order_items (order_id, product_id, qty, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id, order_id, product_id, qty, created_at
		)
		SELECT i.id, i.order_id, p.name as product_name, p.price as price, i.qty, i.created_at
		FROM inserted i
		INNER JOIN products p ON i.product_id = p.id
	`

	args := []interface{}{orderID, productID, qty, now}
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, args...).Scan(
			&orderItem.ID, &orderItem.OrderID, &orderItem.ProductName, &orderItem.Price, &orderItem.Qty, &orderItem.CreatedAt,
		)
	} else {
		err = o.db.QueryRowContext(ctx, q, args...).Scan(
			&orderItem.ID, &orderItem.OrderID, &orderItem.ProductName, &orderItem.Price, &orderItem.Qty, &orderItem.CreatedAt,
		)
	}
	if err != nil {
		return nil, err
	}

	return &orderItem, nil
}

func (o *orderitem) UpdateOrderItemByID(ctx context.Context, tx database.Tx, id, orderID int, input UpdateOrderItemInput) (*model.OrderItem, error) {
	set := []string{}
	var orderItem model.OrderItem

	// build query
	if input.Qty != nil {
		newSet := fmt.Sprintf("qty = %d", *input.Qty)
		set = append(set, newSet)
	}
	if input.ProductID != nil {
		newSet := fmt.Sprintf("product_id = %d", *input.ProductID)
		set = append(set, newSet)
	}

	set = append(set, "updated_at = now()")

	q := `
		WITH updated AS (
			UPDATE order_items
			SET %s
			WHERE id = $1 AND order_id = $2
			RETURNING id, order_id, product_id, qty, created_at, updated_at
		)
		SELECT u.id, u.order_id, p.name as product_name, p.price as price, u.qty, u.created_at, u.updated_at
		FROM updated u
		INNER JOIN products p ON u.product_id = p.id
	`

	q = fmt.Sprintf(q, strings.Join(set, ","))

	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, id, orderID).Scan(&orderItem.ID, &orderItem.OrderID, &orderItem.ProductName, &orderItem.Price, &orderItem.Qty, &orderItem.CreatedAt, &orderItem.UpdatedAt)
	} else {
		err = o.db.QueryRowContext(ctx, q, id, orderID).Scan(&orderItem.ID, &orderItem.OrderID, &orderItem.ProductName, &orderItem.Price, &orderItem.Qty, &orderItem.CreatedAt, &orderItem.UpdatedAt)
	}
	if err != nil {
		return nil, err
	}

	return &orderItem, nil
}

func (o *orderitem) DeleteOrderItemByID(ctx context.Context, id, orderID int) error {
	q := `
		DELETE FROM order_items
		WHERE id = $1 AND order_id = $2
	`

	_, err := o.db.ExecContext(ctx, q, id, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (o *orderitem) DeleteOrderItemsByOrderID(ctx context.Context, tx database.Tx, orderID int) error {
	q := `
		DELETE FROM order_items
		WHERE order_id = $1
	`

	_, err := tx.ExecContext(ctx, q, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (o *orderitem) CreateTempOrderItem(ctx context.Context, tx database.Tx, tempOrderID, productID, qty int) (*model.TempOrderItem, error) {
	now := time.Now()
	var tempOrderItem model.TempOrderItem

	q := `
		WITH inserted AS (
			INSERT INTO temp_order_items (temp_order_id, product_id, qty, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id, temp_order_id, product_id, qty, created_at
		)
		SELECT i.id, i.temp_order_id, p.name as product_name, p.price as price, i.qty, i.created_at
		FROM inserted i
		INNER JOIN products p ON i.product_id = p.id
	`

	err := tx.QueryRowContext(ctx, q, tempOrderID, productID, qty, now).Scan(&tempOrderItem.ID, &tempOrderItem.TempOrderID, &tempOrderItem.ProductName, &tempOrderItem.Price, &tempOrderItem.Qty, &tempOrderItem.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &tempOrderItem, nil
}

func (o *orderitem) GetTempOrderItemsByTempOrderID(ctx context.Context, tempOrderID int) ([]model.TempOrderItem, error) {
	q := `
		SELECT ti.id, ti.temp_order_id, ti.product_id, p.name as product_name, p.price as price, ti.qty, ti.created_at
		FROM temp_order_items ti
		INNER JOIN products p ON ti.product_id = p.id
		WHERE ti.temp_order_id = $1
	`

	rows, err := o.db.QueryContext(ctx, q, tempOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tempOrderItems := []model.TempOrderItem{}
	for rows.Next() {
		var tempOrderItem model.TempOrderItem
		err := rows.Scan(&tempOrderItem.ID, &tempOrderItem.TempOrderID, &tempOrderItem.ProductID, &tempOrderItem.ProductName, &tempOrderItem.Price, &tempOrderItem.Qty, &tempOrderItem.CreatedAt)
		if err != nil {
			return nil, err
		}
		tempOrderItems = append(tempOrderItems, tempOrderItem)
	}

	return tempOrderItems, nil
}

func (o *orderitem) GetOrderItemByProductID(ctx context.Context, productID int, orderID int) (*model.OrderItem, error) {
	q := `
		SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at
		FROM order_items oi
		INNER JOIN products p ON oi.product_id = p.id
		WHERE oi.product_id = $1 AND oi.order_id = $2
	`

	var orderItem model.OrderItem
	err := o.db.QueryRowContext(ctx, q, productID, orderID).Scan(&orderItem.ID, &orderItem.OrderID, &orderItem.ProductName, &orderItem.Price, &orderItem.Qty, &orderItem.CreatedAt, &orderItem.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &orderItem, nil
}
