package store

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/zeirash/recapo/arion/model"
)

func Test_customer_GetCustomerByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		id         int
		shopID     []int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Customer
		wantErr    bool
	}{
		{
			name:   "get customer by ID without shop filter",
			id:     1,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "address", "created_at", "updated_at"}).
					AddRow(1, "John Doe", "1234567890", "123 Main St", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Customer{
				ID:        1,
				Name:      "John Doe",
				Phone:     "1234567890",
				Address:   "123 Main St",
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "get customer by ID with shop filter",
			id:     1,
			shopID: []int{1},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "address", "created_at", "updated_at"}).
					AddRow(1, "Jane Doe", "0987654321", "456 Oak Ave", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE id = \$1\s+AND shop_id = \$2`).
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantResult: &model.Customer{
				ID:        1,
				Name:      "Jane Doe",
				Phone:     "0987654321",
				Address:   "456 Oak Ave",
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "get non-existent customer returns nil",
			id:     9999,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE id = \$1`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:   "get customer with wrong shop ID returns nil",
			id:     1,
			shopID: []int{9999},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE id = \$1\s+AND shop_id = \$2`).
					WithArgs(1, 9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:   "get customer and return error",
			id:     1,
			shopID: nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE id = \$1`).
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
			store := NewCustomerStoreWithDB(db)

			got, gotErr := store.GetCustomerByID(tt.id, tt.shopID...)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetCustomerByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetCustomerByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetCustomerByID() = %v, want %v", got, tt.wantResult)
				return
			}
		})
	}
}

