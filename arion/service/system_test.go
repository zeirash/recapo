package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/store"
)

func Test_sysservice_GetSystemStats(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockSystemStore
		wantErr   bool
		wantMRR   int
	}{
		{
			name: "returns stats successfully",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSystemStore {
				m := mock_store.NewMockSystemStore(ctrl)
				m.EXPECT().GetSystemStats(gomock.Any()).Return(&store.SystemStats{
					TotalShops:    5,
					SubsActive:    2,
					SubsTrialing:  3,
					SubsExpired:   1,
					SubsCancelled: 1,
					MRRIDR:        298000,
				}, nil)
				return m
			},
			wantErr: false,
			wantMRR: 298000,
		},
		{
			name: "returns error on store failure",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSystemStore {
				m := mock_store.NewMockSystemStore(ctrl)
				m.EXPECT().GetSystemStats(gomock.Any()).Return(nil, errors.New("db error"))
				return m
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			systemStore = tt.mockSetup(ctrl)
			svc := &sysservice{}
			got, err := svc.GetSystemStats(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSystemStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.MRRIDR != tt.wantMRR {
				t.Errorf("GetSystemStats() MRRIDR = %v, want %v", got.MRRIDR, tt.wantMRR)
			}
		})
	}
}

func Test_sysservice_GetSystemShops(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	trialEnd := fixedTime.AddDate(0, 0, 14)

	tests := []struct {
		name      string
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockSystemStore
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns shops list",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSystemStore {
				m := mock_store.NewMockSystemStore(ctrl)
				m.EXPECT().GetSystemShops(gomock.Any()).Return([]store.SystemShop{
					{ShopID: 1, ShopName: "Toko Mawar", OwnerName: "Siti", OwnerEmail: "siti@email.com", PlanName: "Starter", SubStatus: "trialing", TrialEndsAt: &trialEnd, PeriodEnd: trialEnd, JoinedAt: fixedTime},
					{ShopID: 2, ShopName: "Toko Melati", OwnerName: "Budi", OwnerEmail: "budi@email.com", PlanName: "Starter", SubStatus: "active", JoinedAt: fixedTime},
				}, nil)
				return m
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "returns empty list",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSystemStore {
				m := mock_store.NewMockSystemStore(ctrl)
				m.EXPECT().GetSystemShops(gomock.Any()).Return([]store.SystemShop{}, nil)
				return m
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "returns error on store failure",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSystemStore {
				m := mock_store.NewMockSystemStore(ctrl)
				m.EXPECT().GetSystemShops(gomock.Any()).Return(nil, errors.New("db error"))
				return m
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			systemStore = tt.mockSetup(ctrl)
			svc := &sysservice{}
			got, err := svc.GetSystemShops(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSystemShops() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("GetSystemShops() returned %d shops, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func Test_sysservice_GetSystemPayments(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockSystemStore
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns payments list",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSystemStore {
				m := mock_store.NewMockSystemStore(ctrl)
				m.EXPECT().GetSystemPayments(gomock.Any(), gomock.Any()).Return([]store.SystemPayment{
					{ShopName: "Toko Mawar", PlanName: "Starter", AmountIDR: 149000, Status: "settlement", MidtransOrderID: "recapo-1-001", PaidAt: &fixedTime, CreatedAt: fixedTime},
					{ShopName: "Toko Melati", PlanName: "Starter", AmountIDR: 149000, Status: "pending", MidtransOrderID: "recapo-2-002", CreatedAt: fixedTime},
				}, nil)
				return m
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "returns empty list",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSystemStore {
				m := mock_store.NewMockSystemStore(ctrl)
				m.EXPECT().GetSystemPayments(gomock.Any(), gomock.Any()).Return([]store.SystemPayment{}, nil)
				return m
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "returns error on store failure",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSystemStore {
				m := mock_store.NewMockSystemStore(ctrl)
				m.EXPECT().GetSystemPayments(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
				return m
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			systemStore = tt.mockSetup(ctrl)
			svc := &sysservice{}
			got, err := svc.GetSystemPayments(context.Background(), model.SystemPaymentFilterOptions{})

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSystemPayments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("GetSystemPayments() returned %d payments, want %d", len(got), tt.wantLen)
			}
		})
	}
}
