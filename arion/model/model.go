package model

import (
	"database/sql"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

/********************* Token ************************/
type (
	JwtCustomClaims struct {
		Name       string `json:"name"`
		UserID     int    `json:"user_id"`
		ShopID     int    `json:"shop_id"`
		SystemMode bool   `json:"system_mode"`
		jwt.RegisteredClaims
	}

	JwtCustomRefreshClaims struct {
		UserID     int  `json:"user_id"`
		ShopID     int  `json:"shop_id"`
		SystemMode bool `json:"system_mode"`
		jwt.RegisteredClaims
	}

	TokenData struct {
		Name       string `json:"name"`
		UserID     int    `json:"user_id"`
		ShopID     int    `json:"shop_id"`
		SystemMode bool   `json:"system_mode"`
	}

	/********************* User ************************/
	User struct {
		ID        int          `db:"id"`
		ShopID    int          `db:"shop_id"`
		Name      string       `db:"name"`
		Email     string       `db:"email"`
		Password  string       `db:"password"`
		Role      string       `db:"role"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt sql.NullTime `db:"updated_at"`
	}

	/********************* Shop ************************/
	Shop struct {
		ID         int          `db:"id"`
		Name       string       `db:"name"`
		ShareToken string       `db:"share_token"`
		CreatedAt  time.Time    `db:"created_at"`
		UpdatedAt  sql.NullTime `db:"updated_at"`
	}

	/******************* Customer *********************/
	Customer struct {
		ID        int          `db:"id"`
		ShopID    int          `db:"shop_id"`
		Name      string       `db:"name"`
		Phone     string       `db:"phone"`
		Address   string       `db:"address"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt sql.NullTime `db:"updated_at"`
	}

	/******************* Product *********************/
	Product struct {
		ID            int          `db:"id"`
		ShopID        int          `db:"shop_id"`
		Name          string       `db:"name"`
		Description   string       `db:"description"`
		Price         int          `db:"price"`
		OriginalPrice int          `db:"original_price"`
		CreatedAt     time.Time    `db:"created_at"`
		UpdatedAt     sql.NullTime `db:"updated_at"`
	}

	/******************** Order **********************/
	// OrderFilterOptions holds optional filters for listing orders.
	// Used by handler and service; store consumes it. Add fields as needed (e.g. Status, CustomerID).
	OrderFilterOptions struct {
		SearchQuery *string
		DateFrom    *time.Time
		DateTo      *time.Time
	}

	Order struct {
		ID           int          `db:"id"`
		ShopID       int          `db:"shop_id"`
		CustomerName string       `db:"customer_name"`
		TotalPrice   int          `db:"total_price"`
		Status       string       `db:"status"`
		Notes        string       `db:"notes"`
		CreatedAt    time.Time    `db:"created_at"`
		UpdatedAt    sql.NullTime `db:"updated_at"`
	}

	OrderItem struct {
		ID          int          `db:"id"`
		OrderID     int          `db:"order_id"`
		ProductName string       `db:"product_name"`
		Price       int          `db:"price"`
		Qty         int          `db:"qty"`
		CreatedAt   time.Time    `db:"created_at"`
		UpdatedAt   sql.NullTime `db:"updated_at"`
	}

	OrderTemp struct {
		ID            int          `db:"id"`
		ShopID        int          `db:"shop_id"`
		CustomerName  string       `db:"customer_name"`
		CustomerPhone string       `db:"customer_phone"`
		TotalPrice    int          `db:"total_price"`
		Status        string       `db:"status"`
		CreatedAt     time.Time    `db:"created_at"`
		UpdatedAt     sql.NullTime `db:"updated_at"`
	}

	OrderTempItem struct {
		ID          int          `db:"id"`
		OrderTempID int          `db:"order_temp_id"`
		ProductName string       `db:"product_name"`
		Price       int          `db:"price"`
		Qty         int          `db:"qty"`
		CreatedAt   time.Time    `db:"created_at"`
	}
)
