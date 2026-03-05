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
)

func newSubscriptionRequest(method, path string, shopID int, body interface{}) *http.Request {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if shopID > 0 {
		ctx := context.WithValue(req.Context(), common.ShopIDKey, shopID)
		req = req.WithContext(ctx)
	}
	return req
}

func TestGetPlansHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldSvc := handler.GetSubscriptionService()
	defer handler.SetSubscriptionService(oldSvc)

	mockSvc := mock_service.NewMockSubscriptionService(ctrl)
	handler.SetSubscriptionService(mockSvc)

	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	_ = fixedTime

	tests := []struct {
		name        string
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "returns active plans",
			mockSetup: func() {
				mockSvc.EXPECT().GetActivePlans().Return([]response.PlanData{
					{ID: 1, Name: "starter", DisplayName: "Starter", PriceIDR: 149000, MaxUsers: 1},
				}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "returns 500 on service error",
			mockSetup: func() {
				mockSvc.EXPECT().GetActivePlans().Return(nil, errors.New("db error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/plans", nil)
			w := httptest.NewRecorder()
			handler.GetPlansHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetPlansHandler() status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if success, _ := resp["success"].(bool); success != tt.wantSuccess {
				t.Errorf("GetPlansHandler() success = %v, want %v", success, tt.wantSuccess)
			}
		})
	}
}

func TestGetSubscriptionHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldSvc := handler.GetSubscriptionService()
	defer handler.SetSubscriptionService(oldSvc)

	mockSvc := mock_service.NewMockSubscriptionService(ctrl)
	handler.SetSubscriptionService(mockSvc)

	now := time.Now()

	tests := []struct {
		name        string
		shopID      int
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:   "returns subscription",
			shopID: 1,
			mockSetup: func() {
				mockSvc.EXPECT().GetSubscriptionByShopID(1).Return(&response.SubscriptionData{
					ID:                 1,
					Status:             "trialing",
					Plan:               response.PlanData{ID: 1, Name: "starter"},
					CurrentPeriodStart: now,
					CurrentPeriodEnd:   now.AddDate(0, 0, 14),
				}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "returns 404 when subscription not found",
			shopID: 2,
			mockSetup: func() {
				mockSvc.EXPECT().GetSubscriptionByShopID(2).Return(nil, errors.New("subscription not found"))
			},
			wantStatus:  http.StatusNotFound,
			wantSuccess: false,
		},
		{
			name:   "returns 500 on service error",
			shopID: 3,
			mockSetup: func() {
				mockSvc.EXPECT().GetSubscriptionByShopID(3).Return(nil, errors.New("unexpected error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := newSubscriptionRequest(http.MethodGet, "/subscription", tt.shopID, nil)
			w := httptest.NewRecorder()
			handler.GetSubscriptionHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetSubscriptionHandler() status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if success, _ := resp["success"].(bool); success != tt.wantSuccess {
				t.Errorf("GetSubscriptionHandler() success = %v, want %v", success, tt.wantSuccess)
			}
		})
	}
}

func TestCheckoutHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldSvc := handler.GetSubscriptionService()
	defer handler.SetSubscriptionService(oldSvc)

	mockSvc := mock_service.NewMockSubscriptionService(ctrl)
	handler.SetSubscriptionService(mockSvc)

	tests := []struct {
		name        string
		shopID      int
		body        interface{}
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:   "checkout success",
			shopID: 1,
			body:   map[string]int{"plan_id": 1},
			mockSetup: func() {
				mockSvc.EXPECT().Checkout(1, 1).Return(&response.CheckoutData{
					OrderID:     "recapo-1-12345",
					RedirectURL: "https://app.sandbox.midtrans.com/snap/v2/vtweb/token",
					SnapToken:   "some-snap-token",
				}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 400 when plan_id missing",
			shopID:      1,
			body:        map[string]int{"plan_id": 0},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:   "returns 404 when plan not found",
			shopID: 1,
			body:   map[string]int{"plan_id": 99},
			mockSetup: func() {
				mockSvc.EXPECT().Checkout(1, 99).Return(nil, errors.New("plan not found"))
			},
			wantStatus:  http.StatusNotFound,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := newSubscriptionRequest(http.MethodPost, "/subscription/checkout", tt.shopID, tt.body)
			w := httptest.NewRecorder()
			handler.CheckoutHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CheckoutHandler() status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if success, _ := resp["success"].(bool); success != tt.wantSuccess {
				t.Errorf("CheckoutHandler() success = %v, want %v", success, tt.wantSuccess)
			}
		})
	}
}

func TestMidtransWebhookHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldSvc := handler.GetSubscriptionService()
	defer handler.SetSubscriptionService(oldSvc)

	mockSvc := mock_service.NewMockSubscriptionService(ctrl)
	handler.SetSubscriptionService(mockSvc)

	tests := []struct {
		name        string
		body        interface{}
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "processes webhook successfully",
			body: map[string]string{
				"order_id":           "recapo-1-12345",
				"status_code":        "200",
				"gross_amount":       "149000.00",
				"signature_key":      "somekey",
				"transaction_status": "settlement",
				"transaction_id":     "txn-123",
			},
			mockSetup: func() {
				mockSvc.EXPECT().HandleMidtransWebhook(gomock.Any()).Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "returns 400 on invalid signature",
			body: map[string]string{
				"order_id":      "recapo-1-12345",
				"signature_key": "bad",
			},
			mockSetup: func() {
				mockSvc.EXPECT().HandleMidtransWebhook(gomock.Any()).Return(errors.New("invalid signature"))
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := newSubscriptionRequest(http.MethodPost, "/webhook/midtrans", 0, tt.body)
			w := httptest.NewRecorder()
			handler.MidtransWebhookHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("MidtransWebhookHandler() status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if success, _ := resp["success"].(bool); success != tt.wantSuccess {
				t.Errorf("MidtransWebhookHandler() success = %v, want %v", success, tt.wantSuccess)
			}
		})
	}
}
