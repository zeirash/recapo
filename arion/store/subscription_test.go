package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeirash/recapo/arion/model"
)

func Test_subscriptionStore_GetActivePlans(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		mockSetup func(mock sqlmock.Sqlmock)
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns active plans",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "display_name", "description_en", "description_id", "price_idr", "max_users", "is_active", "created_at", "updated_at"}).
					AddRow(1, "starter", "Starter", "Full access for 1 user.", "Akses penuh untuk 1 pengguna.", 149000, 1, true, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, name, display_name, description_en, description_id, price_idr, max_users, is_active, created_at, updated_at\s+FROM plans\s+WHERE is_active = TRUE`).
					WillReturnRows(rows)
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "returns error on database failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, display_name, description_en, description_id, price_idr, max_users, is_active, created_at, updated_at\s+FROM plans`).
					WillReturnError(errors.New("database error"))
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			s := &subscriptionStore{db: db}
			got, gotErr := s.GetActivePlans(context.Background())

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetActivePlans() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetActivePlans() succeeded unexpectedly")
			}
			if len(got) != tt.wantLen {
				t.Errorf("GetActivePlans() returned %d plans, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func Test_subscriptionStore_GetSubscriptionByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	trialEnd := fixedTime.AddDate(0, 0, 14)

	tests := []struct {
		name      string
		shopID    int
		mockSetup func(mock sqlmock.Sqlmock)
		wantNil   bool
		wantErr   bool
	}{
		{
			name:   "returns subscription for shop",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "plan_id", "status", "trial_ends_at", "current_period_start", "current_period_end", "cancelled_at", "created_at", "updated_at"}).
					AddRow(1, 1, 1, "trialing", trialEnd, fixedTime, trialEnd, nil, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, plan_id, status, trial_ends_at, current_period_start, current_period_end, cancelled_at, created_at, updated_at\s+FROM subscriptions\s+WHERE shop_id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name:   "returns nil when not found",
			shopID: 999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, plan_id, status, trial_ends_at, current_period_start, current_period_end, cancelled_at, created_at, updated_at\s+FROM subscriptions\s+WHERE shop_id = \$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name:   "returns error on database failure",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, plan_id, status, trial_ends_at, current_period_start, current_period_end, cancelled_at, created_at, updated_at\s+FROM subscriptions\s+WHERE shop_id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			wantNil: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			s := &subscriptionStore{db: db}
			got, gotErr := s.GetSubscriptionByShopID(context.Background(), tt.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetSubscriptionByShopID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetSubscriptionByShopID() succeeded unexpectedly")
			}
			if tt.wantNil && got != nil {
				t.Errorf("GetSubscriptionByShopID() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("GetSubscriptionByShopID() returned nil, want non-nil")
			}
		})
	}
}

func Test_subscriptionStore_CreateTrialSubscription(t *testing.T) {
	trialEnd := time.Now().AddDate(0, 0, 14)

	tests := []struct {
		name      string
		shopID    int
		planID    int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:   "creates trial subscription successfully",
			shopID: 1,
			planID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`INSERT INTO subscriptions`).
					WithArgs(1, 1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:   "returns error on database failure",
			shopID: 1,
			planID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO subscriptions`).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			s := &subscriptionStore{db: db}
			_, gotErr := s.CreateTrialSubscription(context.Background(), tx, tt.shopID, tt.planID, trialEnd)

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateTrialSubscription() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_subscriptionStore_GetPlanByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		planID    int
		mockSetup func(mock sqlmock.Sqlmock)
		want      *model.Plan
		wantNil   bool
		wantErr   bool
	}{
		{
			name:   "returns plan by ID",
			planID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "display_name", "description_en", "description_id", "price_idr", "max_users", "is_active", "created_at", "updated_at"}).
					AddRow(1, "starter", "Starter", "Full access for 1 user.", "Akses penuh untuk 1 pengguna.", 149000, 1, true, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, name, display_name, description_en, description_id, price_idr, max_users, is_active, created_at, updated_at\s+FROM plans\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			want:    &model.Plan{ID: 1, Name: "starter", DisplayName: "Starter", DescriptionEN: "Full access for 1 user.", DescriptionID: "Akses penuh untuk 1 pengguna.", PriceIDR: 149000, MaxUsers: 1, IsActive: true, CreatedAt: fixedTime},
			wantNil: false,
			wantErr: false,
		},
		{
			name:   "returns nil when plan not found",
			planID: 999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, display_name, description_en, description_id, price_idr, max_users, is_active, created_at, updated_at\s+FROM plans\s+WHERE id = \$1`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			want:    nil,
			wantNil: true,
			wantErr: false,
		},
		{
			name:   "returns error on database failure",
			planID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, display_name, description_en, description_id, price_idr, max_users, is_active, created_at, updated_at\s+FROM plans\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			want:    nil,
			wantNil: false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			s := &subscriptionStore{db: db}
			got, gotErr := s.GetPlanByID(context.Background(), tt.planID)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetPlanByID() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetPlanByID() succeeded unexpectedly")
			}
			if tt.wantNil && got != nil {
				t.Errorf("GetPlanByID() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("GetPlanByID() returned nil, want non-nil")
			}
			if got != nil && tt.want != nil && got.ID != tt.want.ID {
				t.Errorf("GetPlanByID() ID = %v, want %v", got.ID, tt.want.ID)
			}
		})
	}
}

