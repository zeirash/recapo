package service

import (
	"errors"
	"testing"

	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/constant"
)

func Test_sfeedback_CreateFeedback(t *testing.T) {
	tests := []struct {
		name         string
		feedbackType string
		title        string
		description  string
		githubFunc   func(githubIssueRequest) error
		wantErr      bool
		wantErrKey   string
	}{
		{
			name:         "creates bug issue successfully",
			feedbackType: "bug",
			title:        "Something is broken",
			description:  "Steps to reproduce...",
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      false,
		},
		{
			name:         "creates enhancement issue successfully",
			feedbackType: "enhancement",
			title:        "Add dark mode",
			description:  "",
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      false,
		},
		{
			name:         "passes correct labels to github",
			feedbackType: "bug",
			title:        "Label check",
			description:  "",
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
			name:         "passes title and body to github",
			feedbackType: "bug",
			title:        "My title",
			description:  "My description",
			githubFunc: func(req githubIssueRequest) error {
				if req.Title != "My title" {
					return errors.New("wrong title")
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
			feedbackType: "bug",
			title:        "",
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      true,
			wantErrKey:   apierr.ErrFeedbackTitleRequired,
		},
		{
			name:         "returns error when type is invalid",
			feedbackType: "feature-request",
			title:        "Some title",
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      true,
			wantErrKey:   apierr.ErrFeedbackTypeInvalid,
		},
		{
			name:         "returns error when type is empty",
			feedbackType: "",
			title:        "Some title",
			githubFunc:   func(githubIssueRequest) error { return nil },
			wantErr:      true,
			wantErrKey:   apierr.ErrFeedbackTypeInvalid,
		},
		{
			name:         "returns error when github call fails",
			feedbackType: "bug",
			title:        "Some title",
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

			s := &sfeedback{}
			gotErr := s.CreateFeedback(tt.feedbackType, tt.title, tt.description)

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
