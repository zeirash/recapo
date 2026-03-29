package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"unicode"

	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/constant"
	emailPkg "github.com/zeirash/recapo/arion/common/email"
	"github.com/zeirash/recapo/arion/common/logger"
	otpPkg "github.com/zeirash/recapo/arion/common/otp"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"

	"golang.org/x/crypto/bcrypt"
)

type (
	UserService interface {
		UserLogin(ctx context.Context, email, password string) (response.TokenResponse, error)
		UserRegister(ctx context.Context, name, email, password string) (response.TokenResponse, error)
		RefreshToken(ctx context.Context, refreshToken string) (response.TokenResponse, error)
		UpdateUser(ctx context.Context, input UpdateUserInput) (response.UserData, error)
		GetUserByID(ctx context.Context, userID int) (*response.UserData, error)
		GetUsers(ctx context.Context) ([]response.UserData, error)
		SendOTP(ctx context.Context, email, lang string) error
		ForgotPassword(ctx context.Context, email, lang string) error
		ResetPassword(ctx context.Context, email, otp, newPassword string) error
		Logout(ctx context.Context, userID int) error
	}

	uservice struct{}

	UpdateUserInput struct {
		ID       int
		Name     *string
		Email    *string
		Password *string
	}
)

func NewUserService() UserService {
	cfg = config.GetConfig()

	if userStore == nil {
		userStore = store.NewUserStore()
	}

	if tokenStore == nil {
		tokenStore = store.NewTokenStore()
	}

	if shopStore == nil {
		shopStore = store.NewShopStore()
	}

	if subscriptionService == nil {
		subscriptionService = NewSubscriptionService()
	}

	return &uservice{}
}

func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New(apierr.ErrPasswordTooWeak)
	}
	hasLetter := false
	hasDigit := false
	for _, c := range password {
		if unicode.IsLetter(c) {
			hasLetter = true
		}
		if unicode.IsDigit(c) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New(apierr.ErrPasswordTooWeak)
	}
	return nil
}

func generateSessionToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (u *uservice) UserLogin(ctx context.Context, email, password string) (response.TokenResponse, error) {
	user, err := userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if user == nil {
		return response.TokenResponse{}, errors.New(apierr.ErrUserNotExist)
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return response.TokenResponse{}, errors.New(apierr.ErrPasswordIncorrect)
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		return response.TokenResponse{}, err
	}

	if err := userStore.SetSessionToken(ctx, user.ID, sessionToken); err != nil {
		return response.TokenResponse{}, err
	}
	user.SessionToken = sql.NullString{String: sessionToken, Valid: true}

	accessToken, err := tokenStore.CreateAccessToken(ctx, user, cfg.SecretKey, 2)
	if err != nil {
		return response.TokenResponse{}, err
	}

	refreshToken, err := tokenStore.CreateRefreshToken(ctx, user, cfg.SecretKey, 168)
	if err != nil {
		return response.TokenResponse{}, err
	}

	return response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *uservice) RefreshToken(ctx context.Context, refreshToken string) (response.TokenResponse, error) {
	// Validate and extract data from refresh token
	tokenData, err := tokenStore.ExtractDataFromToken(ctx, refreshToken, cfg.SecretKey)
	if err != nil {
		return response.TokenResponse{}, errors.New(apierr.ErrInvalidRefreshToken)
	}

	// Get user from database
	user, err := userStore.GetUserByID(ctx, tokenData.UserID)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if user == nil {
		return response.TokenResponse{}, errors.New(apierr.ErrUserNotFound)
	}

	// Validate session token: if DB has a valid session token, it must match the token's claim
	if user.SessionToken.Valid && user.SessionToken.String != tokenData.SessionToken {
		return response.TokenResponse{}, errors.New(apierr.ErrInvalidRefreshToken)
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		return response.TokenResponse{}, err
	}

	if err := userStore.SetSessionToken(ctx, user.ID, sessionToken); err != nil {
		return response.TokenResponse{}, err
	}
	user.SessionToken = sql.NullString{String: sessionToken, Valid: true}

	// Generate new tokens
	accessToken, err := tokenStore.CreateAccessToken(ctx, user, cfg.SecretKey, 2)
	if err != nil {
		return response.TokenResponse{}, err
	}

	newRefreshToken, err := tokenStore.CreateRefreshToken(ctx, user, cfg.SecretKey, 168)
	if err != nil {
		return response.TokenResponse{}, err
	}

	return response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *uservice) UserRegister(ctx context.Context, name, email, password string) (response.TokenResponse, error) {
	if err := validatePasswordStrength(password); err != nil {
		return response.TokenResponse{}, err
	}

	existUser, err := userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if existUser != nil {
		return response.TokenResponse{}, errors.New(apierr.ErrUserAlreadyExists)
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return response.TokenResponse{}, err
	}

	db := dbGetter()

	tx, err := db.Begin()
	if err != nil {
		return response.TokenResponse{}, err
	}
	defer tx.Rollback()

	shopName := fmt.Sprintf("%s's Shop", name)
	shop, err := shopStore.CreateShop(ctx, tx, shopName)
	if err != nil {
		return response.TokenResponse{}, err
	}

	newUser, err := userStore.CreateUser(ctx, tx, name, email, string(encryptedPassword), constant.RoleOwner, shop.ID)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return response.TokenResponse{}, err
	}

	if trialErr := subscriptionService.CreateTrialSubscription(ctx, shop.ID); trialErr != nil {
		logger.WithError(trialErr).Error("failed to create trial subscription")
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

func (u *uservice) Logout(ctx context.Context, userID int) error {
	return userStore.ClearSessionToken(ctx, userID)
}

func (u *uservice) UpdateUser(ctx context.Context, input UpdateUserInput) (response.UserData, error) {
	user, err := userStore.GetUserByID(ctx, input.ID)
	if err != nil {
		return response.UserData{}, err
	}

	if user == nil {
		return response.UserData{}, errors.New(apierr.ErrUserNotFound)
	}

	var password string
	if input.Password != nil {
		if err := validatePasswordStrength(*input.Password); err != nil {
			return response.UserData{}, err
		}

		encryptedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(*input.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return response.UserData{}, err
		}

		password = string(encryptedPassword)
	}

	updateData := store.UpdateUserInput{
		Name:  input.Name,
		Email: input.Email,
	}
	if password != "" {
		updateData.Password = &password
	}

	userData, err := userStore.UpdateUser(ctx, input.ID, updateData)
	if err != nil {
		return response.UserData{}, err
	}

	res := response.UserData{
		ID:        userData.ID,
		Name:      userData.Name,
		Email:     userData.Email,
		CreatedAt: userData.CreatedAt,
	}

	if userData.UpdatedAt.Valid {
		res.UpdatedAt = &userData.UpdatedAt.Time
	}

	return res, nil
}

func (u *uservice) GetUserByID(ctx context.Context, userID int) (*response.UserData, error) {
	user, err := userStore.GetUserByID(ctx, userID)
	if err != nil {
		return &response.UserData{}, err
	}

	if user == nil {
		return nil, errors.New(apierr.ErrUserNotFound)
	}

	res := response.UserData{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}

	if user.UpdatedAt.Valid {
		res.UpdatedAt = &user.UpdatedAt.Time
	}

	return &res, nil
}

func (u *uservice) SendOTP(ctx context.Context, email, lang string) error {
	if !otpPkg.CanResend(email) {
		return errors.New(apierr.ErrOTPCooldown)
	}

	existUser, err := userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	if existUser != nil {
		return errors.New(apierr.ErrUserAlreadyExists)
	}

	code := otpPkg.Generate(email)
	return emailPkg.SendOTP(email, code, lang)
}

// resetOTPKey returns the namespaced OTP key for password reset to avoid
// collisions with registration OTPs.
func resetOTPKey(email string) string {
	return "reset:" + email
}

func (u *uservice) ForgotPassword(ctx context.Context, email, lang string) error {
	if !otpPkg.CanResend(resetOTPKey(email)) {
		return errors.New(apierr.ErrOTPCooldown)
	}

	user, err := userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	// If user doesn't exist, return silently to avoid user enumeration.
	if user == nil {
		return nil
	}

	code := otpPkg.Generate(resetOTPKey(email))
	return emailPkg.SendPasswordResetOTP(email, code, lang)
}

func (u *uservice) ResetPassword(ctx context.Context, email, otpCode, newPassword string) error {
	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	if !otpPkg.Verify(resetOTPKey(email), otpCode) {
		return errors.New(apierr.ErrInvalidOTP)
	}

	user, err := userStore.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New(apierr.ErrUserNotExist)
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashed := string(encryptedPassword)
	_, err = userStore.UpdateUser(ctx, user.ID, store.UpdateUserInput{Password: &hashed})
	if err != nil {
		return err
	}

	otpPkg.Delete(resetOTPKey(email))
	return nil
}

func (u *uservice) GetUsers(ctx context.Context) ([]response.UserData, error) {
	users, err := userStore.GetUsers(ctx)
	if err != nil {
		return []response.UserData{}, err
	}

	usersData := []response.UserData{}
	for _, user := range users {
		res := response.UserData{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		}

		if user.UpdatedAt.Valid {
			t := user.UpdatedAt.Time
			res.UpdatedAt = &t
		}

		usersData = append(usersData, res)
	}

	return usersData, nil
}
