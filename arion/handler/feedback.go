package handler

import (
	"errors"
	"net/http"

	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/logger"
)

type CreateFeedbackRequest struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CreateFeedbackHandler godoc
//
//	@Summary		Submit feedback
//	@Description	Creates a GitHub issue tagged as bug or feature-request.
//	@Tags			feedback
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateFeedbackRequest	true	"feedback payload"
//	@Success		200		{object}	object{}
//	@Failure		400		{object}	ErrorApiResponse
//	@Failure		500		{object}	ErrorApiResponse
//	@Router			/feedback [post]
func CreateFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	inp := CreateFeedbackRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if inp.Title == "" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrFeedbackTitleRequired), "validation")
		return
	}
	if inp.Type != "bug" && inp.Type != "enhancement" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrFeedbackTypeInvalid), "validation")
		return
	}

	if err := feedbackService.CreateFeedback(inp.Type, inp.Title, inp.Description); err != nil {
		logger.WithError(err).Error("create_feedback_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "create_feedback")
		return
	}

	WriteJson(w, http.StatusOK, struct{}{})
}
