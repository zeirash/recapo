package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/model"
)

type (
	OrderStore interface {
		GetOrderByID(id int, shopID ...int) (*model.Order, error)
		GetOrdersByShopID(shopID int, opts model.OrderFilterOptions) ([]model.Order, error)
		CreateOrder(customerID int, shopID int, notes *string) (*model.Order, error)
		UpdateOrder(id int, input UpdateOrderInput) (*model.Order, error)
		DeleteOrderByID(tx database.Tx, id int) error

		CreateTempOrder(tx database.Tx, customerName, customerPhone string, shopID int) (*model.TempOrder, error)
		UpdateTempOrderTotalPrice(tx database.Tx, tempOrderID int, totalPrice int) error
		GetTempOrderByID(id int, shopID ...int) (*model.TempOrder, error)
		GetTempOrdersByShopID(shopID int, opts model.OrderFilterOptions) ([]model.TempOrder, error)
		// UpdateOrderTempByID(id int, input UpdateOrderTempInput) (*model.OrderTemp, error)
		// DeleteOrderTempByID(tx database.Tx, id int) error
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

func (o *order) GetOrdersByShopID(shopID int, opts model.OrderFilterOptions) ([]model.Order, error) {
	q := `
		SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.notes, o.created_at, o.updated_at
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE o.shop_id = $1
	`
	args := []interface{}{shopID}
	argNum := 2

	if opts.SearchQuery != nil && strings.TrimSpace(*opts.SearchQuery) != "" {
		q += fmt.Sprintf(" AND (c.name ILIKE $%d OR c.phone ILIKE $%d)", argNum, argNum)
		args = append(args, "%"+strings.TrimSpace(*opts.SearchQuery)+"%")
		argNum++
	}
	if opts.DateFrom != nil {
		q += fmt.Sprintf(" AND o.created_at >= $%d", argNum)
		args = append(args, *opts.DateFrom)
		argNum++
	}
	if opts.DateTo != nil {
		q += fmt.Sprintf(" AND o.created_at < $%d", argNum)
		args = append(args, *opts.DateTo)
		argNum++
	}

	rows, err := o.db.Query(q, args...)
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

func (o *order) CreateOrder(customerID int, shopID int, notes *string) (*model.Order, error) {
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
	err := o.db.QueryRow(q, 0, constant.OrderStatusCreated, customerID, shopID, notes, now).Scan(
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

func (o *order) CreateTempOrder(tx database.Tx, customerName, customerPhone string, shopID int) (*model.TempOrder, error) {
	now := time.Now()
	var tempOrder model.TempOrder

	q := `
		INSERT INTO temp_orders (customer_name, customer_phone, status, shop_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, customer_name, customer_phone, shop_id, total_price, status, created_at
	`

	err := tx.QueryRow(q, customerName, customerPhone, constant.TempOrderStatusPending, shopID, now).Scan(&tempOrder.ID, &tempOrder.CustomerName, &tempOrder.CustomerPhone, &tempOrder.ShopID, &tempOrder.TotalPrice, &tempOrder.Status, &tempOrder.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &tempOrder, nil
}

func (o *order) UpdateTempOrderTotalPrice(tx database.Tx, tempOrderID int, totalPrice int) error {
	q := `
		UPDATE temp_orders
		SET total_price = $1, updated_at = now()
		WHERE id = $2
	`
	_, err := tx.Exec(q, totalPrice, tempOrderID)
	if err != nil {
		return err
	}
	return nil
}

func (o *order) GetTempOrderByID(id int, shopID ...int) (*model.TempOrder, error) {
	criteria := []interface{}{id}

	q := `
		SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at
		FROM temp_orders
		WHERE id = $1
	`

	if len(shopID) > 0 {
		q += " AND shop_id = $2"
		criteria = append(criteria, shopID[0])
	}

	var tempOrder model.TempOrder
	err := o.db.QueryRow(q, criteria...).Scan(&tempOrder.ID, &tempOrder.ShopID, &tempOrder.CustomerName, &tempOrder.CustomerPhone, &tempOrder.TotalPrice, &tempOrder.Status, &tempOrder.CreatedAt, &tempOrder.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &tempOrder, nil
}

func (o *order) GetTempOrdersByShopID(shopID int, opts model.OrderFilterOptions) ([]model.TempOrder, error) {
	q := `
		SELECT id, shop_id, customer_name, customer_phone, total_price, status, created_at, updated_at
		FROM temp_orders
		WHERE shop_id = $1
	`
	args := []interface{}{shopID}
	argNum := 2

	if opts.SearchQuery != nil && strings.TrimSpace(*opts.SearchQuery) != "" {
		q += fmt.Sprintf(" AND (customer_name ILIKE $%d OR customer_phone ILIKE $%d)", argNum, argNum)
		args = append(args, "%"+strings.TrimSpace(*opts.SearchQuery)+"%")
		argNum++
	}
	if opts.DateFrom != nil {
		q += fmt.Sprintf(" AND created_at >= $%d", argNum)
		args = append(args, *opts.DateFrom)
		argNum++
	}
	if opts.DateTo != nil {
		q += fmt.Sprintf(" AND created_at < $%d", argNum)
		args = append(args, *opts.DateTo)
		argNum++
	}

	rows, err := o.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tempOrders := []model.TempOrder{}
	for rows.Next() {
		var tempOrder model.TempOrder
		err := rows.Scan(&tempOrder.ID, &tempOrder.ShopID, &tempOrder.CustomerName, &tempOrder.CustomerPhone, &tempOrder.TotalPrice, &tempOrder.Status, &tempOrder.CreatedAt, &tempOrder.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tempOrders = append(tempOrders, tempOrder)
	}

	return tempOrders, nil
}
