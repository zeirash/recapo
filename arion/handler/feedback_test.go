package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/handler"
	mock_service "github.com/zeirash/recapo/arion/mock/service"
)

// buildFeedbackRequest builds a multipart/form-data POST with context userID.
// Pass nil imageBytes to omit the image field.
func buildFeedbackRequest(userID int, fields map[string]string, imageBytes []byte) *http.Request {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for k, v := range fields {
		writer.WriteField(k, v)
	}
	if imageBytes != nil {
		part, _ := writer.CreateFormFile("image", "screenshot.jpg")
		part.Write(imageBytes)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/feedback", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req.WithContext(context.WithValue(req.Context(), common.UserIDKey, userID))
}

func TestCreateFeedbackHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	oldSvc := handler.GetFeedbackService()
	defer handler.SetFeedbackService(oldSvc)

	mockSvc := mock_service.NewMockFeedbackService(ctrl)
	handler.SetFeedbackService(mockSvc)

	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}

	tests := []struct {
		name        string
		buildReq    func() *http.Request
		mockSetup   func()
		wantStatus  int
		wantSuccess bool
	}{
		{
			name: "creates bug report without image",
			buildReq: func() *http.Request {
				return buildFeedbackRequest(1, map[string]string{
					"type": "bug", "title": "Login broken", "description": "Can't log in",
				}, nil)
			},
			mockSetup: func() {
				mockSvc.EXPECT().CreateFeedback(gomock.Any(), 1,"bug", "Login broken", "Can't log in", "").Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "creates bug report with image",
			buildReq: func() *http.Request {
				return buildFeedbackRequest(1, map[string]string{
					"type": "bug", "title": "Visual glitch",
				}, jpegBytes)
			},
			mockSetup: func() {
				mockSvc.EXPECT().UploadFeedbackImage(gomock.Any(), gomock.Any()).Return("https://cdn.example.com/feedback/abc.jpg", nil)
				mockSvc.EXPECT().CreateFeedback(gomock.Any(), 1,"bug", "Visual glitch", "", "https://cdn.example.com/feedback/abc.jpg").Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "creates enhancement without image",
			buildReq: func() *http.Request {
				return buildFeedbackRequest(1, map[string]string{
					"type": "enhancement", "title": "Add dark mode",
				}, nil)
			},
			mockSetup: func() {
				mockSvc.EXPECT().CreateFeedback(gomock.Any(), 1,"enhancement", "Add dark mode", "", "").Return(nil)
			},
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name: "returns 400 when title is empty",
			buildReq: func() *http.Request {
				return buildFeedbackRequest(1, map[string]string{"type": "bug", "title": ""}, nil)
			},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "returns 400 when type is invalid",
			buildReq: func() *http.Request {
				return buildFeedbackRequest(1, map[string]string{"type": "invalid", "title": "Some title"}, nil)
			},
			mockSetup:   func() {},
			wantStatus:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name: "returns 500 when image upload fails",
			buildReq: func() *http.Request {
				return buildFeedbackRequest(1, map[string]string{"type": "bug", "title": "Crash"}, jpegBytes)
			},
			mockSetup: func() {
				mockSvc.EXPECT().UploadFeedbackImage(gomock.Any(), gomock.Any()).Return("", errors.New(apierr.ErrFeedbackFailed))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
		{
			name: "returns 500 on feedback service error",
			buildReq: func() *http.Request {
				return buildFeedbackRequest(1, map[string]string{"type": "bug", "title": "Crash"}, nil)
			},
			mockSetup: func() {
				mockSvc.EXPECT().CreateFeedback(gomock.Any(), 1,"bug", "Crash", "", "").Return(errors.New(apierr.ErrFeedbackFailed))
			},
			wantStatus:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := tt.buildReq()
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
