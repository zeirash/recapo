package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	UserStore interface {
		GetUserByID(userID int) (*model.User, error)
		GetUserByEmail(email string) (*model.User, error)
		GetUsers() ([]model.User, error)
		CreateUser(name, email, hashPassword string) (*model.User, error)
		UpdateUser(id int, input UpdateUserInput) (*model.User, error)
	}

	user struct{}

	UpdateUserInput struct {
		Name     *string
		Email    *string
		Password *string
	}
)

func NewUserStore() UserStore {
	return &user{}
}

func (u *user) GetUserByID(userID int) (*model.User, error) {
	db := database.GetDB()
	defer db.Close()

	resp := model.User{}

	q := `
		SELECT id, name, email, password, system_mode, created_at, updated_at
		FROM "user"
		WHERE id = $1
	`

	err := db.QueryRow(q, userID).Scan(&resp.ID, &resp.Name, &resp.Email, &resp.Password, &resp.SystemMode, &resp.CreatedAt, &resp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &resp, nil
}

func (u *user) GetUserByEmail(email string) (*model.User, error) {
	db := database.GetDB()
	if db == nil {
		return nil, sql.ErrNoRows
	}
	defer db.Close()

	resp := model.User{}
	q := `
		SELECT id, name, email, password, system_mode, created_at, updated_at
		FROM "user"
		WHERE email = $1
	`

	err := db.QueryRow(q, email).Scan(&resp.ID, &resp.Name, &resp.Email, &resp.Password, &resp.SystemMode, &resp.CreatedAt, &resp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &resp, nil
}

func (u *user) GetUsers() ([]model.User, error) {
	db := database.GetDB()
	if db == nil {
		return nil, sql.ErrNoRows
	}
	defer db.Close()

	q := `
		SELECT id, name, email, password, system_mode, created_at, updated_at
		FROM "user"
	`

	rows, err := db.Query(q)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.SystemMode, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (u *user) CreateUser(name, email, hashPassword string) (*model.User, error) {
	db := database.GetDB()
	defer db.Close()

	now := time.Now()
	var id int

	q := `
		INSERT INTO "user" (name, email, password, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := db.QueryRow(q, name, email, hashPassword, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: now,
	}, nil
}

func (u *user) UpdateUser(id int, input UpdateUserInput) (*model.User, error) {
	db := database.GetDB()
	defer db.Close()

	set := []string{}
	var user model.User

	// build query
	if input.Name != nil {
		newSet := fmt.Sprintf("name = '%s'", *input.Name)
		set = append(set, newSet)
	}
	if input.Email != nil {
		newSet := fmt.Sprintf("email = '%s'", *input.Email)
		set = append(set, newSet)
	}
	if input.Password != nil {
		newSet := fmt.Sprintf("password = '%s'", *input.Password)
		set = append(set, newSet)
	}

	set = append(set, "updated_at = now()")

	q := `
		UPDATE "user"
		SET %s
		WHERE id = $1
		RETURNING id, name, email, created_at, updated_at
	`

	q = fmt.Sprintf(q, strings.Join(set, ","))

	err := db.QueryRow(q, id).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
