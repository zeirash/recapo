package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
)

type (
	SystemStats struct {
		TotalShops    int
		SubsTrialing  int
		SubsActive    int
		SubsExpired   int
		SubsCancelled int
		MRRIDR        int
	}

	SystemShop struct {
		ShopID        int
		ShopName      string
		OwnerName     string
		OwnerEmail    string
		PlanName      string
		SubStatus     string
		TrialEndsAt   *time.Time
		PeriodEnd     time.Time
		JoinedAt      time.Time
	}

	SystemPayment struct {
		ShopName        string
		PlanName        string
		AmountIDR       int
		Status          string
		MidtransOrderID string
		PaidAt          *time.Time
		CreatedAt       time.Time
	}

	SystemStore interface {
		GetSystemStats(ctx context.Context) (*SystemStats, error)
		GetSystemShops(ctx context.Context) ([]SystemShop, error)
		GetSystemPayments(ctx context.Context) ([]SystemPayment, error)
	}

	systemStore struct {
		db *sql.DB
	}
)

func NewSystemStore() SystemStore {
	return &systemStore{db: database.GetDB()}
}

func NewSystemStoreWithDB(db *sql.DB) SystemStore {
	return &systemStore{db: db}
}

func (s *systemStore) GetSystemStats(ctx context.Context) (*SystemStats, error) {
	stats := &SystemStats{}

	// Total shops and users
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM shops
		WHERE id NOT IN (SELECT shop_id FROM users WHERE role = 'system')
	`).Scan(&stats.TotalShops)
	if err != nil {
		return nil, err
	}

	// Subscription counts by status
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.status, COUNT(*) as cnt
		FROM subscriptions s
		INNER JOIN (
			SELECT shop_id, MAX(created_at) AS max_created_at
			FROM subscriptions
			GROUP BY shop_id
		) latest ON s.shop_id = latest.shop_id AND s.created_at = latest.max_created_at
		WHERE s.shop_id NOT IN (SELECT shop_id FROM users WHERE role = 'system')
		GROUP BY s.status
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		switch status {
		case "trialing":
			stats.SubsTrialing = count
		case "active":
			stats.SubsActive = count
		case "expired":
			stats.SubsExpired = count
		case "cancelled":
			stats.SubsCancelled = count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// MRR: sum of plan prices for all active subscriptions
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(p.price_idr), 0)
		FROM subscriptions s
		INNER JOIN (
			SELECT shop_id, MAX(created_at) AS max_created_at
			FROM subscriptions
			GROUP BY shop_id
		) latest ON s.shop_id = latest.shop_id AND s.created_at = latest.max_created_at
		INNER JOIN plans p ON p.id = s.plan_id
		WHERE s.status = 'active'
		AND s.shop_id NOT IN (SELECT shop_id FROM users WHERE role = 'system')
	`).Scan(&stats.MRRIDR)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *systemStore) GetSystemShops(ctx context.Context) ([]SystemShop, error) {
	q := `
		SELECT
			sh.id,
			sh.name,
			u.name,
			u.email,
			p.display_name,
			sub.status,
			sub.trial_ends_at,
			sub.current_period_end,
			sh.created_at
		FROM shops sh
		INNER JOIN users u ON u.shop_id = sh.id AND u.role = 'owner'
		LEFT JOIN LATERAL (
			SELECT s.status, s.trial_ends_at, s.current_period_end, s.plan_id
			FROM subscriptions s
			WHERE s.shop_id = sh.id
			ORDER BY s.created_at DESC
			LIMIT 1
		) sub ON TRUE
		LEFT JOIN plans p ON p.id = sub.plan_id
		ORDER BY sh.created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shops []SystemShop
	for rows.Next() {
		var shop SystemShop
		var planName sql.NullString
		var subStatus sql.NullString
		var trialEndsAt sql.NullTime
		var periodEnd sql.NullTime

		if err := rows.Scan(
			&shop.ShopID, &shop.ShopName, &shop.OwnerName, &shop.OwnerEmail,
			&planName, &subStatus, &trialEndsAt, &periodEnd, &shop.JoinedAt,
		); err != nil {
			return nil, err
		}

		shop.PlanName = planName.String
		shop.SubStatus = subStatus.String
		if trialEndsAt.Valid {
			t := trialEndsAt.Time
			shop.TrialEndsAt = &t
		}
		if periodEnd.Valid {
			shop.PeriodEnd = periodEnd.Time
		}

		shops = append(shops, shop)
	}
	return shops, rows.Err()
}

func (s *systemStore) GetSystemPayments(ctx context.Context) ([]SystemPayment, error) {
	q := `
		SELECT
			sh.name,
			p.display_name,
			pay.amount_idr,
			pay.status,
			pay.midtrans_order_id,
			pay.paid_at,
			pay.created_at
		FROM payments pay
		INNER JOIN shops sh ON sh.id = pay.shop_id
		INNER JOIN plans p ON p.id = pay.plan_id
		WHERE pay.shop_id NOT IN (SELECT shop_id FROM users WHERE role = 'system')
		ORDER BY pay.created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []SystemPayment
	for rows.Next() {
		var pay SystemPayment
		var paidAt sql.NullTime

		if err := rows.Scan(
			&pay.ShopName, &pay.PlanName, &pay.AmountIDR,
			&pay.Status, &pay.MidtransOrderID, &paidAt, &pay.CreatedAt,
		); err != nil {
			return nil, err
		}

		if paidAt.Valid {
			t := paidAt.Time
			pay.PaidAt = &t
		}

		payments = append(payments, pay)
	}
	return payments, rows.Err()
}
