package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
	"github.com/zeirash/recapo/arion/model"
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
					CreateOrder(gomock.Any(), 1, 1, nil).
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
					CreateOrder(gomock.Any(), 2, 1, &notes).
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
					CreateOrder(gomock.Any(), 1, 1, nil).
					Return(response.OrderData{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:        "create order returns 400 on invalid json",
			body:        "invalid json",
			shopID:      1,
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create order returns 400 on validation failure - missing customer_id",
			body: map[string]interface{}{
				"customer_id": 0,
			},
			shopID:      1,
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create order returns error customer already has an active order",
			body: map[string]interface{}{
				"customer_id": 1,
			},
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					CreateOrder(gomock.Any(), 1, 1, nil).
					Return(response.OrderData{}, errors.New(apierr.ErrActiveOrderExists))
			},
			wantStatus:     http.StatusConflict,
			wantSuccess:    false,
			wantErrMessage: "Customer already has an active order",
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
					GetOrderByID(gomock.Any(), 1, 1).
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
					GetOrderByID(gomock.Any(), 999, 1).
					Return(nil, errors.New(apierr.ErrOrderNotFound))
			},
			wantStatus:     http.StatusNotFound,
			wantSuccess:    false,
			wantErrMessage: "Order not found",
		},
		{
			name:     "get order returns error on service failure",
			shopID:   1,
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderByID(gomock.Any(), 1, 1).
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
			wantErrMessage: "Order ID is required",
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

	type queryOpts struct {
		search        string
		dateFrom      string
		dateTo        string
		status        string
		sort          string
		paymentStatus string
	}

	tests := []struct {
		name           string
		shopID         int
		opts           queryOpts
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:   "successfully get orders with all filters status",
			shopID: 1,
			opts:   queryOpts{status: "all"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{}).
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
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{}).
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
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{}).
					Return(nil, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:        "get orders with search query passes search to service",
			shopID:      1,
			opts:        queryOpts{search: "john"},
			mockSetup: func() {
				q := "john"
				mockOrderService.EXPECT().
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{SearchQuery: &q}).
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
			name:   "get orders with date_from passes filter to service",
			shopID: 1,
			opts:   queryOpts{dateFrom: "2024-01-01"},
			mockSetup: func() {
				df := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				mockOrderService.EXPECT().
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{DateFrom: &df}).
					Return([]response.OrderData{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 10000, Status: "created", CreatedAt: time.Now()},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "get orders with date_to passes filter to service",
			shopID: 1,
			opts:   queryOpts{dateTo: "2024-01-31T23:59:59Z"},
			mockSetup: func() {
				dt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
				mockOrderService.EXPECT().
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{DateTo: &dt}).
					Return([]response.OrderData{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 10000, Status: "created", CreatedAt: time.Now()},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "get orders with date_from and date_to passes filters to service",
			shopID: 1,
			opts:   queryOpts{dateFrom: "2024-01-01", dateTo: "2024-01-31T23:59:59Z"},
			mockSetup: func() {
				df := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				dt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
				mockOrderService.EXPECT().
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{DateFrom: &df, DateTo: &dt}).
					Return([]response.OrderData{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 10000, Status: "created", CreatedAt: time.Now()},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "get orders with status filter passes to service",
			shopID: 1,
			opts:   queryOpts{status: "created"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{Status: []string{"created"}}).
					Return([]response.OrderData{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 10000, Status: "created", CreatedAt: time.Now()},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "get orders with sort order",
			shopID: 1,
			opts:   queryOpts{sort: "created_at,desc"},
			mockSetup: func() {
				sort := "created_at,desc"
				mockOrderService.EXPECT().
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{Sort: &sort}).
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
			name:   "get orders with payment_status filter passes to service",
			shopID: 1,
			opts:   queryOpts{paymentStatus: "paid"},
			mockSetup: func() {
				ps := "paid"
				mockOrderService.EXPECT().
					GetOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{PaymentStatus: &ps}).
					Return([]response.OrderData{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 10000, Status: "done", CreatedAt: time.Now()},
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
			var params []string
			if tt.opts.search != "" {
				params = append(params, "search="+tt.opts.search)
			}
			if tt.opts.dateFrom != "" {
				params = append(params, "date_from="+tt.opts.dateFrom)
			}
			if tt.opts.dateTo != "" {
				params = append(params, "date_to="+tt.opts.dateTo)
			}
			if tt.opts.status != "" {
				params = append(params, "status="+tt.opts.status)
			}
			if tt.opts.sort != "" {
				params = append(params, "sort="+tt.opts.sort)
			}
			if tt.opts.paymentStatus != "" {
				params = append(params, "payment_status="+tt.opts.paymentStatus)
			}
			if len(params) > 0 {
				path += "?" + strings.Join(params, "&")
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
				"status":      "done",
				"total_price": 15000,
			},
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				status := "done"
				totalPrice := 15000
				mockOrderService.EXPECT().
					UpdateOrderByID(gomock.Any(), service.UpdateOrderInput{
						ID:         1,
						TotalPrice: &totalPrice,
						Status:     &status,
					}).
					Return(response.OrderData{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   15000,
						Status:       "done",
						CreatedAt:    time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:     "update order returns error on service failure",
			body:     map[string]interface{}{"status": "done"},
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				status := "done"
				mockOrderService.EXPECT().
					UpdateOrderByID(gomock.Any(), service.UpdateOrderInput{
						ID:     1,
						Status: &status,
					}).
					Return(response.OrderData{}, errors.New(apierr.ErrOrderNotFound))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Order not found",
		},
		{
			name:        "update order returns 400 on invalid json",
			body:        "invalid json",
			pathVars:    map[string]string{"order_id": "1"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:           "update order returns 400 on missing order_id",
			body:           map[string]interface{}{"status": "done"},
			pathVars:       map[string]string{},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Order ID is required",
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
					DeleteOrderByID(gomock.Any(), 1).
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
					DeleteOrderByID(gomock.Any(), 1).
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
			wantErrMessage: "Order ID is required",
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
					CreateOrderItem(gomock.Any(), 1, 1, 2).
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
					CreateOrderItem(gomock.Any(), 1, 1, 2).
					Return(response.OrderItemData{}, errors.New(apierr.ErrProductNotFound))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Product not found",
		},
		{
			name:        "create order item returns 400 on invalid json",
			body:        "invalid json",
			pathVars:    map[string]string{"order_id": "1"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create order item returns 400 on validation failure - missing product_id",
			body: map[string]interface{}{
				"product_id": 0,
				"qty":        2,
			},
			pathVars:    map[string]string{"order_id": "1"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create order item returns 400 on validation failure - invalid qty",
			body: map[string]interface{}{
				"product_id": 1,
				"qty":        0,
			},
			pathVars:    map[string]string{"order_id": "1"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:           "create order item returns 400 on missing order_id",
			body:           map[string]interface{}{"product_id": 1, "qty": 2},
			pathVars:       map[string]string{},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Order ID is required",
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
					UpdateOrderItemByID(gomock.Any(), service.UpdateOrderItemInput{
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
					UpdateOrderItemByID(gomock.Any(), service.UpdateOrderItemInput{
						OrderID:     1,
						OrderItemID: 1,
						Qty:         &qty,
					}).
					Return(response.OrderItemData{}, errors.New(apierr.ErrOrderItemNotFound))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Order item not found",
		},
		{
			name:        "update order item returns 400 on invalid json",
			body:        "invalid json",
			pathVars:    map[string]string{"order_id": "1", "item_id": "1"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:           "update order item returns 400 on missing order_id",
			body:           map[string]interface{}{"qty": 5},
			pathVars:       map[string]string{"item_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Order ID is required",
		},
		{
			name:           "update order item returns 400 on missing item_id",
			body:           map[string]interface{}{"qty": 5},
			pathVars:       map[string]string{"order_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Order item ID is required",
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
					DeleteOrderItemByID(gomock.Any(), 1, 1).
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
					DeleteOrderItemByID(gomock.Any(), 1, 1).
					Return(errors.New(apierr.ErrOrderItemNotFound))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Order item not found",
		},
		{
			name:           "delete order item returns 400 on missing order_id",
			pathVars:       map[string]string{"item_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Order ID is required",
		},
		{
			name:           "delete order item returns 400 on missing item_id",
			pathVars:       map[string]string{"order_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Order item ID is required",
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
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

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
					GetOrderItemByID(gomock.Any(), 1, 1).
					Return(&response.OrderItemData{
						ID:          1,
						OrderID:     1,
						ProductName: "Product A",
						Price:       5000,
						Qty:         2,
						CreatedAt:   fixedTime,
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
					GetOrderItemByID(gomock.Any(), 999, 1).
					Return(nil, errors.New(apierr.ErrOrderItemNotFound))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Order item not found",
		},
		{
			name:           "get order item returns 400 on missing order_id",
			pathVars:       map[string]string{"item_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Order ID is required",
		},
		{
			name:           "get order item returns 400 on missing item_id",
			pathVars:       map[string]string{"order_id": "1"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Order item ID is required",
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
					GetOrderItemsByOrderID(gomock.Any(), 1).
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
					GetOrderItemsByOrderID(gomock.Any(), 1).
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
					GetOrderItemsByOrderID(gomock.Any(), 1).
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
			wantErrMessage: "Order ID is required",
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

func TestMergeTempOrderHandler(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

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
		wantOrder      *response.OrderData
		wantErrMessage string
	}{
		{
			name: "successfully merge temp order (create new order)",
			body: map[string]interface{}{
				"temp_order_id": 10,
				"customer_id":  5,
			},
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					MergeTempOrder(gomock.Any(), 10, 5, 1, (*int)(nil)).
					Return(&response.OrderData{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   3000,
						Status:       "created",
						CreatedAt:    fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantOrder: &response.OrderData{
				ID:           1,
				CustomerName: "John Doe",
				TotalPrice:   3000,
				Status:       "created",
				CreatedAt:    fixedTime,
			},
		},
		{
			name: "successfully merge temp order into active order",
			body: map[string]interface{}{
				"temp_order_id":   10,
				"customer_id":    5,
				"active_order_id": 7,
			},
			shopID: 1,
			mockSetup: func() {
				activeOrderID := 7
				mockOrderService.EXPECT().
					MergeTempOrder(gomock.Any(), 10, 5, 1, &activeOrderID).
					Return(&response.OrderData{
						ID:           7,
						CustomerName: "John Doe",
						TotalPrice:   5000,
						Status:       "in_progress",
						CreatedAt:    fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantOrder: &response.OrderData{
				ID:           7,
				CustomerName: "John Doe",
				TotalPrice:   5000,
				Status:       "in_progress",
				CreatedAt:    fixedTime,
			},
		},
		{
			name: "merge temp order returns 400 on invalid JSON",
			body: "invalid json",
			shopID: 1,
			mockSetup: func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "merge temp order returns 400 when temp_order_id is missing",
			body: map[string]interface{}{
				"temp_order_id": 0,
				"customer_id":  5,
			},
			shopID:         1,
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Temp order ID is required",
		},
		{
			name: "merge temp order returns 400 when customer_id is missing",
			body: map[string]interface{}{
				"temp_order_id": 10,
				"customer_id":  0,
			},
			shopID:         1,
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Customer ID is required",
		},
		{
			name: "merge temp order returns 404 when order not found",
			body: map[string]interface{}{
				"temp_order_id": 10,
				"customer_id":  5,
			},
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					MergeTempOrder(gomock.Any(), 10, 5, 1, (*int)(nil)).
					Return(nil, errors.New(apierr.ErrOrderNotFound))
			},
			wantStatus:     http.StatusNotFound,
			wantSuccess:    false,
			wantErrMessage: "Order not found",
		},
		{
			name: "merge temp order returns 500 on service failure",
			body: map[string]interface{}{
				"temp_order_id": 10,
				"customer_id":  5,
			},
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					MergeTempOrder(gomock.Any(), 10, 5, 1, (*int)(nil)).
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

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			req := newRequestWithShopID("POST", "/temp_orders/merge", bodyBytes, tt.shopID)
			rec := httptest.NewRecorder()

			handler.MergeTempOrderHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("MergeTempOrderHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("MergeTempOrderHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("MergeTempOrderHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
			if tt.wantOrder != nil {
				dataBytes, err := json.Marshal(resp.Data)
				if err != nil {
					t.Fatalf("MergeTempOrderHandler() marshal data: %v", err)
				}
				var actual response.OrderData
				if err := json.Unmarshal(dataBytes, &actual); err != nil {
					t.Fatalf("MergeTempOrderHandler() unmarshal data: %v", err)
				}
				if !reflect.DeepEqual(actual, *tt.wantOrder) {
					t.Errorf("MergeTempOrderHandler() data = %+v, want %+v", actual, *tt.wantOrder)
				}
			}
		})
	}
}

func TestGetTempOrderHandler(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

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
		wantTempOrder  *response.TempOrderData
		wantErrMessage string
	}{
		{
			name:     "successfully get temp order",
			shopID:   1,
			pathVars: map[string]string{"temp_order_id": "10"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetTempOrderByID(gomock.Any(), 10, 1).
					Return(&response.TempOrderData{
						ID:            10,
						CustomerName:  "Jane",
						CustomerPhone: "+62812345678",
						TotalPrice:    2500,
						Status:        "pending",
						TempOrderItems: []response.TempOrderItemData{
							{ID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
							{ID: 2, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
						},
						CreatedAt: fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantTempOrder: &response.TempOrderData{
				ID:            10,
				CustomerName:  "Jane",
				CustomerPhone: "+62812345678",
				TotalPrice:    2500,
				Status:        "pending",
				TempOrderItems: []response.TempOrderItemData{
					{ID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					{ID: 2, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
			},
		},
		{
			name:     "get temp order returns 404 when not found",
			shopID:   1,
			pathVars: map[string]string{"temp_order_id": "999"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetTempOrderByID(gomock.Any(), 999, 1).
					Return(nil, errors.New(apierr.ErrTempOrderNotFound))
			},
			wantStatus:     http.StatusNotFound,
			wantSuccess:    false,
			wantErrMessage: "Temp order not found",
		},
		{
			name:     "get temp order returns 500 on service failure",
			shopID:   1,
			pathVars: map[string]string{"temp_order_id": "10"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetTempOrderByID(gomock.Any(), 10, 1).
					Return(nil, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:           "get temp order returns 400 on missing temp_order_id",
			shopID:         1,
			pathVars:       map[string]string{},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Temp order ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithShopID("GET", "/temp_orders/10", nil, tt.shopID)
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.GetTempOrderHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetTempOrderHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetTempOrderHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetTempOrderHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
			if tt.wantTempOrder != nil {
				dataBytes, err := json.Marshal(resp.Data)
				if err != nil {
					t.Fatalf("GetTempOrderHandler() marshal data: %v", err)
				}
				var actual response.TempOrderData
				if err := json.Unmarshal(dataBytes, &actual); err != nil {
					t.Fatalf("GetTempOrderHandler() unmarshal data: %v", err)
				}
				if !reflect.DeepEqual(actual, *tt.wantTempOrder) {
					t.Errorf("GetTempOrderHandler() data = %+v, want %+v", actual, *tt.wantTempOrder)
				}
			}
		})
	}
}

func TestRejectTempOrderHandler(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

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
		wantTempOrder  *response.TempOrderData
		wantErrMessage string
	}{
		{
			name:     "successfully reject temp order",
			shopID:   1,
			pathVars: map[string]string{"temp_order_id": "10"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					RejectTempOrderByID(gomock.Any(), 10).
					Return(response.TempOrderData{
						ID:            10,
						CustomerName:  "Jane",
						CustomerPhone: "+62812345678",
						TotalPrice:    2500,
						Status:        "rejected",
						TempOrderItems: []response.TempOrderItemData{
							{ID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
						},
						CreatedAt: fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantTempOrder: &response.TempOrderData{
				ID:            10,
				CustomerName:  "Jane",
				CustomerPhone: "+62812345678",
				TotalPrice:    2500,
				Status:        "rejected",
				TempOrderItems: []response.TempOrderItemData{
					{ID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
			},
		},
		{
			name:     "reject temp order returns 500 on service failure",
			shopID:   1,
			pathVars: map[string]string{"temp_order_id": "10"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					RejectTempOrderByID(gomock.Any(), 10).
					Return(response.TempOrderData{}, errors.New(apierr.ErrTempOrderNotFound))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Temp order not found",
		},
		{
			name:           "reject temp order returns 400 on missing temp_order_id",
			shopID:         1,
			pathVars:       map[string]string{},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Temp order ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithShopID("PATCH", "/temp_orders/10/reject", nil, tt.shopID)
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.RejectTempOrderHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("RejectTempOrderHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("RejectTempOrderHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("RejectTempOrderHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
			if tt.wantTempOrder != nil {
				dataBytes, err := json.Marshal(resp.Data)
				if err != nil {
					t.Fatalf("RejectTempOrderHandler() marshal data: %v", err)
				}
				var actual response.TempOrderData
				if err := json.Unmarshal(dataBytes, &actual); err != nil {
					t.Fatalf("RejectTempOrderHandler() unmarshal data: %v", err)
				}
				if !reflect.DeepEqual(actual, *tt.wantTempOrder) {
					t.Errorf("RejectTempOrderHandler() data = %+v, want %+v", actual, *tt.wantTempOrder)
				}
			}
		})
	}
}

func TestExportOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	fakePDF := []byte("%PDF-1.4 fake pdf content")

	tests := []struct {
		name            string
		shopID          int
		pathVars        map[string]string
		body            interface{}
		mockSetup       func()
		wantStatus      int
		wantContentType string
		wantErrMessage  string
	}{
		{
			name:     "successfully export order without message",
			shopID:   1,
			pathVars: map[string]string{"order_id": "1"},
			body:     nil,
			mockSetup: func() {
				mockOrderService.EXPECT().
					GenerateOrderInvoice(gomock.Any(), 1, 1, "").
					Return(fakePDF, nil)
			},
			wantStatus:      http.StatusOK,
			wantContentType: "application/pdf",
		},
		{
			name:     "successfully export order with custom message",
			shopID:   1,
			pathVars: map[string]string{"order_id": "1"},
			body:     map[string]interface{}{"message": "Thank you!\nSee you again."},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GenerateOrderInvoice(gomock.Any(), 1, 1, "Thank you!\nSee you again.").
					Return(fakePDF, nil)
			},
			wantStatus:      http.StatusOK,
			wantContentType: "application/pdf",
		},
		{
			name:     "export order returns 404 when order not found",
			shopID:   1,
			pathVars: map[string]string{"order_id": "999"},
			body:     nil,
			mockSetup: func() {
				mockOrderService.EXPECT().
					GenerateOrderInvoice(gomock.Any(), 999, 1, "").
					Return(nil, errors.New(apierr.ErrOrderNotFound))
			},
			wantStatus:     http.StatusNotFound,
			wantErrMessage: "Order not found",
		},
		{
			name:     "export order returns 500 on service failure",
			shopID:   1,
			pathVars: map[string]string{"order_id": "1"},
			body:     nil,
			mockSetup: func() {
				mockOrderService.EXPECT().
					GenerateOrderInvoice(gomock.Any(), 1, 1, "").
					Return(nil, errors.New("pdf generation error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantErrMessage: "pdf generation error",
		},
		{
			name:           "export order returns 400 on missing order_id",
			shopID:         1,
			pathVars:       map[string]string{},
			body:           nil,
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantErrMessage: "Order ID is required",
		},
		{
			name:           "export order returns 400 on invalid json body",
			shopID:         1,
			pathVars:       map[string]string{"order_id": "1"},
			body:           "invalid json",
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			if tt.body != nil {
				switch b := tt.body.(type) {
				case string:
					bodyBytes = []byte(b)
				default:
					bodyBytes, _ = json.Marshal(b)
				}
			}

			req := newRequestWithShopID("POST", "/orders/1/export", bodyBytes, tt.shopID)
			req = newRequestWithPathVars(req, tt.pathVars)
			rec := httptest.NewRecorder()

			handler.ExportOrderHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("ExportOrderHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			if tt.wantContentType != "" {
				ct := rec.Header().Get("Content-Type")
				if ct != tt.wantContentType {
					t.Errorf("ExportOrderHandler() Content-Type = %v, want %v", ct, tt.wantContentType)
				}
			}

			if tt.wantErrMessage != "" {
				var resp handler.ApiResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp.Message != tt.wantErrMessage {
					t.Errorf("ExportOrderHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
				}
			}
		})
	}
}

func TestGetTempOrdersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	type queryOpts struct {
		status   string
		search   string
		dateFrom string
		dateTo   string
		sort     string
	}

	tests := []struct {
		name           string
		shopID         int
		opts           queryOpts
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantCount      int
		wantErrMessage string
	}{
		{
			name:   "successfully get temp orders with all filters status",
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetTempOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{}).
					Return([]response.TempOrderData{
						{
							ID:            1,
							CustomerName:  "John Doe",
							CustomerPhone: "123456789",
							TotalPrice:    10000,
							Status:        "pending",
							CreatedAt:     time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
		{
			name:   "successfully get temp orders with multiple statuses",
			shopID: 1,
			opts:   queryOpts{status: "accepted,rejected"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetTempOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{Status: []string{"accepted", "rejected"}}).
					Return([]response.TempOrderData{
						{
							ID:            2,
							CustomerName:  "Jane Doe",
							CustomerPhone: "987654321",
							TotalPrice:    20000,
							Status:        "accepted",
							CreatedAt:     time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
		{
			name:   "successfully get temp orders returns empty list",
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetTempOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{}).
					Return([]response.TempOrderData{}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   0,
		},
		{
			name:   "get temp orders with search query passes search to service",
			shopID: 1,
			opts:   queryOpts{search: "john"},
				mockSetup: func() {
				q := "john"
				mockOrderService.EXPECT().
					GetTempOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{SearchQuery: &q}).
					Return([]response.TempOrderData{
						{
							ID:            1,
							CustomerName:  "John Doe",
							CustomerPhone: "123456789",
							TotalPrice:    10000,
							Status:        "pending",
							CreatedAt:     time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
		{
			name:   "get temp orders with date_from passes filter to service",
			shopID: 1,
			opts:   queryOpts{dateFrom: "2024-01-01"},
			mockSetup: func() {
				df := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				mockOrderService.EXPECT().
					GetTempOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{DateFrom: &df}).
					Return([]response.TempOrderData{
						{
							ID:            1,
							CustomerName:  "John Doe",
							CustomerPhone: "123456789",
							TotalPrice:    10000,
							Status:        "pending",
							CreatedAt:     time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
		{
			name:   "get temp orders with date_to passes filter to service",
			shopID: 1,
			opts:   queryOpts{dateTo: "2024-01-31T23:59:59Z"},
			mockSetup: func() {
				dt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
				mockOrderService.EXPECT().
					GetTempOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{DateTo: &dt}).
					Return([]response.TempOrderData{
						{
							ID:            1,
							CustomerName:  "John Doe",
							CustomerPhone: "123456789",
							TotalPrice:    10000,
							Status:        "pending",
							CreatedAt:     time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
		{
			name:   "get temp orders with rejected status filter passes to service",
			shopID: 1,
			opts:   queryOpts{status: "rejected", search: "john", dateFrom: "2024-01-01", dateTo: "2024-01-31T23:59:59Z"},
			mockSetup: func() {
				q := "john"
				df := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				dt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
				mockOrderService.EXPECT().
					GetTempOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{SearchQuery: &q, DateFrom: &df, DateTo: &dt, Status: []string{"rejected"}}).
					Return([]response.TempOrderData{
						{
							ID:            1,
							CustomerName:  "John Doe",
							CustomerPhone: "123456789",
							TotalPrice:    10000,
							Status:        "rejected",
							CreatedAt:     time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
		{
			name:   "get temp orders with sort order",
			shopID: 1,
			opts:   queryOpts{sort: "created_at,desc"},
			mockSetup: func() {
				sort := "created_at,desc"
				mockOrderService.EXPECT().
					GetTempOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{Sort: &sort}).
					Return([]response.TempOrderData{
						{
							ID:            1,
							CustomerName:  "John Doe",
							CustomerPhone: "123456789",
							TotalPrice:    10000,
							Status:        "pending",
							CreatedAt:     time.Now(),
						},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
		{
			name:   "get temp orders returns error on service failure",
			shopID: 1,
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetTempOrdersByShopID(gomock.Any(), 1, model.OrderFilterOptions{}).
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

			path := "/temp_orders"
			var params []string
			if tt.opts.status != "" {
				params = append(params, "status="+tt.opts.status)
			}
			if tt.opts.search != "" {
				params = append(params, "search="+tt.opts.search)
			}
			if tt.opts.dateFrom != "" {
				params = append(params, "date_from="+tt.opts.dateFrom)
			}
			if tt.opts.dateTo != "" {
				params = append(params, "date_to="+tt.opts.dateTo)
			}
			if tt.opts.sort != "" {
				params = append(params, "sort="+tt.opts.sort)
			}
			if len(params) > 0 {
				path += "?" + strings.Join(params, "&")
			}
			req := newRequestWithShopID("GET", path, nil, tt.shopID)
			rec := httptest.NewRecorder()

			handler.GetTempOrdersHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetTempOrdersHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetTempOrdersHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetTempOrdersHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
			if tt.wantCount >= 0 {
				orders, ok := resp.Data.([]interface{})
				if !ok && tt.wantCount > 0 {
					t.Errorf("GetTempOrdersHandler() data = %T, want array", resp.Data)
				} else if ok && len(orders) != tt.wantCount {
					t.Errorf("GetTempOrdersHandler() data count = %v, want %v", len(orders), tt.wantCount)
				}
			}
		})
	}
}

func TestCreateOrderPaymentHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	fixedTime := time.Now()

	tests := []struct {
		name        string
		pathVars    map[string]string
		body        interface{}
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:     "successfully create order payment",
			pathVars: map[string]string{"order_id": "1"},
			body:     map[string]interface{}{"amount": 50000},
			mockSetup: func() {
				mockOrderService.EXPECT().
					CreateOrderPayment(gomock.Any(), 1, 50000).
					Return(response.OrderPaymentData{ID: 1, OrderID: 1, Amount: 50000, CreatedAt: fixedTime}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 400 on missing order_id",
			pathVars:    map[string]string{},
			body:        map[string]interface{}{"amount": 50000},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on invalid json",
			pathVars:    map[string]string{"order_id": "1"},
			body:        "invalid json",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:     "returns 500 on service failure",
			pathVars: map[string]string{"order_id": "1"},
			body:     map[string]interface{}{"amount": 50000},
			mockSetup: func() {
				mockOrderService.EXPECT().
					CreateOrderPayment(gomock.Any(), 1, 50000).
					Return(response.OrderPaymentData{}, errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
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

			req := newRequestWithPathVars(
				newRequestWithShopID("POST", "/orders/1/payment", bodyBytes, 1),
				tt.pathVars,
			)
			rec := httptest.NewRecorder()

			handler.CreateOrderPaymentHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("CreateOrderPaymentHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}
			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("CreateOrderPaymentHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
		})
	}
}

func TestUpdateOrderPaymentAmountHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	fixedTime := time.Now()

	tests := []struct {
		name        string
		pathVars    map[string]string
		body        interface{}
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:     "successfully update order payment amount",
			pathVars: map[string]string{"order_id": "10", "payment_id": "1"},
			body:     map[string]interface{}{"amount": 75000},
			mockSetup: func() {
				mockOrderService.EXPECT().
					UpdateOrderPaymentAmountByID(gomock.Any(), 1, 10, 75000).
					Return(response.OrderPaymentData{ID: 1, OrderID: 10, Amount: 75000, CreatedAt: fixedTime}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 400 on missing order_id",
			pathVars:    map[string]string{"payment_id": "1"},
			body:        map[string]interface{}{"amount": 75000},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on missing payment_id",
			pathVars:    map[string]string{"order_id": "10"},
			body:        map[string]interface{}{"amount": 75000},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on invalid json",
			pathVars:    map[string]string{"order_id": "10", "payment_id": "1"},
			body:        "invalid json",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:     "returns 500 on service failure",
			pathVars: map[string]string{"order_id": "10", "payment_id": "1"},
			body:     map[string]interface{}{"amount": 75000},
			mockSetup: func() {
				mockOrderService.EXPECT().
					UpdateOrderPaymentAmountByID(gomock.Any(), 1, 10, 75000).
					Return(response.OrderPaymentData{}, errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
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

			req := newRequestWithPathVars(
				newRequestWithShopID("PATCH", "/orders/10/payments/1", bodyBytes, 1),
				tt.pathVars,
			)
			rec := httptest.NewRecorder()

			handler.UpdateOrderPaymentAmountHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("UpdateOrderPaymentAmountHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}
			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("UpdateOrderPaymentAmountHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
		})
	}
}

func TestGetOrderPaymentsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	fixedTime := time.Now()

	tests := []struct {
		name        string
		pathVars    map[string]string
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:     "successfully get order payments",
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderPaymentsByOrderID(gomock.Any(), 1).
					Return([]response.OrderPaymentData{
						{ID: 1, OrderID: 1, Amount: 50000, CreatedAt: fixedTime},
						{ID: 2, OrderID: 1, Amount: 25000, CreatedAt: fixedTime},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 400 on missing order_id",
			pathVars:    map[string]string{},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:     "returns 500 on service failure",
			pathVars: map[string]string{"order_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					GetOrderPaymentsByOrderID(gomock.Any(), 1).
					Return(nil, errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithPathVars(
				newRequestWithShopID("GET", "/orders/1/payments", nil, 1),
				tt.pathVars,
			)
			rec := httptest.NewRecorder()

			handler.GetOrderPaymentsHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetOrderPaymentsHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}
			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetOrderPaymentsHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
		})
	}
}

func TestDeleteOrderPaymentHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderService := mock_service.NewMockOrderService(ctrl)
	handler.SetOrderService(mockOrderService)

	tests := []struct {
		name        string
		pathVars    map[string]string
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:     "successfully delete order payment",
			pathVars: map[string]string{"order_id": "10", "payment_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					DeleteOrderPaymentByID(gomock.Any(), 1, 10).
					Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 400 on missing order_id",
			pathVars:    map[string]string{"payment_id": "1"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on missing payment_id",
			pathVars:    map[string]string{"order_id": "10"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:     "returns 500 on service failure",
			pathVars: map[string]string{"order_id": "10", "payment_id": "1"},
			mockSetup: func() {
				mockOrderService.EXPECT().
					DeleteOrderPaymentByID(gomock.Any(), 1, 10).
					Return(errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithPathVars(
				newRequestWithShopID("DELETE", "/orders/10/payments/1", nil, 1),
				tt.pathVars,
			)
			rec := httptest.NewRecorder()

			handler.DeleteOrderPaymentHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("DeleteOrderPaymentHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}
			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("DeleteOrderPaymentHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
		})
	}
}