func Test_customer_GetCustomersByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name        string
		shopID      int
		searchQuery *string
		mockSetup   func(mock sqlmock.Sqlmock)
		wantResult  []model.Customer
		wantErr     bool
	}{
		{
			name:   "get customers by shop ID returns multiple customers",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "address", "created_at", "updated_at"}).
					AddRow(1, "John Doe", "1234567890", "123 Main St", fixedTime, nil).
					AddRow(2, "Jane Doe", "0987654321", "456 Oak Ave", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE shop_id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: []model.Customer{
				{
					ID:        1,
					Name:      "John Doe",
					Phone:     "1234567890",
					Address:   "123 Main St",
					CreatedAt: fixedTime,
				},
				{
					ID:        2,
					Name:      "Jane Doe",
					Phone:     "0987654321",
					Address:   "456 Oak Ave",
					CreatedAt: fixedTime,
				},
			},
			wantErr: false,
		},
		{
			name:   "get customers by shop ID returns empty slice when no customers exist",
			shopID: 9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "address", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE shop_id = \$1`).
					WithArgs(9999).
					WillReturnRows(rows)
			},
			wantResult: []model.Customer{},
			wantErr:    false,
		},
		{
			name:   "get customers by shop ID returns error on database failure",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE shop_id = \$1`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get customers by shop ID returns nil when db returns no rows",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE shop_id = \$1`).
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:        "get customers by shop ID with search query filters by name, phone",
			shopID:      1,
			searchQuery: strPtr("john"),
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "address", "created_at", "updated_at"}).
					AddRow(1, "John Doe", "1234567890", "123 Main St", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE shop_id = \$1\s+AND \(name ILIKE \$2 OR phone ILIKE \$2\)`).
					WithArgs(1, "%john%").
					WillReturnRows(rows)
			},
			wantResult: []model.Customer{
				{
					ID:        1,
					Name:      "John Doe",
					Phone:     "1234567890",
					Address:   "123 Main St",
					CreatedAt: fixedTime,
				},
			},
			wantErr: false,
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
			store := NewCustomerStoreWithDB(db)

			got, gotErr := store.GetCustomersByShopID(tt.shopID, tt.searchQuery)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetCustomersByShopID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetCustomersByShopID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetCustomersByShopID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_customer_CreateCustomer(t *testing.T) {
	tests := []struct {
		name       string
		input      CreateCustomerInput
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Customer
		wantErr    error
	}{
		{
			name: "successfully create customer with address",
			input: CreateCustomerInput{
				Name:    "John Doe",
				Phone:   "1234567890",
				Address: strPtr("123 Main St"),
				ShopID:  1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`INSERT INTO customers \(name, phone, address, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, \$5\)\s+RETURNING id`).
					WithArgs("John Doe", "1234567890", "123 Main St", 1, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.Customer{
				ID:      1,
				Name:    "John Doe",
				Phone:   "1234567890",
				Address: "123 Main St",
			},
			wantErr: nil,
		},
		{
			name: "successfully create customer without address",
			input: CreateCustomerInput{
				Name:    "Jane Doe",
				Phone:   "0987654321",
				Address: nil,
				ShopID:  1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(`INSERT INTO customers \(name, phone, address, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, \$5\)\s+RETURNING id`).
					WithArgs("Jane Doe", "0987654321", "", 1, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantResult: &model.Customer{
				ID:      2,
				Name:    "Jane Doe",
				Phone:   "0987654321",
				Address: "",
			},
			wantErr: nil,
		},
		{
			name: "create customer with duplicate phone returns ErrDuplicatePhone",
			input: CreateCustomerInput{
				Name:    "Jane Doe",
				Phone:   "1234567890",
				Address: strPtr("456 Oak Ave"),
				ShopID:  1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO customers \(name, phone, address, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, \$5\)\s+RETURNING id`).
					WithArgs("Jane Doe", "1234567890", "456 Oak Ave", 1, sqlmock.AnyArg()).
					WillReturnError(&pq.Error{Code: "23505"})
			},
			wantResult: nil,
			wantErr:    ErrDuplicatePhone,
		},
		{
			name: "create customer returns error on database failure",
			input: CreateCustomerInput{
				Name:    "John Doe",
				Phone:   "1234567890",
				Address: strPtr("123 Main St"),
				ShopID:  1,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO customers \(name, phone, address, shop_id, created_at\)\s+VALUES \(\$1, \$2, \$3, \$4, \$5\)\s+RETURNING id`).
					WithArgs("John Doe", "1234567890", "123 Main St", 1, sqlmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    errors.New("database error"),
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
			store := NewCustomerStoreWithDB(db)

			got, gotErr := store.CreateCustomer(tt.input)

			if tt.wantErr != nil {
				if gotErr == nil {
					t.Fatalf("CreateCustomer() expected error %v, got nil", tt.wantErr)
				}
				if tt.wantErr == ErrDuplicatePhone {
					if gotErr != ErrDuplicatePhone {
						t.Errorf("CreateCustomer() error = %v, want %v", gotErr, tt.wantErr)
					}
				}
				return
			}
			if gotErr != nil {
				t.Fatalf("CreateCustomer() unexpected error: %v", gotErr)
			}

			// Set expected CreatedAt to match for DeepEqual comparison
			tt.wantResult.CreatedAt = got.CreatedAt
			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateCustomer() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func Test_customer_UpdateCustomer(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		id         int
		input      UpdateCustomerInput
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Customer
		wantErr    error
	}{
		{
			name: "update customer with all fields",
			id:   1,
			input: UpdateCustomerInput{
				Name:    strPtr("John Updated"),
				Phone:   strPtr("9999999999"),
				Address: strPtr("789 New St"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "address", "created_at", "updated_at"}).
					AddRow(1, "John Updated", "9999999999", "789 New St", fixedTime, updatedTime)
				mock.ExpectQuery(`UPDATE customers\s+SET name = 'John Updated',phone = '9999999999',address = '789 New St',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, phone, address, created_at, updated_at`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Customer{
				ID:        1,
				Name:      "John Updated",
				Phone:     "9999999999",
				Address:   "789 New St",
				CreatedAt: fixedTime,
				UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: nil,
		},
		{
			name: "update customer with name only",
			id:   1,
			input: UpdateCustomerInput{
				Name: strPtr("Jane Updated"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "address", "created_at", "updated_at"}).
					AddRow(1, "Jane Updated", "1234567890", "123 Main St", fixedTime, updatedTime)
				mock.ExpectQuery(`UPDATE customers\s+SET name = 'Jane Updated',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, phone, address, created_at, updated_at`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantResult: &model.Customer{
				ID:        1,
				Name:      "Jane Updated",
				Phone:     "1234567890",
				Address:   "123 Main St",
				CreatedAt: fixedTime,
				UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
			},
			wantErr: nil,
		},
		{
			name: "update customer with duplicate phone returns ErrDuplicatePhone",
			id:   1,
			input: UpdateCustomerInput{
				Phone: strPtr("1234567890"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE customers\s+SET phone = '1234567890',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, phone, address, created_at, updated_at`).
					WithArgs(1).
					WillReturnError(&pq.Error{Code: "23505"})
			},
			wantResult: nil,
			wantErr:    ErrDuplicatePhone,
		},
		{
			name: "update non-existent customer returns error",
			id:   9999,
			input: UpdateCustomerInput{
				Name: strPtr("Ghost"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE customers\s+SET name = 'Ghost',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, phone, address, created_at, updated_at`).
					WithArgs(9999).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    sql.ErrNoRows,
		},
		{
			name: "update customer returns error on database failure",
			id:   1,
			input: UpdateCustomerInput{
				Name: strPtr("John"),
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE customers\s+SET name = 'John',updated_at = now\(\)\s+WHERE id = \$1\s+RETURNING id, name, phone, address, created_at, updated_at`).
					WithArgs(1).
					WillReturnError(errors.New("database error"))
			},
			wantResult: nil,
			wantErr:    errors.New("database error"),
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
			store := NewCustomerStoreWithDB(db)

			got, gotErr := store.UpdateCustomer(tt.id, tt.input)

			if tt.wantErr != nil {
				if gotErr == nil {
					t.Fatalf("UpdateCustomer() expected error %v, got nil", tt.wantErr)
				}
				if tt.wantErr == ErrDuplicatePhone && gotErr != ErrDuplicatePhone {
					t.Errorf("UpdateCustomer() error = %v, want %v", gotErr, tt.wantErr)
				}
				return
			}
			if gotErr != nil {
				t.Fatalf("UpdateCustomer() unexpected error: %v", gotErr)
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateCustomer() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_customer_DeleteCustomerByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		mockSetup func(mock sqlmock.Sqlmock)
		wantErr   bool
	}{
		{
			name: "successfully delete customer",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM customers\s+WHERE id = \$1`).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "delete non-existent customer succeeds",
			id:   9999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM customers\s+WHERE id = \$1`).
					WithArgs(9999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: false,
		},
		{
			name: "delete customer returns error on database failure",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`DELETE FROM customers\s+WHERE id = \$1`).
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
			store := NewCustomerStoreWithDB(db)

			gotErr := store.DeleteCustomerByID(tt.id)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteCustomerByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteCustomerByID() succeeded unexpectedly")
			}
		})
	}
}

