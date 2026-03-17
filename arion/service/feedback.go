package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/store"
)

type (
	FeedbackService interface {
		UploadFeedbackImage(ctx context.Context, file io.Reader) (string, error)
		CreateFeedback(ctx context.Context, userID int, feedbackType, title, description, imageURL string) error
	}

	sfeedback struct{}

	githubIssueRequest struct {
		Title  string   `json:"title"`
		Body   string   `json:"body,omitempty"`
		Labels []string `json:"labels"`
	}
)

// githubAPIFunc is the function used to call the GitHub Issues API.
// Overridable in tests to avoid real HTTP calls.
var githubAPIFunc func(context.Context, githubIssueRequest) error = callGithubAPI

func NewFeedbackService() FeedbackService {
	if userStore == nil {
		userStore = store.NewUserStore()
	}
	return &sfeedback{}
}

func (s *sfeedback) UploadFeedbackImage(ctx context.Context, file io.Reader) (string, error) {
	return uploadImage(file, "feedback")
}

func (s *sfeedback) CreateFeedback(ctx context.Context, userID int, feedbackType, title, description, imageURL string) error {
	if title == "" {
		return errors.New(apierr.ErrFeedbackTitleRequired)
	}
	if feedbackType != constant.GithubLabelBug && feedbackType != constant.GithubLabelEnhancement {
		return errors.New(apierr.ErrFeedbackTypeInvalid)
	}

	user, err := userStore.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New(apierr.ErrUserNotFound)
	}

	body := description
	if imageURL != "" {
		if body != "" {
			body += "\n\n"
		}
		body += fmt.Sprintf("![screenshot](%s)", imageURL)
	}

	reqBody := githubIssueRequest{
		Title:  fmt.Sprintf("[%s] %s", user.Email, title),
		Body:   body,
		Labels: []string{feedbackType, constant.GithubLabelUser},
	}

	return githubAPIFunc(ctx, reqBody)
}

func callGithubAPI(ctx context.Context, reqBody githubIssueRequest) error {
	owner := cfg.GitHubRepoOwner
	repo := cfg.GitHubRepoName
	token := cfg.GitHubToken

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: github returned status %d: %s", apierr.ErrFeedbackFailed, resp.StatusCode, string(respBytes))
	}

	return nil
}
