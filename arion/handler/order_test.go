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
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
	"github.com/zeirash/recapo/arion/service"
)

func TestCreateOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		body           interface{}
		shopID         int
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name: "successfully create order",
			body: map[string]interface{}{
				"customer_id": 1,
			},
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					CreateOrder(1, 1, nil).
					Return(response.OrderData{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   0,
						Status:       "created",
						CreatedAt:    time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "successfully create order with notes",
			body: map[string]interface{}{
				"customer_id": 2,
				"notes":       "Rush delivery",
			},
			shopID: 1,
			mockSetup: func() {
				notes := "Rush delivery"
				mockOrderService.EXPECT().
					CreateOrder(2, 1, &notes).
					Return(response.OrderData{
						ID:           2,
						CustomerName: "Jane Doe",
						TotalPrice:   0,
						Status:       "created",
						Notes:        "Rush delivery",
						CreatedAt:    time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "create order returns error on service failure",
			body: map[string]interface{}{
				"customer_id": 1,
			},
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					CreateOrder(1, 1, nil).
					Return(response.OrderData{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:       "create order returns 400 on invalid json",
			body:       "invalid json",
			shopID:     1,
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create order returns 400 on validation failure - missing customer_id",
			body: map[string]interface{}{
				"customer_id": 0,
			},
			shopID:     1,
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			req := newRequestWithShopID("POST", "/order", bodyBytes, tt.shopID)
			rec := httptest.NewRecorder()

			handler.CreateOrderHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("CreateOrderHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("CreateOrderHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("CreateOrderHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestGetOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		shopID         int
		pathVars       map[string]string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:     "successfully get order",
			shopID:   1,
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderByID(1, 1).
					Return(&response.OrderData{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   10000,
						Status:       "created",
						CreatedAt:    time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:     "get order returns 404 when not found",
			shopID:   1,
			pathVars: map[string]string{"order_id": "999"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderByID(999, 1).
					Return(nil, nil)
			},
			wantStatus:     http.StatusNotFound,
			wantSuccess:    false,
			wantErrMessage: "order not found",
		},
		{
			name:     "get order returns error on service failure",
			shopID:   1,
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderByID(1, 1).
					Return(nil, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:           "get order returns 400 on missing order_id",
			shopID:         1,
			pathVars:       map[string]string{},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithShopID("GET", "/order/1", nil, tt.shopID)
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.GetOrderHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetOrderHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetOrderHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetOrderHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestGetOrdersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		shopID         int
		searchQuery    string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:   "successfully get orders",
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrdersByShopID(1, nil).
					Return([]response.OrderData{
						{
							ID:           1,
							CustomerName: "John Doe",
							TotalPrice:   10000,
							Status:       "created",
							CreatedAt:    time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "get orders returns empty list",
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrdersByShopID(1, nil).
					Return([]response.OrderData{}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "get orders returns error on service failure",
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrdersByShopID(1, nil).
					Return(nil, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:        "get orders with search query passes search to service",
			shopID:      1,
			searchQuery: "john",
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrdersByShopID(1, gomock.Any()).
					Return([]response.OrderData{
						{
							ID:           1,
							CustomerName: "John Doe",
							TotalPrice:   10000,
							Status:       "created",
							CreatedAt:    time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			path := "/orders"
			if tt.searchQuery != "" {
				path += "?search=" + tt.searchQuery
			}
			req := newRequestWithShopID("GET", path, nil, tt.shopID)
			rec := httptest.NewRecorder()

			handler.GetOrdersHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetOrdersHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetOrdersHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetOrdersHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestUpdateOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		body           interface{}
		pathVars       map[string]string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name: "successfully update order",
			body: map[string]interface{}{
				"status":      "completed",
				"total_price": 15000,
			},
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				status := "completed"
				totalPrice := 15000
				mockOrderService.EXPECT().
					UpdateOrderByID(service.UpdateOrderInput{
						ID:         1,
						TotalPrice: &totalPrice,
						Status:     &status,
					}).
					Return(response.OrderData{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   15000,
						Status:       "completed",
						CreatedAt:    time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:     "update order returns error on service failure",
			body:     map[string]interface{}{"status": "completed"},
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				status := "completed"
				mockOrderService.EXPECT().
					UpdateOrderByID(service.UpdateOrderInput{
						ID:     1,
						Status: &status,
					}).
					Return(response.OrderData{}, errors.New("order not found"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "order not found",
		},
		{
			name:           "update order returns 400 on invalid json",
			body:           "invalid json",
			pathVars:       map[string]string{"order_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
		},
		{
			name:           "update order returns 400 on missing order_id",
			body:           map[string]interface{}{"status": "completed"},
			pathVars:       map[string]string{},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			req := httptest.NewRequest("PUT", "/order/1", bytes.NewReader(bodyBytes))
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.UpdateOrderHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("UpdateOrderHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("UpdateOrderHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("UpdateOrderHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestDeleteOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		pathVars       map[string]string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:     "successfully delete order",
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					DeleteOrderByID(1).
					Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:     "delete order returns error on service failure",
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					DeleteOrderByID(1).
					Return(errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:           "delete order returns 400 on missing order_id",
			pathVars:       map[string]string{},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("DELETE", "/order/1", nil)
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.DeleteOrderHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("DeleteOrderHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("DeleteOrderHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("DeleteOrderHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestCreateOrderItemHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		body           interface{}
		pathVars       map[string]string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name: "successfully create order item",
			body: map[string]interface{}{
				"product_id": 1,
				"qty":        2,
			},
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					CreateOrderItem(1, 1, 2).
					Return(response.OrderItemData{
						ID:          1,
						OrderID:     1,
						ProductName: "Product A",
						Price:       5000,
						Qty:         2,
						CreatedAt:   time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "create order item returns error on service failure",
			body: map[string]interface{}{
				"product_id": 1,
				"qty":        2,
			},
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					CreateOrderItem(1, 1, 2).
					Return(response.OrderItemData{}, errors.New("product not found"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "product not found",
		},
		{
			name:           "create order item returns 400 on invalid json",
			body:           "invalid json",
			pathVars:       map[string]string{"order_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
		},
		{
			name: "create order item returns 400 on validation failure - missing product_id",
			body: map[string]interface{}{
				"product_id": 0,
				"qty":        2,
			},
			pathVars:   map[string]string{"order_id": "1"},
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create order item returns 400 on validation failure - invalid qty",
			body: map[string]interface{}{
				"product_id": 1,
				"qty":        0,
			},
			pathVars:   map[string]string{"order_id": "1"},
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:           "create order item returns 400 on missing order_id",
			body:           map[string]interface{}{"product_id": 1, "qty": 2},
			pathVars:       map[string]string{},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			req := httptest.NewRequest("POST", "/order/1/items", bytes.NewReader(bodyBytes))
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.CreateOrderItemHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("CreateOrderItemHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("CreateOrderItemHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("CreateOrderItemHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestUpdateOrderItemHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		body           interface{}
		pathVars       map[string]string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name: "successfully update order item",
			body: map[string]interface{}{
				"qty": 5,
			},
			pathVars: map[string]string{"order_id": "1", "item_id": "1"},
			mockSetup: func() {
				qty := 5
				mockOrderService.EXPECT().
					UpdateOrderItemByID(service.UpdateOrderItemInput{
						OrderID:     1,
						OrderItemID: 1,
						Qty:         &qty,
					}).
					Return(response.OrderItemData{
						ID:          1,
						OrderID:     1,
						ProductName: "Product A",
						Price:       5000,
						Qty:         5,
						CreatedAt:   time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:     "update order item returns error on service failure",
			body:     map[string]interface{}{"qty": 5},
			pathVars: map[string]string{"order_id": "1", "item_id": "1"},
			mockSetup: func() {
				qty := 5
				mockOrderService.EXPECT().
					UpdateOrderItemByID(service.UpdateOrderItemInput{
						OrderID:     1,
						OrderItemID: 1,
						Qty:         &qty,
					}).
					Return(response.OrderItemData{}, errors.New("order item not found"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "order item not found",
		},
		{
			name:           "update order item returns 400 on invalid json",
			body:           "invalid json",
			pathVars:       map[string]string{"order_id": "1", "item_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
		},
		{
			name:           "update order item returns 400 on missing order_id",
			body:           map[string]interface{}{"qty": 5},
			pathVars:       map[string]string{"item_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_id is required",
		},
		{
			name:           "update order item returns 400 on missing item_id",
			body:           map[string]interface{}{"qty": 5},
			pathVars:       map[string]string{"order_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "item_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			req := httptest.NewRequest("PUT", "/order/1/items/1", bytes.NewReader(bodyBytes))
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.UpdateOrderItemHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("UpdateOrderItemHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("UpdateOrderItemHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("UpdateOrderItemHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestDeleteOrderItemHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		pathVars       map[string]string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:     "successfully delete order item",
			pathVars: map[string]string{"order_id": "1", "item_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					DeleteOrderItemByID(1, 1).
					Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:     "delete order item returns error on service failure",
			pathVars: map[string]string{"order_id": "1", "item_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					DeleteOrderItemByID(1, 1).
					Return(errors.New("order item not found"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "order item not found",
		},
		{
			name:           "delete order item returns 400 on missing order_id",
			pathVars:       map[string]string{"item_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_id is required",
		},
		{
			name:           "delete order item returns 400 on missing item_id",
			pathVars:       map[string]string{"order_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "item_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("DELETE", "/order/1/items/1", nil)
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.DeleteOrderItemHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("DeleteOrderItemHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("DeleteOrderItemHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("DeleteOrderItemHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestGetOrderItemHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		pathVars       map[string]string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:     "successfully get order item",
			pathVars: map[string]string{"order_id": "1", "item_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderItemByID(1, 1).
					Return(response.OrderItemData{
						ID:          1,
						OrderID:     1,
						ProductName: "Product A",
						Price:       5000,
						Qty:         2,
						CreatedAt:   time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:     "get order item returns error on service failure",
			pathVars: map[string]string{"order_id": "1", "item_id": "999"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderItemByID(999, 1).
					Return(response.OrderItemData{}, errors.New("order item not found"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "order item not found",
		},
		{
			name:           "get order item returns 400 on missing order_id",
			pathVars:       map[string]string{"item_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_id is required",
		},
		{
			name:           "get order item returns 400 on missing item_id",
			pathVars:       map[string]string{"order_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "item_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/order/1/items/1", nil)
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.GetOrderItemHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetOrderItemHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetOrderItemHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetOrderItemHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestGetOrderItemsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name           string
		pathVars       map[string]string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:     "successfully get order items",
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderItemsByOrderID(1).
					Return([]response.OrderItemData{
						{
							ID:          1,
							ProductName: "Product A",
							Price:       5000,
							Qty:         2,
							CreatedAt:   time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:     "get order items returns empty list",
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderItemsByOrderID(1).
					Return([]response.OrderItemData{}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:     "get order items returns error on service failure",
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderItemsByOrderID(1).
					Return(nil, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:           "get order items returns 400 on missing order_id",
			pathVars:       map[string]string{},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "order_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/order/1/items", nil)
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.GetOrderItemsHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetOrderItemsHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetOrderItemsHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetOrderItemsHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}
