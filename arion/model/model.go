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
		UserID int `json:"user_id"`
		ShopID int `json:"shop_id"`
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
		ID        int          `db:"id"`
		Name      string       `db:"name"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt sql.NullTime `db:"updated_at"`
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
		ID        int          `db:"id"`
		ShopID    int          `db:"shop_id"`
		Name      string       `db:"name"`
		Price     int          `db:"price"`
		CreatedAt time.Time    `db:"created_at"`
		UpdatedAt sql.NullTime `db:"updated_at"`
	}

	/******************** Order **********************/
	Order struct {
		ID           int          `db:"id"`
		ShopID       int          `db:"shop_id"`
		CustomerName string       `db:"customer_name"`
		TotalPrice   int          `db:"total_price"`
		Status       string       `db:"status"`
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
)
