package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
	"github.com/zeirash/recapo/arion/model"
)

func TestGetSystemStatsHandler(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(m *mock_service.MockSystemService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "returns stats successfully",
			mockSetup: func(m *mock_service.MockSystemService) {
				m.EXPECT().GetSystemStats(gomock.Any()).Return(&response.SystemStatsData{
					TotalShops:   5,
					SubsActive:   2,
					SubsTrialing: 3,
					MRRIDR:       298000,
				}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "returns 500 on service error",
			mockSetup: func(m *mock_service.MockSystemService) {
				m.EXPECT().GetSystemStats(gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldSvc := handler.GetSystemService()
			defer handler.SetSystemService(oldSvc)

			mockSvc := mock_service.NewMockSystemService(ctrl)
			handler.SetSystemService(mockSvc)
			tt.mockSetup(mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/system/stats", nil)
			w := httptest.NewRecorder()
			handler.GetSystemStatsHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetSystemShopsHandler(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		mockSetup   func(m *mock_service.MockSystemService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "returns shops list successfully",
			mockSetup: func(m *mock_service.MockSystemService) {
				m.EXPECT().GetSystemShops(gomock.Any()).Return([]response.SystemShopData{
					{ShopID: 1, ShopName: "Toko Mawar", OwnerName: "Siti", OwnerEmail: "siti@email.com", PlanName: "Starter", SubStatus: "trialing", JoinedAt: fixedTime},
				}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "returns empty list",
			mockSetup: func(m *mock_service.MockSystemService) {
				m.EXPECT().GetSystemShops(gomock.Any()).Return([]response.SystemShopData{}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "returns 500 on service error",
			mockSetup: func(m *mock_service.MockSystemService) {
				m.EXPECT().GetSystemShops(gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldSvc := handler.GetSystemService()
			defer handler.SetSystemService(oldSvc)

			mockSvc := mock_service.NewMockSystemService(ctrl)
			handler.SetSystemService(mockSvc)
			tt.mockSetup(mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/system/shops", nil)
			w := httptest.NewRecorder()
			handler.GetSystemShopsHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetSystemPaymentsHandler(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	statusSettlement := "settlement"
	sortOpt := "amount_idr,desc"

	tests := []struct {
		name        string
		query       string
		mockSetup   func(m *mock_service.MockSystemService)
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:  "returns payments list successfully",
			query: "",
			mockSetup: func(m *mock_service.MockSystemService) {
				m.EXPECT().GetSystemPayments(gomock.Any(), gomock.Any()).Return([]response.SystemPaymentData{
					{ShopName: "Toko Mawar", PlanName: "Starter", AmountIDR: 149000, Status: "settlement", MidtransOrderID: "recapo-1-001", PaidAt: &fixedTime, CreatedAt: fixedTime},
				}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:  "returns empty list",
			query: "",
			mockSetup: func(m *mock_service.MockSystemService) {
				m.EXPECT().GetSystemPayments(gomock.Any(), gomock.Any()).Return([]response.SystemPaymentData{}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:  "returns 500 on service error",
			query: "",
			mockSetup: func(m *mock_service.MockSystemService) {
				m.EXPECT().GetSystemPayments(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
		{
			name:  "passes parsed filters to service",
			query: "?date_from=2024-01-01&date_to=2024-01-31&status=settlement&sort=amount_idr,desc",
			mockSetup: func(m *mock_service.MockSystemService) {
				expected := model.SystemPaymentFilterOptions{
					DateFrom: &dateFrom,
					DateTo:   &dateTo,
					Status:   &statusSettlement,
					Sort:     &sortOpt,
				}
				m.EXPECT().GetSystemPayments(gomock.Any(), gomock.Eq(expected)).Return([]response.SystemPaymentData{
					{ShopName: "Toko Mawar", PlanName: "Starter", AmountIDR: 149000, Status: "settlement", MidtransOrderID: "recapo-1-001", PaidAt: &fixedTime, CreatedAt: fixedTime},
				}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldSvc := handler.GetSystemService()
			defer handler.SetSystemService(oldSvc)

			mockSvc := mock_service.NewMockSystemService(ctrl)
			handler.SetSystemService(mockSvc)
			tt.mockSetup(mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/system/payments"+tt.query, nil)
			w := httptest.NewRecorder()
			handler.GetSystemPaymentsHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
