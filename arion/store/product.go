package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	ProductStore interface {
		GetProductByID(productID int) (*model.Product, error)
		GetProductsByShopID(shopID int) ([]model.Product, error)
		CreateProduct(name string, shopID int) (*model.Product, error)
		UpdateProduct(productID int, name string) (*model.Product, error)
		DeleteProductByID(productID int) error
		CreatePrice(productID int, price int) (*model.Price, error)
		UpdatePrice(productID, priceID, price int) (*model.Price, error)
		GetPricesByProductID(productID int) ([]model.Price, error)
		DeletePrice(productID, priceID int) error
		DeletePricesByProductID(productID int) error
	}

	product struct{}
)

func NewProductStore() ProductStore {
	return &product{}
}

func (p *product) GetProductByID(productID int) (*model.Product, error) {
	db := database.GetDB()
	defer db.Close()

	q := `
		SELECT id, name, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product model.Product
	err := db.QueryRow(q, productID).Scan(&product.ID, &product.Name, &product.CreatedAt, &product.UpdatedAt)
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

func (p *product) UpdateProduct(productID int, name string) (*model.Product, error) {
	db := database.GetDB()
	defer db.Close()

	var product model.Product

	q := `
		UPDATE products
		SET name = $1, updated_at = now()
		WHERE id = $2
		RETURNING id, name, created_at, updated_at
	`

	err := db.QueryRow(q, name, productID).Scan(&product.ID, &product.Name, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
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

func (p *product) CreatePrice(productID int, price int) (*model.Price, error) {
	db := database.GetDB()
	defer db.Close()

	now := time.Now()
	var id int

	q := `
		INSERT INTO prices (product_id, price, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := db.QueryRow(q, productID, price, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.Price{
		ID:        id,
		ProductID: productID,
		Price:     price,
		CreatedAt: now,
	}, nil
}

func (p *product) UpdatePrice(productID, priceID, price int) (*model.Price, error) {
	db := database.GetDB()
	defer db.Close()

	var priceData model.Price

	q := `
		UPDATE prices
		SET price = $1, updated_at = now()
		WHERE id = $2 AND product_id = $3
		RETURNING id, product_id, price, created_at, updated_at
	`

	err := db.QueryRow(q, price, priceID, productID).Scan(&priceData.ID, &priceData.ProductID, &priceData.Price, &priceData.CreatedAt, &priceData.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("price not found")
		}
		return nil, err
	}

	return &priceData, nil
}

func (p *product) GetPricesByProductID(productID int) ([]model.Price, error) {
	db := database.GetDB()
	defer db.Close()

	q := `
		SELECT id, product_id, price, created_at, updated_at
		FROM prices
		WHERE product_id = $1
	`

	rows, err := db.Query(q, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	prices := []model.Price{}
	for rows.Next() {
		var price model.Price
		err := rows.Scan(&price.ID, &price.ProductID, &price.Price, &price.CreatedAt, &price.UpdatedAt)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}

	return prices, nil
}

func (p *product) DeletePrice(productID, priceID int) error {
	db := database.GetDB()
	defer db.Close()

	q := `
		DELETE FROM prices
		WHERE id = $1 AND product_id = $2
	`

	_, err := db.Exec(q, priceID, productID)
	if err != nil {
		return err
	}

	return nil
}

func (p *product) DeletePricesByProductID(productID int) error {
	db := database.GetDB()
	defer db.Close()

	q := `
		DELETE FROM prices
		WHERE product_id = $1
	`

	_, err := db.Exec(q, productID)
	if err != nil {
		return err
	}

	return nil
}
