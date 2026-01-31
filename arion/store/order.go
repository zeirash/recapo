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
		CreateOrder(customerID int, shopID int, status string, notes *string) (*model.Order, error)
		UpdateOrder(id int, input UpdateOrderInput) (*model.Order, error)
		DeleteOrderByID(tx database.Tx, id int) error
	}

	order struct {
		db *sql.DB
	}

	UpdateOrderInput struct {
		TotalPrice *int
		Status     *string
		Notes      *string
	}
)

func NewOrderStore() OrderStore {
	return &order{db: database.GetDB()}
}

// NewOrderStoreWithDB creates an OrderStore with a custom db connection (for testing)
func NewOrderStoreWithDB(db *sql.DB) OrderStore {
	return &order{db: db}
}

func (o *order) GetOrderByID(id int, shopID ...int) (*model.Order, error) {
	criteria := []interface{}{id}

	q := `
		SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE o.id = $1
	`

	if len(shopID) > 0 {
		q += " AND o.shop_id = $2"
		criteria = append(criteria, shopID[0])
	}

	var order model.Order
	err := o.db.QueryRow(q, criteria...).Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.TotalPrice, &order.Status, &order.Notes, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &order, nil
}

func (o *order) GetOrdersByShopID(shopID int) ([]model.Order, error) {

	q := `
		SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE o.shop_id = $1
	`

	rows, err := o.db.Query(q, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []model.Order{}
	for rows.Next() {
		var order model.Order
		err := rows.Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.TotalPrice, &order.Status, &order.Notes, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (o *order) CreateOrder(customerID int, shopID int, status string, notes *string) (*model.Order, error) {
	now := time.Now()
	var order model.Order

	q := `
		WITH inserted AS (
			INSERT INTO orders (total_price, status, customer_id, shop_id, notes, created_at)
			VALUES ($1, $2, $3, $4, COALESCE($5, ''), $6)
			RETURNING id, total_price, status, customer_id, shop_id, notes, created_at
		)
		SELECT i.id, i.total_price, i.status, c.name as customer_name, i.shop_id, i.notes, i.created_at
		FROM inserted i
		INNER JOIN customers c ON i.customer_id = c.id
	`

	// total price is 0 as default, it will be calculated later
	err := o.db.QueryRow(q, 0, status, customerID, shopID, notes, now).Scan(
		&order.ID,
		&order.TotalPrice,
		&order.Status,
		&order.CustomerName,
		&order.ShopID,
		&order.Notes,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *order) UpdateOrder(id int, input UpdateOrderInput) (*model.Order, error) {
	set := []string{}
	var order model.Order

	// build query
	if input.TotalPrice != nil {
		newSet := fmt.Sprintf("total_price = %d", *input.TotalPrice)
		set = append(set, newSet)
	}
	if input.Status != nil {
		newSet := fmt.Sprintf("status = '%s'", *input.Status)
		set = append(set, newSet)
	}
	if input.Notes != nil {
		newSet := fmt.Sprintf("notes = '%s'", *input.Notes)
		set = append(set, newSet)
	}

	set = append(set, "updated_at = now()")

	q := `
		WITH updated AS (
			UPDATE orders
			SET %s
			WHERE id = $1
			RETURNING id, shop_id, customer_id, total_price, status, notes, created_at, updated_at
		)
		SELECT u.id, u.shop_id, c.name as customer_name, u.total_price, u.status, u.notes, u.created_at, u.updated_at
		FROM updated u
		INNER JOIN customers c ON u.customer_id = c.id
	`

	q = fmt.Sprintf(q, strings.Join(set, ","))

	err := o.db.QueryRow(q, id).Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.TotalPrice, &order.Status, &order.Notes, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *order) DeleteOrderByID(tx database.Tx, id int) error {
	q := `
		DELETE FROM orders
		WHERE id = $1
	`

	_, err := tx.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}
