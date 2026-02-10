package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
)

func newRequestWithShareToken(method, path string, shareToken string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if shareToken != "" {
		req = mux.SetURLVars(req, map[string]string{"share_token": shareToken})
	}
	return req
}

func TestGetShopProductsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetShopService()
	defer handler.SetShopService(oldService)

	mockShopService := mock_service.NewMockShopService(ctrl)
	handler.SetShopService(mockShopService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		shareToken     string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantCount      int
		wantErrMessage string
	}{
		{
			name:       "successfully get shop products",
			shareToken: "abc123xyz",
			mockSetup: func() {
				mockShopService.EXPECT().
					GetPublicProducts("abc123xyz").
					Return([]response.ProductData{
						{ID: 1, Name: "Product A", Price: 1000, CreatedAt: fixedTime},
						{ID: 2, Name: "Product B", Price: 500, CreatedAt: fixedTime},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   2,
		},
		{
			name:       "successfully get shop products - empty list",
			shareToken: "empty123",
			mockSetup: func() {
				mockShopService.EXPECT().
					GetPublicProducts("empty123").
					Return([]response.ProductData{}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   0,
		},
		{
			name:       "returns 400 when share_token is missing",
			shareToken: "",
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:       "returns 404 when shop not found",
			shareToken: "invalid",
			mockSetup: func() {
				mockShopService.EXPECT().
					GetPublicProducts("invalid").
					Return(nil, errors.New("shop not found"))
			},
			wantStatus:     http.StatusNotFound,
			wantSuccess:    false,
			wantErrMessage: "shop not found",
		},
		{
			name:       "returns 500 on service error",
			shareToken: "token",
			mockSetup: func() {
				mockShopService.EXPECT().
					GetPublicProducts("token").
					Return(nil, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithShareToken("GET", "/public/shops/"+tt.shareToken+"/products", tt.shareToken)
			rec := httptest.NewRecorder()

			handler.GetShopProductsHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetShopProductsHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetShopProductsHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetShopProductsHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
			if tt.wantCount > 0 {
				products, ok := resp.Data.([]interface{})
				if !ok || len(products) != tt.wantCount {
					t.Errorf("GetShopProductsHandler() data count = %v, want %v", len(products), tt.wantCount)
				}
			}
		})
	}
}

func TestGetShopShareTokenHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetShopService()
	defer handler.SetShopService(oldService)

	mockShopService := mock_service.NewMockShopService(ctrl)
	handler.SetShopService(mockShopService)

	tests := []struct {
		name           string
		shopID         int
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantShareToken string
		wantErrMessage string
	}{
		{
			name:   "successfully get share token",
			shopID: 1,
			mockSetup: func() {
				mockShopService.EXPECT().
					GetShareTokenByID(1).
					Return("abc123xyz789", nil)
			},
			wantStatus:     http.StatusOK,
			wantSuccess:    true,
			wantShareToken: "abc123xyz789",
		},
		{
			name:   "returns 404 when shop not found",
			shopID: 999,
			mockSetup: func() {
				mockShopService.EXPECT().
					GetShareTokenByID(999).
					Return("", errors.New("shop not found"))
			},
			wantStatus:     http.StatusNotFound,
			wantSuccess:    false,
			wantErrMessage: "shop not found",
		},
		{
			name:   "returns 500 on service error",
			shopID: 1,
			mockSetup: func() {
				mockShopService.EXPECT().
					GetShareTokenByID(1).
					Return("", errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithShopID("GET", "/shop/share_token", nil, tt.shopID)
			rec := httptest.NewRecorder()

			handler.GetShopShareTokenHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetShopShareTokenHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetShopShareTokenHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetShopShareTokenHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
			if tt.wantShareToken != "" {
				data, ok := resp.Data.(map[string]interface{})
				if !ok {
					t.Errorf("GetShopShareTokenHandler() data = %T, want map", resp.Data)
				} else if data["share_token"] != tt.wantShareToken {
					t.Errorf("GetShopShareTokenHandler() data.share_token = %v, want %v", data["share_token"], tt.wantShareToken)
				}
			}
		})
	}
}
