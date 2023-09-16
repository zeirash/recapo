package handler

import (
	"net/http"
)

type (
	LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	RegisterRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	inp := LoginRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := userService.UserLogin(inp.Email, inp.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "password incorrect" {
			status = http.StatusUnauthorized
		}

		WriteErrorJson(w, status, err, "user_login")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	inp := RegisterRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := userService.UserRegister(inp.Name, inp.Email, inp.Password)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "user_register")
		return
	}

	WriteJson(w, http.StatusOK, res)
}
