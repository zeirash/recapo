package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	UserStore interface {
		GetUserByID(userID int) (*model.User, error)
		GetUserByEmail(email string) (*model.User, error)
		CreateUser(name, email, hashPassword string) (*model.User, error)
	}

	user struct{}
)

func NewUserStore() UserStore {
	return &user{}
}

func (u *user) GetUserByID(userID int) (*model.User, error) {
	db := database.GetDB()
	defer db.Close()

	resp := model.User{}

	q := `select name from "user" where id = $1`

	err := db.QueryRow(q, userID).Scan(&resp.Name)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (u *user) GetUserByEmail(email string) (*model.User, error) {
	db := database.GetDB()
	if db == nil {
		fmt.Println("DB NIL")
		return nil, sql.ErrNoRows
	}
	defer db.Close()

	resp := model.User{}
	q := `
		select id, name, email, password, created_at, updated_at
		from "user"
		where email = $1
	`

	err := db.QueryRow(q, email).Scan(&resp.ID, &resp.Name, &resp.Email, &resp.Password, &resp.CreatedAt, &resp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &resp, nil
}

func (u *user) CreateUser(name, email, hashPassword string) (*model.User, error) {
	db := database.GetDB()
	defer db.Close()

	now := time.Now()
	var id int

	q := `
		insert into "user" (name, email, password, created_at)
		values ($1, $2, $3, $4)
		returning id
	`

	err := db.QueryRow(q, name, email, hashPassword, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        id,
		Name:      name,
		Email:     email,
		Password:  hashPassword,
		CreatedAt: now,
	}, nil
}
