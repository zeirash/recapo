package service

import (
	"context"

	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	SystemService interface {
		GetSystemStats(ctx context.Context) (*response.SystemStatsData, error)
		GetSystemShops(ctx context.Context) ([]response.SystemShopData, error)
		GetSystemPayments(ctx context.Context) ([]response.SystemPaymentData, error)
	}

	sysservice struct{}
)

func NewSystemService() SystemService {
	if systemStore == nil {
		systemStore = store.NewSystemStore()
	}
	return &sysservice{}
}

func (s *sysservice) GetSystemStats(ctx context.Context) (*response.SystemStatsData, error) {
	stats, err := systemStore.GetSystemStats(ctx)
	if err != nil {
		return nil, err
	}
	return &response.SystemStatsData{
		TotalShops:    stats.TotalShops,
		SubsTrialing:  stats.SubsTrialing,
		SubsActive:    stats.SubsActive,
		SubsExpired:   stats.SubsExpired,
		SubsCancelled: stats.SubsCancelled,
		MRRIDR:        stats.MRRIDR,
	}, nil
}

func (s *sysservice) GetSystemShops(ctx context.Context) ([]response.SystemShopData, error) {
	shops, err := systemStore.GetSystemShops(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]response.SystemShopData, 0, len(shops))
	for _, sh := range shops {
		results = append(results, response.SystemShopData{
			ShopID:      sh.ShopID,
			ShopName:    sh.ShopName,
			OwnerName:   sh.OwnerName,
			OwnerEmail:  sh.OwnerEmail,
			PlanName:    sh.PlanName,
			SubStatus:   sh.SubStatus,
			TrialEndsAt: sh.TrialEndsAt,
			PeriodEnd:   sh.PeriodEnd,
			JoinedAt:    sh.JoinedAt,
		})
	}
	return results, nil
}

func (s *sysservice) GetSystemPayments(ctx context.Context) ([]response.SystemPaymentData, error) {
	payments, err := systemStore.GetSystemPayments(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]response.SystemPaymentData, 0, len(payments))
	for _, p := range payments {
		results = append(results, response.SystemPaymentData{
			ShopName:        p.ShopName,
			PlanName:        p.PlanName,
			AmountIDR:       p.AmountIDR,
			Status:          p.Status,
			MidtransOrderID: p.MidtransOrderID,
			PaidAt:          p.PaidAt,
			CreatedAt:       p.CreatedAt,
		})
	}
	return results, nil
}
