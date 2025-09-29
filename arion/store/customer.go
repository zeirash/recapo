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
	CustomerStore interface {
		GetCustomerByID(id int, shopID ...int) (*model.Customer, error)
		GetCustomersByShopID(shopID int) ([]model.Customer, error)
		CreateCustomer(name, phone, address string, shopID int) (*model.Customer, error)
		UpdateCustomer(id int, input UpdateCustomerInput) (*model.Customer, error)
		DeleteCustomerByID(id int) error
	}

	customer struct{}

	UpdateCustomerInput struct {
		Name    *string
		Phone   *string
		Address *string
	}
)

func NewCustomerStore() CustomerStore {
	return &customer{}
}

func (c *customer) GetCustomerByID(id int, shopID ...int) (*model.Customer, error) {
	db := database.GetDB()
	defer db.Close()

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
	err := db.QueryRow(q, criteria...).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Address, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &customer, nil
}

func (c *customer) GetCustomersByShopID(shopID int) ([]model.Customer, error) {
	db := database.GetDB()
	defer db.Close()

	q := `
		SELECT id, name, phone, address, created_at, updated_at
		FROM customers
		WHERE shop_id = $1
	`

	rows, err := db.Query(q, shopID)
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

func (c *customer) CreateCustomer(name, phone, address string, shopID int) (*model.Customer, error) {
	db := database.GetDB()
	defer db.Close()

	now := time.Now()
	var id int

	q := `
		INSERT INTO customers (name, phone, address, shop_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := db.QueryRow(q, name, phone, address, shopID, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.Customer{
		ID:        id,
		Name:      name,
		Phone:     phone,
		Address:   address,
		CreatedAt: now,
	}, nil
}

func (c *customer) UpdateCustomer(id int, input UpdateCustomerInput) (*model.Customer, error) {
	db := database.GetDB()
	defer db.Close()

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

	err := db.QueryRow(q, id).Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Address, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

func (c *customer) DeleteCustomerByID(id int) error {
	db := database.GetDB()
	defer db.Close()

	q := `
		DELETE FROM customers
		WHERE id = $1
	`

	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}
