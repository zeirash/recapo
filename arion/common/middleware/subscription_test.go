package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/middleware"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
)

func TestSubscriptionCheck(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		shopID         int
		mockSetup      func(m *mock_service.MockSubscriptionService)
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "missing shop ID in context returns 401",
			shopID:         0,
			mockSetup:      func(m *mock_service.MockSubscriptionService) {},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:   "active subscription passes through",
			shopID: 1,
			mockSetup: func(m *mock_service.MockSubscriptionService) {
				m.EXPECT().IsSubscriptionActive(gomock.Any(), 1).Return(true, nil)
			},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:   "inactive subscription returns 402",
			shopID: 2,
			mockSetup: func(m *mock_service.MockSubscriptionService) {
				m.EXPECT().IsSubscriptionActive(gomock.Any(), 2).Return(false, nil)
			},
			wantStatus:     http.StatusPaymentRequired,
			wantNextCalled: false,
		},
		{
			name:   "service error returns 500",
			shopID: 3,
			mockSetup: func(m *mock_service.MockSubscriptionService) {
				m.EXPECT().IsSubscriptionActive(gomock.Any(), 3).Return(false, errors.New("db error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			nextCalled = false
			mockSvc := mock_service.NewMockSubscriptionService(ctrl)
			tt.mockSetup(mockSvc)

			oldSvc := handler.GetSubscriptionService()
			handler.SetSubscriptionService(mockSvc)
			defer handler.SetSubscriptionService(oldSvc)

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := context.WithValue(r.Context(), common.ShopIDKey, tt.shopID)
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			middleware.SubscriptionCheck(next).ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if nextCalled != tt.wantNextCalled {
				t.Errorf("next called = %v, want %v", nextCalled, tt.wantNextCalled)
			}
			if tt.wantStatus != http.StatusOK {
				var body map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
					t.Fatalf("response body is not valid JSON: %v", err)
				}
				if body["success"] != false {
					t.Errorf("expected success=false in error response")
				}
			}
		})
	}
}
