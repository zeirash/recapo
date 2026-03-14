package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/otp"
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
					Return(response.TokenResponse{}, errors.New(apierr.ErrPasswordIncorrect))
			},
			wantStatus:     http.StatusUnauthorized,
			wantSuccess:    false,
			wantErrMessage: "Password is incorrect",
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
			name:        "login returns 400 on invalid json",
			body:        "invalid json",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "login returns 400 on validation failure - missing email",
			body: map[string]interface{}{
				"email":    "",
				"password": "password123",
			},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "login returns 400 on validation failure - missing password",
			body: map[string]interface{}{
				"email":    "user@example.com",
				"password": "",
			},
			mockSetup:   func() {},
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

func TestSendOTPHandler(t *testing.T) {
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
			name:      "successfully send OTP",
			body:      map[string]interface{}{"email": "john@example.com"},
			mockSetup: func() { mockUserService.EXPECT().SendOTP("john@example.com", gomock.Any()).Return(nil) },
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:      "returns 400 when email already registered",
			body:      map[string]interface{}{"email": "existing@example.com"},
			mockSetup: func() {
				mockUserService.EXPECT().SendOTP("existing@example.com", gomock.Any()).Return(errors.New(apierr.ErrUserAlreadyExists))
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "returns 500 on service error",
			body:      map[string]interface{}{"email": "john@example.com"},
			mockSetup: func() {
				mockUserService.EXPECT().SendOTP("john@example.com", gomock.Any()).Return(errors.New("smtp error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on invalid json",
			body:        "invalid json",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on missing email",
			body:        map[string]interface{}{"email": ""},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on invalid email format",
			body:        map[string]interface{}{"email": "not-an-email"},
			mockSetup:   func() {},
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

			req := httptest.NewRequest("POST", "/send_otp", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler.SendOTPHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("SendOTPHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("SendOTPHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("SendOTPHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
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
		bodyFn         func(otpCode string) []byte
		otpSetup       func() string // seeds OTP store, returns the code
		mockSetup      func(otpCode string)
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name: "successfully register",
			otpSetup: func() string {
				return otp.Generate("john@example.com")
			},
			mockSetup: func(otpCode string) {
				mockUserService.EXPECT().
					UserRegister("John Doe", "john@example.com", "password123").
					Return(response.TokenResponse{
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
					}, nil)
			},
			bodyFn: func(otpCode string) []byte {
				b, _ := json.Marshal(map[string]interface{}{
					"name":     "John Doe",
					"email":    "john@example.com",
					"password": "password123",
					"otp":      otpCode,
				})
				return b
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "register returns 500 on service error",
			otpSetup: func() string {
				return otp.Generate("john@example.com")
			},
			mockSetup: func(otpCode string) {
				mockUserService.EXPECT().
					UserRegister("John Doe", "john@example.com", "password123").
					Return(response.TokenResponse{}, errors.New("database error"))
			},
			bodyFn: func(otpCode string) []byte {
				b, _ := json.Marshal(map[string]interface{}{
					"name":     "John Doe",
					"email":    "john@example.com",
					"password": "password123",
					"otp":      otpCode,
				})
				return b
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "database error",
		},
		{
			name:      "register returns 400 on invalid json",
			otpSetup:  func() string { return "" },
			mockSetup: func(_ string) {},
			bodyFn:    func(_ string) []byte { return []byte("invalid json") },
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "register returns 400 on validation failure - missing name",
			otpSetup:  func() string { return "" },
			mockSetup: func(_ string) {},
			bodyFn: func(_ string) []byte {
				b, _ := json.Marshal(map[string]interface{}{
					"name":     "",
					"email":    "john@example.com",
					"password": "password123",
					"otp":      "123456",
				})
				return b
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "register returns 400 on validation failure - missing email",
			otpSetup:  func() string { return "" },
			mockSetup: func(_ string) {},
			bodyFn: func(_ string) []byte {
				b, _ := json.Marshal(map[string]interface{}{
					"name":     "John Doe",
					"email":    "",
					"password": "password123",
					"otp":      "123456",
				})
				return b
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "register returns 400 on validation failure - invalid email",
			otpSetup:  func() string { return "" },
			mockSetup: func(_ string) {},
			bodyFn: func(_ string) []byte {
				b, _ := json.Marshal(map[string]interface{}{
					"name":     "John Doe",
					"email":    "invalid-email",
					"password": "password123",
					"otp":      "123456",
				})
				return b
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "register returns 400 on validation failure - missing password",
			otpSetup:  func() string { return "" },
			mockSetup: func(_ string) {},
			bodyFn: func(_ string) []byte {
				b, _ := json.Marshal(map[string]interface{}{
					"name":     "John Doe",
					"email":    "john@example.com",
					"password": "",
					"otp":      "123456",
				})
				return b
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "register returns 400 on validation failure - missing otp",
			otpSetup:  func() string { return "" },
			mockSetup: func(_ string) {},
			bodyFn: func(_ string) []byte {
				b, _ := json.Marshal(map[string]interface{}{
					"name":     "John Doe",
					"email":    "john@example.com",
					"password": "password123",
					"otp":      "",
				})
				return b
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:      "register returns 400 on invalid otp",
			otpSetup:  func() string { return otp.Generate("john@example.com") },
			mockSetup: func(_ string) {},
			bodyFn: func(_ string) []byte {
				b, _ := json.Marshal(map[string]interface{}{
					"name":     "John Doe",
					"email":    "john@example.com",
					"password": "password123",
					"otp":      "000000",
				})
				return b
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otpCode := tt.otpSetup()
			tt.mockSetup(otpCode)
			bodyBytes := tt.bodyFn(otpCode)

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

func TestForgotPasswordHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetUserService()
	defer handler.SetUserService(oldService)

	mockUserService := mock_service.NewMockUserService(ctrl)
	handler.SetUserService(mockUserService)

	tests := []struct {
		name        string
		body        interface{}
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:        "successfully sends reset OTP",
			body:        map[string]interface{}{"email": "user@example.com"},
			mockSetup:   func() { mockUserService.EXPECT().ForgotPassword("user@example.com", gomock.Any()).Return(nil) },
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 200 even when user not found (anti-enumeration)",
			body:        map[string]interface{}{"email": "unknown@example.com"},
			mockSetup:   func() { mockUserService.EXPECT().ForgotPassword("unknown@example.com", gomock.Any()).Return(nil) },
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 500 on service error",
			body:        map[string]interface{}{"email": "user@example.com"},
			mockSetup:   func() { mockUserService.EXPECT().ForgotPassword("user@example.com", gomock.Any()).Return(errors.New("smtp error")) },
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on invalid json",
			body:        "invalid json",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on missing email",
			body:        map[string]interface{}{"email": ""},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on invalid email format",
			body:        map[string]interface{}{"email": "not-an-email"},
			mockSetup:   func() {},
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

			req := httptest.NewRequest("POST", "/forgot_password", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler.ForgotPasswordHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("ForgotPasswordHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("ForgotPasswordHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
		})
	}
}

func TestResetPasswordHandler(t *testing.T) {
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
			name: "successfully resets password",
			body: map[string]interface{}{"email": "user@example.com", "otp": "123456", "password": "newpassword"},
			mockSetup: func() {
				mockUserService.EXPECT().ResetPassword("user@example.com", "123456", "newpassword").Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "returns 400 on invalid OTP",
			body: map[string]interface{}{"email": "user@example.com", "otp": "000000", "password": "newpassword"},
			mockSetup: func() {
				mockUserService.EXPECT().ResetPassword("user@example.com", "000000", "newpassword").Return(errors.New(apierr.ErrInvalidOTP))
			},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Invalid or expired verification code",
		},
		{
			name:        "returns 400 on invalid json",
			body:        "invalid json",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on missing email",
			body:        map[string]interface{}{"email": "", "otp": "123456", "password": "newpassword"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on invalid email format",
			body:        map[string]interface{}{"email": "not-an-email", "otp": "123456", "password": "newpassword"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on missing otp",
			body:        map[string]interface{}{"email": "user@example.com", "otp": "", "password": "newpassword"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 on missing password",
			body:        map[string]interface{}{"email": "user@example.com", "otp": "123456", "password": ""},
			mockSetup:   func() {},
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

			req := httptest.NewRequest("POST", "/reset_password", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler.ResetPasswordHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("ResetPasswordHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("ResetPasswordHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("ResetPasswordHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
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
			name:        "refresh returns 400 on invalid json",
			body:        "invalid json",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "refresh returns 400 when refresh_token missing",
			body: map[string]interface{}{
				"refresh_token": "",
			},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "refresh returns 400 when refresh_token not in body",
			body: map[string]interface{}{
				"other_field": "value",
			},
			mockSetup:   func() {},
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
func TestLogoutHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetUserService()
	defer handler.SetUserService(oldService)

	mockUserService := mock_service.NewMockUserService(ctrl)
	handler.SetUserService(mockUserService)

	tests := []struct {
		name        string
		userID      interface{}
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:   "successfully logout",
			userID: 1,
			mockSetup: func() {
				mockUserService.EXPECT().Logout(1).Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "logout returns 500 on service error",
			userID: 1,
			mockSetup: func() {
				mockUserService.EXPECT().Logout(1).Return(errors.New("db error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("POST", "/logout", nil)
			ctx := context.WithValue(req.Context(), common.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			handler.LogoutHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("LogoutHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("LogoutHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
		})
	}
}
