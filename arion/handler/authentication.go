package handler

import (
	"errors"
	"net/http"
	"net/mail"
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

	RefreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	inp := LoginRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateLogin(inp); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := userService.UserLogin(inp.Email, inp.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "password incorrect" {
			status = http.StatusUnauthorized
		}

		WriteErrorJson(w, r, status, err, "user_login")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	inp := RegisterRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateRegister(inp); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := userService.UserRegister(inp.Name, inp.Email, inp.Password)
	if err != nil {
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "user_register")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	inp := RefreshRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if inp.RefreshToken == "" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New("refresh_token is required"), "validation")
		return
	}

	res, err := userService.RefreshToken(inp.RefreshToken)
	if err != nil {
		WriteErrorJson(w, r, http.StatusUnauthorized, err, "refresh_token")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func validateLogin(input LoginRequest) (bool, error) {
	if input.Email == "" {
		return false, errors.New("email must not empty")
	}

	if input.Password == "" {
		return false, errors.New("password must not empty")
	}

	return true, nil
}

func validateRegister(input RegisterRequest) (bool, error) {
	if input.Name == "" {
		return false, errors.New("name must not empty")
	}

	if input.Email == "" {
		return false, errors.New("email must not empty")
	}

	_, err := mail.ParseAddress(input.Email)
	if err != nil {
		return false, errors.New("email not valid")
	}

	if input.Password == "" {
		return false, errors.New("password must not empty")
	}

	return true, nil
}
