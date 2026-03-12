package handler

import (
	"errors"
	"net/http"
	"net/mail"

	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/i18n"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/common/otp"
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
		OTP      string `json:"otp"`
	}

	SendOTPRequest struct {
		Email string `json:"email"`
	}

	RefreshRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	ForgotPasswordRequest struct {
		Email string `json:"email"`
	}

	ResetPasswordRequest struct {
		Email    string `json:"email"`
		OTP      string `json:"otp"`
		Password string `json:"password"`
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
		logger.WithError(err).Error("login_error")
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
		if err.Error() == apierr.ErrPasswordIncorrect {
			status = http.StatusUnauthorized
		}

		WriteErrorJson(w, r, status, err, "user_login")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// SendOTPHandler godoc
//
//	@Summary		Send OTP
//	@Description	Send a 6-digit verification code to the given email. Used before registration.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SendOTPRequest	true	"Email address"
//	@Success		200		{object}	ApiResponse
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON, validation, or email already registered)"
//	@Failure		500		{object}	ErrorApiResponse	"Internal server error"
//	@Router			/send_otp [post]
func SendOTPHandler(w http.ResponseWriter, r *http.Request) {
	inp := SendOTPRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if inp.Email == "" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrEmailRequired), "validation")
		return
	}

	if _, err := mail.ParseAddress(inp.Email); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrEmailInvalid), "validation")
		return
	}

	if err := userService.SendOTP(inp.Email, i18n.GetLangFromRequest(r)); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == apierr.ErrUserAlreadyExists {
			status = http.StatusBadRequest
		}
		WriteErrorJson(w, r, status, err, "send_otp")
		return
	}

	WriteJson(w, http.StatusOK, nil)
}

// RegisterHandler godoc
//
//	@Summary		Register
//	@Description	Create a new user account after OTP verification. Returns access and refresh tokens.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RegisterRequest	true	"Registration data"
//	@Success		200		{object}	response.TokenResponse
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON, validation, or invalid OTP)"
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

	if !otp.Verify(inp.Email, inp.OTP) {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrInvalidOTP), "otp_verify")
		return
	}

	res, err := userService.UserRegister(inp.Name, inp.Email, inp.Password)
	if err != nil {
		logger.WithError(err).Error("user_register_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "user_register")
		return
	}

	otp.Delete(inp.Email)
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
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrRefreshTokenRequired), "validation")
		return
	}

	res, err := userService.RefreshToken(inp.RefreshToken)
	if err != nil {
		WriteErrorJson(w, r, http.StatusUnauthorized, err, "refresh_token")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// ForgotPasswordHandler godoc
//
//	@Summary		Forgot password
//	@Description	Send a 6-digit OTP to the email address for password reset. Silently succeeds if email is not registered.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		ForgotPasswordRequest	true	"Email address"
//	@Success		200		{object}	ApiResponse
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON or validation)"
//	@Failure		500		{object}	ErrorApiResponse	"Internal server error"
//	@Router			/forgot_password [post]
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	inp := ForgotPasswordRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if inp.Email == "" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrEmailRequired), "validation")
		return
	}

	if _, err := mail.ParseAddress(inp.Email); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrEmailInvalid), "validation")
		return
	}

	if err := userService.ForgotPassword(inp.Email, i18n.GetLangFromRequest(r)); err != nil {
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "forgot_password")
		return
	}

	WriteJson(w, http.StatusOK, nil)
}

// ResetPasswordHandler godoc
//
//	@Summary		Reset password
//	@Description	Reset the user's password after verifying the OTP sent to their email.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		ResetPasswordRequest	true	"Reset password data"
//	@Success		200		{object}	ApiResponse
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON, validation, or invalid OTP)"
//	@Failure		500		{object}	ErrorApiResponse	"Internal server error"
//	@Router			/reset_password [post]
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	inp := ResetPasswordRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateResetPassword(inp); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	if err := userService.ResetPassword(inp.Email, inp.OTP, inp.Password); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == apierr.ErrInvalidOTP {
			status = http.StatusBadRequest
		}
		WriteErrorJson(w, r, status, err, "reset_password")
		return
	}

	WriteJson(w, http.StatusOK, nil)
}

func validateLogin(input LoginRequest) (bool, error) {
	if input.Email == "" {
		return false, errors.New(apierr.ErrEmailRequired)
	}

	if input.Password == "" {
		return false, errors.New(apierr.ErrPasswordRequired)
	}

	return true, nil
}

func validateRegister(input RegisterRequest) (bool, error) {
	if input.Name == "" {
		return false, errors.New(apierr.ErrNameRequired)
	}

	if input.Email == "" {
		return false, errors.New(apierr.ErrEmailRequired)
	}

	_, err := mail.ParseAddress(input.Email)
	if err != nil {
		return false, errors.New(apierr.ErrEmailInvalid)
	}

	if input.Password == "" {
		return false, errors.New(apierr.ErrPasswordRequired)
	}

	if input.OTP == "" {
		return false, errors.New(apierr.ErrOTPRequired)
	}

	return true, nil
}

func validateResetPassword(input ResetPasswordRequest) (bool, error) {
	if input.Email == "" {
		return false, errors.New(apierr.ErrEmailRequired)
	}

	if _, err := mail.ParseAddress(input.Email); err != nil {
		return false, errors.New(apierr.ErrEmailInvalid)
	}

	if input.OTP == "" {
		return false, errors.New(apierr.ErrOTPRequired)
	}

	if input.Password == "" {
		return false, errors.New(apierr.ErrPasswordRequired)
	}

	return true, nil
}
