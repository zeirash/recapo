package store

import (
	"database/sql"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	SubscriptionStore interface {
		GetActivePlans() ([]model.Plan, error)
		GetPlanByID(planID int) (*model.Plan, error)
		GetSubscriptionByShopID(shopID int) (*model.Subscription, error)
		CreateTrialSubscription(tx database.Tx, shopID, planID int, trialEndsAt time.Time) (*model.Subscription, error)
		UpdateSubscriptionStatus(tx database.Tx, subID int, status string, periodEnd *time.Time) error
		CreatePayment(tx database.Tx, shopID, subscriptionID, planID int, midtransOrderID string, amountIDR int) (*model.Payment, error)
		GetPaymentByMidtransOrderID(orderID string) (*model.Payment, error)
		UpdatePaymentSettled(tx database.Tx, paymentID int, midtransTxnID string, paidAt time.Time) error
		UpdatePaymentFailed(tx database.Tx, paymentID int, status string) error
		UpdatePaymentSnapInfo(paymentID int, snapToken, redirectURL string) error
		CancelSubscription(tx database.Tx, subID int) error
		ExpireSubscriptions() (int64, error)
	}

	subscriptionStore struct {
		db *sql.DB
	}
)

func NewSubscriptionStore() SubscriptionStore {
	return &subscriptionStore{db: database.GetDB()}
}

func NewSubscriptionStoreWithDB(db *sql.DB) SubscriptionStore {
	return &subscriptionStore{db: db}
}

func (s *subscriptionStore) GetActivePlans() ([]model.Plan, error) {
	q := `
		SELECT id, name, display_name, description_en, description_id, price_idr, max_users, is_active, created_at, updated_at
		FROM plans
		WHERE is_active = TRUE
		ORDER BY price_idr ASC
	`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []model.Plan
	for rows.Next() {
		var p model.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.DisplayName, &p.DescriptionEN, &p.DescriptionID, &p.PriceIDR, &p.MaxUsers, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (s *subscriptionStore) GetPlanByID(planID int) (*model.Plan, error) {
	q := `
		SELECT id, name, display_name, description_en, description_id, price_idr, max_users, is_active, created_at, updated_at
		FROM plans
		WHERE id = $1
	`
	var p model.Plan
	err := s.db.QueryRow(q, planID).Scan(&p.ID, &p.Name, &p.DisplayName, &p.DescriptionEN, &p.DescriptionID, &p.PriceIDR, &p.MaxUsers, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *subscriptionStore) GetSubscriptionByShopID(shopID int) (*model.Subscription, error) {
	q := `
		SELECT id, shop_id, plan_id, status, trial_ends_at, current_period_start, current_period_end, cancelled_at, created_at, updated_at
		FROM subscriptions
		WHERE shop_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var sub model.Subscription
	err := s.db.QueryRow(q, shopID).Scan(
		&sub.ID, &sub.ShopID, &sub.PlanID, &sub.Status,
		&sub.TrialEndsAt, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd,
		&sub.CancelledAt, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (s *subscriptionStore) CreateTrialSubscription(tx database.Tx, shopID, planID int, trialEndsAt time.Time) (*model.Subscription, error) {
	now := time.Now()
	q := `
		INSERT INTO subscriptions (shop_id, plan_id, status, trial_ends_at, current_period_start, current_period_end, created_at)
		VALUES ($1, $2, 'trialing', $3, $4, $5, $6)
		RETURNING id
	`
	var id int
	err := tx.QueryRow(q, shopID, planID, trialEndsAt, now, trialEndsAt, now).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &model.Subscription{
		ID:                 id,
		ShopID:             shopID,
		PlanID:             planID,
		Status:             "trialing",
		TrialEndsAt:        sql.NullTime{Time: trialEndsAt, Valid: true},
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   trialEndsAt,
		CreatedAt:          now,
	}, nil
}

func (s *subscriptionStore) UpdateSubscriptionStatus(tx database.Tx, subID int, status string, periodEnd *time.Time) error {
	now := time.Now()
	if periodEnd != nil {
		q := `UPDATE subscriptions SET status = $1, current_period_start = $2, current_period_end = $3, updated_at = $2 WHERE id = $4`
		_, err := tx.Exec(q, status, now, *periodEnd, subID)
		return err
	}
	q := `UPDATE subscriptions SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := tx.Exec(q, status, now, subID)
	return err
}

func (s *subscriptionStore) CreatePayment(tx database.Tx, shopID, subscriptionID, planID int, midtransOrderID string, amountIDR int) (*model.Payment, error) {
	now := time.Now()
	q := `
		INSERT INTO payments (shop_id, subscription_id, plan_id, midtrans_order_id, amount_idr, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6)
		RETURNING id
	`
	var id int
	err := tx.QueryRow(q, shopID, subscriptionID, planID, midtransOrderID, amountIDR, now).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &model.Payment{
		ID:              id,
		ShopID:          shopID,
		SubscriptionID:  subscriptionID,
		PlanID:          planID,
		MidtransOrderID: midtransOrderID,
		AmountIDR:       amountIDR,
		Status:          "pending",
		CreatedAt:       now,
	}, nil
}

func (s *subscriptionStore) GetPaymentByMidtransOrderID(orderID string) (*model.Payment, error) {
	q := `
		SELECT id, shop_id, subscription_id, plan_id, midtrans_order_id, COALESCE(midtrans_txn_id, ''),
		       amount_idr, status, COALESCE(snap_token, ''), COALESCE(redirect_url, ''), paid_at, created_at, updated_at
		FROM payments
		WHERE midtrans_order_id = $1
	`
	var p model.Payment
	err := s.db.QueryRow(q, orderID).Scan(
		&p.ID, &p.ShopID, &p.SubscriptionID, &p.PlanID, &p.MidtransOrderID, &p.MidtransTxnID,
		&p.AmountIDR, &p.Status, &p.SnapToken, &p.RedirectURL, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *subscriptionStore) UpdatePaymentSettled(tx database.Tx, paymentID int, midtransTxnID string, paidAt time.Time) error {
	now := time.Now()
	q := `UPDATE payments SET status = 'settlement', midtrans_txn_id = $1, paid_at = $2, updated_at = $3 WHERE id = $4`
	_, err := tx.Exec(q, midtransTxnID, paidAt, now, paymentID)
	return err
}

func (s *subscriptionStore) UpdatePaymentFailed(tx database.Tx, paymentID int, status string) error {
	now := time.Now()
	q := `UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := tx.Exec(q, status, now, paymentID)
	return err
}

func (s *subscriptionStore) ExpireSubscriptions() (int64, error) {
	now := time.Now()
	q := `
		UPDATE subscriptions
		SET status = 'expired', updated_at = $1
		WHERE status = 'active' AND current_period_end < $1
	`
	res, err := s.db.Exec(q, now)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *subscriptionStore) CancelSubscription(tx database.Tx, subID int) error {
	now := time.Now()
	q := `UPDATE subscriptions SET status = 'cancelled', cancelled_at = $1, updated_at = $1 WHERE id = $2`
	_, err := tx.Exec(q, now, subID)
	return err
}

func (s *subscriptionStore) UpdatePaymentSnapInfo(paymentID int, snapToken, redirectURL string) error {
	now := time.Now()
	q := `UPDATE payments SET snap_token = $1, redirect_url = $2, updated_at = $3 WHERE id = $4`
	_, err := s.db.Exec(q, snapToken, redirectURL, now, paymentID)
	return err
}
