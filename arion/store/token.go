package store

import (
	"fmt"
	"time"

	"github.com/zeirash/recapo/arion/model"
	jwt "github.com/golang-jwt/jwt/v4"
)

type (
	TokenStore interface {
		CreateAccessToken(user *model.User, secret string, expiry int) (string, error)
		CreateRefreshToken(user *model.User, secret string, expiry int) (string, error)
		IsAuthorized(requestToken, secret string) (bool, error)
		ExtractIDFromToken(requestToken, secret string) (float64, error)
	}

	token struct {}
)

func NewTokenStore() TokenStore {
	return &token{}
}

func (t *token) CreateAccessToken(user *model.User, secret string, expiry int) (string, error) {
	exp := &jwt.NumericDate{
		time.Now().Add(time.Hour * time.Duration(expiry)),
	}

	claim := &model.JwtCustomClaims{
		Name: user.Name,
		ID:   user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: exp,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (t *token) CreateRefreshToken(user *model.User, secret string, expiry int) (string, error) {
	exp := &jwt.NumericDate{
		time.Now().Add(time.Hour * time.Duration(expiry)),
	}
	claimsRefresh := &model.JwtCustomRefreshClaims{
		ID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: exp,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsRefresh)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func (t *token) IsAuthorized(requestToken, secret string) (bool, error) {
	_, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (t *token) ExtractIDFromToken(requestToken, secret string) (float64, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok && !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	return claims["id"].(float64), nil
}
