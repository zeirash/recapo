package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	InvitationStore interface {
		CreateInvitation(ctx context.Context, shopID, invitedBy int, email, token string) (*model.Invitation, error)
		GetInvitationByToken(ctx context.Context, token string) (*model.Invitation, error)
		GetPendingInvitationByEmail(ctx context.Context, shopID int, email string) (*model.Invitation, error)
		AcceptInvitation(ctx context.Context, id int) error
	}

	invitation struct {
		db *sql.DB
	}
)

func NewInvitationStore() InvitationStore {
	return &invitation{db: database.GetDB()}
}

func NewInvitationStoreWithDB(db *sql.DB) InvitationStore {
	return &invitation{db: db}
}

func (s *invitation) CreateInvitation(ctx context.Context, shopID, invitedBy int, email, token string) (*model.Invitation, error) {
	now := time.Now()
	var id int

	q := `
		INSERT INTO invitations (shop_id, invited_by, email, token, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := s.db.QueryRowContext(ctx, q, shopID, invitedBy, email, token, constant.InvitationStatusPending, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.Invitation{
		ID:        id,
		ShopID:    shopID,
		Email:     email,
		Token:     token,
		Status:    constant.InvitationStatusPending,
		InvitedBy: invitedBy,
		CreatedAt: now,
	}, nil
}

func (s *invitation) GetInvitationByToken(ctx context.Context, token string) (*model.Invitation, error) {
	q := `
		SELECT id, shop_id, email, token, status, invited_by, created_at, updated_at
		FROM invitations
		WHERE token = $1
	`

	var inv model.Invitation
	err := s.db.QueryRowContext(ctx, q, token).Scan(
		&inv.ID, &inv.ShopID, &inv.Email, &inv.Token, &inv.Status,
		&inv.InvitedBy, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &inv, nil
}

func (s *invitation) GetPendingInvitationByEmail(ctx context.Context, shopID int, email string) (*model.Invitation, error) {
	q := `
		SELECT id, shop_id, email, token, status, invited_by, created_at, updated_at
		FROM invitations
		WHERE shop_id = $1 AND email = $2 AND status = $3
	`

	var inv model.Invitation
	err := s.db.QueryRowContext(ctx, q, shopID, email, constant.InvitationStatusPending).Scan(
		&inv.ID, &inv.ShopID, &inv.Email, &inv.Token, &inv.Status,
		&inv.InvitedBy, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &inv, nil
}

func (s *invitation) AcceptInvitation(ctx context.Context, id int) error {
	q := `UPDATE invitations SET status = $1, updated_at = now() WHERE id = $2`
	_, err := s.db.ExecContext(ctx, q, constant.InvitationStatusAccepted, id)
	return err
}
