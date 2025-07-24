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
		SystemMode bool   `json:"system_mode"`
		jwt.RegisteredClaims
	}

	JwtCustomRefreshClaims struct {
		UserID int `json:"user_id"`
		jwt.RegisteredClaims
	}

	TokenData struct {
		Name       string `json:"name"`
		UserID     int    `json:"user_id"`
		SystemMode bool   `json:"system_mode"`
	}

	/********************* User ************************/
	User struct {
		ID         int          `db:"id"`
		ShopID     int          `db:"shop_id"`
		Name       string       `db:"name"`
		Email      string       `db:"email"`
		Password   string       `db:"password"`
		Role       string       `db:"role"`
		CreatedAt  time.Time    `db:"created_at"`
		UpdatedAt  sql.NullTime `db:"updated_at"`
	}

	/******************** Shop *********************/
	Shop struct {
		ID         int          `db:"id"`
		Name       string       `db:"name"`
		CreatedAt  time.Time    `db:"created_at"`
		UpdatedAt  sql.NullTime `db:"updated_at"`
	}
)
