package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
	"github.com/zeirash/recapo/arion/service"
)

func newRequestWithShopID(method, path string, body []byte, shopID int) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	return req.WithContext(context.WithValue(req.Context(), common.ShopIDKey, shopID))
}

func newRequestWithPathVars(r *http.Request, vars map[string]string) *http.Request {
	return mux.SetURLVars(r, vars)
}

func TestCreateProductHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetProductService()
	defer handler.SetProductService(oldService)

	mockProductService := mock_service.NewMockProductService(ctrl)
	handler.SetProductService(mockProductService)

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
			name: "successfully create product",
			body: map[string]interface{}{
				"name":        "Test Product",
				"price":       1000,
				"description": "Test description",
				"original_price": 800,
			},
			shopID: 1,
			mockSetup: func() {
				desc := "Test description"
				orgPrice := 800
				mockProductService.EXPECT().
					CreateProduct(1, "Test Product", &desc, 1000, &orgPrice).
					Return(response.ProductData{
						ID:            1,
						Name:          "Test Product",
						Description:   "Test description",
						Price:         1000,
						OriginalPrice: 800,
						CreatedAt:     time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess:  true,
		},
		{
			name: "create product returns error on service failure",
			body: map[string]interface{}{
				"name":  "Test",
				"price": 100,
			},
			shopID: 1,
			mockSetup: func() {
				mockProductService.EXPECT().
					CreateProduct(1, "Test", nil, 100, nil).
					Return(response.ProductData{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:       "create product returns 400 on invalid json",
			body:       "invalid json",
			shopID:     1,
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create product returns 400 on validation failure - missing name",
			body: map[string]interface{}{
				"name":  "",
				"price": 100,
			},
			shopID:     1,
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create product returns 400 on validation failure - negative price",
			body: map[string]interface{}{
				"name":  "Test",
				"price": -1,
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

			req := newRequestWithShopID("POST", "/product", bodyBytes, tt.shopID)
			rec := httptest.NewRecorder()

			handler.CreateProductHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("CreateProductHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("CreateProductHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("CreateProductHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestGetProductHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetProductService()
	defer handler.SetProductService(oldService)

	mockProductService := mock_service.NewMockProductService(ctrl)
	handler.SetProductService(mockProductService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		productID      string
		shopID         int
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:      "successfully get product by ID",
			productID: "1",
			shopID:    1,
			mockSetup: func() {
				mockProductService.EXPECT().
					GetProductByID(1, 1).
					Return(&response.ProductData{
						ID:          1,
						Name:        "Product A",
						Description: "Desc A",
						Price:       1000,
						CreatedAt:   fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess:  true,
		},
		{
			name:      "get product returns 400 when product_id missing",
			productID: "",
			shopID:    1,
			mockSetup: func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "get product returns 404 when not found",
			productID: "999",
			shopID:    1,
			mockSetup: func() {
				mockProductService.EXPECT().
					GetProductByID(999, 1).
					Return(nil, nil)
			},
			wantStatus:  http.StatusNotFound,
			wantSuccess: false,
		},
		{
			name:      "get product returns 500 on service error",
			productID: "1",
			shopID:    1,
			mockSetup: func() {
				mockProductService.EXPECT().
					GetProductByID(1, 1).
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

			req := newRequestWithShopID("GET", "/products/"+tt.productID, nil, tt.shopID)
			if tt.productID != "" {
				req = newRequestWithPathVars(req, map[string]string{"product_id": tt.productID})
			}
			rec := httptest.NewRecorder()

			handler.GetProductHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetProductHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetProductHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetProductHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestGetProductsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetProductService()
	defer handler.SetProductService(oldService)

	mockProductService := mock_service.NewMockProductService(ctrl)
	handler.SetProductService(mockProductService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		url         string
		shopID      int
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
		wantCount   int
	}{
		{
			name:   "successfully get products list",
			url:    "/products",
			shopID: 1,
			mockSetup: func() {
				mockProductService.EXPECT().
					GetProductsByShopID(1, nil).
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
			name:   "successfully get products with search query",
			url:    "/products?search=widget",
			shopID: 1,
			mockSetup: func() {
				q := "widget"
				mockProductService.EXPECT().
					GetProductsByShopID(1, &q).
					Return([]response.ProductData{
						{ID: 1, Name: "Widget A", Price: 1000, CreatedAt: fixedTime},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
		{
			name:   "get products returns 500 on service error",
			url:    "/products",
			shopID: 1,
			mockSetup: func() {
				mockProductService.EXPECT().
					GetProductsByShopID(1, nil).
					Return(nil, errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithShopID("GET", tt.url, nil, tt.shopID)
			rec := httptest.NewRecorder()

			handler.GetProductsHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetProductsHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetProductsHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantCount > 0 {
				products, ok := resp.Data.([]interface{})
				if !ok || len(products) != tt.wantCount {
					t.Errorf("GetProductsHandler() data count = %v, want %v", len(products), tt.wantCount)
				}
			}
		})
	}
}

func TestUpdateProductHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetProductService()
	defer handler.SetProductService(oldService)

	mockProductService := mock_service.NewMockProductService(ctrl)
	handler.SetProductService(mockProductService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		productID      string
		body           interface{}
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:      "successfully update product",
			productID: "1",
			body: map[string]interface{}{
				"name":           "Updated Product",
				"price":          2000,
				"original_price": 2000,
			},
			mockSetup: func() {
				name := "Updated Product"
				price := 2000
				mockProductService.EXPECT().
					UpdateProduct(service.UpdateProductInput{
						ID:            1,
						Name:          &name,
						Price:         &price,
						OriginalPrice: &price,
					}).
					Return(response.ProductData{
						ID:            1,
						Name:          "Updated Product",
						Price:         2000,
						OriginalPrice: 2000,
						CreatedAt:     fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:      "update product returns 400 when product_id missing",
			productID: "",
			body:      map[string]interface{}{"name": "Test"},
			mockSetup: func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "update product returns 400 on invalid json",
			productID: "1",
			body:      "invalid json",
			mockSetup: func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "update product returns 500 on service error",
			productID: "1",
			body: map[string]interface{}{
				"name": "Test",
			},
			mockSetup: func() {
				mockProductService.EXPECT().
					UpdateProduct(gomock.Any()).
					Return(response.ProductData{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("PATCH", "/products/"+tt.productID, bytes.NewReader(bodyBytes))
			if tt.productID != "" {
				req = newRequestWithPathVars(req, map[string]string{"product_id": tt.productID})
			}
			rec := httptest.NewRecorder()

			handler.UpdateProductHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("UpdateProductHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("UpdateProductHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("UpdateProductHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestDeleteProductHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetProductService()
	defer handler.SetProductService(oldService)

	mockProductService := mock_service.NewMockProductService(ctrl)
	handler.SetProductService(mockProductService)

	tests := []struct {
		name           string
		productID      string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:      "successfully delete product",
			productID: "1",
			mockSetup: func() {
				mockProductService.EXPECT().
					DeleteProductByID(1).
					Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:      "delete product returns 400 when product_id missing",
			productID: "",
			mockSetup: func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "delete product returns 500 on service error",
			productID: "1",
			mockSetup: func() {
				mockProductService.EXPECT().
					DeleteProductByID(1).
					Return(errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("DELETE", "/products/"+tt.productID, nil)
			if tt.productID != "" {
				req = newRequestWithPathVars(req, map[string]string{"product_id": tt.productID})
			}
			rec := httptest.NewRecorder()

			handler.DeleteProductHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("DeleteProductHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("DeleteProductHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("DeleteProductHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}
