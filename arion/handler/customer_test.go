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
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
	"github.com/zeirash/recapo/arion/service"
)

func TestCreateCustomerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetCustomerService()
	defer handler.SetCustomerService(oldService)

	mockCustomerService := mock_service.NewMockCustomerService(ctrl)
	handler.SetCustomerService(mockCustomerService)

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
			name: "successfully create customer",
			body: map[string]interface{}{
				"name":    "John Doe",
				"phone":   "08123456789",
				"address": "123 Main St",
			},
			shopID: 1,
			mockSetup: func() {
				mockCustomerService.EXPECT().
					CreateCustomer("John Doe", "08123456789", "123 Main St", 1).
					Return(response.CustomerData{
						ID:        1,
						Name:      "John Doe",
						Phone:     "08123456789",
						Address:   "123 Main St",
						CreatedAt: time.Now(),
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "create customer returns error on service failure",
			body: map[string]interface{}{
				"name":    "John",
				"phone":   "08123456789",
				"address": "123 Main St",
			},
			shopID: 1,
			mockSetup: func() {
				mockCustomerService.EXPECT().
					CreateCustomer("John", "08123456789", "123 Main St", 1).
					Return(response.CustomerData{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Failed to create customer",
		},
		{
			name:       "create customer returns 400 on invalid json",
			body:       "invalid json",
			shopID:     1,
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create customer returns 400 on validation failure - missing name",
			body: map[string]interface{}{
				"name":    "",
				"phone":   "08123456789",
				"address": "123 Main St",
			},
			shopID:     1,
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create customer returns 400 on validation failure - missing phone",
			body: map[string]interface{}{
				"name":    "John",
				"phone":   "",
				"address": "123 Main St",
			},
			shopID:     1,
			mockSetup:  func() {},
			wantStatus: http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "create customer returns 400 on validation failure - missing address",
			body: map[string]interface{}{
				"name":    "John",
				"phone":   "08123456789",
				"address": "",
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

			req := newRequestWithShopID("POST", "/customer", bodyBytes, tt.shopID)
			rec := httptest.NewRecorder()

			handler.CreateCustomerHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("CreateCustomerHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("CreateCustomerHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("CreateCustomerHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestGetCustomerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetCustomerService()
	defer handler.SetCustomerService(oldService)

	mockCustomerService := mock_service.NewMockCustomerService(ctrl)
	handler.SetCustomerService(mockCustomerService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		customerID     string
		shopID         int
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:       "successfully get customer by ID",
			customerID: "1",
			shopID:     1,
			mockSetup: func() {
				mockCustomerService.EXPECT().
					GetCustomerByID(1, 1).
					Return(&response.CustomerData{
						ID:        1,
						Name:      "John Doe",
						Phone:     "08123456789",
						Address:   "123 Main St",
						CreatedAt: fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:       "get customer returns 400 when customer_id missing",
			customerID: "",
			shopID:     1,
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:       "get customer returns 404 when not found",
			customerID: "999",
			shopID:     1,
			mockSetup: func() {
				mockCustomerService.EXPECT().
					GetCustomerByID(999, 1).
					Return(nil, nil)
			},
			wantStatus:  http.StatusNotFound,
			wantSuccess: false,
		},
		{
			name:       "get customer returns 500 on service error",
			customerID: "1",
			shopID:     1,
			mockSetup: func() {
				mockCustomerService.EXPECT().
					GetCustomerByID(1, 1).
					Return(nil, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Failed to get customer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithShopID("GET", "/customers/"+tt.customerID, nil, tt.shopID)
			if tt.customerID != "" {
				req = newRequestWithPathVars(req, map[string]string{"customer_id": tt.customerID})
			}
			rec := httptest.NewRecorder()

			handler.GetCustomerHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetCustomerHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetCustomerHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetCustomerHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestGetCustomersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetCustomerService()
	defer handler.SetCustomerService(oldService)

	mockCustomerService := mock_service.NewMockCustomerService(ctrl)
	handler.SetCustomerService(mockCustomerService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		shopID      int
		searchQuery string
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
		wantCount   int
	}{
		{
			name:   "successfully get customers list",
			shopID: 1,
			mockSetup: func() {
				mockCustomerService.EXPECT().
					GetCustomersByShopID(1, nil).
					Return([]response.CustomerData{
						{ID: 1, Name: "John Doe", Phone: "08123456789", Address: "123 Main St", CreatedAt: fixedTime},
						{ID: 2, Name: "Jane Doe", Phone: "08987654321", Address: "456 Oak Ave", CreatedAt: fixedTime},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   2,
		},
		{
			name:   "get customers returns 500 on service error",
			shopID: 1,
			mockSetup: func() {
				mockCustomerService.EXPECT().
					GetCustomersByShopID(1, nil).
					Return(nil, errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
		{
			name:        "get customers with search query passes search to service",
			shopID:      1,
			searchQuery: "john",
			mockSetup: func() {
				mockCustomerService.EXPECT().
					GetCustomersByShopID(1, gomock.Any()).
					Return([]response.CustomerData{
						{ID: 1, Name: "John Doe", Phone: "08123456789", Address: "123 Main St", CreatedAt: fixedTime},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			path := "/customers"
			if tt.searchQuery != "" {
				path += "?search=" + tt.searchQuery
			}
			req := newRequestWithShopID("GET", path, nil, tt.shopID)
			rec := httptest.NewRecorder()

			handler.GetCustomersHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetCustomersHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetCustomersHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantCount > 0 {
				customers, ok := resp.Data.([]interface{})
				if !ok || len(customers) != tt.wantCount {
					t.Errorf("GetCustomersHandler() data count = %v, want %v", len(customers), tt.wantCount)
				}
			}
		})
	}
}

func TestUpdateCustomerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetCustomerService()
	defer handler.SetCustomerService(oldService)

	mockCustomerService := mock_service.NewMockCustomerService(ctrl)
	handler.SetCustomerService(mockCustomerService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		customerID     string
		body           interface{}
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:       "successfully update customer",
			customerID: "1",
			body: map[string]interface{}{
				"name":    "Updated Name",
				"phone":   "08111111111",
				"address": "Updated Address",
			},
			mockSetup: func() {
				name := "Updated Name"
				phone := "08111111111"
				address := "Updated Address"
				mockCustomerService.EXPECT().
					UpdateCustomer(service.UpdateCustomerInput{
						ID:      1,
						Name:    &name,
						Phone:   &phone,
						Address: &address,
					}).
					Return(response.CustomerData{
						ID:        1,
						Name:      "Updated Name",
						Phone:     "08111111111",
						Address:   "Updated Address",
						CreatedAt: fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:       "update customer returns 400 when customer_id missing",
			customerID: "",
			body:       map[string]interface{}{"name": "Test"},
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:       "update customer returns 400 on invalid json",
			customerID: "1",
			body:       "invalid json",
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:       "update customer returns 500 on service error",
			customerID: "1",
			body: map[string]interface{}{
				"name": "Test",
			},
			mockSetup: func() {
				mockCustomerService.EXPECT().
					UpdateCustomer(gomock.Any()).
					Return(response.CustomerData{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Failed to update customer",
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

			req := httptest.NewRequest("PATCH", "/customers/"+tt.customerID, bytes.NewReader(bodyBytes))
			req = req.WithContext(context.WithValue(req.Context(), common.ShopIDKey, 1))
			if tt.customerID != "" {
				req = newRequestWithPathVars(req, map[string]string{"customer_id": tt.customerID})
			}
			rec := httptest.NewRecorder()

			handler.UpdateCustomerHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("UpdateCustomerHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("UpdateCustomerHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("UpdateCustomerHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestDeleteCustomerHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetCustomerService()
	defer handler.SetCustomerService(oldService)

	mockCustomerService := mock_service.NewMockCustomerService(ctrl)
	handler.SetCustomerService(mockCustomerService)

	tests := []struct {
		name           string
		customerID     string
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:       "successfully delete customer",
			customerID: "1",
			mockSetup: func() {
				mockCustomerService.EXPECT().
					DeleteCustomerByID(1).
					Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:       "delete customer returns 400 when customer_id missing",
			customerID: "",
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:       "delete customer returns 500 on service error",
			customerID: "1",
			mockSetup: func() {
				mockCustomerService.EXPECT().
					DeleteCustomerByID(1).
					Return(errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Failed to delete customer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("DELETE", "/customers/"+tt.customerID, nil)
			if tt.customerID != "" {
				req = newRequestWithPathVars(req, map[string]string{"customer_id": tt.customerID})
			}
			rec := httptest.NewRecorder()

			handler.DeleteCustomerHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("DeleteCustomerHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("DeleteCustomerHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("DeleteCustomerHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}
