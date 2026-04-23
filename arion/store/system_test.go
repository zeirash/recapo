package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeirash/recapo/arion/model"
)

func Test_systemStore_GetSystemStats(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(mock sqlmock.Sqlmock)
		want      *SystemStats
		wantErr   bool
	}{
		{
			name: "returns stats successfully",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM shops`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
				mock.ExpectQuery(`SELECT s.status, COUNT\(\*\) as cnt`).
					WillReturnRows(sqlmock.NewRows([]string{"status", "cnt"}).
						AddRow("active", 2).
						AddRow("trialing", 3).
						AddRow("expired", 1).
						AddRow("cancelled", 1))
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(p.price_idr\), 0\)`).
					WillReturnRows(sqlmock.NewRows([]string{"mrr"}).AddRow(298000))
			},
			want: &SystemStats{
				TotalShops:    5,
				SubsActive:    2,
				SubsTrialing:  3,
				SubsExpired:   1,
				SubsCancelled: 1,
				MRRIDR:        298000,
			},
			wantErr: false,
		},
		{
			name: "returns error on shops query failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM shops`).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "returns error on subscription status query failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM shops`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
				mock.ExpectQuery(`SELECT s.status, COUNT\(\*\) as cnt`).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "returns error on MRR query failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM shops`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
				mock.ExpectQuery(`SELECT s.status, COUNT\(\*\) as cnt`).
					WillReturnRows(sqlmock.NewRows([]string{"status", "cnt"}))
				mock.ExpectQuery(`SELECT COALESCE\(SUM\(p.price_idr\), 0\)`).
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

			s := &systemStore{db: db}
			got, gotErr := s.GetSystemStats(context.Background())

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetSystemStats() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetSystemStats() succeeded unexpectedly")
			}
			if got.TotalShops != tt.want.TotalShops {
				t.Errorf("TotalShops = %v, want %v", got.TotalShops, tt.want.TotalShops)
			}
			if got.SubsActive != tt.want.SubsActive {
				t.Errorf("SubsActive = %v, want %v", got.SubsActive, tt.want.SubsActive)
			}
			if got.SubsTrialing != tt.want.SubsTrialing {
				t.Errorf("SubsTrialing = %v, want %v", got.SubsTrialing, tt.want.SubsTrialing)
			}
			if got.MRRIDR != tt.want.MRRIDR {
				t.Errorf("MRRIDR = %v, want %v", got.MRRIDR, tt.want.MRRIDR)
			}
		})
	}
}

func Test_systemStore_GetSystemShops(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	trialEnd := fixedTime.AddDate(0, 0, 14)

	tests := []struct {
		name      string
		mockSetup func(mock sqlmock.Sqlmock)
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns shops list",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "owner_name", "owner_email",
					"plan_name", "sub_status", "trial_ends_at", "period_end", "created_at",
				}).
					AddRow(1, "Toko Mawar", "Siti", "siti@email.com", "Starter", "trialing", trialEnd, trialEnd, fixedTime).
					AddRow(2, "Toko Melati", "Budi", "budi@email.com", "Starter", "active", nil, fixedTime.AddDate(0, 1, 0), fixedTime)
				mock.ExpectQuery(`SELECT\s+sh\.id`).WillReturnRows(rows)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "returns empty list when no shops",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{
					"id", "name", "owner_name", "owner_email",
					"plan_name", "sub_status", "trial_ends_at", "period_end", "created_at",
				})
				mock.ExpectQuery(`SELECT\s+sh\.id`).WillReturnRows(rows)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "returns error on database failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.id`).WillReturnError(errors.New("database error"))
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

			s := &systemStore{db: db}
			got, gotErr := s.GetSystemShops(context.Background())

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetSystemShops() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetSystemShops() succeeded unexpectedly")
			}
			if len(got) != tt.wantLen {
				t.Errorf("GetSystemShops() returned %d shops, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func Test_systemStore_GetSystemPayments(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	ptrTime := func(t time.Time) *time.Time { return &t }
	strPtr := func(s string) *string { return &s }

	emptyRows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{
			"shop_name", "plan_name", "amount_idr", "status",
			"midtrans_order_id", "paid_at", "created_at",
		})
	}
	twoRows := func() *sqlmock.Rows {
		return emptyRows().
			AddRow("Toko Mawar", "Starter", 149000, "settlement", "recapo-1-001", fixedTime, fixedTime).
			AddRow("Toko Melati", "Starter", 149000, "pending", "recapo-2-002", nil, fixedTime)
	}

	tests := []struct {
		name      string
		opts      model.SystemPaymentFilterOptions
		mockSetup func(mock sqlmock.Sqlmock)
		wantLen   int
		wantErr   bool
	}{
		{
			name: "returns payments list",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.name`).WillReturnRows(twoRows())
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "returns empty list when no payments",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.name`).WillReturnRows(emptyRows())
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "returns error on database failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.name`).WillReturnError(errors.New("database error"))
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "filters by date_from",
			opts: model.SystemPaymentFilterOptions{
				DateFrom: ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.name`).
					WithArgs(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)).
					WillReturnRows(twoRows())
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "filters by date_to",
			opts: model.SystemPaymentFilterOptions{
				DateTo: ptrTime(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.name`).
					WithArgs(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)).
					WillReturnRows(twoRows())
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "filters by status",
			opts: model.SystemPaymentFilterOptions{
				Status: strPtr("settlement"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.name`).
					WithArgs("settlement").
					WillReturnRows(emptyRows().AddRow("Toko Mawar", "Starter", 149000, "settlement", "recapo-1-001", fixedTime, fixedTime))
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "applies sort order",
			opts: model.SystemPaymentFilterOptions{
				Sort: strPtr("amount_idr,desc"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.name`).WillReturnRows(twoRows())
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "ignores invalid sort column",
			opts: model.SystemPaymentFilterOptions{
				Sort: strPtr("dropped_tables,desc"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.name`).WillReturnRows(twoRows())
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "all filters combined",
			opts: model.SystemPaymentFilterOptions{
				DateFrom: ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				DateTo:   ptrTime(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)),
				Status:   strPtr("settlement"),
				Sort:     strPtr("amount_idr,asc"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT\s+sh\.name`).
					WithArgs(
						time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
						"settlement",
					).
					WillReturnRows(emptyRows().AddRow("Toko Mawar", "Starter", 149000, "settlement", "recapo-1-001", fixedTime, fixedTime))
			},
			wantLen: 1,
			wantErr: false,
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

			s := &systemStore{db: db}
			got, gotErr := s.GetSystemPayments(context.Background(), tt.opts)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetSystemPayments() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetSystemPayments() succeeded unexpectedly")
			}
			if len(got) != tt.wantLen {
				t.Errorf("GetSystemPayments() returned %d payments, want %d", len(got), tt.wantLen)
			}
		})
	}
}
