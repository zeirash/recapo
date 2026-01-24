package store

import (
	"reflect"
	"testing"

	"github.com/zeirash/recapo/arion/model"
)

func Test_token_CreateAccessToken(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		secret  string
		expiry  int
		wantErr bool
	}{
		{
			name: "successfully create access token",
			user: &model.User{
				ID:     1,
				ShopID: 10,
				Name:   "John Doe",
				Role:   "admin",
			},
			secret:  "testsecret",
			expiry:  1,
			wantErr: false,
		},
		{
			name: "create access token for system role",
			user: &model.User{
				ID:     1,
				ShopID: 10,
				Name:   "System User",
				Role:   "system",
			},
			secret:  "testsecret",
			expiry:  24,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var to token
			got, gotErr := to.CreateAccessToken(tt.user, tt.secret, tt.expiry)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateAccessToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateAccessToken() succeeded unexpectedly")
			}

			// Verify token is non-empty
			if got == "" {
				t.Error("CreateAccessToken() returned empty token")
			}

			// Verify the token is valid and can be authorized
			authorized, err := to.IsAuthorized(got, tt.secret)
			if err != nil {
				t.Errorf("Created token is not valid: %v", err)
			}
			if !authorized {
				t.Error("Created token is not authorized")
			}
		})
	}
}

func Test_token_CreateRefreshToken(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		secret  string
		expiry  int
		wantErr bool
	}{
		{
			name: "successfully create refresh token",
			user: &model.User{
				ID:     1,
				ShopID: 10,
				Name:   "John Doe",
				Role:   "admin",
			},
			secret:  "refreshsecret",
			expiry:  168,
			wantErr: false,
		},
		{
			name: "create refresh token for system role",
			user: &model.User{
				ID:     1,
				ShopID: 10,
				Name:   "System User",
				Role:   "system",
			},
			secret:  "refreshsecret",
			expiry:  720,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var to token
			got, gotErr := to.CreateRefreshToken(tt.user, tt.secret, tt.expiry)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateRefreshToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateRefreshToken() succeeded unexpectedly")
			}

			// Verify token is non-empty
			if got == "" {
				t.Error("CreateRefreshToken() returned empty token")
			}

			// Verify the token is valid and can be authorized
			authorized, err := to.IsAuthorized(got, tt.secret)
			if err != nil {
				t.Errorf("Created token is not valid: %v", err)
			}
			if !authorized {
				t.Error("Created token is not authorized")
			}
		})
	}
}

func Test_token_IsAuthorized(t *testing.T) {
	// Create a valid token for testing
	var to token
	user := &model.User{
		ID:     1,
		ShopID: 10,
		Name:   "John Doe",
		Role:   "admin",
	}
	validToken, _ := to.CreateAccessToken(user, "testsecret", 1)

	tests := []struct {
		name         string
		requestToken string
		secret       string
		want         bool
		wantErr      bool
	}{
		{
			name:         "valid token with correct secret",
			requestToken: validToken,
			secret:       "testsecret",
			want:         true,
			wantErr:      false,
		},
		{
			name:         "valid token with wrong secret",
			requestToken: validToken,
			secret:       "wrongsecret",
			want:         false,
			wantErr:      true,
		},
		{
			name:         "invalid token format",
			requestToken: "invalid.token.format",
			secret:       "testsecret",
			want:         false,
			wantErr:      true,
		},
		{
			name:         "empty token",
			requestToken: "",
			secret:       "testsecret",
			want:         false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := to.IsAuthorized(tt.requestToken, tt.secret)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("IsAuthorized() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("IsAuthorized() succeeded unexpectedly")
			}

			if got != tt.want {
				t.Errorf("IsAuthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_token_ExtractDataFromToken(t *testing.T) {
	var to token

	// Create valid tokens for testing
	adminUser := &model.User{
		ID:     1,
		ShopID: 10,
		Name:   "John Doe",
		Role:   "admin",
	}
	systemUser := &model.User{
		ID:     2,
		ShopID: 20,
		Name:   "System User",
		Role:   "system",
	}
	validAccessToken, _ := to.CreateAccessToken(adminUser, "testsecret", 1)
	systemAccessToken, _ := to.CreateAccessToken(systemUser, "testsecret", 1)
	validRefreshToken, _ := to.CreateRefreshToken(adminUser, "testsecret", 168)

	tests := []struct {
		name         string
		requestToken string
		secret       string
		want         model.TokenData
		wantErr      bool
	}{
		{
			name:         "extract data from access token",
			requestToken: validAccessToken,
			secret:       "testsecret",
			want: model.TokenData{
				UserID:     1,
				ShopID:     10,
				Name:       "John Doe",
				SystemMode: false,
			},
			wantErr: false,
		},
		{
			name:         "extract data from system user access token",
			requestToken: systemAccessToken,
			secret:       "testsecret",
			want: model.TokenData{
				UserID:     2,
				ShopID:     20,
				Name:       "System User",
				SystemMode: true,
			},
			wantErr: false,
		},
		{
			name:         "extract data from refresh token",
			requestToken: validRefreshToken,
			secret:       "testsecret",
			want: model.TokenData{
				UserID:     1,
				ShopID:     10,
				Name:       "",
				SystemMode: false,
			},
			wantErr: false,
		},
		{
			name:         "extract data with wrong secret",
			requestToken: validAccessToken,
			secret:       "wrongsecret",
			want:         model.TokenData{},
			wantErr:      true,
		},
		{
			name:         "extract data from invalid token",
			requestToken: "invalid.token",
			secret:       "testsecret",
			want:         model.TokenData{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := to.ExtractDataFromToken(tt.requestToken, tt.secret)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ExtractDataFromToken() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ExtractDataFromToken() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractDataFromToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
