package store

import (
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	ShopStore interface {
		CreateShop(name string) (*model.Shop, error)
	}

	shop struct{}
)

func NewShopStore() ShopStore {
	return &shop{}
}

func (s *shop) CreateShop(name string) (*model.Shop, error) {
	db := database.GetDB()
	defer db.Close()

	now := time.Now()
	var id int

	q := `
		INSERT INTO shops (name, created_at)
		VALUES ($1, $2)
		RETURNING id
	`

	err := db.QueryRow(q, name, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.Shop{
		ID:        id,
		Name:      name,
		CreatedAt: now,
	}, nil
}
