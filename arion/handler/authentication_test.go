package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
)

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetUserService()
	defer handler.SetUserService(oldService)

	mockUserService := mock_service.NewMockUserService(ctrl)
	handler.SetUserService(mockUserService)

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name: "successfully login",
			body: map[string]interface{}{
				"email":    "user@example.com",
				"password": "password123",
			},
			mockSetup: func() {
				mockUserService.EXPECT().
					UserLogin("user@example.com", "password123").
					Return(response.TokenResponse{
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "login returns 401 on incorrect password",
			body: map[string]interface{}{
				"email":    "user@example.com",
				"password": "wrongpassword",
			},
			mockSetup: func() {
				mockUserService.EXPECT().
					UserLogin("user@example.com", "wrongpassword").
					Return(response.TokenResponse{}, errors.New("password incorrect"))
			},
			wantStatus:     http.StatusUnauthorized,
			wantSuccess:    false,
			wantErrMessage: "password incorrect",
		},
		{
			name: "login returns 500 on service error",
			body: map[string]interface{}{
				"email":    "user@example.com",
				"password": "password123",
			},
			mockSetup: func() {
				mockUserService.EXPECT().
					UserLogin("user@example.com", "password123").
					Return(response.TokenResponse{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:       "login returns 400 on invalid json",
			body:       "invalid json",
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "login returns 400 on validation failure - missing email",
			body: map[string]interface{}{
				"email":    "",
				"password": "password123",
			},
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "login returns 400 on validation failure - missing password",
			body: map[string]interface{}{
				"email":    "user@example.com",
				"password": "",
			},
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler.LoginHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("LoginHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("LoginHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("LoginHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetUserService()
	defer handler.SetUserService(oldService)

	mockUserService := mock_service.NewMockUserService(ctrl)
	handler.SetUserService(mockUserService)

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name: "successfully register",
			body: map[string]interface{}{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "password123",
			},
			mockSetup: func() {
				mockUserService.EXPECT().
					UserRegister("John Doe", "john@example.com", "password123").
					Return(response.TokenResponse{
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "register returns 500 on service error",
			body: map[string]interface{}{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "password123",
			},
			mockSetup: func() {
				mockUserService.EXPECT().
					UserRegister("John Doe", "john@example.com", "password123").
					Return(response.TokenResponse{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:       "register returns 400 on invalid json",
			body:       "invalid json",
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "register returns 400 on validation failure - missing name",
			body: map[string]interface{}{
				"name":     "",
				"email":    "john@example.com",
				"password": "password123",
			},
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "register returns 400 on validation failure - missing email",
			body: map[string]interface{}{
				"name":     "John Doe",
				"email":    "",
				"password": "password123",
			},
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "register returns 400 on validation failure - invalid email",
			body: map[string]interface{}{
				"name":     "John Doe",
				"email":    "invalid-email",
				"password": "password123",
			},
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "register returns 400 on validation failure - missing password",
			body: map[string]interface{}{
				"name":     "John Doe",
				"email":    "john@example.com",
				"password": "",
			},
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler.RegisterHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("RegisterHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("RegisterHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("RegisterHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestRefreshHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetUserService()
	defer handler.SetUserService(oldService)

	mockUserService := mock_service.NewMockUserService(ctrl)
	handler.SetUserService(mockUserService)

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name: "successfully refresh token",
			body: map[string]interface{}{
				"refresh_token": "valid-refresh-token",
			},
			mockSetup: func() {
				mockUserService.EXPECT().
					RefreshToken("valid-refresh-token").
					Return(response.TokenResponse{
						AccessToken:  "new-access-token",
						RefreshToken: "new-refresh-token",
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "refresh returns 401 on invalid token",
			body: map[string]interface{}{
				"refresh_token": "invalid-token",
			},
			mockSetup: func() {
				mockUserService.EXPECT().
					RefreshToken("invalid-token").
					Return(response.TokenResponse{}, errors.New("token expired"))
			},
			wantStatus:     http.StatusUnauthorized,
			wantSuccess:    false,
			wantErrMessage: "token expired",
		},
		{
			name:       "refresh returns 400 on invalid json",
			body:       "invalid json",
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "refresh returns 400 when refresh_token missing",
			body: map[string]interface{}{
				"refresh_token": "",
			},
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "refresh returns 400 when refresh_token not in body",
			body: map[string]interface{}{
				"other_field": "value",
			},
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch b := tt.body.(type) {
			case string:
				bodyBytes = []byte(b)
			default:
				bodyBytes, _ = json.Marshal(b)
			}

			req := httptest.NewRequest("POST", "/refresh", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler.RefreshHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("RefreshHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("RefreshHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("RefreshHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}
