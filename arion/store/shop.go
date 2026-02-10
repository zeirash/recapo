package store

import (
	"crypto/rand"
	"database/sql"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

const shareTokenCharset = "abcdefghijklmnopqrstuvwxyz0123456789"
const shareTokenLength = 12

type (
	ShopStore interface {
		CreateShop(tx database.Tx, name string) (*model.Shop, error)
		GetShareTokenByID(shopID int) (string, error)
		GetShopByShareToken(shareToken string) (*model.Shop, error)
	}

	shop struct {
		db *sql.DB
	}
)

func NewShopStore() ShopStore {
	return &shop{db: database.GetDB()}
}

// NewShopStoreWithDB creates a ShopStore with a custom db connection (for testing)
func NewShopStoreWithDB(db *sql.DB) ShopStore {
	return &shop{db: db}
}

func generateShareToken() (string, error) {
	b := make([]byte, shareTokenLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for i := range b {
		b[i] = shareTokenCharset[int(b[i])%len(shareTokenCharset)]
	}
	return string(b), nil
}

func (s *shop) CreateShop(tx database.Tx, name string) (*model.Shop, error) {
	shareToken, err := generateShareToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var id int

	q := `
		INSERT INTO shops (name, share_token, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err = tx.QueryRow(q, name, shareToken, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.Shop{
		ID:         id,
		Name:       name,
		ShareToken: shareToken,
		CreatedAt:  now,
	}, nil
}

func (s *shop) GetShareTokenByID(shopID int) (string, error) {
	q := `SELECT share_token FROM shops WHERE id = $1`

	var token string
	err := s.db.QueryRow(q, shopID).Scan(&token)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return token, nil
}

func (s *shop) GetShopByShareToken(shareToken string) (*model.Shop, error) {
	q := `
		SELECT id, name, share_token, created_at, updated_at
		FROM shops
		WHERE share_token = $1
	`

	var sh model.Shop
	err := s.db.QueryRow(q, shareToken).Scan(&sh.ID, &sh.Name, &sh.ShareToken, &sh.CreatedAt, &sh.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &sh, nil
}