func Test_customer_GetCustomerByPhone(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		phone      string
		shopID     int
		mockSetup  func(mock sqlmock.Sqlmock)
		wantResult *model.Customer
		wantErr    bool
	}{
		{
			name:   "returns customer when found",
			phone:  "08123456789",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "phone", "address", "created_at", "updated_at"}).
					AddRow(1, "John Doe", "08123456789", "123 Main St", fixedTime, nil)
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE phone = \$1 AND shop_id = \$2`).
					WithArgs("08123456789", 1).
					WillReturnRows(rows)
			},
			wantResult: &model.Customer{
				ID:        1,
				Name:      "John Doe",
				Phone:     "08123456789",
				Address:   "123 Main St",
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "returns nil nil when no rows",
			phone:  "08000000000",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE phone = \$1 AND shop_id = \$2`).
					WithArgs("08000000000", 1).
					WillReturnError(sql.ErrNoRows)
			},
			wantResult: nil,
			wantErr:    false,
		},
		{
			name:   "returns error on database failure",
			phone:  "08123456789",
			shopID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT id, name, phone, address, created_at, updated_at\s+FROM customers\s+WHERE phone = \$1 AND shop_id = \$2`).
					WithArgs("08123456789", 1).
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
			s := NewCustomerStoreWithDB(db)

			got, gotErr := s.GetCustomerByPhone(tt.phone, tt.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetCustomerByPhone() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetCustomerByPhone() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetCustomerByPhone() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}
