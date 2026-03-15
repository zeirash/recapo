package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
)

func newFeedbackRequest(body interface{}) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestCreateFeedbackHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldSvc := handler.GetFeedbackService()
	defer handler.SetFeedbackService(oldSvc)

	mockSvc := mock_service.NewMockFeedbackService(ctrl)
	handler.SetFeedbackService(mockSvc)

	tests := []struct {
		name        string
		body        interface{}
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "creates bug report successfully",
			body: map[string]string{"type": "bug", "title": "Login broken", "description": "Can't log in"},
			mockSetup: func() {
				mockSvc.EXPECT().CreateFeedback("bug", "Login broken", "Can't log in").Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "creates enhancement successfully",
			body: map[string]string{"type": "enhancement", "title": "Add dark mode"},
			mockSetup: func() {
				mockSvc.EXPECT().CreateFeedback("enhancement", "Add dark mode", "").Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 400 when title is empty",
			body:        map[string]string{"type": "bug", "title": ""},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 when type is invalid",
			body:        map[string]string{"type": "invalid", "title": "Some title"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 when type is missing",
			body:        map[string]string{"title": "Some title"},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 when invalid json",
			body:        "invalid json",
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "returns 500 on service error",
			body: map[string]string{"type": "bug", "title": "Crash on load"},
			mockSetup: func() {
				mockSvc.EXPECT().CreateFeedback("bug", "Crash on load", "").Return(errors.New(apierr.ErrFeedbackFailed))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := newFeedbackRequest(tt.body)
			w := httptest.NewRecorder()
			handler.CreateFeedbackHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateFeedbackHandler() status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if success, _ := resp["success"].(bool); success != tt.wantSuccess {
				t.Errorf("CreateFeedbackHandler() success = %v, want %v", success, tt.wantSuccess)
			}
		})
	}
}
