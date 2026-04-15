package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"regexp"

	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/constant"
	emailPkg "github.com/zeirash/recapo/arion/common/email"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"

	"golang.org/x/crypto/bcrypt"
)

type (
	InvitationService interface {
		InviteAdmin(ctx context.Context, shopID, userID int, email, lang string) error
		ValidateInviteToken(ctx context.Context, token string) (*response.InvitationData, error)
		AcceptInvite(ctx context.Context, token, name, password string) (response.TokenResponse, error)
	}

	iservice struct{}
)

func NewInvitationService() InvitationService {
	cfg = config.GetConfig()

	if userStore == nil {
		userStore = store.NewUserStore()
	}
	if shopStore == nil {
		shopStore = store.NewShopStore()
	}
	if tokenStore == nil {
		tokenStore = store.NewTokenStore()
	}
	if invitationStore == nil {
		invitationStore = store.NewInvitationStore()
	}
	if subscriptionStore == nil {
		subscriptionStore = store.NewSubscriptionStore()
	}

	return &iservice{}
}

var emailRegexp = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func generateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *iservice) InviteAdmin(ctx context.Context, shopID, userID int, email, lang string) error {
	if !emailRegexp.MatchString(email) {
		return errors.New(apierr.ErrEmailInvalid)
	}

	// Verify caller is owner
	caller, err := userStore.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if caller == nil || caller.Role != constant.RoleOwner {
		return errors.New(apierr.ErrNotOwner)
	}

	// Check plan user limit
	sub, err := subscriptionStore.GetSubscriptionByShopID(ctx, shopID)
	if err != nil {
		return err
	}
	if sub != nil {
		plan, err := subscriptionStore.GetPlanByID(ctx, sub.PlanID)
		if err != nil {
			return err
		}
		if plan != nil && plan.MaxUsers > 0 {
			count, err := userStore.CountUsersByShopID(ctx, shopID)
			if err != nil {
				return err
			}
			if count >= plan.MaxUsers {
				return errors.New(apierr.ErrMaxUsersReached)
			}
		}
	}

	// Check email not already registered
	existingUser, err := userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New(apierr.ErrUserAlreadyExists)
	}

	// Check no pending invite already exists
	existing, err := invitationStore.GetPendingInvitationByEmail(ctx, shopID, email)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New(apierr.ErrInvitationAlreadySent)
	}

	// Get shop for name in email
	shop, err := shopStore.GetShopByID(ctx, shopID)
	if err != nil {
		return err
	}
	if shop == nil {
		return errors.New(apierr.ErrShopNotFound)
	}

	token, err := generateInviteToken()
	if err != nil {
		return err
	}

	if _, err := invitationStore.CreateInvitation(ctx, shopID, userID, email, token); err != nil {
		return err
	}

	inviteURL := cfg.FrontendURL + "/accept-invite?token=" + token
	return emailPkg.SendInvitation(email, caller.Name, shop.Name, inviteURL, lang)
}

func (s *iservice) ValidateInviteToken(ctx context.Context, token string) (*response.InvitationData, error) {
	inv, err := invitationStore.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil || inv.Status != constant.InvitationStatusPending {
		return nil, errors.New(apierr.ErrInvitationNotFound)
	}

	shop, err := shopStore.GetShopByID(ctx, inv.ShopID)
	if err != nil {
		return nil, err
	}
	if shop == nil {
		return nil, errors.New(apierr.ErrShopNotFound)
	}

	return &response.InvitationData{
		Email:    inv.Email,
		ShopName: shop.Name,
	}, nil
}

func (s *iservice) AcceptInvite(ctx context.Context, token, name, password string) (response.TokenResponse, error) {
	inv, err := invitationStore.GetInvitationByToken(ctx, token)
	if err != nil {
		return response.TokenResponse{}, err
	}
	if inv == nil {
		return response.TokenResponse{}, errors.New(apierr.ErrInvitationNotFound)
	}
	if inv.Status != constant.InvitationStatusPending {
		return response.TokenResponse{}, errors.New(apierr.ErrInvitationAlreadyAccepted)
	}

	// Check plan user limit before creating the user
	sub, err := subscriptionStore.GetSubscriptionByShopID(ctx, inv.ShopID)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if sub == nil {
		return response.TokenResponse{}, errors.New(apierr.ErrSubscriptionNotFound)
	}

	plan, err := subscriptionStore.GetPlanByID(ctx, sub.PlanID)
	if err != nil {
		return response.TokenResponse{}, err
	}
	if plan != nil && plan.MaxUsers > 0 {
		count, err := userStore.CountUsersByShopID(ctx, inv.ShopID)
		if err != nil {
			return response.TokenResponse{}, err
		}
		if count >= plan.MaxUsers {
			return response.TokenResponse{}, errors.New(apierr.ErrMaxUsersReached)
		}
	}

	if err := validatePasswordStrength(password); err != nil {
		return response.TokenResponse{}, err
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return response.TokenResponse{}, err
	}

	db := dbGetter()
	tx, err := db.Begin()
	if err != nil {
		return response.TokenResponse{}, err
	}
	defer tx.Rollback()

	newUser, err := userStore.CreateUser(ctx, tx, name, inv.Email, string(encryptedPassword), constant.RoleAdmin, inv.ShopID)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return response.TokenResponse{}, err
	}

	if err := invitationStore.AcceptInvitation(ctx, inv.ID); err != nil {
		return response.TokenResponse{}, err
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		return response.TokenResponse{}, err
	}

	if err := userStore.SetSessionToken(ctx, newUser.ID, sessionToken); err != nil {
		return response.TokenResponse{}, err
	}
	newUser.SessionToken = sql.NullString{String: sessionToken, Valid: true}

	accessToken, err := tokenStore.CreateAccessToken(ctx, newUser, cfg.SecretKey, 2)
	if err != nil {
		return response.TokenResponse{}, err
	}

	refreshToken, err := tokenStore.CreateRefreshToken(ctx, newUser, cfg.SecretKey, 168)
	if err != nil {
		return response.TokenResponse{}, err
	}

	return response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
