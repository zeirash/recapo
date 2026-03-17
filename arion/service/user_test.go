package service

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/database"
	otpPkg "github.com/zeirash/recapo/arion/common/otp"
	"github.com/zeirash/recapo/arion/common/response"
	mock_database "github.com/zeirash/recapo/arion/mock/database"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/store"
	"golang.org/x/crypto/bcrypt"
)

// due to circular dependency, we need to create a noop subscription service
type noopSubscriptionService struct{}

type errTrialSubscriptionService struct{ noopSubscriptionService }

func (e *errTrialSubscriptionService) CreateTrialSubscription(context.Context, int) error {
	return errors.New("trial error")
}

func (n *noopSubscriptionService) GetActivePlans(context.Context) ([]response.PlanData, error) { return nil, nil }
func (n *noopSubscriptionService) GetSubscriptionByShopID(context.Context, int) (*response.SubscriptionData, error) {
	return nil, nil
}
func (n *noopSubscriptionService) CreateTrialSubscription(context.Context, int) error             { return nil }
func (n *noopSubscriptionService) Checkout(context.Context, int, int) (*response.CheckoutData, error) { return nil, nil }
func (n *noopSubscriptionService) HandleMidtransWebhook(context.Context, MidtransWebhookPayload) error { return nil }
func (n *noopSubscriptionService) IsSubscriptionActive(context.Context, int) (bool, error)        { return false, nil }
func (n *noopSubscriptionService) CancelSubscription(context.Context, int) error                  { return nil }
func (n *noopSubscriptionService) ExpireSubscriptions(context.Context) error                    { return nil }

