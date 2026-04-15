package store

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/zeirash/recapo/arion/model"
)

func Test_user_GetUserByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		userID     int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.User
		wantErr    bool
	}{
		{
			name:   "get user by ID",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "email", "password", "role", "session_token", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", "john@example.com", "hashedpass", "admin", nil, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at\s+FROM users\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.User{
				ID:        1,
				ShopID:    10,
				Name:      "John Doe",
				Email:     "john@example.com",
				Password:  "hashedpass",
				Role:      "admin",
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "get non-existent user returns nil",
			userID: 9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at\s+FROM users\s+WHERE id = \$1`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:   "get user returns error on database failure",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at\s+FROM users\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewUserStoreWithDB(db)

			got, gotErr := store.GetUserByID(context.Background(), tt.userID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetUserByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetUserByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetUserByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_user_GetUserByEmail(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		email      string
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.User
		wantErr    bool
	}{
		{
			name:  "get user by email",
			email: "john@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "email", "password", "role", "session_token", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", "john@example.com", "hashedpass", "admin", nil, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at\s+FROM users\s+WHERE email = \$1`).
					WithArgs("john@example.com").
					WillReturnRows(rows)
			},
			wantResult: &model.User{
				ID:        1,
				ShopID:    10,
				Name:      "John Doe",
				Email:     "john@example.com",
				Password:  "hashedpass",
				Role:      "admin",
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:  "get non-existent email returns nil",
			email: "notfound@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at\s+FROM users\s+WHERE email = \$1`).
					WithArgs("notfound@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:  "get user by email returns error on database failure",
			email: "john@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at\s+FROM users\s+WHERE email = \$1`).
					WithArgs("john@example.com").
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewUserStoreWithDB(db)

			got, gotErr := store.GetUserByEmail(context.Background(), tt.email)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetUserByEmail() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetUserByEmail() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetUserByEmail() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_user_GetUsers(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult []model.User
		wantErr    bool
	}{
		{
			name: "get users returns multiple users",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"}).
					AddRow(1, "John Doe", "john@example.com", "hashedpass1", "admin", fixedTime, nil).
					AddRow(2, "Jane Doe", "jane@example.com", "hashedpass2", "owner", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, name, email, password, role, created_at, updated_at\s+FROM users`).
					WillReturnRows(rows)
			},
			wantResult: []model.User{
				{ID: 1, Name: "John Doe", Email: "john@example.com", Password: "hashedpass1", Role: "admin", CreatedAt: fixedTime},
				{ID: 2, Name: "Jane Doe", Email: "jane@example.com", Password: "hashedpass2", Role: "owner", CreatedAt: fixedTime},
			},
			wantErr: false,
		},
		{
			name: "get users returns empty slice when no users exist",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT id, name, email, password, role, created_at, updated_at\s+FROM users`).
					WillReturnRows(rows)
			},
			wantResult: []model.User{},
			wantErr:    false,
		},
		{
			name: "get users returns nil when db returns no rows",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, role, created_at, updated_at\s+FROM users`).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name: "get users returns error on database failure",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, email, password, role, created_at, updated_at\s+FROM users`).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewUserStoreWithDB(db)

			got, gotErr := store.GetUsers(context.Background())

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetUsers() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetUsers() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetUsers() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_user_CreateUser(t *testing.T) {
	type input struct {
		name         string
		email        string
		hashPassword string
		role         string
		shopID       int
	}

	tests := []struct {
		name       string
		input      input
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.User
		wantErr    bool
	}{
		{
			name: "successfully create user",
			input: input{
				name:         "John Doe",
				email:        "john@example.com",
				hashPassword: "hashedpass",
				role:         "admin",
				shopID:       10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`INSERT INTO users \(name, email, password, role, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)\s+RETURNING id`).
					WithArgs("John Doe", "john@example.com", "hashedpass", "admin", 10, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.User{
				ID:     1,
				ShopID: 10,
				Name:   "John Doe",
				Email:  "john@example.com",
				Role:   "admin",
			},
			wantErr: false,
		},
		{
			name: "create user returns error on database failure",
			input: input{
				name:         "John Doe",
				email:        "john@example.com",
				hashPassword: "hashedpass",
				role:         "admin",
				shopID:       10,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO users \(name, email, password, role, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)\s+RETURNING id`).
					WithArgs("John Doe", "john@example.com", "hashedpass", "admin", 10, sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("failed to begin transaction: %v", err)
			}

			var u user
			got, gotErr := u.CreateUser(context.Background(), tx, tt.input.name, tt.input.email, tt.input.hashPassword, tt.input.role, tt.input.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateUser() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateUser() succeeded unexpectedly")
			}

			// Set expected CreatedAt to match for DeepEqual comparison
			tt.wantResult.CreatedAt = got.CreatedAt

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateUser() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_user_UpdateUser(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		id         int
		input      UpdateUserInput
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.User
		wantErr    bool
	}{
		{
			name: "update user with name",
			id:   1,
			input: UpdateUserInput{
				Name: strPtr("Updated Name"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
					AddRow(1, "Updated Name", "john@example.com", fixedTime, updatedTime)
				mock.ExpectQuery(`UPDATE users\s+SET name = \$2,updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, email, created_at, updated_at`).
					WithArgs(1, "Updated Name").
					WillReturnRows(rows)
			},
			wantResult: &model.User{
				ID:        1,
				Name:      "Updated Name",
				Email:     "john@example.com",
				CreatedAt: fixedTime,
				UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name: "update user with email",
			id:   1,
			input: UpdateUserInput{
				Email: strPtr("newemail@example.com"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
					AddRow(1, "John Doe", "newemail@example.com", fixedTime, updatedTime)
				mock.ExpectQuery(`UPDATE users\s+SET email = \$2,updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, email, created_at, updated_at`).
					WithArgs(1, "newemail@example.com").
					WillReturnRows(rows)
			},
			wantResult: &model.User{
				ID:        1,
				Name:      "John Doe",
				Email:     "newemail@example.com",
				CreatedAt: fixedTime,
				UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name: "update user with password",
			id:   1,
			input: UpdateUserInput{
				Password: strPtr("newhashedpassword"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"}).
					AddRow(1, "John Doe", "john@example.com", fixedTime, updatedTime)
				mock.ExpectQuery(`UPDATE users\s+SET password = \$2,updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, email, created_at, updated_at`).
					WithArgs(1, "newhashedpassword").
					WillReturnRows(rows)
			},
			wantResult: &model.User{
				ID:        1,
				Name:      "John Doe",
				Email:     "john@example.com",
				CreatedAt: fixedTime,
				UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: false,
		},
		{
			name: "update non-existent user returns error",
			id:   9999,
			input: UpdateUserInput{
				Name: strPtr("New Name"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE users\s+SET name = \$2,updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, email, created_at, updated_at`).
					WithArgs(9999, "New Name").
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "update user returns error on database failure",
			id:   1,
			input: UpdateUserInput{
				Name: strPtr("New Name"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE users\s+SET name = \$2,updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, email, created_at, updated_at`).
					WithArgs(1, "New Name").
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewUserStoreWithDB(db)

			got, gotErr := store.UpdateUser(context.Background(), tt.id, tt.input)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateUser() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateUser() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateUser() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_user_SetSessionToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		token     string
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:   "set session token successfully",
			userID: 1,
			token:  "abc123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET session_token = \$1, updated_at = now\(\) WHERE id = \$2`).
					WithArgs("abc123", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:   "set session token returns error on database failure",
			userID: 1,
			token:  "abc123",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET session_token = \$1, updated_at = now\(\) WHERE id = \$2`).
					WithArgs("abc123", 1).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewUserStoreWithDB(db)

			gotErr := store.SetSessionToken(context.Background(), tt.userID, tt.token)

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("SetSessionToken() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_user_ClearSessionToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name:   "clear session token successfully",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET session_token = NULL, updated_at = now\(\) WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:   "clear session token returns error on database failure",
			userID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE users SET session_token = NULL, updated_at = now\(\) WHERE id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewUserStoreWithDB(db)

			gotErr := store.ClearSessionToken(context.Background(), tt.userID)

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("ClearSessionToken() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func Test_user_IsValidRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{
			name: "valid role - system",
			role: "system",
			want: true,
		},
		{
			name: "valid role - owner",
			role: "owner",
			want: true,
		},
		{
			name: "valid role - admin",
			role: "admin",
			want: true,
		},
		{
			name: "invalid role - user",
			role: "user",
			want: false,
		},
		{
			name: "invalid role - empty string",
			role: "",
			want: false,
		},
		{
			name: "invalid role - random string",
			role: "superadmin",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u user
			got := u.IsValidRole(tt.role)

			if got != tt.want {
				t.Errorf("IsValidRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_user_GetUserByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		shopID     int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.User
		wantErr    bool
	}{
		{
			name:   "get user by shop ID",
			shopID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "shop_id", "name", "email", "password", "role", "session_token", "created_at", "updated_at"}).
					AddRow(1, 10, "John Doe", "john@example.com", "hashedpass", "admin", nil, fixedTime, nil)
				mock.ExpectQuery(`SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at\s+FROM users\s+WHERE shop_id = \$1`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			wantResult: &model.User{
				ID:        1,
				ShopID:    10,
				Name:      "John Doe",
				Email:     "john@example.com",
				Password:  "hashedpass",
				Role:      "admin",
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "get non-existent shop ID returns nil",
			shopID: 9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at\s+FROM users\s+WHERE shop_id = \$1`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:   "get user by shop ID returns error on database failure",
			shopID: 10,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, shop_id, name, email, password, role, session_token, created_at, updated_at\s+FROM users\s+WHERE shop_id = \$1`).
					WithArgs(10).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)
			store := NewUserStoreWithDB(db)

			got, gotErr := store.GetUserByShopID(context.Background(), tt.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetUserByShopID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetUserByShopID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetUserByShopID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_user_CountUsersByShopID(t *testing.T) {
	tests := []struct {
		name      string
		shopID    int
		mockSetup func(mock sqlmock.Sqlmock)
		want      int
		wantErr   bool
	}{
		{
			name:   "returns count of users for shop",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(3)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE shop_id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: 3,
		},
		{
			name:   "returns zero when no users in shop",
			shopID: 99,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE shop_id = \$1`).
					WithArgs(99).
					WillReturnRows(rows)
			},
			want: 0,
		},
		{
			name:   "returns error on database failure",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE shop_id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			u := &user{db: db}
			got, gotErr := u.CountUsersByShopID(context.Background(), tt.shopID)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CountUsersByShopID() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CountUsersByShopID() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("CountUsersByShopID() = %v, want %v", got, tt.want)
			}
		})
	}
}
