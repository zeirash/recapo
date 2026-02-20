package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

var ErrDuplicateProductName = errors.New("product with this name already exists")

type (
	ProductStore interface {
		GetProductByID(productID int, shopID ...int) (*model.Product, error)
		GetProductsByShopID(shopID int, searchQuery *string) ([]model.Product, error)
		CreateProduct(name string, description *string, price int, shopID int, originalPrice *int) (*model.Product, error)
		UpdateProduct(productID int, input UpdateProductInput) (*model.Product, error)
		DeleteProductByID(productID int) error
		GetProductsListByActiveOrders(shopID int) ([]model.PurchaseProduct, error)
	}

	product struct {
		db *sql.DB
	}

	UpdateProductInput struct {
		Name          *string
		Description   *string
		Price         *int
		OriginalPrice *int
	}
)

func NewProductStore() ProductStore {
	return &product{db: database.GetDB()}
}

// NewProductStoreWithDB creates a ProductStore with a custom db connection (for testing)
func NewProductStoreWithDB(db *sql.DB) ProductStore {
	return &product{db: db}
}

func (p *product) GetProductByID(productID int, shopID ...int) (*model.Product, error) {
	criteria := []interface{}{productID}

	q := `
		SELECT id, shop_id, name, description, price, original_price, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	if len(shopID) > 0 {
		q += " AND shop_id = $2"
		criteria = append(criteria, shopID[0])
	}

	var product model.Product
	err := p.db.QueryRow(q, criteria...).Scan(&product.ID, &product.ShopID, &product.Name, &product.Description, &product.Price, &product.OriginalPrice, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func (p *product) GetProductsByShopID(shopID int, searchQuery *string) ([]model.Product, error) {
	q := `
		SELECT id, shop_id, name, description, price, original_price, created_at, updated_at
		FROM products
		WHERE shop_id = $1
	`
	args := []interface{}{shopID}

	if searchQuery != nil && strings.TrimSpace(*searchQuery) != "" {
		q += ` AND name ILIKE $2`
		args = append(args, "%"+strings.TrimSpace(*searchQuery)+"%")
	}

	rows, err := p.db.Query(q, args...)
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
		err := rows.Scan(&product.ID, &product.ShopID, &product.Name, &product.Description, &product.Price, &product.OriginalPrice, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (p *product) CreateProduct(name string, description *string, price int, shopID int, originalPrice *int) (*model.Product, error) {
	now := time.Now()
	origPrice := price
	if originalPrice != nil {
		origPrice = *originalPrice
	}

	var id int
	var desc string

	q := `
		INSERT INTO products (name, description, price, original_price, shop_id, created_at)
		VALUES ($1, COALESCE($2, ''), $3, $4, $5, $6)
		RETURNING id, description
	`

	err := p.db.QueryRow(q, name, description, price, origPrice, shopID, now).Scan(&id, &desc)
	if err != nil {
		if isProductUniqueViolation(err) {
			return nil, ErrDuplicateProductName
		}
		return nil, err
	}

	return &model.Product{
		ID:            id,
		Name:          name,
		Description:   desc,
		Price:         price,
		OriginalPrice: origPrice,
		ShopID:        shopID,
		CreatedAt:     now,
	}, nil
}

func (p *product) UpdateProduct(productID int, input UpdateProductInput) (*model.Product, error) {
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
	if input.OriginalPrice != nil {
		newSet := fmt.Sprintf("original_price = %d", *input.OriginalPrice)
		set = append(set, newSet)
	}

	set = append(set, "updated_at = now()")

	q := `
		UPDATE products
		SET %s
		WHERE id = $1
		RETURNING id, shop_id, name, description, price, original_price, created_at, updated_at
	`

	q = fmt.Sprintf(q, strings.Join(set, ","))

	err := p.db.QueryRow(q, productID).Scan(&product.ID, &product.ShopID, &product.Name, &product.Description, &product.Price, &product.OriginalPrice, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if isProductUniqueViolation(err) {
			return nil, ErrDuplicateProductName
		}
		return nil, err
	}

	return &product, nil
}

func (p *product) DeleteProductByID(productID int) error {
	q := `
		DELETE FROM products
		WHERE id = $1
	`

	_, err := p.db.Exec(q, productID)
	if err != nil {
		return err
	}

	return nil
}

func (p *product) GetProductsListByActiveOrders(shopID int) ([]model.PurchaseProduct, error) {
	q := `
		SELECT p.name, p.price, COALESCE(SUM(oi.qty), 0)::int AS qty
		FROM products p
		INNER JOIN order_items oi ON p.id = oi.product_id
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.shop_id = $1 AND o.status IN ($2, $3)
		GROUP BY p.id, p.name, p.price
	`
	rows, err := p.db.Query(q, shopID, constant.OrderStatusCreated, constant.OrderStatusInProgress)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var list []model.PurchaseProduct
	for rows.Next() {
		var pp model.PurchaseProduct
		if err := rows.Scan(&pp.ProductName, &pp.Price, &pp.Qty); err != nil {
			return nil, err
		}
		list = append(list, pp)
	}

	return list, nil
}

// isProductUniqueViolation checks if the error is a PostgreSQL unique constraint violation
func isProductUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
