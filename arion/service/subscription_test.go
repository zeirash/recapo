package service

import (
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/database"
	mock_database "github.com/zeirash/recapo/arion/mock/database"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
)

func Test_subscriptionService_GetActivePlans(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns active plans",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetActivePlans().Return([]model.Plan{
					{ID: 1, Name: "starter", DisplayName: "Starter", PriceIDR: 149000, MaxUsers: 1, CreatedAt: fixedTime},
				}, nil)
				return m
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "returns error",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetActivePlans().Return(nil, errors.New("db error"))
				return m
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			subscriptionStore = tt.mockSetup(ctrl)
			svc := &ssubscription{}
			got, err := svc.GetActivePlans()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetActivePlans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("GetActivePlans() returned %d plans, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func Test_subscriptionService_IsSubscriptionActive(t *testing.T) {
	now := time.Now()
	futurePeriodEnd := now.Add(10 * 24 * time.Hour)
	pastPeriodEnd := now.Add(-10 * 24 * time.Hour)

	tests := []struct {
		name      string
		shopID    int
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore
		want      bool
		wantErr   bool
	}{
		{
			name:   "active subscription within period",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{
					ID:                 1,
					ShopID:             1,
					Status:             "active",
					CurrentPeriodStart: now,
					CurrentPeriodEnd:   futurePeriodEnd,
				}, nil)
				return m
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "active subscription past period end",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{
					ID:                 1,
					ShopID:             1,
					Status:             "active",
					CurrentPeriodStart: pastPeriodEnd,
					CurrentPeriodEnd:   pastPeriodEnd,
				}, nil)
				return m
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "trialing subscription within trial",
			shopID: 2,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetSubscriptionByShopID(2).Return(&model.Subscription{
					ID:                 2,
					ShopID:             2,
					Status:             "trialing",
					TrialEndsAt:        sql.NullTime{Time: futurePeriodEnd, Valid: true},
					CurrentPeriodStart: now,
					CurrentPeriodEnd:   futurePeriodEnd,
				}, nil)
				return m
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "trialing subscription expired",
			shopID: 2,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetSubscriptionByShopID(2).Return(&model.Subscription{
					ID:                 2,
					ShopID:             2,
					Status:             "trialing",
					TrialEndsAt:        sql.NullTime{Time: pastPeriodEnd, Valid: true},
					CurrentPeriodStart: pastPeriodEnd,
					CurrentPeriodEnd:   pastPeriodEnd,
				}, nil)
				return m
			},
			want:    false,
			wantErr: false,
		},
		{
			name:   "no subscription returns false",
			shopID: 3,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetSubscriptionByShopID(3).Return(nil, nil)
				return m
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			subscriptionStore = tt.mockSetup(ctrl)
			svc := &ssubscription{}
			got, err := svc.IsSubscriptionActive(tt.shopID)

			if (err != nil) != tt.wantErr {
				t.Errorf("IsSubscriptionActive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsSubscriptionActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_subscriptionService_CreateTrialSubscription(t *testing.T) {
	tests := []struct {
		name      string
		shopID    int
		mockSetup func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB)
		wantErr   bool
	}{
		{
			name:   "creates trial subscription successfully",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				subMock := mock_store.NewMockSubscriptionStore(ctrl)
				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)

				subMock.EXPECT().GetActivePlans().Return([]model.Plan{
					{ID: 1, Name: "starter", PriceIDR: 149000, MaxUsers: 1},
				}, nil)

				dbMock.EXPECT().Begin().Return(txMock, nil)
				subMock.EXPECT().CreateTrialSubscription(txMock, 1, 1, gomock.Any()).Return(&model.Subscription{ID: 1}, nil)
				txMock.EXPECT().Commit().Return(nil)
				txMock.EXPECT().Rollback().Return(nil)

				return subMock, dbMock
			},
			wantErr: false,
		},
		{
			name:   "returns error when no plans found",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				subMock := mock_store.NewMockSubscriptionStore(ctrl)
				dbMock := mock_database.NewMockDB(ctrl)

				subMock.EXPECT().GetActivePlans().Return([]model.Plan{}, nil)

				return subMock, dbMock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			subMock, dbMock := tt.mockSetup(ctrl)
			subscriptionStore = subMock
			oldDBGetter := dbGetter
			defer func() { dbGetter = oldDBGetter }()
			dbGetter = func() database.DB { return dbMock }

			svc := &ssubscription{}
			err := svc.CreateTrialSubscription(tt.shopID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTrialSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ssubscription_GetSubscriptionByShopID(t *testing.T) {
	now := time.Now()
	trialEnd := now.Add(14 * 24 * time.Hour)

	tests := []struct {
		name      string
		shopID    int
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore
		wantNil   bool
		wantErr   bool
	}{
		{
			name:   "returns subscription with plan",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{
					ID:                 1,
					ShopID:             1,
					PlanID:             1,
					Status:             "trialing",
					TrialEndsAt:        sql.NullTime{Time: trialEnd, Valid: true},
					CurrentPeriodStart: now,
					CurrentPeriodEnd:   trialEnd,
				}, nil)
				m.EXPECT().GetPlanByID(1).Return(&model.Plan{
					ID: 1, Name: "starter", DisplayName: "Starter", PriceIDR: 149000, MaxUsers: 1,
				}, nil)
				return m
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name:   "returns error when subscription not found",
			shopID: 2,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetSubscriptionByShopID(2).Return(nil, nil)
				return m
			},
			wantNil: false,
			wantErr: true,
		},
		{
			name:   "returns error when plan not found",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{
					ID: 1, ShopID: 1, PlanID: 99,
				}, nil)
				m.EXPECT().GetPlanByID(99).Return(nil, nil)
				return m
			},
			wantNil: false,
			wantErr: true,
		},
		{
			name:   "returns error on store failure",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetSubscriptionByShopID(1).Return(nil, errors.New("db error"))
				return m
			},
			wantNil: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			subscriptionStore = tt.mockSetup(ctrl)
			svc := &ssubscription{}
			got, gotErr := svc.GetSubscriptionByShopID(tt.shopID)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetSubscriptionByShopID() failed: %v", gotErr)
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

func Test_ssubscription_Checkout(t *testing.T) {
	tests := []struct {
		name      string
		shopID    int
		planID    int
		mockSetup func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB)
		snapFunc  func(string, int, int) (*midtransSnapResponse, error)
		wantErr   bool
	}{
		{
			name:   "returns error when plan not found",
			shopID: 1,
			planID: 99,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPlanByID(99).Return(nil, nil)
				return m, mock_database.NewMockDB(ctrl)
			},
			wantErr: true,
		},
		{
			name:   "returns error when subscription not found",
			shopID: 1,
			planID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPlanByID(1).Return(&model.Plan{ID: 1, PriceIDR: 149000}, nil)
				m.EXPECT().GetSubscriptionByShopID(1).Return(nil, nil)
				return m, mock_database.NewMockDB(ctrl)
			},
			wantErr: true,
		},
		{
			name:   "returns error on db.Begin failure",
			shopID: 1,
			planID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPlanByID(1).Return(&model.Plan{ID: 1, PriceIDR: 149000}, nil)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{ID: 1, ShopID: 1, PlanID: 1}, nil)

				dbMock := mock_database.NewMockDB(ctrl)
				dbMock.EXPECT().Begin().Return(nil, errors.New("begin error"))
				return m, dbMock
			},
			wantErr: true,
		},
		{
			name:   "returns error on CreatePayment failure",
			shopID: 1,
			planID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPlanByID(1).Return(&model.Plan{ID: 1, PriceIDR: 149000}, nil)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{ID: 1, ShopID: 1, PlanID: 1}, nil)
				m.EXPECT().CreatePayment(gomock.Any(), 1, 1, 1, gomock.Any(), 149000).Return(nil, errors.New("db error"))

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			wantErr: true,
		},
		{
			name:   "returns error on tx.Commit failure",
			shopID: 1,
			planID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPlanByID(1).Return(&model.Plan{ID: 1, PriceIDR: 149000}, nil)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{ID: 1, ShopID: 1, PlanID: 1}, nil)
				m.EXPECT().CreatePayment(gomock.Any(), 1, 1, 1, gomock.Any(), 149000).Return(&model.Payment{ID: 1}, nil)

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(errors.New("commit error"))
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			wantErr: true,
		},
		{
			name:   "returns error when callMidtransSnap fails",
			shopID: 1,
			planID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPlanByID(1).Return(&model.Plan{ID: 1, PriceIDR: 149000}, nil)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{ID: 1, ShopID: 1, PlanID: 1}, nil)
				m.EXPECT().CreatePayment(gomock.Any(), 1, 1, 1, gomock.Any(), 149000).Return(&model.Payment{ID: 1}, nil)

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			snapFunc: func(string, int, int) (*midtransSnapResponse, error) {
				return nil, errors.New("midtrans unavailable")
			},
			wantErr: true,
		},
		{
			name:   "returns checkout data even when UpdatePaymentSnapInfo fails",
			shopID: 1,
			planID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPlanByID(1).Return(&model.Plan{ID: 1, PriceIDR: 149000}, nil)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{ID: 1, ShopID: 1, PlanID: 1}, nil)
				m.EXPECT().CreatePayment(gomock.Any(), 1, 1, 1, gomock.Any(), 149000).Return(&model.Payment{ID: 1}, nil)
				m.EXPECT().UpdatePaymentSnapInfo(1, "snap-token", "https://snap.example.com").Return(errors.New("db error"))

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			snapFunc: func(string, int, int) (*midtransSnapResponse, error) {
				return &midtransSnapResponse{Token: "snap-token", RedirectURL: "https://snap.example.com"}, nil
			},
			wantErr: false,
		},
		{
			name:   "successfully completes checkout",
			shopID: 1,
			planID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPlanByID(1).Return(&model.Plan{ID: 1, PriceIDR: 149000}, nil)
				m.EXPECT().GetSubscriptionByShopID(1).Return(&model.Subscription{ID: 1, ShopID: 1, PlanID: 1}, nil)
				m.EXPECT().CreatePayment(gomock.Any(), 1, 1, 1, gomock.Any(), 149000).Return(&model.Payment{ID: 1}, nil)
				m.EXPECT().UpdatePaymentSnapInfo(1, "snap-token", "https://snap.example.com").Return(nil)

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			snapFunc: func(string, int, int) (*midtransSnapResponse, error) {
				return &midtransSnapResponse{Token: "snap-token", RedirectURL: "https://snap.example.com"}, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			subMock, dbMock := tt.mockSetup(ctrl)
			subscriptionStore = subMock
			oldDBGetter := dbGetter
			defer func() { dbGetter = oldDBGetter }()
			dbGetter = func() database.DB { return dbMock }

			oldSnapFunc := midtransSnapFunc
			defer func() { midtransSnapFunc = oldSnapFunc }()
			if tt.snapFunc != nil {
				midtransSnapFunc = tt.snapFunc
			}

			svc := &ssubscription{}
			got, gotErr := svc.Checkout(tt.shopID, tt.planID)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Checkout() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Checkout() succeeded unexpectedly")
			}
			if got == nil {
				t.Error("Checkout() returned nil, want non-nil")
			}
		})
	}
}

