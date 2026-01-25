package store

import (
	"database/sql"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	ShopStore interface {
		CreateShop(tx database.Tx, name string) (*model.Shop, error)
	}

	shop struct {
		db *sql.DB
	}
)

func NewShopStore() ShopStore {
	return &shop{}
}

// NewShopStoreWithDB creates a ShopStore with a custom db connection (for testing)
func NewShopStoreWithDB(db *sql.DB) ShopStore {
	return &shop{db: db}
}

func (s *shop) CreateShop(tx database.Tx, name string) (*model.Shop, error) {
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
