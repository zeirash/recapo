package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/model"
)

func Test_invitation_CreateInvitation(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	_ = fixedTime

	tests := []struct {
		name      string
		shopID    int
		invitedBy int
		email     string
		token     string
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:      "successfully create invitation",
			shopID:    1,
			invitedBy: 2,
			email:     "invite@example.com",
			token:     "abc123token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(10)
				mock.ExpectQuery(`INSERT INTO invitations`).
					WithArgs(1, 2, "invite@example.com", "abc123token", constant.InvitationStatusPending, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:      "returns error on database failure",
			shopID:    1,
			invitedBy: 2,
			email:     "invite@example.com",
			token:     "abc123token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO invitations`).
					WithArgs(1, 2, "invite@example.com", "abc123token", constant.InvitationStatusPending, sqlmock.AnyArg()).
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

			s := &invitation{db: db}
			got, gotErr := s.CreateInvitation(context.Background(), tt.shopID, tt.invitedBy, tt.email, tt.token)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateInvitation() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateInvitation() succeeded unexpectedly")
			}

			if got.ID != 10 || got.ShopID != tt.shopID || got.Email != tt.email || got.Token != tt.token || got.Status != constant.InvitationStatusPending {
				t.Errorf("CreateInvitation() = %+v", got)
			}
		})
	}
}

func Test_invitation_GetInvitationByToken(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		token      string
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Invitation
		wantErr    bool
	}{
		{
			name:  "successfully get invitation by token",
			token: "abc123token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "email", "token", "status", "invited_by", "created_at", "updated_at"}).
					AddRow(1, 5, "invite@example.com", "abc123token", "pending", 2, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, email, token, status, invited_by, created_at, updated_at\s+FROM invitations\s+WHERE token = \$1`).
					WithArgs("abc123token").
					WillReturnRows(rows)
			},
			wantResult: &model.Invitation{
				ID:        1,
				ShopID:    5,
				Email:     "invite@example.com",
				Token:     "abc123token",
				Status:    "pending",
				InvitedBy: 2,
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:  "returns nil when not found",
			token: "notexist",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, email, token, status, invited_by, created_at, updated_at\s+FROM invitations\s+WHERE token = \$1`).
					WithArgs("notexist").
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:  "returns error on database failure",
			token: "abc123token",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, email, token, status, invited_by, created_at, updated_at\s+FROM invitations\s+WHERE token = \$1`).
					WithArgs("abc123token").
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
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

			s := &invitation{db: db}
			got, gotErr := s.GetInvitationByToken(context.Background(), tt.token)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetInvitationByToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetInvitationByToken() succeeded unexpectedly")
			}

			if tt.wantResult == nil {
				if got != nil {
					t.Errorf("GetInvitationByToken() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("GetInvitationByToken() = nil, want non-nil")
			}
			if got.ID != tt.wantResult.ID || got.ShopID != tt.wantResult.ShopID ||
				got.Email != tt.wantResult.Email || got.Token != tt.wantResult.Token ||
				got.Status != tt.wantResult.Status || got.InvitedBy != tt.wantResult.InvitedBy {
				t.Errorf("GetInvitationByToken() = %+v, want %+v", got, tt.wantResult)
			}
		})
	}
}

func Test_invitation_GetPendingInvitationByEmail(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		shopID     int
		email      string
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Invitation
		wantErr    bool
	}{
		{
			name:   "successfully get pending invitation by email",
			shopID: 5,
			email:  "invite@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "email", "token", "status", "invited_by", "created_at", "updated_at"}).
					AddRow(3, 5, "invite@example.com", "tokenxyz", "pending", 2, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, email, token, status, invited_by, created_at, updated_at\s+FROM invitations\s+WHERE shop_id = \$1 AND email = \$2 AND status = \$3`).
					WithArgs(5, "invite@example.com", constant.InvitationStatusPending).
					WillReturnRows(rows)
			},
			wantResult: &model.Invitation{
				ID:        3,
				ShopID:    5,
				Email:     "invite@example.com",
				Token:     "tokenxyz",
				Status:    "pending",
				InvitedBy: 2,
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "returns nil when no pending invite",
			shopID: 5,
			email:  "noinvite@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, email, token, status, invited_by, created_at, updated_at\s+FROM invitations\s+WHERE shop_id = \$1 AND email = \$2 AND status = \$3`).
					WithArgs(5, "noinvite@example.com", constant.InvitationStatusPending).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:   "returns error on database failure",
			shopID: 5,
			email:  "invite@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, email, token, status, invited_by, created_at, updated_at\s+FROM invitations\s+WHERE shop_id = \$1 AND email = \$2 AND status = \$3`).
					WithArgs(5, "invite@example.com", constant.InvitationStatusPending).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
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

			s := &invitation{db: db}
			got, gotErr := s.GetPendingInvitationByEmail(context.Background(), tt.shopID, tt.email)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetPendingInvitationByEmail() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetPendingInvitationByEmail() succeeded unexpectedly")
			}

			if tt.wantResult == nil {
				if got != nil {
					t.Errorf("GetPendingInvitationByEmail() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("GetPendingInvitationByEmail() = nil, want non-nil")
			}
			if got.ID != tt.wantResult.ID || got.Email != tt.wantResult.Email || got.Status != tt.wantResult.Status {
				t.Errorf("GetPendingInvitationByEmail() = %+v, want %+v", got, tt.wantResult)
			}
		})
	}
}

func Test_invitation_AcceptInvitation(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "successfully accept invitation",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE invitations SET status = \$1, updated_at = now\(\) WHERE id = \$2`).
					WithArgs(constant.InvitationStatusAccepted, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "returns error on database failure",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE invitations SET status = \$1, updated_at = now\(\) WHERE id = \$2`).
					WithArgs(constant.InvitationStatusAccepted, 1).
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

			s := &invitation{db: db}
			gotErr := s.AcceptInvitation(context.Background(), tt.id)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("AcceptInvitation() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("AcceptInvitation() succeeded unexpectedly")
			}
		})
	}
}
