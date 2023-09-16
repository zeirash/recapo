package model

import (
	"database/sql"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

/********************* Token ************************/
type (
	JwtCustomClaims struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
		jwt.RegisteredClaims
	}

	JwtCustomRefreshClaims struct {
		ID int `json:"id"`
		jwt.RegisteredClaims
	}

	/********************* User ************************/
	User struct {
		ID        int          `db:"id"`
		Name      string       `db:"name"`
		Email     string       `db:"email"`
		Password  string       `db:"password"`
		CreatedAt time.Time	   `db:"created_at"`
		UpdatedAt sql.NullTime `db:"updated_at"`
	}
	
)
