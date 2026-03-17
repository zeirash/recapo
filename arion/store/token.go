package store

import (
	"context"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/model"
)

type (
	TokenStore interface {
		CreateAccessToken(ctx context.Context, user *model.User, secret string, expiry int) (string, error)
		CreateRefreshToken(ctx context.Context, user *model.User, secret string, expiry int) (string, error)
		IsAuthorized(ctx context.Context, requestToken, secret string) (bool, error)
		ExtractDataFromToken(ctx context.Context, requestToken, secret string) (model.TokenData, error)
	}

	token struct{}
)

func NewTokenStore() TokenStore {
	return &token{}
}

func (t *token) CreateAccessToken(ctx context.Context, user *model.User, secret string, expiry int) (string, error) {
	exp := &jwt.NumericDate{
		Time: time.Now().Add(time.Hour * time.Duration(expiry)),
	}

	claim := &model.JwtCustomClaims{
		Name:         user.Name,
		UserID:       user.ID,
		ShopID:       user.ShopID,
		SystemMode:   user.Role == constant.RoleSystem,
		SessionToken: user.SessionToken.String,
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

func (t *token) CreateRefreshToken(ctx context.Context, user *model.User, secret string, expiry int) (string, error) {
	exp := &jwt.NumericDate{
		Time: time.Now().Add(time.Hour * time.Duration(expiry)),
	}
	claimsRefresh := &model.JwtCustomRefreshClaims{
		UserID:       user.ID,
		ShopID:       user.ShopID,
		SystemMode:   user.Role == constant.RoleSystem,
		SessionToken: user.SessionToken.String,
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

func (t *token) IsAuthorized(ctx context.Context, requestToken, secret string) (bool, error) {
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

func (t *token) ExtractDataFromToken(ctx context.Context, requestToken, secret string) (model.TokenData, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return model.TokenData{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok && !token.Valid {
		return model.TokenData{}, fmt.Errorf("invalid token")
	}

	tokenData := model.TokenData{
		UserID:     int(claims["user_id"].(float64)),
		ShopID:     int(claims["shop_id"].(float64)),
		SystemMode: claims["system_mode"].(bool),
	}

	// Handle optional fields (not present in all token types)
	if name, ok := claims["name"].(string); ok {
		tokenData.Name = name
	}
	if sessionToken, ok := claims["session_token"].(string); ok {
		tokenData.SessionToken = sessionToken
	}

	return tokenData, nil
}