func makeSignature(orderID, statusCode, grossAmount, serverKey string) string {
	raw := orderID + statusCode + grossAmount + serverKey
	h := sha512.Sum512([]byte(raw))
	return fmt.Sprintf("%x", h)
}

func Test_ssubscription_HandleMidtransWebhook(t *testing.T) {
	const testServerKey = "test-server-key"

	tests := []struct {
		name      string
		payload   MidtransWebhookPayload
		mockSetup func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB)
		wantErr   bool
	}{
		{
			name: "returns error on invalid signature",
			payload: MidtransWebhookPayload{
				OrderID:      "recapo-1-111",
				StatusCode:   "200",
				GrossAmount:  "149000.00",
				SignatureKey: "bad-signature",
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				return mock_store.NewMockSubscriptionStore(ctrl), mock_database.NewMockDB(ctrl)
			},
			wantErr: true,
		},
		{
			name: "returns error when payment not found",
			payload: MidtransWebhookPayload{
				OrderID:           "recapo-1-111",
				StatusCode:        "200",
				GrossAmount:       "149000.00",
				TransactionStatus: "settlement",
				SignatureKey:      makeSignature("recapo-1-111", "200", "149000.00", testServerKey),
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPaymentByMidtransOrderID("recapo-1-111").Return(nil, nil)
				return m, mock_database.NewMockDB(ctrl)
			},
			wantErr: true,
		},
		{
			name: "settlement activates subscription",
			payload: MidtransWebhookPayload{
				OrderID:           "recapo-1-111",
				StatusCode:        "200",
				GrossAmount:       "149000.00",
				TransactionStatus: "settlement",
				TransactionID:     "txn-abc",
				SignatureKey:      makeSignature("recapo-1-111", "200", "149000.00", testServerKey),
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPaymentByMidtransOrderID("recapo-1-111").Return(&model.Payment{ID: 1, SubscriptionID: 1}, nil)
				m.EXPECT().UpdatePaymentSettled(gomock.Any(), 1, "txn-abc", gomock.Any()).Return(nil)
				m.EXPECT().UpdateSubscriptionStatus(gomock.Any(), 1, "active", gomock.Any()).Return(nil)

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			wantErr: false,
		},
		{
			name: "capture with accept fraud status activates subscription",
			payload: MidtransWebhookPayload{
				OrderID:           "recapo-1-111",
				StatusCode:        "200",
				GrossAmount:       "149000.00",
				TransactionStatus: "capture",
				FraudStatus:       "accept",
				TransactionID:     "txn-abc",
				SignatureKey:      makeSignature("recapo-1-111", "200", "149000.00", testServerKey),
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPaymentByMidtransOrderID("recapo-1-111").Return(&model.Payment{ID: 1, SubscriptionID: 1}, nil)
				m.EXPECT().UpdatePaymentSettled(gomock.Any(), 1, "txn-abc", gomock.Any()).Return(nil)
				m.EXPECT().UpdateSubscriptionStatus(gomock.Any(), 1, "active", gomock.Any()).Return(nil)

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			wantErr: false,
		},
		{
			name: "capture with non-accept fraud status is a no-op",
			payload: MidtransWebhookPayload{
				OrderID:           "recapo-1-111",
				StatusCode:        "200",
				GrossAmount:       "149000.00",
				TransactionStatus: "capture",
				FraudStatus:       "challenge",
				SignatureKey:      makeSignature("recapo-1-111", "200", "149000.00", testServerKey),
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPaymentByMidtransOrderID("recapo-1-111").Return(&model.Payment{ID: 1}, nil)

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			wantErr: false,
		},
		{
			name: "deny marks payment as failed",
			payload: MidtransWebhookPayload{
				OrderID:           "recapo-1-111",
				StatusCode:        "202",
				GrossAmount:       "149000.00",
				TransactionStatus: "deny",
				SignatureKey:      makeSignature("recapo-1-111", "202", "149000.00", testServerKey),
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().GetPaymentByMidtransOrderID("recapo-1-111").Return(&model.Payment{ID: 1}, nil)
				m.EXPECT().UpdatePaymentFailed(gomock.Any(), 1, "deny").Return(nil)

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			subMock, dbMock := tt.mockSetup(ctrl)
			subscriptionStore = subMock
			oldDBGetter := dbGetter
			defer func() { dbGetter = oldDBGetter }()
			dbGetter = func() database.DB { return dbMock }
			cfg.MidtransServerKey = testServerKey

			svc := &ssubscription{}
			gotErr := svc.HandleMidtransWebhook(tt.payload)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("HandleMidtransWebhook() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("HandleMidtransWebhook() succeeded unexpectedly")
			}
		})
	}
}

func Test_ssubscription_ExpireSubscriptions(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore
		wantErr   bool
	}{
		{
			name: "successfully expires active subscriptions",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().ExpireSubscriptions().Return(int64(3), nil)
				return m
			},
			wantErr: false,
		},
		{
			name: "returns nil error when no subscriptions to expire",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().ExpireSubscriptions().Return(int64(0), nil)
				return m
			},
			wantErr: false,
		},
		{
			name: "returns error on store failure",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockSubscriptionStore {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().ExpireSubscriptions().Return(int64(0), errors.New("database error"))
				return m
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			subscriptionStore = tt.mockSetup(ctrl)

			svc := &ssubscription{}
			gotErr := svc.ExpireSubscriptions()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ExpireSubscriptions() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ExpireSubscriptions() succeeded unexpectedly")
			}
		})
	}
}

