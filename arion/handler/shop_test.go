package handler_test

import (
	"bytes"
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

func TestCreateShopOrderTempHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldOrderService := handler.GetOrderService()
	defer handler.SetOrderService(oldOrderService)

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		shareToken     string
		body           interface{}
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:       "successfully create shop order temp",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_name":  "Jane Doe",
				"customer_phone": "+62812345678",
				"order_items": []interface{}{
					map[string]interface{}{"product_id": 10, "qty": 2},
					map[string]interface{}{"product_id": 20, "qty": 1},
				},
			},
			mockSetup: func() {
				mockOrderService.EXPECT().
					CreateOrderTemp("Jane Doe", "+62812345678", "share-abc123", gomock.Any()).
					Return(response.OrderTempData{
						ID:            1,
						CustomerName:  "Jane Doe",
						CustomerPhone: "+62812345678",
						TotalPrice:    2500,
						Status:        "pending",
						OrderTempItems: []response.OrderItemTempData{
							{ID: 1, OrderTempID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
							{ID: 2, OrderTempID: 1, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
						},
						CreatedAt: fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:       "returns 400 when share_token is missing",
			shareToken: "",
			body: map[string]interface{}{
				"customer_name":  "Jane Doe",
				"customer_phone": "+62812345678",
				"order_items":    []interface{}{},
			},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "share_token is required",
		},
		{
			name:       "returns 400 when body is invalid JSON",
			shareToken: "share-abc123",
			body:       "{ invalid json }",
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:       "returns 400 when customer_name is missing",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_phone": "+62812345678",
				"order_items":     []interface{}{},
			},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "customer_name is required",
		},
		{
			name:       "returns 400 when customer_phone is missing",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_name": "Jane Doe",
				"order_items":    []interface{}{},
			},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "customer_phone is required",
		},
		{
			name:       "returns 400 when customer_name is empty string",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_name":  "",
				"customer_phone": "+62812345678",
				"order_items":     []interface{}{map[string]interface{}{"product_id": 10, "qty": 1}},
			},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "customer_name is required",
		},
		{
			name:       "returns 400 when customer_phone is empty string",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_name":  "Jane Doe",
				"customer_phone": "",
				"order_items":    []interface{}{map[string]interface{}{"product_id": 10, "qty": 1}},
			},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "customer_phone is required",
		},
		{
			name:       "returns 400 when order_items is missing",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_name":  "Jane Doe",
				"customer_phone": "+62812345678",
			},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_items is required",
		},
		{
			name:       "returns 400 when order_items is empty",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_name":  "Jane Doe",
				"customer_phone": "+62812345678",
				"order_items":     []interface{}{},
			},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_items is required",
		},
		{
			name:       "returns 400 when product_id is zero",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_name":  "Jane Doe",
				"customer_phone": "+62812345678",
				"order_items":    []interface{}{map[string]interface{}{"product_id": 0, "qty": 1}},
			},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "product_id is required",
		},
		{
			name:       "returns 400 when qty is zero",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_name":  "Jane Doe",
				"customer_phone": "+62812345678",
				"order_items":    []interface{}{map[string]interface{}{"product_id": 10, "qty": 0}},
			},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "qty is required",
		},
		{
			name:       "returns 500 when order service fails",
			shareToken: "share-abc123",
			body: map[string]interface{}{
				"customer_name":  "Jane Doe",
				"customer_phone": "+62812345678",
				"order_items":    []interface{}{map[string]interface{}{"product_id": 10, "qty": 1}},
			},
			mockSetup: func() {
				mockOrderService.EXPECT().
					CreateOrderTemp("Jane Doe", "+62812345678", "share-abc123", gomock.Any()).
					Return(response.OrderTempData{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			path := "/public/shops/" + tt.shareToken + "/orders"
			if tt.shareToken == "" {
				path = "/public/shops//orders"
			}
			var bodyBytes []byte
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest("POST", path, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.shareToken != "" {
				req = mux.SetURLVars(req, map[string]string{"share_token": tt.shareToken})
			}

			rec := httptest.NewRecorder()
			handler.CreateShopOrderTempHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("CreateShopOrderTempHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("CreateShopOrderTempHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("CreateShopOrderTempHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}
