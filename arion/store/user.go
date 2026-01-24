package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/model"
)

type (
	UserStore interface {
		GetUserByID(userID int) (*model.User, error)
		GetUserByEmail(email string) (*model.User, error)
		GetUsers() ([]model.User, error)
		CreateUser(tx *sql.Tx, name, email, hashPassword, role string, shop_id int) (*model.User, error)
		UpdateUser(id int, input UpdateUserInput) (*model.User, error)
		Roles() []string
		IsValidRole(role string) bool
	}

	user struct {
		db *sql.DB
	}

	UpdateUserInput struct {
		Name     *string
		Email    *string
		Password *string
	}
)

func NewUserStore() UserStore {
	return &user{db: database.GetDB()}
}

// NewUserStoreWithDB creates a UserStore with a custom db connection (for testing)
func NewUserStoreWithDB(db *sql.DB) UserStore {
	return &user{db: db}
}

func (u *user) GetUserByID(userID int) (*model.User, error) {
	resp := model.User{}

	q := `
		SELECT id, shop_id, name, email, password, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := u.db.QueryRow(q, userID).Scan(&resp.ID, &resp.ShopID, &resp.Name, &resp.Email, &resp.Password, &resp.Role, &resp.CreatedAt, &resp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &resp, nil
}

func (u *user) GetUserByEmail(email string) (*model.User, error) {
	resp := model.User{}
	q := `
		SELECT id, shop_id,name, email, password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := u.db.QueryRow(q, email).Scan(&resp.ID, &resp.ShopID, &resp.Name, &resp.Email, &resp.Password, &resp.Role, &resp.CreatedAt, &resp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &resp, nil
}

func (u *user) GetUsers() ([]model.User, error) {
	q := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
	`

	rows, err := u.db.Query(q)
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
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (u *user) CreateUser(tx *sql.Tx, name, email, hashPassword, role string, shop_id int) (*model.User, error) {
	now := time.Now()
	var id int

	q := `
		INSERT INTO users (name, email, password, role, shop_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := tx.QueryRow(q, name, email, hashPassword, role, shop_id, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        id,
		ShopID:    shop_id,
		Name:      name,
		Email:     email,
		Role:      role,
		CreatedAt: now,
	}, nil
}

func (u *user) UpdateUser(id int, input UpdateUserInput) (*model.User, error) {
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
		UPDATE users
		SET %s
		WHERE id = $1
		RETURNING id, name, email, created_at, updated_at
	`

	q = fmt.Sprintf(q, strings.Join(set, ","))

	err := u.db.QueryRow(q, id).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *user) Roles() []string {
	return []string{
		constant.RoleSystem,
		constant.RoleOwner,
		constant.RoleAdmin,
	}
}

func (u *user) IsValidRole(role string) bool {
	for _, validRole := range u.Roles() {
		if role == validRole {
			return true
		}
	}
	return false
}
