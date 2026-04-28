package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

var ErrDuplicateProductName = errors.New(apierr.ErrProductNameExists)

type (
	ProductStore interface {
		GetProductByID(ctx context.Context, productID int, shopID ...int) (*model.Product, error)
		GetProductsByShopID(ctx context.Context, shopID int, filter model.FilterOptions) ([]model.Product, error)
		CreateProduct(ctx context.Context, name string, description *string, price int, shopID int, originalPrice *int, imageURL *string) (*model.Product, error)
		UpdateProduct(ctx context.Context, productID int, input UpdateProductInput) (*model.Product, error)
		DeleteProductByID(ctx context.Context, productID int) error
		GetProductsListByActiveOrders(ctx context.Context, shopID int) ([]model.PurchaseProduct, error)
	}

	product struct {
		db *sql.DB
	}

	UpdateProductInput struct {
		Name          *string
		Description   *string
		Price         *int
		OriginalPrice *int
		ImageURL      *string
		IsActive      *bool
	}
)

func NewProductStore() ProductStore {
	return &product{db: database.GetDB()}
}

// NewProductStoreWithDB creates a ProductStore with a custom db connection (for testing)
func NewProductStoreWithDB(db *sql.DB) ProductStore {
	return &product{db: db}
}

func (p *product) GetProductByID(ctx context.Context, productID int, shopID ...int) (*model.Product, error) {
	criteria := []interface{}{productID}

	q := `
		SELECT id, shop_id, name, description, price, original_price, image_url, is_active, created_at, updated_at, deleted_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`

	if len(shopID) > 0 {
		q += " AND shop_id = $2"
		criteria = append(criteria, shopID[0])
	}

	var product model.Product
	err := p.db.QueryRowContext(ctx, q, criteria...).Scan(&product.ID, &product.ShopID, &product.Name, &product.Description, &product.Price, &product.OriginalPrice, &product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

func (p *product) GetProductsByShopID(ctx context.Context, shopID int, filter model.FilterOptions) ([]model.Product, error) {
	q := `
		SELECT id, shop_id, name, description, price, original_price, image_url, is_active, created_at, updated_at, deleted_at
		FROM products
		WHERE shop_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{shopID}

	argNum := 2
	if filter.SearchQuery != nil && strings.TrimSpace(*filter.SearchQuery) != "" {
		q += fmt.Sprintf(` AND name ILIKE $%d`, argNum)
		args = append(args, "%"+strings.TrimSpace(*filter.SearchQuery)+"%")
		argNum++
	}
	if filter.IsActive != nil {
		q += fmt.Sprintf(` AND is_active = $%d`, argNum)
		args = append(args, *filter.IsActive)
		argNum++
	}
	if filter.Sort != nil {
		sort := strings.Split(*filter.Sort, ",")
		if len(sort) == 2 {
			col, dir := sort[0], strings.ToUpper(sort[1])
			allowedCols := map[string]bool{"id": true, "name": true, "price": true, "created_at": true, "updated_at": true}
			if dir != "ASC" && dir != "DESC" {
				dir = "ASC"
			}
			if allowedCols[col] {
				nullsOrder := "NULLS LAST"
				if dir == "ASC" {
					nullsOrder = "NULLS FIRST"
				}
				textCols := map[string]bool{"name": true}
				if textCols[col] {
					q += fmt.Sprintf(" ORDER BY LOWER(%s) %s %s, id ASC", col, dir, nullsOrder)
				} else {
					q += fmt.Sprintf(" ORDER BY %s %s %s, id ASC", col, dir, nullsOrder)
				}
			}
		}
	} else {
		q += " ORDER BY id ASC"
	}

	rows, err := p.db.QueryContext(ctx, q, args...)
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
		err := rows.Scan(&product.ID, &product.ShopID, &product.Name, &product.Description, &product.Price, &product.OriginalPrice, &product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (p *product) CreateProduct(ctx context.Context, name string, description *string, price int, shopID int, originalPrice *int, imageURL *string) (*model.Product, error) {
	now := time.Now()
	origPrice := price
	if originalPrice != nil {
		origPrice = *originalPrice
	}

	var id int
	var desc string
	var imgURL string
	var isActive bool

	q := `
		INSERT INTO products (name, description, price, original_price, shop_id, image_url, created_at)
		VALUES ($1, COALESCE($2, ''), $3, $4, $5, COALESCE($6, ''), $7)
		RETURNING id, description, image_url, is_active
	`

	err := p.db.QueryRowContext(ctx, q, name, description, price, origPrice, shopID, imageURL, now).Scan(&id, &desc, &imgURL, &isActive)
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
		ImageURL:      imgURL,
		IsActive:      isActive,
		ShopID:        shopID,
		CreatedAt:     now,
	}, nil
}

func (p *product) UpdateProduct(ctx context.Context, productID int, input UpdateProductInput) (*model.Product, error) {
	set := []string{}
	args := []interface{}{productID}
	argNum := 2
	var product model.Product

	// build query
	if input.Name != nil {
		set = append(set, fmt.Sprintf("name = $%d", argNum))
		args = append(args, *input.Name)
		argNum++
	}
	if input.Description != nil {
		set = append(set, fmt.Sprintf("description = $%d", argNum))
		args = append(args, *input.Description)
		argNum++
	}
	if input.Price != nil {
		set = append(set, fmt.Sprintf("price = $%d", argNum))
		args = append(args, *input.Price)
		argNum++
	}
	if input.OriginalPrice != nil {
		set = append(set, fmt.Sprintf("original_price = $%d", argNum))
		args = append(args, *input.OriginalPrice)
		argNum++
	}
	if input.ImageURL != nil {
		set = append(set, fmt.Sprintf("image_url = $%d", argNum))
		args = append(args, *input.ImageURL)
		argNum++
	}
	if input.IsActive != nil {
		set = append(set, fmt.Sprintf("is_active = $%d", argNum))
		args = append(args, *input.IsActive)
		argNum++
	}

	set = append(set, "updated_at = now()")

	q := fmt.Sprintf(`
		UPDATE products
		SET %s
		WHERE id = $1
		RETURNING id, shop_id, name, description, price, original_price, image_url, is_active, created_at, updated_at
	`, strings.Join(set, ","))

	err := p.db.QueryRowContext(ctx, q, args...).Scan(&product.ID, &product.ShopID, &product.Name, &product.Description, &product.Price, &product.OriginalPrice, &product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if isProductUniqueViolation(err) {
			return nil, ErrDuplicateProductName
		}
		return nil, err
	}

	return &product, nil
}

func (p *product) DeleteProductByID(ctx context.Context, productID int) error {
	q := `
		UPDATE products
		SET deleted_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := p.db.ExecContext(ctx, q, productID)
	if err != nil {
		return err
	}

	return nil
}

func (p *product) GetProductsListByActiveOrders(ctx context.Context, shopID int) ([]model.PurchaseProduct, error) {
	q := `
		SELECT p.name, p.price, COALESCE(SUM(oi.qty), 0)::int AS qty
		FROM products p
		INNER JOIN order_items oi ON p.id = oi.product_id
		INNER JOIN orders o ON oi.order_id = o.id
		WHERE o.shop_id = $1 AND o.status IN ($2, $3)
		GROUP BY p.id, p.name, p.price
	`
	rows, err := p.db.QueryContext(ctx, q, shopID, constant.OrderStatusCreated, constant.OrderStatusInProgress)
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
