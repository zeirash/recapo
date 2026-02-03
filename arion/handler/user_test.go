package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
	"github.com/zeirash/recapo/arion/service"
)

func newRequestWithUserID(method, path string, body []byte, userID int) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	return req.WithContext(context.WithValue(req.Context(), common.UserIDKey, userID))
}

func TestGetUserHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetUserService()
	defer handler.SetUserService(oldService)

	mockUserService := mock_service.NewMockUserService(ctrl)
	handler.SetUserService(mockUserService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		userID         int
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:   "successfully get user by ID",
			userID: 1,
			mockSetup: func() {
				mockUserService.EXPECT().
					GetUserByID(1).
					Return(&response.UserData{
						ID:        1,
						Name:      "John Doe",
						Email:     "john@example.com",
						CreatedAt: fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "get user returns 404 when not found",
			userID: 999,
			mockSetup: func() {
				mockUserService.EXPECT().
					GetUserByID(999).
					Return(nil, nil)
			},
			wantStatus:  http.StatusNotFound,
			wantSuccess: false,
		},
		{
			name:   "get user returns 500 on service error",
			userID: 1,
			mockSetup: func() {
				mockUserService.EXPECT().
					GetUserByID(1).
					Return(nil, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Failed to get user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := newRequestWithUserID("GET", "/user", nil, tt.userID)
			rec := httptest.NewRecorder()

			handler.GetUserHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetUserHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetUserHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("GetUserHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}

func TestGetUsersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetUserService()
	defer handler.SetUserService(oldService)

	mockUserService := mock_service.NewMockUserService(ctrl)
	handler.SetUserService(mockUserService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
		wantCount   int
	}{
		{
			name: "successfully get users list",
			mockSetup: func() {
				mockUserService.EXPECT().
					GetUsers().
					Return([]response.UserData{
						{ID: 1, Name: "John Doe", Email: "john@example.com", CreatedAt: fixedTime},
						{ID: 2, Name: "Jane Doe", Email: "jane@example.com", CreatedAt: fixedTime},
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
			wantCount:   2,
		},
		{
			name: "get users returns 500 on service error",
			mockSetup: func() {
				mockUserService.EXPECT().
					GetUsers().
					Return(nil, errors.New("database error"))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/users", nil)
			rec := httptest.NewRecorder()

			handler.GetUsersHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("GetUsersHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("GetUsersHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantCount > 0 {
				users, ok := resp.Data.([]interface{})
				if !ok || len(users) != tt.wantCount {
					t.Errorf("GetUsersHandler() data count = %v, want %v", len(users), tt.wantCount)
				}
			}
		})
	}
}

func TestUpdateUserHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldService := handler.GetUserService()
	defer handler.SetUserService(oldService)

	mockUserService := mock_service.NewMockUserService(ctrl)
	handler.SetUserService(mockUserService)

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name           string
		userID         int
		body           interface{}
		mockSetup      func()
		wantStatus     int
		wantSuccess    bool
		wantErrMessage string
	}{
		{
			name:   "successfully update user",
			userID: 1,
			body: map[string]interface{}{
				"name":  "Updated Name",
				"email": "updated@example.com",
			},
			mockSetup: func() {
				name := "Updated Name"
				email := "updated@example.com"
				mockUserService.EXPECT().
					UpdateUser(service.UpdateUserInput{
						ID:    1,
						Name:  &name,
						Email: &email,
					}).
					Return(response.UserData{
						ID:        1,
						Name:      "Updated Name",
						Email:     "updated@example.com",
						CreatedAt: fixedTime,
					}, nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:       "update user returns 400 on invalid json",
			userID:     1,
			body:       "invalid json",
			mockSetup:  func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:   "update user returns 500 on service error",
			userID: 1,
			body: map[string]interface{}{
				"name": "Test",
			},
			mockSetup: func() {
				mockUserService.EXPECT().
					UpdateUser(gomock.Any()).
					Return(response.UserData{}, errors.New("database error"))
			},
			wantStatus:     http.StatusInternalServerError,
			wantSuccess:    false,
			wantErrMessage: "Failed to update user",
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

			req := newRequestWithUserID("PATCH", "/user", bodyBytes, tt.userID)
			rec := httptest.NewRecorder()

			handler.UpdateUserHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("UpdateUserHandler() status = %v, want %v", rec.Code, tt.wantStatus)
			}

			var resp handler.ApiResponse
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.Success != tt.wantSuccess {
				t.Errorf("UpdateUserHandler() success = %v, want %v", resp.Success, tt.wantSuccess)
			}
			if tt.wantErrMessage != "" && resp.Message != tt.wantErrMessage {
				t.Errorf("UpdateUserHandler() message = %v, want %v", resp.Message, tt.wantErrMessage)
			}
		})
	}
}
