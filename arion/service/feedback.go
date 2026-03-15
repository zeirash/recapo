package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/constant"
)

type (
	FeedbackService interface {
		CreateFeedback(feedbackType, title, description string) error
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
var githubAPIFunc = callGithubAPI

func NewFeedbackService() FeedbackService {
	return &sfeedback{}
}

func (s *sfeedback) CreateFeedback(feedbackType, title, description string) error {
	if title == "" {
		return errors.New(apierr.ErrFeedbackTitleRequired)
	}
	if feedbackType != constant.GithubLabelBug && feedbackType != constant.GithubLabelEnhancement {
		return errors.New(apierr.ErrFeedbackTypeInvalid)
	}

	reqBody := githubIssueRequest{
		Title:  title,
		Body:   description,
		Labels: []string{feedbackType, constant.GithubLabelUser},
	}

	return githubAPIFunc(reqBody)
}

func callGithubAPI(reqBody githubIssueRequest) error {
	owner := cfg.GitHubRepoOwner
	repo := cfg.GitHubRepoName
	token := cfg.GitHubToken

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
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
