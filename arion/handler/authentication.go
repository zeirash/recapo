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

// LoginHandler godoc
//
//	@Summary		Login
//	@Description	Authenticate with email and password. Returns access and refresh tokens.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginRequest	true	"Login credentials"
//	@Success		200		{object}	response.TokenResponse
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON or validation)"
//	@Failure		401		{object}	ErrorApiResponse	"Invalid credentials"
//	@Failure		500		{object}	ErrorApiResponse	"Internal server error"
//	@Router			/login [post]
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

// RegisterHandler godoc
//
//	@Summary		Register
//	@Description	Create a new user account. Returns access and refresh tokens.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RegisterRequest	true	"Registration data"
//	@Success		200		{object}	response.TokenResponse
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON or validation)"
//	@Failure		500		{object}	ErrorApiResponse	"Internal server error"
//	@Router			/register [post]
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

// RefreshHandler godoc
//
//	@Summary		Refresh token
//	@Description	Exchange a valid refresh token for new access and refresh tokens.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	response.TokenResponse
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON or refresh_token required)"
//	@Failure		401		{object}	ErrorApiResponse	"Invalid or expired refresh token"
//	@Router			/refresh [post]
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
