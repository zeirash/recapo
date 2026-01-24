package store

import (
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeirash/recapo/arion/model"
)

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
				mock.ExpectQuery(`INSERT INTO shops \(name, created_at\)\s+VALUES \(\$1, \$2\)\s+RETURNING id`).
					WithArgs("My Shop", sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.Shop{
				ID:   1,
				Name: "My Shop",
			},
			wantErr: false,
		},
		{
			name:     "create shop returns error on database failure",
			shopName: "My Shop",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO shops \(name, created_at\)\s+VALUES \(\$1, \$2\)\s+RETURNING id`).
					WithArgs("My Shop", sqlmock.AnyArg()).
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

			// Set expected CreatedAt to match for DeepEqual comparison
			tt.wantResult.CreatedAt = got.CreatedAt

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateShop() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