func Test_uservice_UserLogin(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	// Create a hashed password for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	type input struct {
		email    string
		password string
	}

	tests := []struct {
		name       string
		input      input
		mockSetup  func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore)
		wantResult response.TokenResponse
		wantErr    bool
	}{
		{
			name: "successfully login user",
			input: input{
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				user := &model.User{
					ID:        1,
					ShopID:    10,
					Name:      "John Doe",
					Email:     "john@example.com",
					Password:  string(hashedPassword),
					Role:      "admin",
					CreatedAt: fixedTime,
				}

				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(user, nil)

				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)

				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), user, gomock.Any(), 2).
					Return("access_token_123", nil)

				mockToken.EXPECT().
					CreateRefreshToken(gomock.Any(), user, gomock.Any(), 168).
					Return("refresh_token_123", nil)

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_123",
			},
			wantErr: false,
		},
		{
			name: "login user not found returns error",
			input: input{
				email:    "notfound@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "notfound@example.com").
					Return(nil, nil)

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "login with wrong password returns error",
			input: input{
				email:    "john@example.com",
				password: "wrongpassword",
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(&model.User{
						ID:       1,
						Email:    "john@example.com",
						Password: string(hashedPassword),
					}, nil)

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "login returns error on database failure",
			input: input{
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, errors.New("database error"))

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "login returns error when access token creation fails",
			input: input{
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				user := &model.User{
					ID:       1,
					Email:    "john@example.com",
					Password: string(hashedPassword),
				}

				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(user, nil)

				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)

				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), user, gomock.Any(), 2).
					Return("", errors.New("token error"))

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "login returns error when refresh token creation fails",
			input: input{
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				user := &model.User{
					ID:       1,
					Email:    "john@example.com",
					Password: string(hashedPassword),
				}

				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(user, nil)

				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)

				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), user, gomock.Any(), 2).
					Return("access_token_123", nil)

				mockToken.EXPECT().
					CreateRefreshToken(gomock.Any(), user, gomock.Any(), 168).
					Return("", errors.New("refresh token error"))

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldUserStore, oldTokenStore := userStore, tokenStore
			defer func() { userStore, tokenStore = oldUserStore, oldTokenStore }()

			mockUser, mockToken := tt.mockSetup(ctrl)
			userStore = mockUser
			tokenStore = mockToken
			cfg = config.Config{SecretKey: "testsecret"}

			var u uservice
			got, gotErr := u.UserLogin(context.Background(), tt.input.email, tt.input.password)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UserLogin() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UserLogin() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UserLogin() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_uservice_RefreshToken(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name         string
		refreshToken string
		mockSetup    func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore)
		wantResult   response.TokenResponse
		wantErr      bool
	}{
		{
			name:         "successfully refresh token",
			refreshToken: "valid_refresh_token",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				user := &model.User{
					ID:        1,
					ShopID:    10,
					Name:      "John Doe",
					Email:     "john@example.com",
					Role:      "admin",
					CreatedAt: fixedTime,
				}

				mockToken.EXPECT().
					ExtractDataFromToken(gomock.Any(), "valid_refresh_token", gomock.Any()).
					Return(model.TokenData{UserID: 1, ShopID: 10}, nil)

				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(user, nil)

				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)

				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), user, gomock.Any(), 2).
					Return("new_access_token", nil)

				mockToken.EXPECT().
					CreateRefreshToken(gomock.Any(), user, gomock.Any(), 168).
					Return("new_refresh_token", nil)

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
			},
			wantErr: false,
		},
		{
			name:         "refresh token with mismatched session returns error",
			refreshToken: "valid_refresh_token",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				mockToken.EXPECT().
					ExtractDataFromToken(gomock.Any(), "valid_refresh_token", gomock.Any()).
					Return(model.TokenData{UserID: 1, ShopID: 10, SessionToken: "old_token"}, nil)

				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(&model.User{
						ID:           1,
						ShopID:       10,
						SessionToken: sql.NullString{String: "new_token", Valid: true},
					}, nil)

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr:    true,
		},
		{
			name:         "refresh token with invalid token returns error",
			refreshToken: "invalid_token",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				mockToken.EXPECT().
					ExtractDataFromToken(gomock.Any(), "invalid_token", gomock.Any()).
					Return(model.TokenData{}, errors.New("invalid token"))

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name:         "refresh token with user not found returns error",
			refreshToken: "valid_refresh_token",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				mockToken.EXPECT().
					ExtractDataFromToken(gomock.Any(), "valid_refresh_token", gomock.Any()).
					Return(model.TokenData{UserID: 9999, ShopID: 10}, nil)

				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 9999).
					Return(nil, nil)

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name:         "refresh token returns error on database failure",
			refreshToken: "valid_refresh_token",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				mockToken.EXPECT().
					ExtractDataFromToken(gomock.Any(), "valid_refresh_token", gomock.Any()).
					Return(model.TokenData{UserID: 1, ShopID: 10}, nil)

				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(nil, errors.New("database error"))

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name:         "refresh token returns error when access token creation fails",
			refreshToken: "valid_refresh_token",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				user := &model.User{
					ID:     1,
					ShopID: 10,
				}

				mockToken.EXPECT().
					ExtractDataFromToken(gomock.Any(), "valid_refresh_token", gomock.Any()).
					Return(model.TokenData{UserID: 1, ShopID: 10}, nil)

				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(user, nil)

				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)

				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), user, gomock.Any(), 2).
					Return("", errors.New("token error"))

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name:         "refresh token returns error when refresh token creation fails",
			refreshToken: "valid_refresh_token",
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockUserStore, *mock_store.MockTokenStore) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockToken := mock_store.NewMockTokenStore(ctrl)

				user := &model.User{
					ID:     1,
					ShopID: 10,
				}

				mockToken.EXPECT().
					ExtractDataFromToken(gomock.Any(), "valid_refresh_token", gomock.Any()).
					Return(model.TokenData{UserID: 1, ShopID: 10}, nil)

				mockUser.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(user, nil)

				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)

				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), user, gomock.Any(), 2).
					Return("new_access_token", nil)

				mockToken.EXPECT().
					CreateRefreshToken(gomock.Any(), user, gomock.Any(), 168).
					Return("", errors.New("refresh token error"))

				return mockUser, mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldUserStore, oldTokenStore := userStore, tokenStore
			defer func() { userStore, tokenStore = oldUserStore, oldTokenStore }()

			mockUser, mockToken := tt.mockSetup(ctrl)
			userStore = mockUser
			tokenStore = mockToken
			cfg = config.Config{SecretKey: "testsecret"}

			var u uservice
			got, gotErr := u.RefreshToken(context.Background(), tt.refreshToken)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("RefreshToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("RefreshToken() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("RefreshToken() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_uservice_UserRegister(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	type input struct {
		name     string
		email    string
		password string
	}

	tests := []struct {
		name       string
		input      input
		mockSetup  func(ctrl *gomock.Controller)
		wantResult response.TokenResponse
		wantErr    bool
	}{
		{
			name: "register user already exists returns error",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(&model.User{ID: 1, Email: "john@example.com"}, nil)
				userStore = mockUser
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "register returns error on database lookup failure",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, errors.New("database error"))
				userStore = mockUser
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "register returns error on db.Begin failure",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, nil)
				userStore = mockUser

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().
					Begin().
					Return(nil, errors.New("transaction error"))
				dbGetter = func() database.DB { return mockDB }
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "register returns error on CreateShop failure",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, nil)
				userStore = mockUser

				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().
					Begin().
					Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					CreateShop(gomock.Any(), mockTx, "John Doe's Shop").
					Return(nil, errors.New("shop creation error"))
				shopStore = mockShop
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "register returns error on CreateUser failure",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, nil)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "John Doe", "john@example.com", gomock.Any(), "owner", 1).
					Return(nil, errors.New("user creation error"))
				userStore = mockUser

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().
					Begin().
					Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					CreateShop(gomock.Any(), mockTx, "John Doe's Shop").
					Return(&model.Shop{ID: 1, Name: "John Doe's Shop", CreatedAt: fixedTime}, nil)
				shopStore = mockShop
			},
			wantResult: response.TokenResponse{},
			wantErr:    true,
		},
		{
			name: "register returns error on tx.Commit failure",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(errors.New("commit error"))
				mockTx.EXPECT().Rollback().Return(nil)

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, nil)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "John Doe", "john@example.com", gomock.Any(), "owner", 1).
					Return(&model.User{ID: 1, ShopID: 1, Name: "John Doe", Email: "john@example.com", Role: "owner", CreatedAt: fixedTime}, nil)
				userStore = mockUser

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().
					Begin().
					Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					CreateShop(gomock.Any(), mockTx, "John Doe's Shop").
					Return(&model.Shop{ID: 1, Name: "John Doe's Shop", CreatedAt: fixedTime}, nil)
				shopStore = mockShop
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "register returns error on CreateAccessToken failure",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, nil)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "John Doe", "john@example.com", gomock.Any(), "owner", 1).
					Return(&model.User{ID: 1, ShopID: 1, Name: "John Doe", Email: "john@example.com", Role: "owner", CreatedAt: fixedTime}, nil)
				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)
				userStore = mockUser

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().
					Begin().
					Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					CreateShop(gomock.Any(), mockTx, "John Doe's Shop").
					Return(&model.Shop{ID: 1, Name: "John Doe's Shop", CreatedAt: fixedTime}, nil)
				shopStore = mockShop

				mockToken := mock_store.NewMockTokenStore(ctrl)
				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), gomock.Any(), "testsecret", 2).
					Return("", errors.New("token error"))
				tokenStore = mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "register returns error on CreateRefreshToken failure",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, nil)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "John Doe", "john@example.com", gomock.Any(), "owner", 1).
					Return(&model.User{ID: 1, ShopID: 1, Name: "John Doe", Email: "john@example.com", Role: "owner", CreatedAt: fixedTime}, nil)
				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)
				userStore = mockUser

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().
					Begin().
					Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					CreateShop(gomock.Any(), mockTx, "John Doe's Shop").
					Return(&model.Shop{ID: 1, Name: "John Doe's Shop", CreatedAt: fixedTime}, nil)
				shopStore = mockShop

				mockToken := mock_store.NewMockTokenStore(ctrl)
				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), gomock.Any(), "testsecret", 2).
					Return("access-token", nil)
				mockToken.EXPECT().
					CreateRefreshToken(gomock.Any(), gomock.Any(), "testsecret", 168).
					Return("", errors.New("refresh token error"))
				tokenStore = mockToken
			},
			wantResult: response.TokenResponse{},
			wantErr: true,
		},
		{
			name: "successfully register user",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, nil)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "John Doe", "john@example.com", gomock.Any(), "owner", 1).
					Return(&model.User{ID: 1, ShopID: 1, Name: "John Doe", Email: "john@example.com", Role: "owner", CreatedAt: fixedTime}, nil)
				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)
				userStore = mockUser

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().
					Begin().
					Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					CreateShop(gomock.Any(), mockTx, "John Doe's Shop").
					Return(&model.Shop{ID: 1, Name: "John Doe's Shop", CreatedAt: fixedTime}, nil)
				shopStore = mockShop

				mockToken := mock_store.NewMockTokenStore(ctrl)
				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), gomock.Any(), "testsecret", 2).
					Return("access-token", nil)
				mockToken.EXPECT().
					CreateRefreshToken(gomock.Any(), gomock.Any(), "testsecret", 168).
					Return("refresh-token", nil)
				tokenStore = mockToken
			},
			wantResult: response.TokenResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			},
			wantErr: false,
		},
		{
			name: "register succeeds even when CreateTrialSubscription returns error",
			input: input{
				name:     "John Doe",
				email:    "john@example.com",
				password: "password123",
			},
			mockSetup: func(ctrl *gomock.Controller) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)

				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					GetUserByEmail(gomock.Any(), "john@example.com").
					Return(nil, nil)
				mockUser.EXPECT().
					CreateUser(gomock.Any(), mockTx, "John Doe", "john@example.com", gomock.Any(), "owner", 1).
					Return(&model.User{ID: 1, ShopID: 1, Name: "John Doe", Email: "john@example.com", Role: "owner", CreatedAt: fixedTime}, nil)
				mockUser.EXPECT().
					SetSessionToken(gomock.Any(), 1, gomock.Any()).
					Return(nil)
				userStore = mockUser

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().
					Begin().
					Return(mockTx, nil)
				dbGetter = func() database.DB { return mockDB }

				mockShop := mock_store.NewMockShopStore(ctrl)
				mockShop.EXPECT().
					CreateShop(gomock.Any(), mockTx, "John Doe's Shop").
					Return(&model.Shop{ID: 1, Name: "John Doe's Shop", CreatedAt: fixedTime}, nil)
				shopStore = mockShop

				mockToken := mock_store.NewMockTokenStore(ctrl)
				mockToken.EXPECT().
					CreateAccessToken(gomock.Any(), gomock.Any(), "testsecret", 2).
					Return("access-token", nil)
				mockToken.EXPECT().
					CreateRefreshToken(gomock.Any(), gomock.Any(), "testsecret", 168).
					Return("refresh-token", nil)
				tokenStore = mockToken

				subscriptionService = &errTrialSubscriptionService{}
			},
			wantResult: response.TokenResponse{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldUserStore, oldShopStore, oldTokenStore := userStore, shopStore, tokenStore
			oldDBGetter := dbGetter
			oldSubscriptionService := subscriptionService
			defer func() {
				userStore, shopStore, tokenStore = oldUserStore, oldShopStore, oldTokenStore
				dbGetter = oldDBGetter
				subscriptionService = oldSubscriptionService
			}()
			subscriptionService = &noopSubscriptionService{}

			tt.mockSetup(ctrl)
			cfg = config.Config{SecretKey: "testsecret"}

			var u uservice
			got, gotErr := u.UserRegister(context.Background(), tt.input.name, tt.input.email, tt.input.password)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UserRegister() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UserRegister() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UserRegister() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_uservice_UpdateUser(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		input      UpdateUserInput
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockUserStore
		wantResult response.UserData
		wantErr    bool
	}{
		{
			name: "successfully update user name",
			input: UpdateUserInput{
				ID:   1,
				Name: strPtr("Updated Name"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(&model.User{ID: 1, Name: "John Doe", Email: "john@example.com"}, nil)
				mock.EXPECT().
					UpdateUser(gomock.Any(), 1, store.UpdateUserInput{Name: strPtr("Updated Name")}).
					Return(&model.User{
						ID:        1,
						Name:      "Updated Name",
						Email:     "john@example.com",
						CreatedAt: fixedTime,
						UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.UserData{
				ID:        1,
				Name:      "Updated Name",
				Email:     "john@example.com",
				CreatedAt: fixedTime,
				UpdatedAt: &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "successfully update user with password",
			input: UpdateUserInput{
				ID:       1,
				Name:     strPtr("Updated Name"),
				Password: strPtr("newpassword123"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(&model.User{ID: 1, Name: "John Doe", Email: "john@example.com"}, nil)
				// Use gomock.Any() since bcrypt generates different hashes each time
				mock.EXPECT().
					UpdateUser(gomock.Any(), 1, gomock.Any()).
					Return(&model.User{
						ID:        1,
						Name:      "Updated Name",
						Email:     "john@example.com",
						CreatedAt: fixedTime,
						UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.UserData{
				ID:        1,
				Name:      "Updated Name",
				Email:     "john@example.com",
				CreatedAt: fixedTime,
				UpdatedAt: &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "update user not found returns error",
			input: UpdateUserInput{
				ID:   9999,
				Name: strPtr("Updated Name"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByID(gomock.Any(), 9999).
					Return(nil, nil)
				return mock
			},
			wantResult: response.UserData{},
			wantErr: true,
		},
		{
			name: "update user returns error on get failure",
			input: UpdateUserInput{
				ID:   1,
				Name: strPtr("Updated Name"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.UserData{},
			wantErr:    true,
		},
		{
			name: "update user returns error on update failure",
			input: UpdateUserInput{
				ID:   1,
				Name: strPtr("Updated Name"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(&model.User{ID: 1, Name: "John Doe"}, nil)
				mock.EXPECT().
					UpdateUser(gomock.Any(), 1, store.UpdateUserInput{Name: strPtr("Updated Name")}).
					Return(nil, errors.New("update error"))
				return mock
			},
			wantResult: response.UserData{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := userStore
			defer func() { userStore = oldStore }()
			userStore = tt.mockSetup(ctrl)

			var u uservice
			got, gotErr := u.UpdateUser(context.Background(), tt.input)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateUser() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateUser() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateUser() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_uservice_GetUserByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		userID     int
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockUserStore
		wantResult *response.UserData
		wantErr    bool
	}{
		{
			name:   "successfully get user by ID",
			userID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(&model.User{
						ID:        1,
						Name:      "John Doe",
						Email:     "john@example.com",
						CreatedAt: fixedTime,
						UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: &response.UserData{
				ID:        1,
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: fixedTime,
				UpdatedAt: &updatedTime,
			},
			wantErr: false,
		},
		{
			name:   "get user by ID not found returns error",
			userID: 9999,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByID(gomock.Any(), 9999).
					Return(nil, nil)
				return mock
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get user by ID returns error on database failure",
			userID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByID(gomock.Any(), 1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: &response.UserData{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := userStore
			defer func() { userStore = oldStore }()
			userStore = tt.mockSetup(ctrl)

			var u uservice
			got, gotErr := u.GetUserByID(context.Background(), tt.userID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetUserByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetUserByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetUserByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_uservice_GetUsers(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockUserStore
		wantResult []response.UserData
		wantErr    bool
	}{
		{
			name: "successfully get users",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUsers(gomock.Any()).
					Return([]model.User{
						{ID: 1, Name: "John Doe", Email: "john@example.com", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{Time: fixedTime, Valid: true}},
						{ID: 2, Name: "Jane Doe", Email: "jane@example.com", CreatedAt: fixedTime},
					}, nil)
				return mock
			},
			wantResult: []response.UserData{
				{ID: 1, Name: "John Doe", Email: "john@example.com", CreatedAt: fixedTime, UpdatedAt: &fixedTime},
				{ID: 2, Name: "Jane Doe", Email: "jane@example.com", CreatedAt: fixedTime},
			},
			wantErr: false,
		},
		{
			name: "get users returns empty slice",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUsers(gomock.Any()).
					Return([]model.User{}, nil)
				return mock
			},
			wantResult: []response.UserData{},
			wantErr:    false,
		},
		{
			name: "get users returns error on database failure",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUsers(gomock.Any()).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: []response.UserData{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := userStore
			defer func() { userStore = oldStore }()
			userStore = tt.mockSetup(ctrl)

			var u uservice
			got, gotErr := u.GetUsers(context.Background())

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetUsers() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetUsers() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetUsers() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_uservice_SendOTP(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockUserStore
		wantErr   bool
	}{
		{
			name:  "successfully sends OTP when user does not exist",
			email: "new@example.com",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByEmail(gomock.Any(), "new@example.com").
					Return(nil, nil)
				return mock
			},
			wantErr: false,
		},
		{
			name:  "returns error when user already exists",
			email: "existing@example.com",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByEmail(gomock.Any(), "existing@example.com").
					Return(&model.User{ID: 1, Email: "existing@example.com"}, nil)
				return mock
			},
			wantErr: true,
		},
		{
			name:  "returns error on database failure",
			email: "new@example.com",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByEmail(gomock.Any(), "new@example.com").
					Return(nil, errors.New("database error"))
				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := userStore
			defer func() { userStore = oldStore }()
			userStore = tt.mockSetup(ctrl)

			var u uservice
			gotErr := u.SendOTP(context.Background(), tt.email, "en")

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("SendOTP() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_uservice_ForgotPassword(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockUserStore
		wantErr   bool
	}{
		{
			name:  "successfully sends reset OTP when user exists",
			email: "user@example.com",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&model.User{ID: 1, Email: "user@example.com"}, nil)
				return mock
			},
			wantErr: false,
		},
		{
			name:  "silently succeeds when user does not exist",
			email: "unknown@example.com",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByEmail(gomock.Any(), "unknown@example.com").
					Return(nil, nil)
				return mock
			},
			wantErr: false,
		},
		{
			name:  "returns error on database failure",
			email: "user@example.com",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(nil, errors.New("database error"))
				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := userStore
			defer func() { userStore = oldStore }()
			userStore = tt.mockSetup(ctrl)

			var u uservice
			gotErr := u.ForgotPassword(context.Background(), tt.email, "en")

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("ForgotPassword() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_uservice_ResetPassword(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		email     string
		password  string
		otpSetup  func(email string) string // seeds OTP, returns code
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockUserStore
		wantErr   bool
	}{
		{
			name:     "successfully resets password",
			email:    "user@example.com",
			password: "newpassword123",
			otpSetup: func(email string) string {
				return otpPkg.Generate(resetOTPKey(email))
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&model.User{ID: 1, Email: "user@example.com", CreatedAt: fixedTime}, nil)
				mock.EXPECT().
					UpdateUser(gomock.Any(), 1, gomock.Any()).
					Return(&model.User{ID: 1, Email: "user@example.com", CreatedAt: fixedTime}, nil)
				return mock
			},
			wantErr: false,
		},
		{
			name:     "returns error on invalid OTP",
			email:    "user@example.com",
			password: "newpassword123",
			otpSetup: func(email string) string {
				otpPkg.Generate(resetOTPKey(email))
				return "000000" // wrong code
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				return mock_store.NewMockUserStore(ctrl)
			},
			wantErr: true,
		},
		{
			name:     "returns error when user not found after OTP verification",
			email:    "user@example.com",
			password: "newpassword123",
			otpSetup: func(email string) string {
				return otpPkg.Generate(resetOTPKey(email))
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(nil, nil)
				return mock
			},
			wantErr: true,
		},
		{
			name:     "returns error on database failure during update",
			email:    "user@example.com",
			password: "newpassword123",
			otpSetup: func(email string) string {
				return otpPkg.Generate(resetOTPKey(email))
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mock := mock_store.NewMockUserStore(ctrl)
				mock.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&model.User{ID: 1, Email: "user@example.com", CreatedAt: fixedTime}, nil)
				mock.EXPECT().
					UpdateUser(gomock.Any(), 1, gomock.Any()).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := userStore
			defer func() { userStore = oldStore }()
			userStore = tt.mockSetup(ctrl)

			otpCode := tt.otpSetup(tt.email)

			var u uservice
			gotErr := u.ResetPassword(context.Background(), tt.email, otpCode, tt.password)

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("ResetPassword() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_uservice_Logout(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockUserStore
		wantErr   bool
	}{
		{
			name:   "successfully logout user",
			userID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					ClearSessionToken(gomock.Any(), 1).
					Return(nil)
				return mockUser
			},
			wantErr: false,
		},
		{
			name:   "logout returns error when ClearSessionToken fails",
			userID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockUserStore {
				mockUser := mock_store.NewMockUserStore(ctrl)
				mockUser.EXPECT().
					ClearSessionToken(gomock.Any(), 1).
					Return(errors.New("db error"))
				return mockUser
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := userStore
			defer func() { userStore = oldStore }()
			userStore = tt.mockSetup(ctrl)

			var u uservice
			gotErr := u.Logout(context.Background(), tt.userID)

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Logout() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}
