package service

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/response"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
)

func Test_shopService_GetPublicProducts(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		shareToken string
		mockSetup func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockProductStore)
		want      []response.ProductData
		wantErr   bool
	}{
		{
			name:       "success - returns products",
			shareToken: "abc123xyz",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockProductStore) {
				shopMock := mock_store.NewMockShopStore(ctrl)
				productMock := mock_store.NewMockProductStore(ctrl)

				shopMock.EXPECT().
					GetShopByShareToken("abc123xyz").
					Return(&model.Shop{
						ID:         5,
						Name:       "Test Shop",
						ShareToken: "abc123xyz",
						CreatedAt:  fixedTime,
					}, nil)

				productMock.EXPECT().
					GetProductsByShopID(5, (*string)(nil)).
					Return([]model.Product{
						{
							ID:            1,
							Name:          "Product A",
							Description:   "Desc A",
							Price:         1000,
							OriginalPrice: 1200,
							CreatedAt:     fixedTime,
							UpdatedAt:     sql.NullTime{Time: fixedTime, Valid: true},
						},
					}, nil)

				return shopMock, productMock
			},
			want: []response.ProductData{
				{
					ID:            1,
					Name:          "Product A",
					Description:   "Desc A",
					Price:         1000,
					OriginalPrice: 1200,
					CreatedAt:     fixedTime,
					UpdatedAt:     &fixedTime,
				},
			},
			wantErr: false,
		},
		{
			name:       "success - empty products",
			shareToken: "empty123",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockProductStore) {
				shopMock := mock_store.NewMockShopStore(ctrl)
				productMock := mock_store.NewMockProductStore(ctrl)

				shopMock.EXPECT().
					GetShopByShareToken("empty123").
					Return(&model.Shop{ID: 1, Name: "Shop", ShareToken: "empty123", CreatedAt: fixedTime}, nil)

				productMock.EXPECT().
					GetProductsByShopID(1, (*string)(nil)).
					Return([]model.Product{}, nil)

				return shopMock, productMock
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:       "shop not found",
			shareToken: "invalid",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockProductStore) {
				shopMock := mock_store.NewMockShopStore(ctrl)

				shopMock.EXPECT().
					GetShopByShareToken("invalid").
					Return(nil, nil)

				return shopMock, nil
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:       "GetShopByShareToken returns error",
			shareToken: "token",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockProductStore) {
				shopMock := mock_store.NewMockShopStore(ctrl)

				shopMock.EXPECT().
					GetShopByShareToken("token").
					Return(nil, errors.New("db error"))

				return shopMock, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:       "GetProductsByShopID returns error",
			shareToken: "token",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockProductStore) {
				shopMock := mock_store.NewMockShopStore(ctrl)
				productMock := mock_store.NewMockProductStore(ctrl)

				shopMock.EXPECT().
					GetShopByShareToken("token").
					Return(&model.Shop{ID: 1, ShareToken: "token", CreatedAt: fixedTime}, nil)

				productMock.EXPECT().
					GetProductsByShopID(1, (*string)(nil)).
					Return(nil, errors.New("query failed"))

				return shopMock, productMock
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			shopMock, productMock := tt.mockSetup(ctrl)

			oldShop := shopStore
			oldProduct := productStore
			defer func() {
				shopStore = oldShop
				productStore = oldProduct
			}()
			shopStore = shopMock
			productStore = productMock

			var s shopService
			got, gotErr := s.GetPublicProducts(tt.shareToken)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetPublicProducts() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetPublicProducts() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublicProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}
