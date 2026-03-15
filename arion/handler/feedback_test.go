package handler_test

import (
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

func newFeedbackRequest(body interface{}, userID int) *http.Request {
	b, _ := json.Marshal(body)
	return newRequestWithUserID(http.MethodPost, "/feedback", b, userID)
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
		userID      int
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name:   "creates bug report successfully",
			body:   map[string]string{"type": "bug", "title": "Login broken", "description": "Can't log in"},
			userID: 1,
			mockSetup: func() {
				mockSvc.EXPECT().CreateFeedback(1, "bug", "Login broken", "Can't log in").Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:   "creates enhancement successfully",
			body:   map[string]string{"type": "enhancement", "title": "Add dark mode"},
			userID: 1,
			mockSetup: func() {
				mockSvc.EXPECT().CreateFeedback(1, "enhancement", "Add dark mode", "").Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "returns 400 when title is empty",
			body:        map[string]string{"type": "bug", "title": ""},
			userID:      1,
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 when type is invalid",
			body:        map[string]string{"type": "invalid", "title": "Some title"},
			userID:      1,
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 when type is missing",
			body:        map[string]string{"title": "Some title"},
			userID:      1,
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "returns 400 when invalid json",
			body:        "invalid json",
			userID:      1,
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:   "returns 500 on service error",
			body:   map[string]string{"type": "bug", "title": "Crash on load"},
			userID: 1,
			mockSetup: func() {
				mockSvc.EXPECT().CreateFeedback(1, "bug", "Crash on load", "").Return(errors.New(apierr.ErrFeedbackFailed))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := newFeedbackRequest(tt.body, tt.userID)
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
