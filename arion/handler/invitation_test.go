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
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
)

func newRequestWithUserAndShopID(method, path string, body []byte, userID, shopID int) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	ctx := context.WithValue(req.Context(), common.UserIDKey, userID)
	ctx = context.WithValue(ctx, common.ShopIDKey, shopID)
	return req.WithContext(ctx)
}

func TestInviteAdminHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetInvitationService()
	defer handler.SetInvitationService(oldService)

	mockInvitationService := mock_service.NewMockInvitationService(ctrl)
	handler.SetInvitationService(mockInvitationService)

	tests := []struct {
		name           string
		body           interface{}
		userID         int
		shopID         int
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:           "returns 400 when email is missing",
			body:           map[string]interface{}{},
			userID:         2,
			shopID:         1,
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Email is required",
		},
		{
			name:           "returns 400 when email is empty string",
			body:           map[string]interface{}{"email": ""},
			userID:         2,
			shopID:         1,
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Email is required",
		},
		{
			name:   "returns 400 on invalid email",
			body:   map[string]interface{}{"email": "invite@example.com"},
			userID: 2,
			shopID: 1,
			mockSetup: func() {
				mockInvitationService.EXPECT().
					InviteAdmin(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(errors.New(apierr.ErrEmailInvalid))
			},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Email is not valid",
		},
		{
			name:   "returns 403 when caller is not owner",
			body:   map[string]interface{}{"email": "invite@example.com"},
			userID: 2,
			shopID: 1,
			mockSetup: func() {
				mockInvitationService.EXPECT().
					InviteAdmin(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(errors.New(apierr.ErrNotOwner))
			},
			wantStatus:  http.StatusForbidden,
			wantSuccess: false,
		},
		{
			name:   "returns 403 when plan user limit reached",
			body:   map[string]interface{}{"email": "invite@example.com"},
			userID: 2,
			shopID: 1,
			mockSetup: func() {
				mockInvitationService.EXPECT().
					InviteAdmin(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(errors.New(apierr.ErrMaxUsersReached))
			},
			wantStatus:  http.StatusForbidden,
			wantSuccess: false,
		},
		{
			name:   "returns 409 when user already exists",
			body:   map[string]interface{}{"email": "invite@example.com"},
			userID: 2,
			shopID: 1,
			mockSetup: func() {
				mockInvitationService.EXPECT().
					InviteAdmin(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(errors.New(apierr.ErrUserAlreadyExists))
			},
			wantStatus:  http.StatusConflict,
			wantSuccess: false,
		},
		{
			name:   "returns 409 when invitation already sent",
			body:   map[string]interface{}{"email": "invite@example.com"},
			userID: 2,
			shopID: 1,
			mockSetup: func() {
				mockInvitationService.EXPECT().
					InviteAdmin(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(errors.New(apierr.ErrInvitationAlreadySent))
			},
			wantStatus:  http.StatusConflict,
			wantSuccess: false,
		},
		{
			name:   "returns 404 when shop not found",
			body:   map[string]interface{}{"email": "invite@example.com"},
			userID: 2,
			shopID: 1,
			mockSetup: func() {
				mockInvitationService.EXPECT().
					InviteAdmin(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(errors.New(apierr.ErrShopNotFound))
			},
			wantStatus:  http.StatusNotFound,
			wantSuccess: false,
		},
		{
			name:   "returns 500 on unexpected service error",
			body:   map[string]interface{}{"email": "invite@example.com"},
			userID: 2,
			shopID: 1,
			mockSetup: func() {
				mockInvitationService.EXPECT().
					InviteAdmin(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
		{
			name:   "successfully invite admin",
			body:   map[string]interface{}{"email": "invite@example.com"},
			userID: 2,
			shopID: 1,
			mockSetup: func() {
				mockInvitationService.EXPECT().
					InviteAdmin(gomock.Any(), 1, 2, "invite@example.com", gomock.Any()).
					Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := newRequestWithUserAndShopID("POST", "/invite", bodyBytes, tt.userID, tt.shopID)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.InviteAdminHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("InviteAdminHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("InviteAdminHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("InviteAdminHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestValidateInviteHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetInvitationService()
	defer handler.SetInvitationService(oldService)

	mockInvitationService := mock_service.NewMockInvitationService(ctrl)
	handler.SetInvitationService(mockInvitationService)

	tests := []struct {
		name        string
		token       string
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:        "returns 400 when token is missing",
			token:       "",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:  "returns 400 when invitation not found",
			token: "badtoken",
			mockSetup: func() {
				mockInvitationService.EXPECT().
					ValidateInviteToken(gomock.Any(), "badtoken").
					Return(nil, errors.New(apierr.ErrInvitationNotFound))
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:  "returns 400 when shop not found",
			token: "sometoken",
			mockSetup: func() {
				mockInvitationService.EXPECT().
					ValidateInviteToken(gomock.Any(), "sometoken").
					Return(nil, errors.New(apierr.ErrShopNotFound))
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:  "returns 500 on unexpected service error",
			token: "sometoken",
			mockSetup: func() {
				mockInvitationService.EXPECT().
					ValidateInviteToken(gomock.Any(), "sometoken").
					Return(nil, errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
		{
			name:  "successfully validate invite token",
			token: "validtoken",
			mockSetup: func() {
				mockInvitationService.EXPECT().
					ValidateInviteToken(gomock.Any(), "validtoken").
					Return(&response.InvitationData{
						Email:    "invite@example.com",
						ShopName: "My Shop",
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			path := "/invite/validate"
			if tt.token != "" {
				path += "?token=" + tt.token
			}
			req := httptest.NewRequest("GET", path, nil)
			rec := httptest.NewRecorder()

			handler.ValidateInviteHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("ValidateInviteHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("ValidateInviteHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
		})
	}
}

func TestAcceptInviteHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetInvitationService()
	defer handler.SetInvitationService(oldService)

	mockInvitationService := mock_service.NewMockInvitationService(ctrl)
	handler.SetInvitationService(mockInvitationService)

	tests := []struct {
		name           string
		body           interface{}
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:           "returns 400 when token is missing",
			body:           map[string]interface{}{"name": "New Admin", "password": "pass1234"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Invitation not found or already accepted",
		},
		{
			name:           "returns 400 when name is missing",
			body:           map[string]interface{}{"token": "validtoken", "password": "pass1234"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Name is required",
		},
		{
			name:           "returns 400 when password is missing",
			body:           map[string]interface{}{"token": "validtoken", "name": "New Admin"},
			mockSetup:      func() {},
			wantStatus:     http.StatusBadRequest,
			wantSuccess:    false,
			wantErrMessage: "Password is required",
		},
		{
			name:  "returns 400 when invitation not found",
			body:  map[string]interface{}{"token": "badtoken", "name": "New Admin", "password": "pass1234"},
			mockSetup: func() {
				mockInvitationService.EXPECT().
					AcceptInvite(gomock.Any(), "badtoken", "New Admin", "pass1234").
					Return(response.TokenResponse{}, errors.New(apierr.ErrInvitationNotFound))
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:  "returns 409 when invitation already accepted",
			body:  map[string]interface{}{"token": "usedtoken", "name": "New Admin", "password": "pass1234"},
			mockSetup: func() {
				mockInvitationService.EXPECT().
					AcceptInvite(gomock.Any(), "usedtoken", "New Admin", "pass1234").
					Return(response.TokenResponse{}, errors.New(apierr.ErrInvitationAlreadyAccepted))
			},
			wantStatus:  http.StatusConflict,
			wantSuccess: false,
		},
		{
			name:  "returns 400 when password is too weak",
			body:  map[string]interface{}{"token": "validtoken", "name": "New Admin", "password": "pass1234"},
			mockSetup: func() {
				mockInvitationService.EXPECT().
					AcceptInvite(gomock.Any(), "validtoken", "New Admin", "pass1234").
					Return(response.TokenResponse{}, errors.New(apierr.ErrPasswordTooWeak))
			},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:  "returns 403 when plan user limit reached",
			body:  map[string]interface{}{"token": "validtoken", "name": "New Admin", "password": "pass1234"},
			mockSetup: func() {
				mockInvitationService.EXPECT().
					AcceptInvite(gomock.Any(), "validtoken", "New Admin", "pass1234").
					Return(response.TokenResponse{}, errors.New(apierr.ErrMaxUsersReached))
			},
			wantStatus:  http.StatusForbidden,
			wantSuccess: false,
		},
		{
			name:  "returns 500 on unexpected service error",
			body:  map[string]interface{}{"token": "validtoken", "name": "New Admin", "password": "pass1234"},
			mockSetup: func() {
				mockInvitationService.EXPECT().
					AcceptInvite(gomock.Any(), "validtoken", "New Admin", "pass1234").
					Return(response.TokenResponse{}, errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
		{
			name: "successfully accept invite",
			body: map[string]interface{}{"token": "validtoken", "name": "New Admin", "password": "pass1234"},
			mockSetup: func() {
				mockInvitationService.EXPECT().
					AcceptInvite(gomock.Any(), "validtoken", "New Admin", "pass1234").
					Return(response.TokenResponse{
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest("POST", "/invite/accept", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.AcceptInviteHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("AcceptInviteHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("AcceptInviteHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("AcceptInviteHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}
