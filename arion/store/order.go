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
	OrderStore interface {
		GetOrderByID(id int, shopID ...int) (*model.Order, error)
		GetOrdersByShopID(shopID int) ([]model.Order, error)
		CreateOrder(customerID int, shopID int, status string) (*model.Order, error)
		UpdateOrder(id int, input UpdateOrderInput) (*model.Order, error)
		DeleteOrderByID(id int) error
	}

	order struct{}

	UpdateOrderInput struct {
		CustomerID *int
		TotalPrice *int
		Status     *string
	}
)

func NewOrderStore() OrderStore {
	return &order{}
}

func (o *order) GetOrderByID(id int, shopID ...int) (*model.Order, error) {
	db := database.GetDB()
	defer db.Close()

	criteria := []interface{}{id}

	q := `
		SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.created_at, o.updated_at
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE o.id = $1
	`

	if len(shopID) > 0 {
		q += " AND o.shop_id = $2"
		criteria = append(criteria, shopID[0])
	}

	var order model.Order
	err := db.QueryRow(q, criteria...).Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &order, nil
}

func (o *order) GetOrdersByShopID(shopID int) ([]model.Order, error) {
	db := database.GetDB()
	defer db.Close()

	q := `
		SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.created_at, o.updated_at
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE o.shop_id = $1
	`

	rows, err := db.Query(q, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []model.Order{}
	for rows.Next() {
		var order model.Order
		err := rows.Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (o *order) CreateOrder(customerID int, shopID int, status string) (*model.Order, error) {
	db := database.GetDB()
	defer db.Close()

	now := time.Now()
	var order model.Order

	q := `
		WITH inserted AS (
			INSERT INTO orders (total_price, status, customer_id, shop_id, created_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, total_price, status, customer_id, shop_id, created_at
		)
		SELECT i.id, i.total_price, i.status, c.name as customer_name, i.shop_id, i.created_at
		FROM inserted i
		INNER JOIN customers c ON i.customer_id = c.id
	`

	// total price is 0 as default, it will be calculated later
	err := db.QueryRow(q, 0, status, customerID, shopID, now).Scan(
		&order.ID,
		&order.TotalPrice,
		&order.Status,
		&order.CustomerName,
		&order.ShopID,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *order) UpdateOrder(id int, input UpdateOrderInput) (*model.Order, error) {
	db := database.GetDB()
	defer db.Close()

	set := []string{}
	var order model.Order

	// build query
	if input.CustomerID != nil {
		newSet := fmt.Sprintf("customer_id = %d", *input.CustomerID)
		set = append(set, newSet)
	}
	if input.TotalPrice != nil {
		newSet := fmt.Sprintf("total_price = %d", *input.TotalPrice)
		set = append(set, newSet)
	}
	if input.Status != nil {
		newSet := fmt.Sprintf("status = '%s'", *input.Status)
		set = append(set, newSet)
	}

	set = append(set, "updated_at = now()")

	q := `
		WITH updated AS (
			UPDATE orders
			SET %s
			WHERE id = $1
			RETURNING id, shop_id, customer_id, total_price, status, created_at, updated_at
		)
		SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.created_at, u.updated_at
		FROM updated u
		INNER JOIN customers c ON u.customer_id = c.id
	`

	q = fmt.Sprintf(q, strings.Join(set, ","))

	err := db.QueryRow(q, id).Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.TotalPrice, &order.Status, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *order) DeleteOrderByID(id int) error {
	db := database.GetDB()
	defer db.Close()

	q := `
		DELETE FROM orders
		WHERE id = $1
	`

	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}
