package handler

import (
	"net/http"

	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/service"
)

type (
	UpdateUserRequest struct {
		Name     *string `json:"name"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
)

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(common.UserIDKey).(float64)

	inp := UpdateUserRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := userService.UpdateUser(service.UpdateUserInput{
		ID:       int(userID),
		Name:     inp.Name,
		Email:    inp.Email,
		Password: inp.Password,
	})
	if err != nil {
		status := http.StatusInternalServerError
		WriteErrorJson(w, status, err, "update_user")
		return
	}

	WriteJson(w, http.StatusOK, res)
}
