package store

import (
	"context"
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
		GetUserByID(ctx context.Context, userID int) (*model.User, error)
		GetOwnerByShopID(ctx context.Context, shopID int) (*model.User, error)
		GetUserByEmail(ctx context.Context, email string) (*model.User, error)
		GetUsers(ctx context.Context) ([]model.User, error)
		GetUsersByShopID(ctx context.Context, shopID int) ([]model.User, error)
		CountUsersByShopID(ctx context.Context, shopID int) (int, error)
		CreateUser(ctx context.Context, tx database.Tx, name, email, hashPassword, role string, shop_id int) (*model.User, error)
		UpdateUser(ctx context.Context, id int, input UpdateUserInput) (*model.User, error)
		SetSessionToken(ctx context.Context, userID int, sessionToken string) error
		ClearSessionToken(ctx context.Context, userID int) error
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

func (u *user) GetUserByID(ctx context.Context, userID int) (*model.User, error) {
	resp := model.User{}

	q := `
		SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := u.db.QueryRowContext(ctx, q, userID).Scan(&resp.ID, &resp.ShopID, &resp.Name, &resp.Email, &resp.Password, &resp.Role, &resp.SessionToken, &resp.CreatedAt, &resp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &resp, nil
}

func (u *user) GetOwnerByShopID(ctx context.Context, shopID int) (*model.User, error) {
	resp := model.User{}

	q := `
		SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at
		FROM users
		WHERE shop_id = $1 AND role = 'owner'
	`

	err := u.db.QueryRowContext(ctx, q, shopID).Scan(&resp.ID, &resp.ShopID, &resp.Name, &resp.Email, &resp.Password, &resp.Role, &resp.SessionToken, &resp.CreatedAt, &resp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &resp, nil
}

func (u *user) CountUsersByShopID(ctx context.Context, shopID int) (int, error) {
	var count int
	q := `SELECT COUNT(*) FROM users WHERE shop_id = $1`
	err := u.db.QueryRowContext(ctx, q, shopID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (u *user) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	resp := model.User{}
	q := `
		SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	err := u.db.QueryRowContext(ctx, q, email).Scan(&resp.ID, &resp.ShopID, &resp.Name, &resp.Email, &resp.Password, &resp.Role, &resp.SessionToken, &resp.CreatedAt, &resp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &resp, nil
}

func (u *user) GetUsers(ctx context.Context) ([]model.User, error) {
	q := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
	`

	rows, err := u.db.QueryContext(ctx, q)
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

func (u *user) GetUsersByShopID(ctx context.Context, shopID int) ([]model.User, error) {
	q := `
		SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at
		FROM users
		WHERE shop_id = $1
		ORDER BY created_at ASC
	`

	rows, err := u.db.QueryContext(ctx, q, shopID)
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
		err := rows.Scan(&user.ID, &user.ShopID, &user.Name, &user.Email, &user.Password, &user.Role, &user.SessionToken, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (u *user) CreateUser(ctx context.Context, tx database.Tx, name, email, hashPassword, role string, shop_id int) (*model.User, error) {
	now := time.Now()
	var id int

	q := `
		INSERT INTO users (name, email, password, role, shop_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := tx.QueryRowContext(ctx, q, name, email, hashPassword, role, shop_id, now).Scan(&id)
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

func (u *user) UpdateUser(ctx context.Context, id int, input UpdateUserInput) (*model.User, error) {
	set := []string{}
	args := []interface{}{id}
	argNum := 2
	var user model.User

	// build query
	if input.Name != nil {
		set = append(set, fmt.Sprintf("name = $%d", argNum))
		args = append(args, *input.Name)
		argNum++
	}
	if input.Email != nil {
		set = append(set, fmt.Sprintf("email = $%d", argNum))
		args = append(args, *input.Email)
		argNum++
	}
	if input.Password != nil {
		set = append(set, fmt.Sprintf("password = $%d", argNum))
		args = append(args, *input.Password)
		argNum++
	}

	set = append(set, "updated_at = now()")

	q := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $1
		RETURNING id, name, email, created_at, updated_at
	`, strings.Join(set, ","))

	err := u.db.QueryRowContext(ctx, q, args...).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *user) SetSessionToken(ctx context.Context, userID int, sessionToken string) error {
	q := `UPDATE users SET session_token = $1, updated_at = now() WHERE id = $2`
	_, err := u.db.ExecContext(ctx, q, sessionToken, userID)
	return err
}

func (u *user) ClearSessionToken(ctx context.Context, userID int) error {
	q := `UPDATE users SET session_token = NULL, updated_at = now() WHERE id = $1`
	_, err := u.db.ExecContext(ctx, q, userID)
	return err
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
