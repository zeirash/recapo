package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	OrderItemStore interface {
		GetOrderItemByID(id int) (*model.OrderItem, error)
		GetOrderItemsByOrderID(orderID int) ([]model.OrderItem, error)
		CreateOrderItem(orderID, productID, qty int) (*model.OrderItem, error)
		UpdateOrderItemByID(id, orderID int, input UpdateOrderItemInput) (*model.OrderItem, error)
		DeleteOrderItemByID(id, orderID int) error
		DeleteOrderItemsByOrderID(tx *sql.Tx, orderID int) error
	}

	orderitem struct{}

	UpdateOrderItemInput struct {
		ProductID *int
		Qty       *int
	}
)

func NewOrderItemStore() OrderItemStore {
	return &orderitem{}
}

func (o *orderitem) GetOrderItemByID(id int) (*model.OrderItem, error) {
	db := database.GetDB()
	defer db.Close()

	q := `
		SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at
		FROM order_items oi
		INNER JOIN products p ON oi.product_id = p.id
		WHERE oi.id = $1
	`

	var orderItem model.OrderItem
	err := db.QueryRow(q, id).Scan(&orderItem.ID, &orderItem.OrderID, &orderItem.ProductName, &orderItem.Price, &orderItem.Qty, &orderItem.CreatedAt, &orderItem.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &orderItem, nil
}

func (o *orderitem) GetOrderItemsByOrderID(orderID int) ([]model.OrderItem, error) {
	db := database.GetDB()
	defer db.Close()

	q := `
		SELECT oi.id, oi.order_id, p.name as product_name, p.price as price, oi.qty, oi.created_at, oi.updated_at
		FROM order_items oi
		INNER JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
	`

	rows, err := db.Query(q, orderID)
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

func (o *orderitem) CreateOrderItem(orderID, productID, qty int) (*model.OrderItem, error) {
	db := database.GetDB()
	defer db.Close()

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

	err := db.QueryRow(q, orderID, productID, qty, now).Scan(
		&orderItem.ID,
		&orderItem.OrderID,
		&orderItem.ProductName,
		&orderItem.Price,
		&orderItem.Qty,
		&orderItem.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &orderItem, nil
}

func (o *orderitem) UpdateOrderItemByID(id, orderID int, input UpdateOrderItemInput) (*model.OrderItem, error) {
	db := database.GetDB()
	defer db.Close()

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

	err := db.QueryRow(q, id, orderID).Scan(&orderItem.ID, &orderItem.OrderID, &orderItem.ProductName, &orderItem.Price, &orderItem.Qty, &orderItem.CreatedAt, &orderItem.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &orderItem, nil
}

func (o *orderitem) DeleteOrderItemByID(id, orderID int) error {
	db := database.GetDB()
	defer db.Close()

	q := `
		DELETE FROM order_items
		WHERE id = $1 AND order_id = $2
	`

	_, err := db.Exec(q, id, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (o *orderitem) DeleteOrderItemsByOrderID(tx *sql.Tx, orderID int) error {
	q := `
		DELETE FROM order_items
		WHERE order_id = $1
	`

	_, err := tx.Exec(q, orderID)
	if err != nil {
		return err
	}

	return nil
}
