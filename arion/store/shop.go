package store

import (
	"database/sql"
	"time"

	"github.com/zeirash/recapo/arion/model"
)

type (
	ShopStore interface {
		CreateShop(tx *sql.Tx, name string) (*model.Shop, error)
	}

	shop struct{}
)

func NewShopStore() ShopStore {
	return &shop{}
}

func (s *shop) CreateShop(tx *sql.Tx, name string) (*model.Shop, error) {
	now := time.Now()
	var id int

	q := `
		INSERT INTO shops (name, created_at)
		VALUES ($1, $2)
		RETURNING id
	`

	err := tx.QueryRow(q, name, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.Shop{
		ID:        id,
		Name:      name,
		CreatedAt: now,
	}, nil
}
