package service

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/constant"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
)

func Test_sfeedback_CreateFeedback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserStore := mock_store.NewMockUserStore(ctrl)

	testUser := &model.User{ID: 1, Email: "user@example.com"}

	tests := []struct {
		name         string
		userID       int
		feedbackType string
		title        string
		description  string
		storeMock    func()
		githubFunc   func(githubIssueRequest) error
		wantErr      bool
		wantErrKey   string
	}{
		{
			name:         "creates bug issue successfully",
			userID:       1,
			feedbackType: "bug",
			title:        "Something is broken",
			description:  "Steps to reproduce...",
			storeMock:    func() { mockUserStore.EXPECT().GetUserByID(1).Return(testUser, nil) },
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      false,
		},
		{
			name:         "creates enhancement issue successfully",
			userID:       1,
			feedbackType: "enhancement",
			title:        "Add dark mode",
			description:  "",
			storeMock:    func() { mockUserStore.EXPECT().GetUserByID(1).Return(testUser, nil) },
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      false,
		},
		{
			name:         "passes correct labels to github",
			userID:       1,
			feedbackType: "bug",
			title:        "Label check",
			description:  "",
			storeMock:    func() { mockUserStore.EXPECT().GetUserByID(1).Return(testUser, nil) },
			githubFunc: func(req githubIssueRequest) error {
				if len(req.Labels) != 2 {
					return errors.New("expected 2 labels")
				}
				if req.Labels[0] != "bug" {
					return errors.New("expected first label to be bug")
				}
				if req.Labels[1] != constant.GithubLabelUser {
					return errors.New("expected second label to be user label")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:         "prefixes title with user email",
			userID:       1,
			feedbackType: "bug",
			title:        "My title",
			description:  "My description",
			storeMock:    func() { mockUserStore.EXPECT().GetUserByID(1).Return(testUser, nil) },
			githubFunc: func(req githubIssueRequest) error {
				if req.Title != "[user@example.com] My title" {
					return errors.New("wrong title: " + req.Title)
				}
				if req.Body != "My description" {
					return errors.New("wrong body")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:         "returns error when title is empty",
			userID:       1,
			feedbackType: "bug",
			title:        "",
			storeMock:    func() {},
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      true,
			wantErrKey:   apierr.ErrFeedbackTitleRequired,
		},
		{
			name:         "returns error when type is invalid",
			userID:       1,
			feedbackType: "feature-request",
			title:        "Some title",
			storeMock:    func() {},
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      true,
			wantErrKey:   apierr.ErrFeedbackTypeInvalid,
		},
		{
			name:         "returns error when type is empty",
			userID:       1,
			feedbackType: "",
			title:        "Some title",
			storeMock:    func() {},
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      true,
			wantErrKey:   apierr.ErrFeedbackTypeInvalid,
		},
		{
			name:         "returns error when user not found",
			userID:       99,
			feedbackType: "bug",
			title:        "Some title",
			storeMock:    func() { mockUserStore.EXPECT().GetUserByID(99).Return(nil, errors.New("not found")) },
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      true,
			wantErrKey:   apierr.ErrUserNotFound,
		},
		{
			name:         "returns error when github call fails",
			userID:       1,
			feedbackType: "bug",
			title:        "Some title",
			storeMock:    func() { mockUserStore.EXPECT().GetUserByID(1).Return(testUser, nil) },
			githubFunc:   func(githubIssueRequest) error { return errors.New(apierr.ErrFeedbackFailed) },
			wantErr:      true,
			wantErrKey:   apierr.ErrFeedbackFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldFunc := githubAPIFunc
			defer func() { githubAPIFunc = oldFunc }()
			githubAPIFunc = tt.githubFunc

			oldUserStore := userStore
			defer func() { userStore = oldUserStore }()
			userStore = mockUserStore

			tt.storeMock()

			s := &sfeedback{}
			gotErr := s.CreateFeedback(tt.userID, tt.feedbackType, tt.title, tt.description)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateFeedback() failed: %v", gotErr)
				}
				if tt.wantErrKey != "" && gotErr.Error() != tt.wantErrKey {
					t.Errorf("CreateFeedback() error = %v, want %v", gotErr.Error(), tt.wantErrKey)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateFeedback() succeeded unexpectedly")
			}
		})
	}
}
