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
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	inp := LoginRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateLogin(inp); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
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

	if valid, err := validateRegister(inp); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := userService.UserRegister(inp.Name, inp.Email, inp.Password)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "user_register")
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

	//TODO: validate password value
	if input.Password == "" {
		return false, errors.New("password must not empty")
	}

	return true, nil
}
