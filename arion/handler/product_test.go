package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
	"github.com/zeirash/recapo/arion/model"
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
					CreateProduct(gomock.Any(), 1, "Test Product", &desc, 1000, &orgPrice, nil).
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
					CreateProduct(gomock.Any(), 1, "Test", nil, 100, nil, nil).
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
					GetProductByID(gomock.Any(), 1, 1).
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
					GetProductByID(gomock.Any(), 999, 1).
					Return(nil, errors.New(apierr.ErrProductNotFound))
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
					GetProductByID(gomock.Any(), 1, 1).
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
					GetProductsByShopID(gomock.Any(), 1, model.FilterOptions{}).
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
					GetProductsByShopID(gomock.Any(), 1, model.FilterOptions{SearchQuery: &q}).
					Return([]response.ProductData{
						{ID: 1, Name: "Widget A", Price: 1000, CreatedAt: fixedTime},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
		{
			name:   "successfully get products with sort param",
			url:    "/products?sort=name,asc",
			shopID: 1,
			mockSetup: func() {
				s := "name,asc"
				mockProductService.EXPECT().
					GetProductsByShopID(gomock.Any(), 1, model.FilterOptions{Sort: &s}).
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
					GetProductsByShopID(gomock.Any(), 1, model.FilterOptions{}).
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
					UpdateProduct(gomock.Any(), service.UpdateProductInput{
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
					UpdateProduct(gomock.Any(), gomock.Any()).
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
					DeleteProductByID(gomock.Any(), 1).
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
					DeleteProductByID(gomock.Any(), 1).
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

func TestPurchaseListProductHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetProductService()
	defer handler.SetProductService(oldService)

	mockProductService := mock_service.NewMockProductService(ctrl)
	handler.SetProductService(mockProductService)

	tests := []struct {
		name           string
		shopID         int
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantCount      int
		wantErrMessage string
	}{
		{
			name:   "returns purchase list products",
			shopID: 10,
			mockSetup: func() {
				mockProductService.EXPECT().
					GetPurchaseListProducts(gomock.Any(), 10).
					Return([]response.PurchaseListProductData{
						{ProductName: "Product A", Price: 1000, Qty: 5},
						{ProductName: "Product B", Price: 2000, Qty: 3},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   2,
		},
		{
			name:   "returns empty list when no active order products",
			shopID: 20,
			mockSetup: func() {
				mockProductService.EXPECT().
					GetPurchaseListProducts(gomock.Any(), 20).
					Return([]response.PurchaseListProductData{}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   0,
		},
		{
			name:   "returns 500 on service error",
			shopID: 10,
			mockSetup: func() {
				mockProductService.EXPECT().
					GetPurchaseListProducts(gomock.Any(), 10).
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

			req := newRequestWithShopID("GET", "/products/purchase_list", nil, tt.shopID)
			rec := httptest.NewRecorder()

			handler.PurchaseListProductHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("PurchaseListProductHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("PurchaseListProductHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("PurchaseListProductHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
			if tt.wantCount >= 0 {
				products, ok := resp.Data.([]interface{})
				if !ok && tt.wantCount > 0 {
					t.Errorf("PurchaseListProductHandler() data = %T, want array", resp.Data)
				} else if ok && len(products) != tt.wantCount {
					t.Errorf("PurchaseListProductHandler() data count = %v, want %v", len(products), tt.wantCount)
				}
			}
		})
	}
}

func TestUploadProductImageHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetProductService()
	defer handler.SetProductService(oldService)

	mockProductService := mock_service.NewMockProductService(ctrl)
	handler.SetProductService(mockProductService)

	buildMultipartRequest := func(fieldName, filename string, content []byte) *http.Request {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile(fieldName, filename)
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		part.Write(content)
		writer.Close()

		req := httptest.NewRequest("POST", "/products/image", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		return req
	}

	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

	tests := []struct {
		name        string
		buildReq    func() *http.Request
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "successfully upload jpeg image",
			buildReq: func() *http.Request {
				return buildMultipartRequest("image", "test.jpg", jpegBytes)
			},
			mockSetup: func() {
				mockProductService.EXPECT().
					UploadProductImage(gomock.Any(), gomock.Any()).
					Return("/uploads/products/abc123.jpg", nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "returns 400 when image field is missing",
			buildReq: func() *http.Request {
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				writer.WriteField("other_field", "value")
				writer.Close()
				req := httptest.NewRequest("POST", "/products/image", &body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "returns 400 when service returns unsupported image type error",
			buildReq: func() *http.Request {
				return buildMultipartRequest("image", "test.txt", []byte("hello world plain text"))
			},
			mockSetup: func() {
				mockProductService.EXPECT().
					UploadProductImage(gomock.Any(), gomock.Any()).
					Return("", errors.New(apierr.ErrUnsupportedImageType))
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "returns 500 on service internal error",
			buildReq: func() *http.Request {
				return buildMultipartRequest("image", "test.jpg", jpegBytes)
			},
			mockSetup: func() {
				mockProductService.EXPECT().
					UploadProductImage(gomock.Any(), gomock.Any()).
					Return("", errors.New("failed to save file"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := tt.buildReq()
			rec := httptest.NewRecorder()

			handler.UploadProductImageHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("UploadProductImageHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("UploadProductImageHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}

			if tt.wantSuccess {
				data, ok := resp.Data.(map[string]interface{})
				if !ok {
					t.Errorf("UploadProductImageHandler() data is not a map, got %T", resp.Data)
					return
				}
				imageURL, ok := data["image_url"].(string)
				if !ok || imageURL == "" {
					t.Errorf("UploadProductImageHandler() image_url missing or empty in response")
				}
			}
		})
	}
}

func TestDeleteProductImageHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetProductService()
	defer handler.SetProductService(oldService)

	mockProductService := mock_service.NewMockProductService(ctrl)
	handler.SetProductService(mockProductService)

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name: "successfully delete image",
			body: map[string]interface{}{"image_url": "/uploads/products/abc123.jpg"},
			mockSetup: func() {
				mockProductService.EXPECT().
					DeleteProductImage(gomock.Any(), "/uploads/products/abc123.jpg").
					Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 400 when image_url is missing",
			body:        map[string]interface{}{},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on invalid json",
			body:        "invalid json",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "returns 400 when service returns invalid URL error",
			body: map[string]interface{}{"image_url": "/other/path/image.jpg"},
			mockSetup: func() {
				mockProductService.EXPECT().
					DeleteProductImage(gomock.Any(), "/other/path/image.jpg").
					Return(errors.New(apierr.ErrInvalidImageURL))
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "returns 404 when image not found",
			body: map[string]interface{}{"image_url": "/uploads/products/missing.jpg"},
			mockSetup: func() {
				mockProductService.EXPECT().
					DeleteProductImage(gomock.Any(), "/uploads/products/missing.jpg").
					Return(errors.New(apierr.ErrImageNotFound))
			},
			wantStatus:  http.StatusNotFound,
			wantSuccess: false,
		},
		{
			name: "returns 500 on service internal error",
			body: map[string]interface{}{"image_url": "/uploads/products/abc123.jpg"},
			mockSetup: func() {
				mockProductService.EXPECT().
					DeleteProductImage(gomock.Any(), "/uploads/products/abc123.jpg").
					Return(errors.New("permission denied"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "permission denied",
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

			req := httptest.NewRequest("DELETE", "/products/image", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler.DeleteProductImageHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("DeleteProductImageHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("DeleteProductImageHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("DeleteProductImageHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}
