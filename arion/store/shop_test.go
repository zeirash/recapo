package store

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeirash/recapo/arion/model"
)

func Test_shop_GetShopByShareToken(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		shareToken string
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Shop
		wantErr    bool
	}{
		{
			name:       "successfully get shop by share token",
			shareToken: "abc123xyz789",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "share_token", "created_at", "updated_at"}).
					AddRow(1, "My Shop", "abc123xyz789", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, name, share_token, created_at, updated_at\s+FROM shops\s+WHERE share_token = \$1`).
					WithArgs("abc123xyz789").
					WillReturnRows(rows)
			},
			wantResult: &model.Shop{
				ID:         1,
				Name:       "My Shop",
				ShareToken: "abc123xyz789",
				CreatedAt:  fixedTime,
			},
			wantErr: false,
		},
		{
			name:       "returns nil when shop not found",
			shareToken: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, share_token, created_at, updated_at\s+FROM shops\s+WHERE share_token = \$1`).
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			s := &shop{db: db}
			got, gotErr := s.GetShopByShareToken(tt.shareToken)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetShopByShareToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetShopByShareToken() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetShopByShareToken() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_shop_CreateShop(t *testing.T) {
	tests := []struct {
		name       string
		shopName   string
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Shop
		wantErr    bool
	}{
		{
			name:     "successfully create shop",
			shopName: "My Shop",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`INSERT INTO shops \(name, share_token, created_at\)\s+VALUES \(\$1, \$2, \$3\)\s+RETURNING id`).
					WithArgs("My Shop", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.Shop{
				ID:         1,
				Name:       "My Shop",
				ShareToken: "abc123def456", // Will be overwritten in assertion
			},
			wantErr: false,
		},
		{
			name:     "create shop returns error on database failure",
			shopName: "My Shop",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO shops \(name, share_token, created_at\)\s+VALUES \(\$1, \$2, \$3\)\s+RETURNING id`).
					WithArgs("My Shop", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			var s shop
			got, gotErr := s.CreateShop(tx, tt.shopName)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateShop() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateShop() succeeded unexpectedly")
			}

			// Set expected CreatedAt and ShareToken to match for DeepEqual comparison
			tt.wantResult.CreatedAt = got.CreatedAt
			tt.wantResult.ShareToken = got.ShareToken

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateShop() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
