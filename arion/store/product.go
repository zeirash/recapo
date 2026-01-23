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

var ErrDuplicateProductName = errors.New("product with this name already exists")

type (
	ProductStore interface {
		GetProductByID(productID int, shopID ...int) (*model.Product, error)
		GetProductsByShopID(shopID int) ([]model.Product, error)
		CreateProduct(name string, description *string, price int, shopID int) (*model.Product, error)
		UpdateProduct(productID int, input UpdateProductInput) (*model.Product, error)
		DeleteProductByID(productID int) error
	}

	product struct{}

	UpdateProductInput struct {
		Name        *string
		Description *string
		Price       *int
	}
)

func NewProductStore() ProductStore {
	return &product{}
}

func (p *product) GetProductByID(productID int, shopID ...int) (*model.Product, error) {
	db := database.GetDB()
	defer db.Close()

	criteria := []interface{}{productID}

	q := `
		SELECT id, name, description, price, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	if len(shopID) > 0 {
		q += " AND shop_id = $2"
		criteria = append(criteria, shopID[0])
	}

	var product model.Product
	err := db.QueryRow(q, criteria...).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func (p *product) GetProductsByShopID(shopID int) ([]model.Product, error) {
	db := database.GetDB()
	defer db.Close()

	q := `
		SELECT id, name, description, price, created_at, updated_at
		FROM products
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

	products := []model.Product{}
	for rows.Next() {
		var product model.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (p *product) CreateProduct(name string, description *string, price int, shopID int) (*model.Product, error) {
	db := database.GetDB()
	defer db.Close()

	now := time.Now()
	var id int

	q := `
		INSERT INTO products (name, description, price, shop_id, created_at)
		VALUES ($1, COALESCE($2, ''), $3, $4, $5)
		RETURNING id, description
	`

	var desc string
	err := db.QueryRow(q, name, description, price, shopID, now).Scan(&id, &desc)
	if err != nil {
		if isProductUniqueViolation(err) {
			return nil, ErrDuplicateProductName
		}
		return nil, err
	}

	return &model.Product{
		ID:          id,
		Name:        name,
		Description: desc,
		Price:       price,
		ShopID:      shopID,
		CreatedAt:   now,
	}, nil
}

func (p *product) UpdateProduct(productID int, input UpdateProductInput) (*model.Product, error) {
	db := database.GetDB()
	defer db.Close()

	set := []string{}
	var product model.Product

	// build query
	if input.Name != nil {
		newSet := fmt.Sprintf("name = '%s'", *input.Name)
		set = append(set, newSet)
	}
	if input.Description != nil {
		newSet := fmt.Sprintf("description = '%s'", *input.Description)
		set = append(set, newSet)
	}
	if input.Price != nil {
		newSet := fmt.Sprintf("price = %d", *input.Price)
		set = append(set, newSet)
	}

	set = append(set, "updated_at = now()")

	q := `
		UPDATE products
		SET %s
		WHERE id = $1
		RETURNING id, name, description, price, created_at, updated_at
	`

	q = fmt.Sprintf(q, strings.Join(set, ","))

	err := db.QueryRow(q, productID).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if isProductUniqueViolation(err) {
			return nil, ErrDuplicateProductName
		}
		return nil, err
	}

	return &product, nil
}

func (p *product) DeleteProductByID(productID int) error {
	db := database.GetDB()
	defer db.Close()

	q := `
		DELETE FROM products
		WHERE id = $1
	`

	_, err := db.Exec(q, productID)
	if err != nil {
		return err
	}

	return nil
}

// isProductUniqueViolation checks if the error is a PostgreSQL unique constraint violation
func isProductUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
