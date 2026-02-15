//go:generate mockgen -source=customer.go -destination=mock/mock_customer.go -package=mock

package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

var ErrDuplicatePhone = errors.New("customer with this phone number already exists")

type (
	CustomerStore interface {
		GetCustomerByID(id int, shopID ...int) (*model.Customer, error)
		GetCustomerByPhone(phone string, shopID int) (*model.Customer, error)
		GetCustomersByShopID(shopID int, searchQuery *string) ([]model.Customer, error)
		CreateCustomer(input CreateCustomerInput) (*model.Customer, error)
		UpdateCustomer(id int, input UpdateCustomerInput) (*model.Customer, error)
		DeleteCustomerByID(id int) error
	}

	customer struct {
		db *sql.DB
	}

	CreateCustomerInput struct {
		ShopID  int
		Name    string
		Phone   string
		Address *string
	}

	UpdateCustomerInput struct {
		Name    *string
		Phone   *string
		Address *string
	}
)

func NewCustomerStore() CustomerStore {
	return &customer{db: database.GetDB()}
}

// NewCustomerStoreWithDB creates a CustomerStore with a custom db connection (for testing)
func NewCustomerStoreWithDB(db *sql.DB) CustomerStore {
	return &customer{db: db}
}

func (c *customer) GetCustomerByID(id int, shopID ...int) (*model.Customer, error) {
	criteria := []interface{}{id}

	q := `
		SELECT id, name, phone, address, created_at, updated_at
		FROM customers
		WHERE id = $1
	`

	if len(shopID) > 0 {
		q += " AND shop_id = $2"
		criteria = append(criteria, shopID[0])
	}

	var customer model.Customer
	err := c.db.QueryRow(q, criteria...).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Address, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &customer, nil
}

func (c *customer) GetCustomerByPhone(phone string, shopID int) (*model.Customer, error) {
	q := `
		SELECT id, name, phone, address, created_at, updated_at
		FROM customers
		WHERE phone = $1 AND shop_id = $2
	`
	var customer model.Customer
	err := c.db.QueryRow(q, phone, shopID).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Address, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &customer, nil
}

func (c *customer) GetCustomersByShopID(shopID int, searchQuery *string) ([]model.Customer, error) {
	q := `
		SELECT id, name, phone, address, created_at, updated_at
		FROM customers
		WHERE shop_id = $1
	`
	args := []interface{}{shopID}

	if searchQuery != nil && strings.TrimSpace(*searchQuery) != "" {
		q += ` AND (name ILIKE $2 OR phone ILIKE $2)`
		args = append(args, "%"+strings.TrimSpace(*searchQuery)+"%")
	}

	rows, err := c.db.Query(q, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	customers := []model.Customer{}
	for rows.Next() {
		var customer model.Customer
		err := rows.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Address, &customer.CreatedAt, &customer.UpdatedAt)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	return customers, nil
}

func (c *customer) CreateCustomer(input CreateCustomerInput) (*model.Customer, error) {
	now := time.Now()
	var id int

	address := ""
	if input.Address != nil {
		address = *input.Address
	}

	q := `
		INSERT INTO customers (name, phone, address, shop_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := c.db.QueryRow(q, input.Name, input.Phone, address, input.ShopID, now).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicatePhone
		}
		return nil, err
	}

	return &model.Customer{
		ID:        id,
		Name:      input.Name,
		Phone:     input.Phone,
		Address:   address,
		CreatedAt: now,
	}, nil
}

func (c *customer) UpdateCustomer(id int, input UpdateCustomerInput) (*model.Customer, error) {
	set := []string{}
	var customer model.Customer

	// build query
	if input.Name != nil {
		newSet := fmt.Sprintf("name = '%s'", *input.Name)
		set = append(set, newSet)
	}
	if input.Phone != nil {
		newSet := fmt.Sprintf("phone = '%s'", *input.Phone)
		set = append(set, newSet)
	}
	if input.Address != nil {
		newSet := fmt.Sprintf("address = '%s'", *input.Address)
		set = append(set, newSet)
	}

	set = append(set, "updated_at = now()")

	q := `
		UPDATE customers
		SET %s
		WHERE id = $1
		RETURNING id, name, phone, address, created_at, updated_at
	`

	q = fmt.Sprintf(q, strings.Join(set, ","))

	err := c.db.QueryRow(q, id).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Address, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicatePhone
		}
		return nil, err
	}

	return &customer, nil
}

func (c *customer) DeleteCustomerByID(id int) error {
	q := `
		DELETE FROM customers
		WHERE id = $1
	`

	_, err := c.db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation
func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
