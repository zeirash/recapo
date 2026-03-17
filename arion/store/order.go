package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	OrderStore interface {
		GetOrderByID(ctx context.Context, id int, shopID ...int) (*model.Order, error)
		GetOrdersByShopID(ctx context.Context, shopID int, opts model.OrderFilterOptions) ([]model.Order, error)
		GetActiveOrderByCustomerID(ctx context.Context, customerID int, shopID int) (*model.Order, error)
		CreateOrder(ctx context.Context, tx database.Tx, customerID int, shopID int, notes *string, totalPrice *int) (*model.Order, error)
		UpdateOrder(ctx context.Context, tx database.Tx, id int, input UpdateOrderInput) (*model.Order, error)
		DeleteOrderByID(ctx context.Context, tx database.Tx, id int) error

		CreateTempOrder(ctx context.Context, tx database.Tx, customerName, customerPhone string, shopID int) (*model.TempOrder, error)
		UpdateTempOrderTotalPrice(ctx context.Context, tx database.Tx, tempOrderID int, totalPrice int) error
		GetTempOrderByID(ctx context.Context, id int, shopID ...int) (*model.TempOrder, error)
		GetTempOrdersByShopID(ctx context.Context, shopID int, opts model.OrderFilterOptions) ([]model.TempOrder, error)
		UpdateTempOrderStatus(ctx context.Context, tx database.Tx, tempOrderID int, status string) error
	}

	order struct {
		db *sql.DB
	}

	UpdateOrderInput struct {
		TotalPrice    *int
		Status        *string
		PaymentStatus *string
		Notes         *string
	}
)

func NewOrderStore() OrderStore {
	return &order{db: database.GetDB()}
}

// NewOrderStoreWithDB creates an OrderStore with a custom db connection (for testing)
func NewOrderStoreWithDB(db *sql.DB) OrderStore {
	return &order{db: db}
}

func (o *order) GetOrderByID(ctx context.Context, id int, shopID ...int) (*model.Order, error) {
	criteria := []interface{}{id}

	q := `
		SELECT o.id, o.shop_id, c.name as customer_name, (c.deleted_at IS NOT NULL) as is_customer_deleted, o.total_price, o.status, o.payment_status, o.notes, o.created_at, o.updated_at
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE o.id = $1
	`

	if len(shopID) > 0 {
		q += " AND o.shop_id = $2"
		criteria = append(criteria, shopID[0])
	}

	var order model.Order
	err := o.db.QueryRowContext(ctx, q, criteria...).Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.IsCustomerDeleted, &order.TotalPrice, &order.Status, &order.PaymentStatus, &order.Notes, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &order, nil
}

func (o *order) GetOrdersByShopID(ctx context.Context, shopID int, opts model.OrderFilterOptions) ([]model.Order, error) {
	q := `
		SELECT o.id, o.shop_id, c.name as customer_name, (c.deleted_at IS NOT NULL) as is_customer_deleted, o.total_price, o.status, o.payment_status, o.notes, o.created_at, o.updated_at
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
	if len(opts.Status) > 0 {
		q += fmt.Sprintf(" AND o.status = ANY($%d)", argNum)
		args = append(args, pq.Array(opts.Status))
		argNum++
	}
	if opts.PaymentStatus != nil {
		q += fmt.Sprintf(" AND o.payment_status = $%d", argNum)
		args = append(args, *opts.PaymentStatus)
		argNum++
	}
	if opts.Sort != nil {
		sort := strings.Split(*opts.Sort, ",")
		if len(sort) == 2 {
			q += fmt.Sprintf(" ORDER BY %s %s", sort[0], sort[1])
		}
	}

	rows, err := o.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []model.Order{}
	for rows.Next() {
		var order model.Order
		err := rows.Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.IsCustomerDeleted, &order.TotalPrice, &order.Status, &order.PaymentStatus, &order.Notes, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (o *order) GetActiveOrderByCustomerID(ctx context.Context, customerID int, shopID int) (*model.Order, error) {
	q := `
		SELECT o.id, o.shop_id, c.name as customer_name, o.total_price, o.status, o.payment_status, o.notes, o.created_at, o.updated_at
		FROM orders o
		INNER JOIN customers c ON o.customer_id = c.id
		WHERE o.customer_id = $1 AND o.shop_id = $2 AND o.status IN ($3, $4)
		ORDER BY o.created_at DESC
		LIMIT 1
	`
	var order model.Order
	err := o.db.QueryRowContext(ctx, q, customerID, shopID, constant.OrderStatusCreated, constant.OrderStatusInProgress).Scan(
		&order.ID,
		&order.ShopID,
		&order.CustomerName,
		&order.TotalPrice,
		&order.Status,
		&order.PaymentStatus,
		&order.Notes,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (o *order) CreateOrder(ctx context.Context, tx database.Tx, customerID int, shopID int, notes *string, totalPrice *int) (*model.Order, error) {
	now := time.Now()
	var order model.Order

	totalPriceVal := 0
	if totalPrice != nil {
		totalPriceVal = *totalPrice
	}

	q := `
		WITH inserted AS (
			INSERT INTO orders (total_price, status, payment_status, customer_id, shop_id, notes, created_at)
			VALUES ($1, $2, $3, $4, $5, COALESCE($6, ''), $7)
			RETURNING id, total_price, status, payment_status, customer_id, shop_id, notes, created_at
		)
		SELECT i.id, i.total_price, i.status, i.payment_status, c.name as customer_name, i.shop_id, i.notes, i.created_at
		FROM inserted i
		INNER JOIN customers c ON i.customer_id = c.id
	`

	args := []interface{}{totalPriceVal, constant.OrderStatusCreated, constant.OrderPaymentStatusOutstanding, customerID, shopID, notes, now}
	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, args...).Scan(
			&order.ID, &order.TotalPrice, &order.Status, &order.PaymentStatus, &order.CustomerName, &order.ShopID, &order.Notes, &order.CreatedAt,
		)
	} else {
		err = o.db.QueryRowContext(ctx, q, args...).Scan(
			&order.ID, &order.TotalPrice, &order.Status, &order.PaymentStatus, &order.CustomerName, &order.ShopID, &order.Notes, &order.CreatedAt,
		)
	}
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *order) UpdateOrder(ctx context.Context, tx database.Tx, id int, input UpdateOrderInput) (*model.Order, error) {
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
	if input.PaymentStatus != nil {
		newSet := fmt.Sprintf("payment_status = '%s'", *input.PaymentStatus)
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
			RETURNING id, shop_id, customer_id, total_price, status, payment_status, notes, created_at, updated_at
		)
		SELECT u.id, u.shop_id, c.name as customer_name, (c.deleted_at IS NOT NULL) as is_customer_deleted, u.total_price, u.status, u.payment_status, u.notes, u.created_at, u.updated_at
		FROM updated u
		INNER JOIN customers c ON u.customer_id = c.id
	`

	q = fmt.Sprintf(q, strings.Join(set, ","))

	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, q, id).Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.IsCustomerDeleted, &order.TotalPrice, &order.Status, &order.PaymentStatus, &order.Notes, &order.CreatedAt, &order.UpdatedAt)
	} else {
		err = o.db.QueryRowContext(ctx, q, id).Scan(&order.ID, &order.ShopID, &order.CustomerName, &order.IsCustomerDeleted, &order.TotalPrice, &order.Status, &order.PaymentStatus, &order.Notes, &order.CreatedAt, &order.UpdatedAt)
	}
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (o *order) DeleteOrderByID(ctx context.Context, tx database.Tx, id int) error {
	q := `
		DELETE FROM orders
		WHERE id = $1
	`

	_, err := tx.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	return nil
}

