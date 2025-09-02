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

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(common.UserIDKey).(int)

	// params := mux.Vars(r)
	// if params["user_id"] != "" {
	// 	userIDInt, _ := strconv.Atoi(params["user_id"])
	// 	userID = userIDInt
	// }

	res, err := userService.GetUserByID(int(userID))
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_user")
		return
	}

	if res == nil {
		WriteErrorJson(w, http.StatusNotFound, err, "get_user")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	res, err := userService.GetUsers()
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_users")
		return
	}

	if res == nil {
		WriteErrorJson(w, http.StatusNotFound, err, "get_users")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(common.UserIDKey).(int)

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
		WriteErrorJson(w, http.StatusInternalServerError, err, "update_user")
		return
	}

	WriteJson(w, http.StatusOK, res)
}