func Test_subscriptionStore_UpdateSubscriptionStatus(t *testing.T) {
	futureTime := time.Now().Add(30 * 24 * time.Hour)

	tests := []struct {
		name      string
		subID     int
		status    string
		periodEnd *time.Time
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "updates status with period end",
			subID:     1,
			status:    "active",
			periodEnd: &futureTime,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE subscriptions SET status = \$1, current_period_start = \$2, current_period_end = \$3, updated_at = \$2 WHERE id = \$4`).
					WithArgs("active", sqlmock.AnyArg(), futureTime, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:      "updates status without period end",
			subID:     1,
			status:    "cancelled",
			periodEnd: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE subscriptions SET status = \$1, updated_at = \$2 WHERE id = \$3`).
					WithArgs("cancelled", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:      "returns error on database failure",
			subID:     1,
			status:    "active",
			periodEnd: &futureTime,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE subscriptions SET status`).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			s := &subscriptionStore{db: db}
			gotErr := s.UpdateSubscriptionStatus(context.Background(), tx, tt.subID, tt.status, tt.periodEnd)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateSubscriptionStatus() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateSubscriptionStatus() succeeded unexpectedly")
			}
		})
	}
}

func Test_subscriptionStore_CreatePayment(t *testing.T) {
	tests := []struct {
		name            string
		shopID          int
		subscriptionID  int
		planID          int
		midtransOrderID string
		amountIDR       int
		mockSetup       func(mock sqlmock.Sqlmock)
		wantID          int
		wantErr         bool
	}{
		{
			name:            "creates payment successfully",
			shopID:          1,
			subscriptionID:  1,
			planID:          1,
			midtransOrderID: "recapo-1-12345",
			amountIDR:       149000,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(5)
				mock.ExpectQuery(`INSERT INTO payments`).
					WithArgs(1, 1, 1, "recapo-1-12345", 149000, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantID:  5,
			wantErr: false,
		},
		{
			name:            "returns error on database failure",
			shopID:          1,
			subscriptionID:  1,
			planID:          1,
			midtransOrderID: "recapo-1-12345",
			amountIDR:       149000,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO payments`).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			s := &subscriptionStore{db: db}
			got, gotErr := s.CreatePayment(context.Background(), tx, tt.shopID, tt.subscriptionID, tt.planID, tt.midtransOrderID, tt.amountIDR)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreatePayment() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreatePayment() succeeded unexpectedly")
			}
			if got.ID != tt.wantID {
				t.Errorf("CreatePayment() ID = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}

func Test_subscriptionStore_GetPaymentByMidtransOrderID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		orderID   string
		mockSetup func(mock sqlmock.Sqlmock)
		wantNil   bool
		wantID    int
		wantErr   bool
	}{
		{
			name:    "returns payment by order ID",
			orderID: "recapo-1-12345",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "subscription_id", "plan_id", "midtrans_order_id", "midtrans_txn_id", "amount_idr", "status", "snap_token", "redirect_url", "paid_at", "created_at", "updated_at"}).
					AddRow(5, 1, 1, 1, "recapo-1-12345", "", 149000, "pending", "", "", nil, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, subscription_id, plan_id, midtrans_order_id`).
					WithArgs("recapo-1-12345").
					WillReturnRows(rows)
			},
			wantNil: false,
			wantID:  5,
			wantErr: false,
		},
		{
			name:    "returns nil when not found",
			orderID: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, subscription_id, plan_id, midtrans_order_id`).
					WithArgs("nonexistent").
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "returns error on database failure",
			orderID: "recapo-1-12345",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, subscription_id, plan_id, midtrans_order_id`).
					WithArgs("recapo-1-12345").
					WillReturnError(errors.New("database error"))
			},
			wantNil: false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			s := &subscriptionStore{db: db}
			got, gotErr := s.GetPaymentByMidtransOrderID(context.Background(), tt.orderID)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetPaymentByMidtransOrderID() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetPaymentByMidtransOrderID() succeeded unexpectedly")
			}
			if tt.wantNil && got != nil {
				t.Errorf("GetPaymentByMidtransOrderID() = %v, want nil", got)
			}
			if !tt.wantNil && got == nil {
				t.Error("GetPaymentByMidtransOrderID() returned nil, want non-nil")
			}
			if got != nil && got.ID != tt.wantID {
				t.Errorf("GetPaymentByMidtransOrderID() ID = %v, want %v", got.ID, tt.wantID)
			}
		})
	}
}

func Test_subscriptionStore_UpdatePaymentSettled(t *testing.T) {
	paidAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		paymentID     int
		midtransTxnID string
		paidAt        time.Time
		mockSetup     func(mock sqlmock.Sqlmock)
		wantErr       bool
	}{
		{
			name:          "marks payment as settled",
			paymentID:     1,
			midtransTxnID: "txn-abc123",
			paidAt:        paidAt,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE payments SET status = 'settlement', midtrans_txn_id = \$1, paid_at = \$2, updated_at = \$3 WHERE id = \$4`).
					WithArgs("txn-abc123", paidAt, sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:          "returns error on database failure",
			paymentID:     1,
			midtransTxnID: "txn-abc123",
			paidAt:        paidAt,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE payments SET status = 'settlement'`).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			s := &subscriptionStore{db: db}
			gotErr := s.UpdatePaymentSettled(context.Background(), tx, tt.paymentID, tt.midtransTxnID, tt.paidAt)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdatePaymentSettled() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdatePaymentSettled() succeeded unexpectedly")
			}
		})
	}
}

func Test_subscriptionStore_UpdatePaymentFailed(t *testing.T) {
	tests := []struct {
		name      string
		paymentID int
		status    string
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "marks payment as failed",
			paymentID: 1,
			status:    "deny",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE payments SET status = \$1, updated_at = \$2 WHERE id = \$3`).
					WithArgs("deny", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:      "returns error on database failure",
			paymentID: 1,
			status:    "expire",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE payments SET status`).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			s := &subscriptionStore{db: db}
			gotErr := s.UpdatePaymentFailed(context.Background(), tx, tt.paymentID, tt.status)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdatePaymentFailed() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdatePaymentFailed() succeeded unexpectedly")
			}
		})
	}
}

func Test_subscriptionStore_UpdatePaymentSnapInfo(t *testing.T) {
	tests := []struct {
		name        string
		paymentID   int
		snapToken   string
		redirectURL string
		mockSetup   func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name:        "updates snap info successfully",
			paymentID:   1,
			snapToken:   "snap-token-abc",
			redirectURL: "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-token-abc",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE payments SET snap_token = \$1, redirect_url = \$2, updated_at = \$3 WHERE id = \$4`).
					WithArgs("snap-token-abc", "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-token-abc", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:        "returns error on database failure",
			paymentID:   1,
			snapToken:   "snap-token-abc",
			redirectURL: "https://app.sandbox.midtrans.com/snap/v2/vtweb/snap-token-abc",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE payments SET snap_token`).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			s := &subscriptionStore{db: db}
			gotErr := s.UpdatePaymentSnapInfo(context.Background(), tt.paymentID, tt.snapToken, tt.redirectURL)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdatePaymentSnapInfo() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdatePaymentSnapInfo() succeeded unexpectedly")
			}
		})
	}
}

func Test_subscriptionStore_ExpireSubscriptions(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mock sqlmock.Sqlmock)
		want      int64
		wantErr   bool
	}{
		{
			name: "expires active subscriptions past period end and returns count",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE subscriptions\s+SET status = 'expired', updated_at = \$1\s+WHERE status = 'active' AND current_period_end < \$1`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 3))
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "returns zero when no subscriptions to expire",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE subscriptions\s+SET status = 'expired'`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "returns error on database failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE subscriptions\s+SET status = 'expired'`).
					WithArgs(sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			s := &subscriptionStore{db: db}
			got, gotErr := s.ExpireSubscriptions(context.Background())
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ExpireSubscriptions() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ExpireSubscriptions() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("ExpireSubscriptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_subscriptionStore_CancelSubscription(t *testing.T) {
	tests := []struct {
		name      string
		subID     int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:  "cancels subscription successfully",
			subID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE subscriptions SET status = 'cancelled', cancelled_at = \$1, updated_at = \$1 WHERE id = \$2`).
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:  "returns error on database failure",
			subID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE subscriptions SET status = 'cancelled'`).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			s := &subscriptionStore{db: db}
			gotErr := s.CancelSubscription(context.Background(), tx, tt.subID)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CancelSubscription() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CancelSubscription() succeeded unexpectedly")
			}
		})
	}
}
