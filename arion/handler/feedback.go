package handler

import (
	"errors"
	"net/http"

	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/logger"
)

// CreateFeedbackHandler godoc
//
//	@Summary		Submit feedback
//	@Description	Creates a GitHub issue tagged as bug or feature-request. Accepts multipart/form-data with an optional image attachment.
//	@Tags			feedback
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			type		formData	string	true	"bug or enhancement"
//	@Param			title		formData	string	true	"Short title"
//	@Param			description	formData	string	false	"Detailed description"
//	@Param			image		formData	file	false	"Optional screenshot (jpeg/png/webp, max 5MB)"
//	@Success		200			{object}	object{}
//	@Failure		400			{object}	ErrorApiResponse
//	@Failure		500			{object}	ErrorApiResponse
//	@Router			/feedback [post]
func CreateFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_form")
		return
	}

	feedbackType := r.FormValue("type")
	title := r.FormValue("title")
	description := r.FormValue("description")

	if title == "" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrFeedbackTitleRequired), "validation")
		return
	}
	if feedbackType != "bug" && feedbackType != "enhancement" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrFeedbackTypeInvalid), "validation")
		return
	}

	var imageURL string
	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		url, err := feedbackService.UploadFeedbackImage(file)
		if err != nil {
			logger.WithError(err).Error("upload_feedback_image_error")
			WriteErrorJson(w, r, http.StatusInternalServerError, err, "upload_feedback_image")
			return
		}
		imageURL = url
	}

	userID := r.Context().Value(common.UserIDKey).(int)

	if err := feedbackService.CreateFeedback(userID, feedbackType, title, description, imageURL); err != nil {
		logger.WithError(err).Error("create_feedback_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "create_feedback")
		return
	}

	WriteJson(w, http.StatusOK, struct{}{})
}