func (o *order) CreateTempOrder(ctx context.Context, tx database.Tx, customerName, customerPhone string, shopID int) (*model.TempOrder, error) {
	now := time.Now()
	var tempOrder model.TempOrder

	q := `
		INSERT INTO temp_orders (customer_name, customer_phone, status, shop_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, customer_name, customer_phone, shop_id, total_price, status, created_at
	`

	err := tx.QueryRowContext(ctx, q, customerName, customerPhone, constant.TempOrderStatusPending, shopID, now).Scan(&tempOrder.ID, &tempOrder.CustomerName, &tempOrder.CustomerPhone, &tempOrder.ShopID, &tempOrder.TotalPrice, &tempOrder.Status, &tempOrder.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &tempOrder, nil
}

func (o *order) UpdateTempOrderTotalPrice(ctx context.Context, tx database.Tx, tempOrderID int, totalPrice int) error {
	q := `
		UPDATE temp_orders
		SET total_price = $1, updated_at = now()
		WHERE id = $2
	`
	_, err := tx.ExecContext(ctx, q, totalPrice, tempOrderID)
	if err != nil {
		return err
	}
	return nil
}

func (o *order) GetTempOrderByID(ctx context.Context, id int, shopID ...int) (*model.TempOrder, error) {
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
	err := o.db.QueryRowContext(ctx, q, criteria...).Scan(&tempOrder.ID, &tempOrder.ShopID, &tempOrder.CustomerName, &tempOrder.CustomerPhone, &tempOrder.TotalPrice, &tempOrder.Status, &tempOrder.CreatedAt, &tempOrder.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &tempOrder, nil
}

func (o *order) GetTempOrdersByShopID(ctx context.Context, shopID int, opts model.OrderFilterOptions) ([]model.TempOrder, error) {
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
	if len(opts.Status) > 0 {
		q += fmt.Sprintf(" AND status = ANY($%d)", argNum)
		args = append(args, pq.Array(opts.Status))
		argNum++
	}
	if opts.Sort != nil {
		sort := strings.Split(*opts.Sort, ",")
		if len(sort) == 2 {
			q += fmt.Sprintf(" ORDER BY %s %s", sort[0], sort[1])
		}
	}

	rows, err := o.db.QueryContext(ctx, q, args...)
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

func (o *order) UpdateTempOrderStatus(ctx context.Context, tx database.Tx, tempOrderID int, status string) error {
	q := `
		UPDATE temp_orders
		SET status = $1, updated_at = now()
		WHERE id = $2
	`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, q, status, tempOrderID)
	} else {
		_, err = o.db.ExecContext(ctx, q, status, tempOrderID)
	}
	if err != nil {
		return err
	}
	return nil
}
