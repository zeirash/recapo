package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/middleware"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/store"
)

func TestAuthentication(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func(ctrl *gomock.Controller) *mock_store.MockTokenStore
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:       "missing Authorization header returns 401",
			authHeader: "",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockTokenStore {
				return mock_store.NewMockTokenStore(ctrl)
			},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:       "single token without Bearer prefix returns 401",
			authHeader: "sometoken",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockTokenStore {
				return mock_store.NewMockTokenStore(ctrl)
			},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:       "too many parts in header returns 401",
			authHeader: "Bearer token extra",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockTokenStore {
				return mock_store.NewMockTokenStore(ctrl)
			},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:       "IsAuthorized returns error returns 401",
			authHeader: "Bearer sometoken",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockTokenStore {
				mock := mock_store.NewMockTokenStore(ctrl)
				mock.EXPECT().
					IsAuthorized("sometoken", gomock.Any()).
					Return(false, errors.New("token expired"))
				return mock
			},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:       "IsAuthorized returns false returns 401",
			authHeader: "Bearer sometoken",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockTokenStore {
				mock := mock_store.NewMockTokenStore(ctrl)
				mock.EXPECT().
					IsAuthorized("sometoken", gomock.Any()).
					Return(false, nil)
				return mock
			},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:       "ExtractDataFromToken returns error returns 500",
			authHeader: "Bearer sometoken",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockTokenStore {
				mock := mock_store.NewMockTokenStore(ctrl)
				mock.EXPECT().
					IsAuthorized("sometoken", gomock.Any()).
					Return(true, nil)
				mock.EXPECT().
					ExtractDataFromToken("sometoken", gomock.Any()).
					Return(model.TokenData{}, errors.New("malformed claims"))
				return mock
			},
			wantStatus:     http.StatusInternalServerError,
			wantNextCalled: false,
		},
		{
			name:       "valid token passes through to next handler",
			authHeader: "Bearer sometoken",
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockTokenStore {
				mock := mock_store.NewMockTokenStore(ctrl)
				mock.EXPECT().
					IsAuthorized("sometoken", gomock.Any()).
					Return(true, nil)
				mock.EXPECT().
					ExtractDataFromToken("sometoken", gomock.Any()).
					Return(model.TokenData{UserID: 10, ShopID: 3, SystemMode: false}, nil)
				return mock
			},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			nextCalled = false

			oldFn := middleware.NewTokenStoreFunc
			defer func() { middleware.NewTokenStoreFunc = oldFn }()

			mockTS := tt.mockSetup(ctrl)
			middleware.NewTokenStoreFunc = func() store.TokenStore { return mockTS }

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				r.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			middleware.Authentication(next).ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if nextCalled != tt.wantNextCalled {
				t.Errorf("next called = %v, want %v", nextCalled, tt.wantNextCalled)
			}
			if tt.wantStatus != http.StatusOK {
				var body map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
					t.Fatalf("response body is not valid JSON: %v", err)
				}
				if body["success"] != false {
					t.Errorf("expected success=false in error response")
				}
			}
		})
	}
}

func TestCheckSystemMode(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		systemMode     bool
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "system mode true passes through",
			systemMode:     true,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
		{
			name:           "system mode false returns 401",
			systemMode:     false,
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled = false
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := context.WithValue(r.Context(), common.SystemModeKey, tt.systemMode)
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			middleware.CheckSystemMode(next).ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if nextCalled != tt.wantNextCalled {
				t.Errorf("next called = %v, want %v", nextCalled, tt.wantNextCalled)
			}
		})
	}
}
