package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/handler"
)

func TestWriteJson(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       interface{}
		wantStatus int
	}{
		{
			name:       "writes 200 with body",
			status:     http.StatusOK,
			body:       map[string]string{"key": "value"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "writes custom status code",
			status:     http.StatusCreated,
			body:       "hello",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "writes nil body",
			status:     http.StatusOK,
			body:       nil,
			wantStatus: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.WriteJson(rec, tt.status, tt.body)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", rec.Code, tt.wantStatus)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", ct)
			}
			var resp handler.ApiResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("body is not valid JSON: %v", err)
			}
			if !resp.Success {
				t.Errorf("success = false, want true")
			}
		})
	}
}

func TestWriteErrorJson(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		err         error
		code        string
		acceptLang  string
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "writes status, success false, and code",
			status:      http.StatusBadRequest,
			err:         errors.New(apierr.ErrEmailRequired),
			code:        "validation",
			wantStatus:  http.StatusBadRequest,
			wantCode:    "validation",
			wantMessage: "Email is required",
		},
		{
			name:        "translates err key via en.json",
			status:      http.StatusUnauthorized,
			err:         errors.New(apierr.ErrPasswordIncorrect),
			code:        "user_login",
			wantStatus:  http.StatusUnauthorized,
			wantMessage: "Password is incorrect",
		},
		{
			name:        "falls back to raw string for unknown key",
			status:      http.StatusInternalServerError,
			err:         errors.New("database connection failed"),
			code:        "get_user",
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "database connection failed",
		},
		{
			name:        "nil error uses code translation",
			status:      http.StatusInternalServerError,
			err:         nil,
			code:        "internal_server_error",
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "An unexpected error occurred",
		},
		{
			name:        "sets Content-Type header",
			status:      http.StatusBadRequest,
			err:         errors.New(apierr.ErrNameRequired),
			code:        "validation",
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Name is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.acceptLang != "" {
				req.Header.Set("Accept-Language", tt.acceptLang)
			}

			handler.WriteErrorJson(rec, req, tt.status, tt.err, tt.code)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", rec.Code, tt.wantStatus)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", ct)
			}
			var resp handler.ApiResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("body is not valid JSON: %v", err)
			}
			if resp.Success {
				t.Errorf("success = true, want false")
			}
			if tt.wantCode != "" && resp.Code != tt.wantCode {
				t.Errorf("code = %v, want %v", resp.Code, tt.wantCode)
			}
			if tt.wantMessage != "" && resp.Message != tt.wantMessage {
				t.Errorf("message = %v, want %v", resp.Message, tt.wantMessage)
			}
		})
	}
}
