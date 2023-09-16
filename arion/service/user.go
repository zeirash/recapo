package service

import (
	"errors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"

	"golang.org/x/crypto/bcrypt"
)

type (
	UserService interface {
		UserLogin(email, password string) (response.TokenResponse, error)
		UserRegister(name, email, password string) (response.TokenResponse, error)
	}

	uservice struct {}
)

func NewUserService() UserService {
	cfg = config.GetConfig()

	if user == nil {
		user = store.NewUserStore()
	}

	if token == nil {
		token = store.NewTokenStore()
	}

	return &uservice{}
}

func (u *uservice) UserLogin(email, password string) (response.TokenResponse, error) {
	user, err := user.GetUserByEmail(email)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if user == nil {
		return response.TokenResponse{}, errors.New("user doesn't exist")
	}
	
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return response.TokenResponse{}, errors.New("password incorrect")
	}

	accessToken, err := token.CreateAccessToken(user, cfg.SecretKey, 2)
	if err != nil {
		return response.TokenResponse{}, err
	}

	refreshToken, err := token.CreateRefreshToken(user, cfg.SecretKey, 168)
	if err != nil {
		return response.TokenResponse{}, err
	}

	return response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *uservice) UserRegister(name, email, password string) (response.TokenResponse, error) {
	existUser, err := user.GetUserByEmail(email)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if existUser != nil {
		return response.TokenResponse{}, errors.New("user already exist")
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return response.TokenResponse{}, err
	}

	newUser, err := user.CreateUser(name, email, string(encryptedPassword))
	if err != nil {
		return response.TokenResponse{}, err
	}

	accessToken, err := token.CreateAccessToken(newUser, cfg.SecretKey, 2)
	if err != nil {
		return response.TokenResponse{}, err
	}

	refreshToken, err := token.CreateRefreshToken(newUser, cfg.SecretKey, 168)
	if err != nil {
		return response.TokenResponse{}, err
	}

	return response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func GetUser() {
	// userId, err := uuid.FromString(c.Param("userId"))
	// if err != nil {
	// 	helpers.RespondWithError(c, http.StatusBadRequest, "invalid user id")
	// 	return
	// }

	// user, err := repository.GetUserById(context.Background(), &userId)
	// if err != nil {
	// 	config.Logger.Errorf("Error while searching: %s", err)
	// 	helpers.RespondWithError(c, http.StatusInternalServerError, "internal server error")
	// 	return
	// }
	// if user.ID.IsNil() {
	// 	helpers.RespondWithError(c, http.StatusNotFound, "user not found")
	// 	return
	// }

	// err = utils.WriteJSON(c.Writer, c)
	// if err != nil {
	// 	helpers.RespondWithError(c, http.StatusInternalServerError, "internal server error")
	// 	return
	// }

	// c.JSON(http.StatusOK, user)
}
