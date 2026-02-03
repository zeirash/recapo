package service

import (
	"errors"
	"fmt"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"

	"golang.org/x/crypto/bcrypt"
)

type (
	UserService interface {
		UserLogin(email, password string) (response.TokenResponse, error)
		UserRegister(name, email, password string) (response.TokenResponse, error)
		RefreshToken(refreshToken string) (response.TokenResponse, error)
		UpdateUser(input UpdateUserInput) (response.UserData, error)
		GetUserByID(userID int) (*response.UserData, error)
		GetUsers() ([]response.UserData, error)
	}

	uservice struct{}

	UpdateUserInput struct {
		ID       int
		Name     *string
		Email    *string
		Password *string
	}
)

func NewUserService() UserService {
	cfg = config.GetConfig()

	if userStore == nil {
		userStore = store.NewUserStore()
	}

	if tokenStore == nil {
		tokenStore = store.NewTokenStore()
	}

	if shopStore == nil {
		shopStore = store.NewShopStore()
	}

	return &uservice{}
}

func (u *uservice) UserLogin(email, password string) (response.TokenResponse, error) {
	user, err := userStore.GetUserByEmail(email)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if user == nil {
		return response.TokenResponse{}, errors.New("user doesn't exist")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return response.TokenResponse{}, errors.New("password incorrect")
	}

	accessToken, err := tokenStore.CreateAccessToken(user, cfg.SecretKey, 2)
	if err != nil {
		return response.TokenResponse{}, err
	}

	refreshToken, err := tokenStore.CreateRefreshToken(user, cfg.SecretKey, 168)
	if err != nil {
		return response.TokenResponse{}, err
	}

	return response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *uservice) RefreshToken(refreshToken string) (response.TokenResponse, error) {
	// Validate and extract data from refresh token
	tokenData, err := tokenStore.ExtractDataFromToken(refreshToken, cfg.SecretKey)
	if err != nil {
		return response.TokenResponse{}, errors.New("invalid refresh token")
	}

	// Get user from database
	user, err := userStore.GetUserByID(tokenData.UserID)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if user == nil {
		return response.TokenResponse{}, errors.New("user not found")
	}

	// Generate new tokens
	accessToken, err := tokenStore.CreateAccessToken(user, cfg.SecretKey, 2)
	if err != nil {
		return response.TokenResponse{}, err
	}

	newRefreshToken, err := tokenStore.CreateRefreshToken(user, cfg.SecretKey, 168)
	if err != nil {
		return response.TokenResponse{}, err
	}

	return response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (u *uservice) UserRegister(name, email, password string) (response.TokenResponse, error) {
	existUser, err := userStore.GetUserByEmail(email)
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

	db := dbGetter()

	tx, err := db.Begin()
	if err != nil {
		return response.TokenResponse{}, err
	}
	defer tx.Rollback()

	shopName := fmt.Sprintf("%s's Shop", name)
	shop, err := shopStore.CreateShop(tx, shopName)
	if err != nil {
		return response.TokenResponse{}, err
	}

	newUser, err := userStore.CreateUser(tx, name, email, string(encryptedPassword), constant.RoleOwner, shop.ID)
	if err != nil {
		return response.TokenResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return response.TokenResponse{}, err
	}

	accessToken, err := tokenStore.CreateAccessToken(newUser, cfg.SecretKey, 2)
	if err != nil {
		return response.TokenResponse{}, err
	}

	refreshToken, err := tokenStore.CreateRefreshToken(newUser, cfg.SecretKey, 168)
	if err != nil {
		return response.TokenResponse{}, err
	}

	return response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *uservice) UpdateUser(input UpdateUserInput) (response.UserData, error) {
	user, err := userStore.GetUserByID(input.ID)
	if err != nil {
		return response.UserData{}, err
	}

	if user == nil {
		return response.UserData{}, errors.New("user not found")
	}

	var password string
	if input.Password != nil {
		encryptedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(*input.Password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return response.UserData{}, err
		}

		password = string(encryptedPassword)
	}

	updateData := store.UpdateUserInput{
		Name:  input.Name,
		Email: input.Email,
	}
	if password != "" {
		updateData.Password = &password
	}

	userData, err := userStore.UpdateUser(input.ID, updateData)
	if err != nil {
		return response.UserData{}, err
	}

	res := response.UserData{
		ID:        userData.ID,
		Name:      userData.Name,
		Email:     userData.Email,
		CreatedAt: userData.CreatedAt,
	}

	if userData.UpdatedAt.Valid {
		res.UpdatedAt = &userData.UpdatedAt.Time
	}

	return res, nil
}

func (u *uservice) GetUserByID(userID int) (*response.UserData, error) {
	user, err := userStore.GetUserByID(userID)
	if err != nil {
		return &response.UserData{}, err
	}

	if user == nil {
		return nil, errors.New("user doesn't exist")
	}

	res := response.UserData{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	if user.UpdatedAt.Valid {
		res.UpdatedAt = &user.UpdatedAt.Time
	}

	return &res, nil
}

func (u *uservice) GetUsers() ([]response.UserData, error) {
	users, err := userStore.GetUsers()
	if err != nil {
		return []response.UserData{}, err
	}

	var usersData []response.UserData
	for _, user := range users {
		res := response.UserData{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		}

		if user.UpdatedAt.Valid {
			t := user.UpdatedAt.Time
			res.UpdatedAt = &t
		}

		usersData = append(usersData, res)
	}

	return usersData, nil
}
