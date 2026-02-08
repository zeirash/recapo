package handler

import (
	"errors"
	"net/http"

	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/service"
)

type (
	UpdateUserRequest struct {
		Name     *string `json:"name"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
)

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(common.UserIDKey).(int)

	res, err := userService.GetUserByID(int(userID))
	if err != nil {
		logger.WithError(err).Error("get_user_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_user")
		return
	}

	if res == nil {
		WriteErrorJson(w, r, http.StatusNotFound, errors.New("user not found"), "user_not_found")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetUsersHandler godoc
//
//	@Summary		List users
//	@Description	Get all users. Requires appropriate permissions.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}	response.UserData
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/users [get]
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	res, err := userService.GetUsers()
	if err != nil {
		logger.WithError(err).Error("get_users_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_users")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// UpdateUserHandler godoc
//
//	@Summary		Update current user
//	@Description	Update the authenticated user's profile. Only provided fields are updated.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		UpdateUserRequest	true	"Fields to update"
//	@Success		200		{object}	response.UserData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid JSON)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/user [patch]
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(common.UserIDKey).(int)

	inp := UpdateUserRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := userService.UpdateUser(service.UpdateUserInput{
		ID:       int(userID),
		Name:     inp.Name,
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		logger.WithError(err).Error("update_user_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "update_user")
		return
	}

	WriteJson(w, http.StatusOK, res)
}
