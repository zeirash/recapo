package handler

import (
	"errors"
	"net/http"

	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/i18n"
	"github.com/zeirash/recapo/arion/common/logger"

)

type (
	InviteAdminRequest struct {
		Email string `json:"email"`
	}

	AcceptInviteRequest struct {
		Token    string `json:"token"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
)

func InviteAdminHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(common.UserIDKey).(int)
	shopID := ctx.Value(common.ShopIDKey).(int)
	lang := i18n.GetLangFromRequest(r)

	var req InviteAdminRequest
	if err := ParseJson(r.Body, &req); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, nil, "bad_request")
		return
	}

	if valid, err := validateInviteAdminRequest(req); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	err := invitationService.InviteAdmin(ctx, shopID, userID, req.Email, lang)
	if err != nil {
		switch err.Error() {
		case apierr.ErrEmailInvalid:
			WriteErrorJson(w, r, http.StatusBadRequest, err, err.Error())
		case apierr.ErrNotOwner:
			WriteErrorJson(w, r, http.StatusForbidden, err, err.Error())
		case apierr.ErrMaxUsersReached:
			WriteErrorJson(w, r, http.StatusForbidden, err, err.Error())
		case apierr.ErrUserAlreadyExists:
			WriteErrorJson(w, r, http.StatusConflict, err, err.Error())
		case apierr.ErrInvitationAlreadySent:
			WriteErrorJson(w, r, http.StatusConflict, err, err.Error())
		case apierr.ErrShopNotFound:
			WriteErrorJson(w, r, http.StatusNotFound, err, err.Error())
		default:
			logger.WithError(err).Error("invite_admin_error")
			WriteErrorJson(w, r, http.StatusInternalServerError, err, "invite_admin")
		}
		return
	}

	WriteJson(w, http.StatusOK, struct{}{})
}

func ValidateInviteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := r.URL.Query().Get("token")

	if token == "" {
		WriteErrorJson(w, r, http.StatusBadRequest, nil, apierr.ErrInvitationNotFound)
		return
	}

	data, err := invitationService.ValidateInviteToken(ctx, token)
	if err != nil {
		switch err.Error() {
		case apierr.ErrInvitationNotFound, apierr.ErrShopNotFound:
			WriteErrorJson(w, r, http.StatusBadRequest, err, err.Error())
		default:
			logger.WithError(err).Error("validate_invite_error")
			WriteErrorJson(w, r, http.StatusInternalServerError, err, "validate_invite")
		}
		return
	}

	WriteJson(w, http.StatusOK, data)
}

func AcceptInviteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AcceptInviteRequest
	if err := ParseJson(r.Body, &req); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, nil, "bad_request")
		return
	}

	if valid, err := validateAcceptInviteRequest(req); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	tokens, err := invitationService.AcceptInvite(ctx, req.Token, req.Name, req.Password)
	if err != nil {
		switch err.Error() {
		case apierr.ErrInvitationNotFound:
			WriteErrorJson(w, r, http.StatusBadRequest, err, err.Error())
		case apierr.ErrMaxUsersReached:
			WriteErrorJson(w, r, http.StatusForbidden, err, err.Error())
		case apierr.ErrInvitationAlreadyAccepted:
			WriteErrorJson(w, r, http.StatusConflict, err, err.Error())
		case apierr.ErrPasswordTooWeak:
			WriteErrorJson(w, r, http.StatusBadRequest, err, err.Error())
		default:
			logger.WithError(err).Error("accept_invite_error")
			WriteErrorJson(w, r, http.StatusInternalServerError, err, "accept_invite")
		}
		return
	}

	WriteJson(w, http.StatusOK, tokens)
}

func validateInviteAdminRequest(req InviteAdminRequest) (bool, error) {
	if req.Email == "" {
		return false, errors.New(apierr.ErrEmailRequired)
	}
	return true, nil
}

func validateAcceptInviteRequest(req AcceptInviteRequest) (bool, error) {
	if req.Token == "" {
		return false, errors.New(apierr.ErrInvitationNotFound)
	}
	if req.Name == "" {
		return false, errors.New(apierr.ErrNameRequired)
	}
	if req.Password == "" {
		return false, errors.New(apierr.ErrPasswordRequired)
	}
	return true, nil
}
