package store

import (
	"database/sql"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	ProductStore interface {
		GetProductByID(id int) (*model.Product, error)
		GetProductsByShopID(shopID int) ([]model.Product, error)
		CreateProduct(name string, shopID int) (*model.Product, error)
		UpdateProduct(id int, name string) (*model.Product, error)
		DeleteProductByID(id int) error
	}

	product struct{}
)

func NewProductStore() ProductStore {
	return &product{}
}

func (p *product) GetProductByID(id int) (*model.Product, error) {
	db := database.GetDB()
	defer db.Close()

	q := `
		SELECT id, name, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product model.Product
	err := db.QueryRow(q, id).Scan(&product.ID, &product.Name, &product.CreatedAt, &product.UpdatedAt)
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
		SELECT id, name, created_at, updated_at
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
		err := rows.Scan(&product.ID, &product.Name, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (p *product) CreateProduct(name string, shopID int) (*model.Product, error) {
	db := database.GetDB()
	defer db.Close()

	now := time.Now()
	var id int

	q := `
		INSERT INTO products (name, shop_id, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := db.QueryRow(q, name, shopID, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.Product{
		ID:        id,
		Name:      name,
		ShopID:    shopID,
		CreatedAt: now,
	}, nil
}

func (p *product) UpdateProduct(id int, name string) (*model.Product, error) {
	db := database.GetDB()
	defer db.Close()

	var product model.Product

	q := `
		UPDATE products
		SET name = $1, updated_at = now()
		WHERE id = $2
		RETURNING id, name, created_at, updated_at
	`

	err := db.QueryRow(q, name, id).Scan(&product.ID, &product.Name, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (p *product) DeleteProductByID(id int) error {
	db := database.GetDB()
	defer db.Close()

	q := `
		DELETE FROM products
		WHERE id = $1
	`

	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}
