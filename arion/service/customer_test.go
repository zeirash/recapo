package service

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/response"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/store"
)

func strPtr(s string) *string {
	return &s
}

func Test_cservice_CreateCustomer(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	type input struct {
		name    string
		phone   string
		address string
		shopID  int
	}

	tests := []struct {
		name       string
		input      input
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockCustomerStore
		wantResult response.CustomerData
		wantErr    bool
	}{
		{
			name: "successfully create customer",
			input: input{
				name:    "John Doe",
				phone:   "1234567890",
				address: "123 Main St",
				shopID:  10,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					CreateCustomer(gomock.Eq(store.CreateCustomerInput{
						Name:    "John Doe",
						Phone:   "1234567890",
						Address: strPtr("123 Main St"),
						ShopID:  10,
					})).
					Return(&model.Customer{
						ID:        1,
						Name:      "John Doe",
						Phone:     "1234567890",
						Address:   "123 Main St",
						CreatedAt: fixedTime,
					}, nil)
				return mock
			},
			wantResult: response.CustomerData{
				ID:        1,
				Name:      "John Doe",
				Phone:     "1234567890",
				Address:   "123 Main St",
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name: "successfully create customer without address",
			input: input{
				name:    "Jane Doe",
				phone:   "0987654321",
				address: "",
				shopID:  10,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					CreateCustomer(gomock.Eq(store.CreateCustomerInput{
						Name:    "Jane Doe",
						Phone:   "0987654321",
						Address: strPtr(""),
						ShopID:  10,
					})).
					Return(&model.Customer{
						ID:        2,
						Name:      "Jane Doe",
						Phone:     "0987654321",
						Address:   "",
						CreatedAt: fixedTime,
					}, nil)
				return mock
			},
			wantResult: response.CustomerData{
				ID:        2,
				Name:      "Jane Doe",
				Phone:     "0987654321",
				Address:   "",
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name: "create customer with duplicate phone returns error",
			input: input{
				name:    "John Doe",
				phone:   "1234567890",
				address: "123 Main St",
				shopID:  10,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					CreateCustomer(gomock.Eq(store.CreateCustomerInput{
						Name:    "John Doe",
						Phone:   "1234567890",
						Address: strPtr("123 Main St"),
						ShopID:  10,
					})).
					Return(nil, store.ErrDuplicatePhone)
				return mock
			},
			wantResult: response.CustomerData{},
			wantErr:    true,
		},
		{
			name: "create customer returns error on database failure",
			input: input{
				name:    "John Doe",
				phone:   "1234567890",
				address: "123 Main St",
				shopID:  10,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					CreateCustomer(gomock.Eq(store.CreateCustomerInput{
						Name:    "John Doe",
						Phone:   "1234567890",
						Address: strPtr("123 Main St"),
						ShopID:  10,
					})).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.CustomerData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := customerStore
			defer func() { customerStore = oldStore }()
			customerStore = tt.mockSetup(ctrl)

			var c cservice
			got, gotErr := c.CreateCustomer(tt.input.name, tt.input.phone, tt.input.address, tt.input.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateCustomer() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateCustomer() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateCustomer() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_cservice_GetCustomerByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	type input struct {
		customerID int
		shopID     []int
	}

	tests := []struct {
		name       string
		input      input
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockCustomerStore
		wantResult *response.CustomerData
		wantErr    bool
	}{
		{
			name: "get customer by ID without shop filter",
			input: input{
				customerID: 1,
				shopID:     nil,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomerByID(1).
					Return(&model.Customer{
						ID:        1,
						Name:      "John Doe",
						Phone:     "1234567890",
						Address:   "123 Main St",
						CreatedAt: fixedTime,
					}, nil)
				return mock
			},
			wantResult: &response.CustomerData{
				ID:        1,
				Name:      "John Doe",
				Phone:     "1234567890",
				Address:   "123 Main St",
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name: "get customer by ID with shop filter",
			input: input{
				customerID: 1,
				shopID:     []int{10},
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomerByID(1, 10).
					Return(&model.Customer{
						ID:        1,
						Name:      "John Doe",
						Phone:     "1234567890",
						Address:   "123 Main St",
						CreatedAt: fixedTime,
						UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: &response.CustomerData{
				ID:        1,
				Name:      "John Doe",
				Phone:     "1234567890",
				Address:   "123 Main St",
				CreatedAt: fixedTime,
				UpdatedAt: &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "get customer not found returns error",
			input: input{
				customerID: 9999,
				shopID:     nil,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomerByID(9999).
					Return(nil, nil)
				return mock
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name: "get customer returns error on database failure",
			input: input{
				customerID: 1,
				shopID:     nil,
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomerByID(1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := customerStore
			defer func() { customerStore = oldStore }()
			customerStore = tt.mockSetup(ctrl)

			var c cservice
			got, gotErr := c.GetCustomerByID(tt.input.customerID, tt.input.shopID...)

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
			}
		})
	}
}

func Test_cservice_GetCustomersByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name        string
		shopID      int
		searchQuery *string
		mockSetup   func(ctrl *gomock.Controller) *mock_store.MockCustomerStore
		wantResult  []response.CustomerData
		wantErr     bool
	}{
		{
			name:   "get customers by shop ID returns multiple customers",
			shopID: 10,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomersByShopID(10, nil).
					Return([]model.Customer{
						{ID: 1, Name: "John Doe", Phone: "1234567890", Address: "123 Main St", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{Time: fixedTime, Valid: true}},
						{ID: 2, Name: "Jane Doe", Phone: "0987654321", Address: "456 Oak Ave", CreatedAt: fixedTime},
					}, nil)
				return mock
			},
			wantResult: []response.CustomerData{
				{ID: 1, Name: "John Doe", Phone: "1234567890", Address: "123 Main St", CreatedAt: fixedTime, UpdatedAt: &fixedTime},
				{ID: 2, Name: "Jane Doe", Phone: "0987654321", Address: "456 Oak Ave", CreatedAt: fixedTime},
			},
			wantErr: false,
		},
		{
			name:   "get customers by shop ID returns empty slice",
			shopID: 10,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomersByShopID(10, nil).
					Return([]model.Customer{}, nil)
				return mock
			},
			wantResult: []response.CustomerData{},
			wantErr:    false,
		},
		{
			name:   "get customers by shop ID returns error on database failure",
			shopID: 10,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomersByShopID(10, nil).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: []response.CustomerData{},
			wantErr:    true,
		},
		{
			name:        "get customers by shop ID with search query returns filtered customers",
			shopID:      10,
			searchQuery: strPtr("john"),
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomersByShopID(10, gomock.Any()).
					Return([]model.Customer{
						{ID: 1, Name: "John Doe", Phone: "1234567890", Address: "123 Main St", CreatedAt: fixedTime},
					}, nil)
				return mock
			},
			wantResult: []response.CustomerData{
				{ID: 1, Name: "John Doe", Phone: "1234567890", Address: "123 Main St", CreatedAt: fixedTime},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := customerStore
			defer func() { customerStore = oldStore }()
			customerStore = tt.mockSetup(ctrl)

			var c cservice
			got, gotErr := c.GetCustomersByShopID(tt.shopID, tt.searchQuery)

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

func Test_cservice_UpdateCustomer(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		input      UpdateCustomerInput
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockCustomerStore
		wantResult response.CustomerData
		wantErr    bool
	}{
		{
			name: "successfully update customer",
			input: UpdateCustomerInput{
				ID:   1,
				Name: strPtr("Updated Name"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomerByID(1).
					Return(&model.Customer{
						ID:        1,
						Name:      "John Doe",
						Phone:     "1234567890",
						Address:   "123 Main St",
						CreatedAt: fixedTime,
					}, nil)
				mock.EXPECT().
					UpdateCustomer(1, store.UpdateCustomerInput{Name: strPtr("Updated Name")}).
					Return(&model.Customer{
						ID:        1,
						Name:      "Updated Name",
						Phone:     "1234567890",
						Address:   "123 Main St",
						CreatedAt: fixedTime,
						UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.CustomerData{
				ID:        1,
				Name:      "Updated Name",
				Phone:     "1234567890",
				Address:   "123 Main St",
				CreatedAt: fixedTime,
				UpdatedAt: &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "update customer not found returns error",
			input: UpdateCustomerInput{
				ID:   9999,
				Name: strPtr("Updated Name"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomerByID(9999).
					Return(nil, nil)
				return mock
			},
			wantResult: response.CustomerData{},
			wantErr:    true,
		},
		{
			name: "update customer returns error on get failure",
			input: UpdateCustomerInput{
				ID:   1,
				Name: strPtr("Updated Name"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomerByID(1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.CustomerData{},
			wantErr:    true,
		},
		{
			name: "update customer returns error on update failure",
			input: UpdateCustomerInput{
				ID:    1,
				Phone: strPtr("9999999999"),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					GetCustomerByID(1).
					Return(&model.Customer{
						ID:        1,
						Name:      "John Doe",
						Phone:     "1234567890",
						Address:   "123 Main St",
						CreatedAt: fixedTime,
					}, nil)
				mock.EXPECT().
					UpdateCustomer(1, store.UpdateCustomerInput{Phone: strPtr("9999999999")}).
					Return(nil, store.ErrDuplicatePhone)
				return mock
			},
			wantResult: response.CustomerData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := customerStore
			defer func() { customerStore = oldStore }()
			customerStore = tt.mockSetup(ctrl)

			var c cservice
			got, gotErr := c.UpdateCustomer(tt.input)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateCustomer() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateCustomer() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateCustomer() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_cservice_DeleteCustomerByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		mockSetup func(ctrl *gomock.Controller) *mock_store.MockCustomerStore
		wantErr   bool
	}{
		{
			name: "successfully delete customer",
			id:   1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					DeleteCustomerByID(1).
					Return(nil)
				return mock
			},
			wantErr: false,
		},
		{
			name: "delete customer returns error on database failure",
			id:   1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockCustomerStore {
				mock := mock_store.NewMockCustomerStore(ctrl)
				mock.EXPECT().
					DeleteCustomerByID(1).
					Return(errors.New("database error"))
				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := customerStore
			defer func() { customerStore = oldStore }()
			customerStore = tt.mockSetup(ctrl)

			var c cservice
			gotErr := c.DeleteCustomerByID(tt.id)

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

func Test_cservice_HasActiveOrders(t *testing.T) {
	tests := []struct {
		name       string
		customerID int
		shopID     int
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockOrderStore
		want       response.CustomerHasActiveOrdersData
		wantErr    bool
	}{
		{
			name:       "returns has_active_orders true when customer has active order",
			customerID: 10,
			shopID:     5,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					HasActiveOrdersByCustomerID(10, 5).
					Return(true, nil)
				return mock
			},
			want:    response.CustomerHasActiveOrdersData{HasActiveOrders: true},
			wantErr: false,
		},
		{
			name:       "returns has_active_orders false when customer has no active order",
			customerID: 99,
			shopID:     5,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					HasActiveOrdersByCustomerID(99, 5).
					Return(false, nil)
				return mock
			},
			want:    response.CustomerHasActiveOrdersData{HasActiveOrders: false},
			wantErr: false,
		},
		{
			name:       "returns error when store fails",
			customerID: 10,
			shopID:     5,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					HasActiveOrdersByCustomerID(10, 5).
					Return(false, errors.New("database error"))
				return mock
			},
			want:    response.CustomerHasActiveOrdersData{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderStore
			defer func() { orderStore = oldStore }()
			orderStore = tt.mockSetup(ctrl)

			var c cservice
			got, gotErr := c.HasActiveOrders(tt.customerID, tt.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("HasActiveOrders() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("HasActiveOrders() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HasActiveOrders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cservice_CheckActiveOrderByPhone(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		phone      string
		nameParam  string
		shopID     int
		mockSetup  func(ctrl *gomock.Controller) (*mock_store.MockCustomerStore, *mock_store.MockOrderStore)
		want       response.CustomerCheckActiveOrderByPhone
		wantErr    bool
	}{
		{
			name:      "customer found with active orders",
			phone:     "08123456789",
			nameParam: "John Doe",
			shopID:    1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockCustomerStore, *mock_store.MockOrderStore) {
				cust := mock_store.NewMockCustomerStore(ctrl)
				ord := mock_store.NewMockOrderStore(ctrl)
				cust.EXPECT().GetCustomerByPhone("08123456789", 1).
					Return(&model.Customer{ID: 1, Name: "John Doe", Phone: "08123456789", Address: "", CreatedAt: fixedTime}, nil)
				ord.EXPECT().HasActiveOrdersByCustomerID(1, 1).Return(true, nil)
				return cust, ord
			},
			want:    response.CustomerCheckActiveOrderByPhone{CustomerID: 1, HasActiveOrders: true},
			wantErr: false,
		},
		{
			name:      "customer found without active orders",
			phone:     "08987654321",
			nameParam: "Jane Doe",
			shopID:    1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockCustomerStore, *mock_store.MockOrderStore) {
				cust := mock_store.NewMockCustomerStore(ctrl)
				ord := mock_store.NewMockOrderStore(ctrl)
				cust.EXPECT().GetCustomerByPhone("08987654321", 1).
					Return(&model.Customer{ID: 2, Name: "Jane Doe", Phone: "08987654321", Address: "", CreatedAt: fixedTime}, nil)
				ord.EXPECT().HasActiveOrdersByCustomerID(2, 1).Return(false, nil)
				return cust, ord
			},
			want:    response.CustomerCheckActiveOrderByPhone{CustomerID: 2, HasActiveOrders: false},
			wantErr: false,
		},
		{
			name:      "customer not found creates customer and returns no active orders",
			phone:     "08000000000",
			nameParam: "New User",
			shopID:    1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockCustomerStore, *mock_store.MockOrderStore) {
				cust := mock_store.NewMockCustomerStore(ctrl)
				ord := mock_store.NewMockOrderStore(ctrl)
				cust.EXPECT().GetCustomerByPhone("08000000000", 1).Return(nil, nil)
				cust.EXPECT().CreateCustomer(gomock.Eq(store.CreateCustomerInput{
					Name: "New User", Phone: "08000000000", Address: nil, ShopID: 1,
				})).Return(&model.Customer{ID: 3, Name: "New User", Phone: "08000000000", Address: "", CreatedAt: fixedTime}, nil)
				ord.EXPECT().HasActiveOrdersByCustomerID(3, 1).Return(false, nil)
				return cust, ord
			},
			want:    response.CustomerCheckActiveOrderByPhone{CustomerID: 3, HasActiveOrders: false},
			wantErr: false,
		},
		{
			name:      "GetCustomerByPhone returns error",
			phone:     "08123456789",
			nameParam: "John",
			shopID:    1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockCustomerStore, *mock_store.MockOrderStore) {
				cust := mock_store.NewMockCustomerStore(ctrl)
				ord := mock_store.NewMockOrderStore(ctrl)
				cust.EXPECT().GetCustomerByPhone("08123456789", 1).Return(nil, errors.New("database error"))
				return cust, ord
			},
			want:    response.CustomerCheckActiveOrderByPhone{},
			wantErr: true,
		},
		{
			name:      "customer not found and CreateCustomer returns error",
			phone:     "08000000000",
			nameParam: "New User",
			shopID:    1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockCustomerStore, *mock_store.MockOrderStore) {
				cust := mock_store.NewMockCustomerStore(ctrl)
				ord := mock_store.NewMockOrderStore(ctrl)
				cust.EXPECT().GetCustomerByPhone("08000000000", 1).Return(nil, nil)
				cust.EXPECT().CreateCustomer(gomock.Eq(store.CreateCustomerInput{
					Name: "New User", Phone: "08000000000", Address: nil, ShopID: 1,
				})).Return(nil, store.ErrDuplicatePhone)
				return cust, ord
			},
			want:    response.CustomerCheckActiveOrderByPhone{},
			wantErr: true,
		},
		{
			name:      "HasActiveOrdersByCustomerID returns error",
			phone:     "08123456789",
			nameParam: "John Doe",
			shopID:    1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockCustomerStore, *mock_store.MockOrderStore) {
				cust := mock_store.NewMockCustomerStore(ctrl)
				ord := mock_store.NewMockOrderStore(ctrl)
				cust.EXPECT().GetCustomerByPhone("08123456789", 1).
					Return(&model.Customer{ID: 1, Name: "John Doe", Phone: "08123456789", Address: "", CreatedAt: fixedTime}, nil)
				ord.EXPECT().HasActiveOrdersByCustomerID(1, 1).Return(false, errors.New("database error"))
				return cust, ord
			},
			want:    response.CustomerCheckActiveOrderByPhone{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCust, mockOrder := tt.mockSetup(ctrl)
			oldCust, oldOrder := customerStore, orderStore
			defer func() { customerStore, orderStore = oldCust, oldOrder }()
			customerStore = mockCust
			orderStore = mockOrder

			var c cservice
			got, gotErr := c.CheckActiveOrderByPhone(tt.phone, tt.nameParam, tt.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CheckActiveOrderByPhone() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CheckActiveOrderByPhone() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckActiveOrderByPhone() = %v, want %v", got, tt.want)
			}
		})
	}
}