func Test_ssubscription_CancelSubscription(t *testing.T) {
	tests := []struct {
		name      string
		shopID    int
		mockSetup func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB)
		wantErr   bool
	}{
		{
			name:   "successfully cancels active subscription",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().
					GetSubscriptionByShopID(1).
					Return(&model.Subscription{ID: 5, ShopID: 1, Status: "active"}, nil)
				m.EXPECT().
					CancelSubscription(gomock.Any(), 5).
					Return(nil)

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			wantErr: false,
		},
		{
			name:   "returns error when GetSubscriptionByShopID fails",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().
					GetSubscriptionByShopID(1).
					Return(nil, errors.New("database error"))
				return m, mock_database.NewMockDB(ctrl)
			},
			wantErr: true,
		},
		{
			name:   "returns error when subscription not found",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().
					GetSubscriptionByShopID(1).
					Return(nil, nil)
				return m, mock_database.NewMockDB(ctrl)
			},
			wantErr: true,
		},
		{
			name:   "returns error when subscription is not active",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().
					GetSubscriptionByShopID(1).
					Return(&model.Subscription{ID: 5, ShopID: 1, Status: "trialing"}, nil)
				return m, mock_database.NewMockDB(ctrl)
			},
			wantErr: true,
		},
		{
			name:   "returns error on db.Begin failure",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().
					GetSubscriptionByShopID(1).
					Return(&model.Subscription{ID: 5, ShopID: 1, Status: "active"}, nil)

				dbMock := mock_database.NewMockDB(ctrl)
				dbMock.EXPECT().Begin().Return(nil, errors.New("begin error"))
				return m, dbMock
			},
			wantErr: true,
		},
		{
			name:   "returns error when CancelSubscription store call fails",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().
					GetSubscriptionByShopID(1).
					Return(&model.Subscription{ID: 5, ShopID: 1, Status: "active"}, nil)
				m.EXPECT().
					CancelSubscription(gomock.Any(), 5).
					Return(errors.New("database error"))

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			wantErr: true,
		},
		{
			name:   "returns error when tx.Commit fails",
			shopID: 1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockSubscriptionStore, *mock_database.MockDB) {
				m := mock_store.NewMockSubscriptionStore(ctrl)
				m.EXPECT().
					GetSubscriptionByShopID(1).
					Return(&model.Subscription{ID: 5, ShopID: 1, Status: "active"}, nil)
				m.EXPECT().
					CancelSubscription(gomock.Any(), 5).
					Return(nil)

				dbMock := mock_database.NewMockDB(ctrl)
				txMock := mock_database.NewMockTx(ctrl)
				dbMock.EXPECT().Begin().Return(txMock, nil)
				txMock.EXPECT().Commit().Return(errors.New("commit error"))
				txMock.EXPECT().Rollback().Return(nil)
				return m, dbMock
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			subMock, dbMock := tt.mockSetup(ctrl)
			subscriptionStore = subMock
			oldDBGetter := dbGetter
			defer func() { dbGetter = oldDBGetter }()
			dbGetter = func() database.DB { return dbMock }

			svc := &ssubscription{}
			gotErr := svc.CancelSubscription(tt.shopID)
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
