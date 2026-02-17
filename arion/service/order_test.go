package service

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/common/response"
	mock_database "github.com/zeirash/recapo/arion/mock/database"
	mock_store "github.com/zeirash/recapo/arion/mock/store"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/store"
)

func Test_oservice_CreateOrder(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		customerID int
		shopID     int
		notes      *string
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockOrderStore
		wantResult response.OrderData
		wantErr    bool
		wantErrMsg string // if non-empty, error must contain this string
	}{
		{
			name:       "successfully create order",
			customerID: 1,
			shopID:     1,
			notes:      nil,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetActiveOrderByCustomerID(1, 1).
					Return(nil, nil)
				mock.EXPECT().
					CreateOrder(nil, 1, 1, nil, nil).
					Return(&model.Order{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   0,
						Status:       constant.OrderStatusCreated,
						CreatedAt:    fixedTime,
					}, nil)
				return mock
			},
			wantResult: response.OrderData{
				ID:           1,
				CustomerName: "John Doe",
				TotalPrice:   0,
				Status:       constant.OrderStatusCreated,
				CreatedAt:    fixedTime,
			},
			wantErr: false,
		},
		{
			name:       "create order returns error when customer has active order",
			customerID: 1,
			shopID:     1,
			notes:      nil,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetActiveOrderByCustomerID(1, 1).
					Return(&model.Order{ID: 1}, nil)
				return mock
			},
			wantResult: response.OrderData{},
			wantErr:    true,
			wantErrMsg: "customer already has an active order",
		},
		{
			name:       "create order returns error when GetActiveOrderByCustomerID fails",
			customerID: 1,
			shopID:     1,
			notes:      nil,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetActiveOrderByCustomerID(1, 1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.OrderData{},
			wantErr:    true,
		},
		{
			name:       "create order returns error on CreateOrder store failure",
			customerID: 1,
			shopID:     1,
			notes:      nil,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetActiveOrderByCustomerID(1, 1).
					Return(nil, nil)
				mock.EXPECT().
					CreateOrder(nil, 1, 1, nil, nil).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.OrderData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderStore
			defer func() { orderStore = oldStore }()
			orderStore = tt.mockSetup(ctrl)

			var o oservice
			got, gotErr := o.CreateOrder(tt.customerID, tt.shopID, tt.notes)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateOrder() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				if tt.wantErrMsg != "" && !strings.Contains(gotErr.Error(), tt.wantErrMsg) {
					t.Errorf("CreateOrder() error = %v, want message containing %q", gotErr, tt.wantErrMsg)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateOrder() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateOrder() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_GetOrderByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		id         int
		shopID     []int
		mockSetup  func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore)
		wantResult *response.OrderData
		wantErr    bool
	}{
		{
			name:   "successfully get order by ID",
			id:     1,
			shopID: []int{1},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore) {
				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					GetOrderByID(1, 1).
					Return(&model.Order{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   100,
						Status:       constant.OrderStatusCreated,
						CreatedAt:    fixedTime,
					}, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				mockOrderItem.EXPECT().
					GetOrderItemsByOrderID(1).
					Return([]model.OrderItem{
						{ID: 1, ProductName: "Product 1", Price: 50, Qty: 2, CreatedAt: fixedTime},
					}, nil)

				return mockOrder, mockOrderItem
			},
			wantResult: &response.OrderData{
				ID:           1,
				CustomerName: "John Doe",
				TotalPrice:   100,
				Status:       constant.OrderStatusCreated,
				OrderItems: []response.OrderItemData{
					{ID: 1, ProductName: "Product 1", Price: 50, Qty: 2, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:   "get order by ID not found returns error",
			id:     999,
			shopID: []int{1},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore) {
				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					GetOrderByID(999, 1).
					Return(nil, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				return mockOrder, mockOrderItem
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get order by ID returns error on store failure",
			id:     1,
			shopID: []int{1},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore) {
				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					GetOrderByID(1, 1).
					Return(nil, errors.New("database error"))

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				return mockOrder, mockOrderItem
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get order by ID returns error on order items failure",
			id:     1,
			shopID: []int{1},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore) {
				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					GetOrderByID(1, 1).
					Return(&model.Order{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   100,
						Status:       constant.OrderStatusCreated,
						CreatedAt:    fixedTime,
					}, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				mockOrderItem.EXPECT().
					GetOrderItemsByOrderID(1).
					Return(nil, errors.New("order items error"))

				return mockOrder, mockOrderItem
			},
			wantResult: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldOrderStore, oldOrderItemStore := orderStore, orderItemStore
			defer func() { orderStore, orderItemStore = oldOrderStore, oldOrderItemStore }()

			mockOrder, mockOrderItem := tt.mockSetup(ctrl)
			orderStore = mockOrder
			orderItemStore = mockOrderItem

			var o oservice
			got, gotErr := o.GetOrderByID(tt.id, tt.shopID...)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetOrderByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetOrderByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetOrderByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_GetOrdersByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	strPtr := func(s string) *string { return &s }
	ptrTime := func(t time.Time) *time.Time { return &t }

	tests := []struct {
		name       string
		shopID     int
		opts       model.OrderFilterOptions
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockOrderStore
		wantResult []response.OrderData
		wantErr    bool
	}{
		{
			name:   "successfully get orders by shop ID",
			shopID: 1,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetOrdersByShopID(1, model.OrderFilterOptions{}).
					Return([]model.Order{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
						{ID: 2, CustomerName: "Jane Doe", TotalPrice: 200, Status: constant.OrderStatusDone, CreatedAt: fixedTime, UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true}},
					}, nil)
				return mock
			},
			wantResult: []response.OrderData{
				{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
				{ID: 2, CustomerName: "Jane Doe", TotalPrice: 200, Status: constant.OrderStatusDone, CreatedAt: fixedTime, UpdatedAt: &updatedTime},
			},
			wantErr: false,
		},
		{
			name:   "get orders by shop ID returns empty slice",
			shopID: 1,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetOrdersByShopID(1, model.OrderFilterOptions{}).
					Return([]model.Order{}, nil)
				return mock
			},
			wantResult: []response.OrderData{},
			wantErr:    false,
		},
		{
			name:   "get orders by shop ID returns error on store failure",
			shopID: 1,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetOrdersByShopID(1, model.OrderFilterOptions{}).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: []response.OrderData{},
			wantErr:    true,
		},
		{
			name:   "get orders by shop ID with search query returns filtered orders",
			shopID: 1,
			opts:   model.OrderFilterOptions{SearchQuery: strPtr("john")},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetOrdersByShopID(1, model.OrderFilterOptions{SearchQuery: strPtr("john")}).
					Return([]model.Order{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
					}, nil)
				return mock
			},
			wantResult: []response.OrderData{
				{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
			},
			wantErr: false,
		},
		{
			name:   "get orders by shop ID with date_from returns filtered orders",
			shopID: 1,
			opts:   model.OrderFilterOptions{DateFrom: ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				mock.EXPECT().GetOrdersByShopID(1, model.OrderFilterOptions{DateFrom: &dateFrom}).
					Return([]model.Order{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
					}, nil)
				return mock
			},
			wantResult: []response.OrderData{
				{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
			},
			wantErr: false,
		},
		{
			name:   "get orders by shop ID with date_to returns filtered orders",
			shopID: 1,
			opts:   model.OrderFilterOptions{DateTo: ptrTime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				dateTo := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
				mock.EXPECT().GetOrdersByShopID(1, model.OrderFilterOptions{DateTo: &dateTo}).
					Return([]model.Order{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
					}, nil)
				return mock
			},
			wantResult: []response.OrderData{
				{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
			},
			wantErr: false,
		},
		{
			name:   "get orders by shop ID with date_from and date_to returns filtered orders",
			shopID: 1,
			opts: model.OrderFilterOptions{
				DateFrom: ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				DateTo:   ptrTime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				dateFrom := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				dateTo := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
				mock.EXPECT().GetOrdersByShopID(1, model.OrderFilterOptions{DateFrom: &dateFrom, DateTo: &dateTo}).
					Return([]model.Order{
						{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
					}, nil)
				return mock
			},
			wantResult: []response.OrderData{
				{ID: 1, CustomerName: "John Doe", TotalPrice: 100, Status: constant.OrderStatusCreated, CreatedAt: fixedTime},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderStore
			defer func() { orderStore = oldStore }()
			orderStore = tt.mockSetup(ctrl)

			var o oservice
			got, gotErr := o.GetOrdersByShopID(tt.shopID, tt.opts)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetOrdersByShopID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetOrdersByShopID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetOrdersByShopID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_UpdateOrderByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	intPtr := func(i int) *int { return &i }
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name       string
		input      UpdateOrderInput
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockOrderStore
		wantResult response.OrderData
		wantErr    bool
	}{
		{
			name: "successfully update order",
			input: UpdateOrderInput{
				ID:     1,
				Status: strPtr(constant.OrderStatusDone),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetOrderByID(1).
					Return(&model.Order{ID: 1, CustomerName: "John Doe", Status: constant.OrderStatusCreated}, nil)
				mock.EXPECT().
					UpdateOrder(nil, 1, store.UpdateOrderInput{Status: strPtr(constant.OrderStatusDone)}).
					Return(&model.Order{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   100,
						Status:       constant.OrderStatusDone,
						CreatedAt:    fixedTime,
						UpdatedAt:    sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.OrderData{
				ID:           1,
				CustomerName: "John Doe",
				TotalPrice:   100,
				Status:       constant.OrderStatusDone,
				CreatedAt:    fixedTime,
				UpdatedAt:    &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "update order with multiple fields",
			input: UpdateOrderInput{
				ID:         1,
				TotalPrice: intPtr(500),
				Status:     strPtr(constant.OrderStatusDone),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetOrderByID(1).
					Return(&model.Order{ID: 1, CustomerName: "John Doe", Status: constant.OrderStatusCreated}, nil)
				mock.EXPECT().
					UpdateOrder(nil, 1, store.UpdateOrderInput{TotalPrice: intPtr(500), Status: strPtr(constant.OrderStatusDone)}).
					Return(&model.Order{
						ID:           1,
						CustomerName: "Jane Doe",
						TotalPrice:   500,
						Status:       constant.OrderStatusDone,
						CreatedAt:    fixedTime,
						UpdatedAt:    sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.OrderData{
				ID:           1,
				CustomerName: "Jane Doe",
				TotalPrice:   500,
				Status:       constant.OrderStatusDone,
				CreatedAt:    fixedTime,
				UpdatedAt:    &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "update order not found returns error",
			input: UpdateOrderInput{
				ID:     999,
				Status: strPtr(constant.OrderStatusDone),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetOrderByID(999).
					Return(nil, nil)
				return mock
			},
			wantResult: response.OrderData{},
			wantErr:    true,
		},
		{
			name: "update order returns error on get failure",
			input: UpdateOrderInput{
				ID:     1,
				Status: strPtr(constant.OrderStatusDone),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetOrderByID(1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.OrderData{},
			wantErr:    true,
		},
		{
			name: "update order returns error on update failure",
			input: UpdateOrderInput{
				ID:     1,
				Status: strPtr(constant.OrderStatusDone),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetOrderByID(1).
					Return(&model.Order{ID: 1, CustomerName: "John Doe"}, nil)
				mock.EXPECT().
					UpdateOrder(nil, 1, store.UpdateOrderInput{Status: strPtr(constant.OrderStatusDone)}).
					Return(nil, errors.New("update error"))
				return mock
			},
			wantResult: response.OrderData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderStore
			defer func() { orderStore = oldStore }()
			orderStore = tt.mockSetup(ctrl)

			var o oservice
			got, gotErr := o.UpdateOrderByID(tt.input)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateOrderByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateOrderByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateOrderByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_DeleteOrderByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		mockSetup func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB, *mock_database.MockTx)
		wantErr   bool
	}{
		{
			name: "successfully delete order",
			id:   1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB, *mock_database.MockTx) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				mockOrderItem.EXPECT().
					DeleteOrderItemsByOrderID(mockTx, 1).
					Return(nil)

				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					DeleteOrderByID(mockTx, 1).
					Return(nil)

				return mockOrder, mockOrderItem, mockDB, mockTx
			},
			wantErr: false,
		},
		{
			name: "delete order returns error on db.Begin failure",
			id:   1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB, *mock_database.MockTx) {
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(nil, errors.New("begin error"))

				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)

				return mockOrder, mockOrderItem, mockDB, nil
			},
			wantErr: true,
		},
		{
			name: "delete order returns error on delete order items failure",
			id:   1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB, *mock_database.MockTx) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				mockOrderItem.EXPECT().
					DeleteOrderItemsByOrderID(mockTx, 1).
					Return(errors.New("delete items error"))

				mockOrder := mock_store.NewMockOrderStore(ctrl)

				return mockOrder, mockOrderItem, mockDB, mockTx
			},
			wantErr: true,
		},
		{
			name: "delete order returns error on delete order failure",
			id:   1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB, *mock_database.MockTx) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				mockOrderItem.EXPECT().
					DeleteOrderItemsByOrderID(mockTx, 1).
					Return(nil)

				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					DeleteOrderByID(mockTx, 1).
					Return(errors.New("delete order error"))

				return mockOrder, mockOrderItem, mockDB, mockTx
			},
			wantErr: true,
		},
		{
			name: "delete order returns error on commit failure",
			id:   1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB, *mock_database.MockTx) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(errors.New("commit error"))
				mockTx.EXPECT().Rollback().Return(nil)

				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				mockOrderItem.EXPECT().
					DeleteOrderItemsByOrderID(mockTx, 1).
					Return(nil)

				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					DeleteOrderByID(mockTx, 1).
					Return(nil)

				return mockOrder, mockOrderItem, mockDB, mockTx
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldOrderStore, oldOrderItemStore := orderStore, orderItemStore
			oldDBGetter := dbGetter
			defer func() {
				orderStore, orderItemStore = oldOrderStore, oldOrderItemStore
				dbGetter = oldDBGetter
			}()

			mockOrder, mockOrderItem, mockDB, _ := tt.mockSetup(ctrl)
			orderStore = mockOrder
			orderItemStore = mockOrderItem
			dbGetter = func() database.DB { return mockDB }

			var o oservice
			gotErr := o.DeleteOrderByID(tt.id)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteOrderByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteOrderByID() succeeded unexpectedly")
			}
		})
	}
}

func Test_oservice_CreateOrderItem(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name       string
		orderID    int
		productID  int
		qty        int
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore
		wantResult response.OrderItemData
		wantErr    bool
	}{
		{
			name:      "successfully create order item",
			orderID:   1,
			productID: 10,
			qty:       2,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					CreateOrderItem(nil, 1, 10, 2).
					Return(&model.OrderItem{
						ID:          1,
						OrderID:     1,
						ProductName: "Product 1",
						Price:       50,
						Qty:         2,
						CreatedAt:   fixedTime,
					}, nil)
				return mock
			},
			wantResult: response.OrderItemData{
				ID:          1,
				OrderID:     1,
				ProductName: "Product 1",
				Price:       50,
				Qty:         2,
				CreatedAt:   fixedTime,
			},
			wantErr: false,
		},
		{
			name:      "create order item returns error on store failure",
			orderID:   1,
			productID: 10,
			qty:       2,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					CreateOrderItem(nil, 1, 10, 2).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.OrderItemData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderItemStore
			defer func() { orderItemStore = oldStore }()
			orderItemStore = tt.mockSetup(ctrl)

			var o oservice
			got, gotErr := o.CreateOrderItem(tt.orderID, tt.productID, tt.qty)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateOrderItem() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateOrderItem() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateOrderItem() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_UpdateOrderItemByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name       string
		input      UpdateOrderItemInput
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore
		wantResult response.OrderItemData
		wantErr    bool
	}{
		{
			name: "successfully update order item",
			input: UpdateOrderItemInput{
				OrderID:     1,
				OrderItemID: 1,
				Qty:         intPtr(5),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemByID(1).
					Return(&model.OrderItem{ID: 1, OrderID: 1, ProductName: "Product 1", Qty: 2}, nil)
				mock.EXPECT().
					UpdateOrderItemByID(nil, 1, 1, store.UpdateOrderItemInput{Qty: intPtr(5)}).
					Return(&model.OrderItem{
						ID:          1,
						OrderID:     1,
						ProductName: "Product 1",
						Price:       50,
						Qty:         5,
						CreatedAt:   fixedTime,
						UpdatedAt:   sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.OrderItemData{
				ID:          1,
				OrderID:     1,
				ProductName: "Product 1",
				Price:       50,
				Qty:         5,
				CreatedAt:   fixedTime,
				UpdatedAt:   &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "update order item with product change",
			input: UpdateOrderItemInput{
				OrderID:     1,
				OrderItemID: 1,
				ProductID:   intPtr(20),
				Qty:         intPtr(3),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemByID(1).
					Return(&model.OrderItem{ID: 1, OrderID: 1, ProductName: "Product 1", Qty: 2}, nil)
				mock.EXPECT().
					UpdateOrderItemByID(nil, 1, 1, store.UpdateOrderItemInput{ProductID: intPtr(20), Qty: intPtr(3)}).
					Return(&model.OrderItem{
						ID:          1,
						OrderID:     1,
						ProductName: "Product 2",
						Price:       100,
						Qty:         3,
						CreatedAt:   fixedTime,
						UpdatedAt:   sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.OrderItemData{
				ID:          1,
				OrderID:     1,
				ProductName: "Product 2",
				Price:       100,
				Qty:         3,
				CreatedAt:   fixedTime,
				UpdatedAt:   &updatedTime,
			},
			wantErr: false,
		},
		{
			name: "update order item not found returns error",
			input: UpdateOrderItemInput{
				OrderID:     1,
				OrderItemID: 999,
				Qty:         intPtr(5),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemByID(999).
					Return(nil, nil)
				return mock
			},
			wantResult: response.OrderItemData{},
			wantErr:    true,
		},
		{
			name: "update order item returns error on get failure",
			input: UpdateOrderItemInput{
				OrderID:     1,
				OrderItemID: 1,
				Qty:         intPtr(5),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemByID(1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.OrderItemData{},
			wantErr:    true,
		},
		{
			name: "update order item returns error on update failure",
			input: UpdateOrderItemInput{
				OrderID:     1,
				OrderItemID: 1,
				Qty:         intPtr(5),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemByID(1).
					Return(&model.OrderItem{ID: 1, OrderID: 1}, nil)
				mock.EXPECT().
					UpdateOrderItemByID(nil, 1, 1, store.UpdateOrderItemInput{Qty: intPtr(5)}).
					Return(nil, errors.New("update error"))
				return mock
			},
			wantResult: response.OrderItemData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderItemStore
			defer func() { orderItemStore = oldStore }()
			orderItemStore = tt.mockSetup(ctrl)

			var o oservice
			got, gotErr := o.UpdateOrderItemByID(tt.input)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateOrderItemByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateOrderItemByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("UpdateOrderItemByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_DeleteOrderItemByID(t *testing.T) {
	tests := []struct {
		name        string
		orderItemID int
		orderID     int
		mockSetup   func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore
		wantErr     bool
	}{
		{
			name:        "successfully delete order item",
			orderItemID: 1,
			orderID:     1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					DeleteOrderItemByID(1, 1).
					Return(nil)
				return mock
			},
			wantErr: false,
		},
		{
			name:        "delete order item returns error on store failure",
			orderItemID: 1,
			orderID:     1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					DeleteOrderItemByID(1, 1).
					Return(errors.New("delete error"))
				return mock
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderItemStore
			defer func() { orderItemStore = oldStore }()
			orderItemStore = tt.mockSetup(ctrl)

			var o oservice
			gotErr := o.DeleteOrderItemByID(tt.orderItemID, tt.orderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteOrderItemByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteOrderItemByID() succeeded unexpectedly")
			}
		})
	}
}

func Test_oservice_GetOrderItemByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		orderItemID int
		orderID     int
		mockSetup   func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore
		wantResult  response.OrderItemData
		wantErr     bool
	}{
		{
			name:        "successfully get order item by ID",
			orderItemID: 1,
			orderID:     1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemByID(1).
					Return(&model.OrderItem{
						ID:          1,
						OrderID:     1,
						ProductName: "Product 1",
						Price:       50,
						Qty:         2,
						CreatedAt:   fixedTime,
						UpdatedAt:   sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				return mock
			},
			wantResult: response.OrderItemData{
				ID:          1,
				OrderID:     1,
				ProductName: "Product 1",
				Price:       50,
				Qty:         2,
				CreatedAt:   fixedTime,
				UpdatedAt:   &updatedTime,
			},
			wantErr: false,
		},
		{
			name:        "get order item by ID not found returns nil result",
			orderItemID: 999,
			orderID:     1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemByID(999).
					Return(nil, nil)
				return mock
			},
			wantResult: response.OrderItemData{},
			wantErr:    false,
		},
		{
			name:        "get order item by ID returns error on store failure",
			orderItemID: 1,
			orderID:     1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemByID(1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: response.OrderItemData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderItemStore
			defer func() { orderItemStore = oldStore }()
			orderItemStore = tt.mockSetup(ctrl)

			var o oservice
			got, gotErr := o.GetOrderItemByID(tt.orderItemID, tt.orderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetOrderItemByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetOrderItemByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetOrderItemByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_GetOrderItemsByOrderID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		orderID    int
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore
		wantResult []response.OrderItemData
		wantErr    bool
	}{
		{
			name:    "successfully get order items by order ID",
			orderID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemsByOrderID(1).
					Return([]model.OrderItem{
						{ID: 1, ProductName: "Product 1", Price: 50, Qty: 2, CreatedAt: fixedTime},
						{ID: 2, ProductName: "Product 2", Price: 100, Qty: 1, CreatedAt: fixedTime, UpdatedAt: sql.NullTime{Time: updatedTime, Valid: true}},
					}, nil)
				return mock
			},
			wantResult: []response.OrderItemData{
				{ID: 1, ProductName: "Product 1", Price: 50, Qty: 2, CreatedAt: fixedTime},
				{ID: 2, ProductName: "Product 2", Price: 100, Qty: 1, CreatedAt: fixedTime, UpdatedAt: &updatedTime},
			},
			wantErr: false,
		},
		{
			name:    "get order items returns empty slice",
			orderID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemsByOrderID(1).
					Return([]model.OrderItem{}, nil)
				return mock
			},
			wantResult: []response.OrderItemData{},
			wantErr:    false,
		},
		{
			name:    "get order items returns error on store failure",
			orderID: 1,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderItemStore {
				mock := mock_store.NewMockOrderItemStore(ctrl)
				mock.EXPECT().
					GetOrderItemsByOrderID(1).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: []response.OrderItemData{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderItemStore
			defer func() { orderItemStore = oldStore }()
			orderItemStore = tt.mockSetup(ctrl)

			var o oservice
			got, gotErr := o.GetOrderItemsByOrderID(tt.orderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetOrderItemsByOrderID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetOrderItemsByOrderID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetOrderItemsByOrderID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_CreateTempOrder(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		customerName  string
		customerPhone string
		shareToken    string
		items         []CreateTempOrderItemInput
		mockSetup     func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB)
		wantResult    response.TempOrderData
		wantErr       bool
	}{
		{
			name:          "successfully create temp order with no items",
			customerName:  "Jane Doe",
			customerPhone: "+62812345678",
			shareToken:    "share-abc123",
			items:         nil,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				shopMock := mock_store.NewMockShopStore(ctrl)
				shopMock.EXPECT().
					GetShopByShareToken("share-abc123").
					Return(&model.Shop{ID: 5, Name: "Test Shop", ShareToken: "share-abc123", CreatedAt: fixedTime}, nil)
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)
				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					CreateTempOrder(gomock.Any(), "Jane Doe", "+62812345678", 5).
					Return(&model.TempOrder{
						ID:            1,
						CustomerName:  "Jane Doe",
						CustomerPhone: "+62812345678",
						ShopID:        5,
						TotalPrice:    0,
						Status:        "pending",
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderTotalPrice(gomock.Any(), 1, 0).
					Return(nil)
				return shopMock, orderMock, nil, mockDB
			},
			wantResult: response.TempOrderData{
				ID:             1,
				CustomerName:   "Jane Doe",
				CustomerPhone:  "+62812345678",
				TotalPrice:     0,
				Status:         "pending",
				TempOrderItems: []response.TempOrderItemData{},
				CreatedAt:      fixedTime,
				UpdatedAt:      &updatedTime,
			},
			wantErr: false,
		},
		{
			name:          "successfully create temp order with items",
			customerName:  "Jane Doe",
			customerPhone: "+62812345678",
			shareToken:    "share-abc123",
			items:         []CreateTempOrderItemInput{{ProductID: 10, Qty: 2}, {ProductID: 20, Qty: 1}},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				shopMock := mock_store.NewMockShopStore(ctrl)
				shopMock.EXPECT().
					GetShopByShareToken("share-abc123").
					Return(&model.Shop{ID: 5, Name: "Test Shop", ShareToken: "share-abc123", CreatedAt: fixedTime}, nil)
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)
				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					CreateTempOrder(gomock.Any(), "Jane Doe", "+62812345678", 5).
					Return(&model.TempOrder{
						ID:            1,
						CustomerName:  "Jane Doe",
						CustomerPhone: "+62812345678",
						ShopID:        5,
						TotalPrice:    0,
						Status:        "pending",
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{Time: updatedTime, Valid: true},
					}, nil)
				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					CreateTempOrderItem(gomock.Any(), 1, 10, 2).
					Return(&model.TempOrderItem{ID: 1, TempOrderID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime}, nil)
				orderItemMock.EXPECT().
					CreateTempOrderItem(gomock.Any(), 1, 20, 1).
					Return(&model.TempOrderItem{ID: 2, TempOrderID: 1, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime}, nil)
				orderMock.EXPECT().
					UpdateTempOrderTotalPrice(gomock.Any(), 1, 2500).
					Return(nil)
				return shopMock, orderMock, orderItemMock, mockDB
			},
			wantResult: response.TempOrderData{
				ID:            1,
				CustomerName:  "Jane Doe",
				CustomerPhone: "+62812345678",
				TotalPrice:    0,
				Status:        "pending",
				TempOrderItems: []response.TempOrderItemData{
					{ID: 1, TempOrderID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					{ID: 2, TempOrderID: 1, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
				UpdatedAt: &updatedTime,
			},
			wantErr: false,
		},
		{
			name:          "create temp order returns error when shop not found",
			customerName:  "Jane Doe",
			customerPhone: "+62812345678",
			shareToken:    "invalid-token",
			items:         nil,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				shopMock := mock_store.NewMockShopStore(ctrl)
				shopMock.EXPECT().
					GetShopByShareToken("invalid-token").
					Return(nil, errors.New("shop not found"))
				return shopMock, nil, nil, nil
			},
			wantResult: response.TempOrderData{},
			wantErr:    true,
		},
		{
			name:          "create temp order returns error when shop is nil",
			customerName:  "Jane Doe",
			customerPhone: "+62812345678",
			shareToken:    "share-abc123",
			items:         nil,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				shopMock := mock_store.NewMockShopStore(ctrl)
				shopMock.EXPECT().
					GetShopByShareToken("share-abc123").
					Return(nil, nil)
				return shopMock, nil, nil, nil
			},
			wantResult: response.TempOrderData{},
			wantErr:    true,
		},
		{
			name:          "create temp order returns error on order store failure",
			customerName:  "Jane Doe",
			customerPhone: "+62812345678",
			shareToken:    "share-abc123",
			items:         nil,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				shopMock := mock_store.NewMockShopStore(ctrl)
				shopMock.EXPECT().
					GetShopByShareToken("share-abc123").
					Return(&model.Shop{ID: 5, Name: "Test Shop", ShareToken: "share-abc123", CreatedAt: fixedTime}, nil)
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)
				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					CreateTempOrder(gomock.Any(), "Jane Doe", "+62812345678", 5).
					Return(nil, errors.New("database error"))
				return shopMock, orderMock, nil, mockDB
			},
			wantResult: response.TempOrderData{},
			wantErr:    true,
		},
		{
			name:          "create temp order returns error when UpdateTempOrderTotalPrice fails",
			customerName:  "Jane Doe",
			customerPhone: "+62812345678",
			shareToken:    "share-abc123",
			items:         nil,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockShopStore, *mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				shopMock := mock_store.NewMockShopStore(ctrl)
				shopMock.EXPECT().
					GetShopByShareToken("share-abc123").
					Return(&model.Shop{ID: 5, Name: "Test Shop", ShareToken: "share-abc123", CreatedAt: fixedTime}, nil)
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)
				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					CreateTempOrder(gomock.Any(), "Jane Doe", "+62812345678", 5).
					Return(&model.TempOrder{
						ID:            1,
						CustomerName:  "Jane Doe",
						CustomerPhone: "+62812345678",
						ShopID:        5,
						TotalPrice:    0,
						Status:        "pending",
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderTotalPrice(gomock.Any(), 1, 0).
					Return(errors.New("update total price error"))
				return shopMock, orderMock, nil, mockDB
			},
			wantResult: response.TempOrderData{},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			shopMock, orderMock, orderItemMock, mockDB := tt.mockSetup(ctrl)
			oldShopStore := shopStore
			oldOrderStore := orderStore
			oldOrderItemStore := orderItemStore
			oldDBGetter := dbGetter
			defer func() {
				shopStore = oldShopStore
				orderStore = oldOrderStore
				orderItemStore = oldOrderItemStore
				dbGetter = oldDBGetter
			}()
			shopStore = shopMock
			orderStore = orderMock
			if orderItemMock != nil {
				orderItemStore = orderItemMock
			}
			if mockDB != nil {
				dbGetter = func() database.DB { return mockDB }
			}

			var o oservice
			got, gotErr := o.CreateTempOrder(tt.customerName, tt.customerPhone, tt.shareToken, tt.items)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateTempOrder() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateTempOrder() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("CreateTempOrder() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_GetTempOrdersByShopID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)
	strPtr := func(s string) *string { return &s }
	ptrTime := func(t time.Time) *time.Time { return &t }

	tests := []struct {
		name       string
		shopID     int
		opts       model.OrderFilterOptions
		mockSetup  func(ctrl *gomock.Controller) *mock_store.MockOrderStore
		wantResult []response.TempOrderData
		wantErr    bool
	}{
		{
			name:   "successfully get temp orders by shop ID",
			shopID: 5,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetTempOrdersByShopID(5, model.OrderFilterOptions{}).
					Return([]model.TempOrder{
						{ID: 1, ShopID: 5, CustomerName: "Jane Doe", CustomerPhone: "+62812345678", TotalPrice: 2500, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{}},
						{ID: 2, ShopID: 5, CustomerName: "John Doe", CustomerPhone: "+62887654321", TotalPrice: 1000, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{}},
					}, nil)
				return mock
			},
			wantResult: []response.TempOrderData{
				{ID: 1, CustomerName: "Jane Doe", CustomerPhone: "+62812345678", TotalPrice: 2500, Status: "pending", CreatedAt: fixedTime, UpdatedAt: nil},
				{ID: 2, CustomerName: "John Doe", CustomerPhone: "+62887654321", TotalPrice: 1000, Status: "pending", CreatedAt: fixedTime, UpdatedAt: nil},
			},
			wantErr: false,
		},
		{
			name:   "get temp orders returns empty slice when none exist",
			shopID: 99,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetTempOrdersByShopID(99, model.OrderFilterOptions{}).
					Return([]model.TempOrder{}, nil)
				return mock
			},
			wantResult: []response.TempOrderData{},
			wantErr:    false,
		},
		{
			name:   "get temp orders returns error on store failure",
			shopID: 5,
			opts:   model.OrderFilterOptions{},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetTempOrdersByShopID(5, model.OrderFilterOptions{}).
					Return(nil, errors.New("database error"))
				return mock
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get temp orders with search and date filters passes opts to store",
			shopID: 5,
			opts: model.OrderFilterOptions{
				SearchQuery: strPtr("62812"),
				DateFrom:    ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				DateTo:      ptrTime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
			},
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					GetTempOrdersByShopID(5, model.OrderFilterOptions{
						SearchQuery: strPtr("62812"),
						DateFrom:    ptrTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						DateTo:      ptrTime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
					}).
					Return([]model.TempOrder{
						{ID: 1, ShopID: 5, CustomerName: "Jane Doe", CustomerPhone: "+62812345678", TotalPrice: 2500, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{Valid: true, Time: updatedTime}},
					}, nil)
				return mock
			},
			wantResult: []response.TempOrderData{
				{ID: 1, CustomerName: "Jane Doe", CustomerPhone: "+62812345678", TotalPrice: 2500, Status: "pending", CreatedAt: fixedTime, UpdatedAt: &updatedTime},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldStore := orderStore
			defer func() { orderStore = oldStore }()
			orderStore = tt.mockSetup(ctrl)

			var o oservice
			got, gotErr := o.GetTempOrdersByShopID(tt.shopID, tt.opts)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTempOrdersByShopID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTempOrdersByShopID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetTempOrdersByShopID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_GetTempOrderByID(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 1, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		id         int
		shopID     []int
		mockSetup  func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore)
		wantResult *response.TempOrderData
		wantErr    bool
	}{
		{
			name:   "successfully get temp order by ID with items",
			id:     1,
			shopID: []int{5},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore) {
				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					GetTempOrderByID(1, 5).
					Return(&model.TempOrder{
						ID:            1,
						ShopID:        5,
						CustomerName:  "Jane Doe",
						CustomerPhone: "+62812345678",
						TotalPrice:    2500,
						Status:        "pending",
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{},
					}, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				mockOrderItem.EXPECT().
					GetTempOrderItemsByTempOrderID(1).
					Return([]model.TempOrderItem{
						{ID: 1, TempOrderID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
						{ID: 2, TempOrderID: 1, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
					}, nil)

				return mockOrder, mockOrderItem
			},
			wantResult: &response.TempOrderData{
				ID:            1,
				CustomerName:  "Jane Doe",
				CustomerPhone: "+62812345678",
				TotalPrice:    2500,
				Status:        "pending",
				TempOrderItems: []response.TempOrderItemData{
					{ID: 1, TempOrderID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					{ID: 2, TempOrderID: 1, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
				UpdatedAt: nil,
			},
			wantErr: false,
		},
		{
			name:   "get temp order by ID not found returns error",
			id:     999,
			shopID: []int{5},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore) {
				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					GetTempOrderByID(999, 5).
					Return(nil, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				return mockOrder, mockOrderItem
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get temp order by ID returns error on order store failure",
			id:     1,
			shopID: []int{5},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore) {
				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					GetTempOrderByID(1, 5).
					Return(nil, errors.New("database error"))

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				return mockOrder, mockOrderItem
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get temp order by ID returns error on order items store failure",
			id:     1,
			shopID: []int{5},
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore) {
				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					GetTempOrderByID(1, 5).
					Return(&model.TempOrder{
						ID:            1,
						ShopID:        5,
						CustomerName:  "Jane Doe",
						CustomerPhone: "+62812345678",
						TotalPrice:    2500,
						Status:        "pending",
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{},
					}, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				mockOrderItem.EXPECT().
					GetTempOrderItemsByTempOrderID(1).
					Return(nil, errors.New("database error"))

				return mockOrder, mockOrderItem
			},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "get temp order by ID without shop ID and with UpdatedAt",
			id:     1,
			shopID: nil,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore) {
				mockOrder := mock_store.NewMockOrderStore(ctrl)
				mockOrder.EXPECT().
					GetTempOrderByID(1).
					Return(&model.TempOrder{
						ID:            1,
						ShopID:        5,
						CustomerName:  "Jane Doe",
						CustomerPhone: "+62812345678",
						TotalPrice:    2500,
						Status:        "pending",
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{Valid: true, Time: updatedTime},
					}, nil)

				mockOrderItem := mock_store.NewMockOrderItemStore(ctrl)
				mockOrderItem.EXPECT().
					GetTempOrderItemsByTempOrderID(1).
					Return([]model.TempOrderItem{}, nil)

				return mockOrder, mockOrderItem
			},
			wantResult: &response.TempOrderData{
				ID:             1,
				CustomerName:   "Jane Doe",
				CustomerPhone:  "+62812345678",
				TotalPrice:     2500,
				Status:         "pending",
				TempOrderItems: []response.TempOrderItemData{},
				CreatedAt:      fixedTime,
				UpdatedAt:      &updatedTime,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oldOrderStore, oldOrderItemStore := orderStore, orderItemStore
			defer func() { orderStore, orderItemStore = oldOrderStore, oldOrderItemStore }()

			mockOrder, mockOrderItem := tt.mockSetup(ctrl)
			orderStore = mockOrder
			orderItemStore = mockOrderItem

			var o oservice
			got, gotErr := o.GetTempOrderByID(tt.id, tt.shopID...)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetTempOrderByID() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetTempOrderByID() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.wantResult) {
				t.Errorf("GetTempOrderByID() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func Test_oservice_createOrderFromTempOrder(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		tempOrderID int
		customerID  int
		shopID      int
		mockSetup   func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB)
		want        *response.OrderData
		wantErr     bool
	}{
		{
			name:        "successfully create order from temp order with items",
			tempOrderID: 10,
			customerID:  5,
			shopID:      1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID:            10,
						ShopID:        1,
						CustomerName:  "Jane Doe",
						CustomerPhone: "+62812345678",
						TotalPrice:    2500,
						Status:        "pending",
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					CreateOrder(gomock.Any(), 5, 1, nil, gomock.Any()).
					Return(&model.Order{
						ID:           1,
						CustomerName: "John Doe",
						TotalPrice:   0,
						Status:       constant.OrderStatusCreated,
						CreatedAt:    fixedTime,
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderStatus(gomock.Any(), 10, constant.TempOrderStatusAccepted).
					Return(nil)

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{
						{ID: 1, TempOrderID: 10, ProductID: 10, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
						{ID: 2, TempOrderID: 10, ProductID: 20, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					CreateOrderItem(gomock.Any(), 1, 10, 2).
					Return(&model.OrderItem{ID: 1, OrderID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime}, nil)
				orderItemMock.EXPECT().
					CreateOrderItem(gomock.Any(), 1, 20, 1).
					Return(&model.OrderItem{ID: 2, OrderID: 1, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want: &response.OrderData{
				ID:           1,
				CustomerName: "John Doe",
				TotalPrice:   0,
				Status:       constant.OrderStatusCreated,
				OrderItems: []response.OrderItemData{
					{ID: 1, OrderID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					{ID: 2, OrderID: 1, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:        "successfully create order from temp order with no items",
			tempOrderID: 20,
			customerID:  3,
			shopID:      2,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(20, 2).
					Return(&model.TempOrder{
						ID:            20,
						ShopID:        2,
						CustomerName:  "Alice",
						CustomerPhone: "+62811111111",
						TotalPrice:    0,
						Status:        "pending",
						CreatedAt:     fixedTime,
						UpdatedAt:     sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					CreateOrder(gomock.Any(), 3, 2, nil, gomock.Any()).
					Return(&model.Order{
						ID:           2,
						CustomerName: "Alice",
						TotalPrice:   0,
						Status:       constant.OrderStatusCreated,
						CreatedAt:    fixedTime,
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderStatus(gomock.Any(), 20, constant.TempOrderStatusAccepted).
					Return(nil)

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(20).
					Return([]model.TempOrderItem{}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want: &response.OrderData{
				ID:           2,
				CustomerName: "Alice",
				TotalPrice:   0,
				Status:       constant.OrderStatusCreated,
				OrderItems:   []response.OrderItemData{},
				CreatedAt:    fixedTime,
			},
			wantErr: false,
		},
		{
			name:        "returns error when GetTempOrderByID fails",
			tempOrderID: 10,
			customerID:  5,
			shopID:      1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(nil, errors.New("database error"))
				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				return orderMock, orderItemMock, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:        "returns error when temp order not found",
			tempOrderID: 999,
			customerID:  5,
			shopID:      1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(999, 1).
					Return(nil, nil)
				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				return orderMock, orderItemMock, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:        "returns error when CreateOrder fails",
			tempOrderID: 10,
			customerID:  5,
			shopID:      1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					CreateOrder(gomock.Any(), 5, 1, nil, gomock.Any()).
					Return(nil, errors.New("create order failed"))

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:        "returns error when CreateOrderItem fails",
			tempOrderID: 10,
			customerID:  5,
			shopID:      1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					CreateOrder(gomock.Any(), 5, 1, nil, gomock.Any()).
					Return(&model.Order{ID: 1, CustomerName: "John", TotalPrice: 0, Status: constant.OrderStatusCreated, CreatedAt: fixedTime}, nil)

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{
						{ID: 1, TempOrderID: 10, ProductID: 10, ProductName: "A", Price: 100, Qty: 1, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					CreateOrderItem(gomock.Any(), 1, 10, 1).
					Return(nil, errors.New("create order item failed"))

				return orderMock, orderItemMock, mockDB
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:        "returns error when UpdateTempOrderStatus fails",
			tempOrderID: 10,
			customerID:  5,
			shopID:      1,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					CreateOrder(gomock.Any(), 5, 1, nil, gomock.Any()).
					Return(&model.Order{ID: 1, CustomerName: "John", TotalPrice: 0, Status: constant.OrderStatusCreated, CreatedAt: fixedTime}, nil)
				orderMock.EXPECT().
					UpdateTempOrderStatus(gomock.Any(), 10, constant.TempOrderStatusAccepted).
					Return(errors.New("update status failed"))

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			orderMock, orderItemMock, mockDB := tt.mockSetup(ctrl)
			oldOrderStore := orderStore
			oldOrderItemStore := orderItemStore
			oldDBGetter := dbGetter
			defer func() {
				orderStore = oldOrderStore
				orderItemStore = oldOrderItemStore
				dbGetter = oldDBGetter
			}()
			orderStore = orderMock
			orderItemStore = orderItemMock
			if mockDB != nil {
				dbGetter = func() database.DB { return mockDB }
			}

			var o oservice
			got, gotErr := o.createOrderFromTempOrder(tt.tempOrderID, tt.customerID, tt.shopID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("createOrderFromTempOrder() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("createOrderFromTempOrder() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createOrderFromTempOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_oservice_resolveActiveOrderConflict(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name          string
		tempOrderID   int
		shopID        int
		activeOrderID int
		mockSetup     func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB)
		want          *response.OrderData
		wantErr       bool
	}{
		{
			name:          "successfully merges temp order items into active order (new item)",
			tempOrderID:   10,
			shopID:        1,
			activeOrderID: 7,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					GetOrderByID(7, 1).
					Return(&model.Order{
						ID: 7, ShopID: 1, CustomerName: "John Doe", TotalPrice: 0, Status: constant.OrderStatusInProgress, CreatedAt: fixedTime,
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderStatus(gomock.Any(), 10, constant.TempOrderStatusAccepted).
					Return(nil)

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{
						{ID: 1, TempOrderID: 10, ProductID: 20, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					GetOrderItemsByOrderID(7).
					Return([]model.OrderItem{
						{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					GetOrderItemByProductID(20, 7).
					Return(nil, nil)
				orderItemMock.EXPECT().
					CreateOrderItem(gomock.Any(), 7, 20, 1).
					Return(&model.OrderItem{ID: 2, OrderID: 7, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime}, nil)
				orderItemMock.EXPECT().
					GetOrderItemsByOrderID(7).
					Return([]model.OrderItem{
						{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
						{ID: 2, OrderID: 7, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
					}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want: &response.OrderData{
				ID:           7,
				CustomerName: "John Doe",
				TotalPrice:   2500, // 1000*2 + 500*1
				Status:       constant.OrderStatusInProgress,
				OrderItems: []response.OrderItemData{
					{ID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					{ID: 2, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:          "successfully merges temp order item that exists in active order (update qty)",
			tempOrderID:   10,
			shopID:        1,
			activeOrderID: 7,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					GetOrderByID(7, 1).
					Return(&model.Order{
						ID: 7, ShopID: 1, CustomerName: "John Doe", TotalPrice: 0, Status: constant.OrderStatusInProgress, CreatedAt: fixedTime,
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderStatus(gomock.Any(), 10, constant.TempOrderStatusAccepted).
					Return(nil)

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{
						{ID: 1, TempOrderID: 10, ProductID: 10, ProductName: "Product A", Price: 1000, Qty: 3, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					GetOrderItemsByOrderID(7).
					Return([]model.OrderItem{
						{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					GetOrderItemByProductID(10, 7).
					Return(&model.OrderItem{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime}, nil)
				// Use gomock.Any() for input: actual call uses &qty (local var); pointer comparison would fail with intPtr(5)
				orderItemMock.EXPECT().
					UpdateOrderItemByID(gomock.Any(), 1, 7, gomock.Any()).
					Return(&model.OrderItem{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 5, CreatedAt: fixedTime}, nil)
				orderItemMock.EXPECT().
					GetOrderItemsByOrderID(7).
					Return([]model.OrderItem{
						{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 5, CreatedAt: fixedTime},
					}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want: &response.OrderData{
				ID:           7,
				CustomerName: "John Doe",
				TotalPrice:   5000, // 1000*5
				Status:       constant.OrderStatusInProgress,
				OrderItems: []response.OrderItemData{
					{ID: 1, ProductName: "Product A", Price: 1000, Qty: 5, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:          "successfully merges temp order with some items existing and some new",
			tempOrderID:   10,
			shopID:        1,
			activeOrderID: 7,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					GetOrderByID(7, 1).
					Return(&model.Order{
						ID: 7, ShopID: 1, CustomerName: "John Doe", TotalPrice: 0, Status: constant.OrderStatusInProgress, CreatedAt: fixedTime,
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderStatus(gomock.Any(), 10, constant.TempOrderStatusAccepted).
					Return(nil)

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				// Temp order has Product A (10) qty 3, Product B (20) qty 1
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{
						{ID: 1, TempOrderID: 10, ProductID: 10, ProductName: "Product A", Price: 1000, Qty: 3, CreatedAt: fixedTime},
						{ID: 2, TempOrderID: 10, ProductID: 20, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					GetOrderItemsByOrderID(7).
					Return([]model.OrderItem{
						{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					}, nil)
				// First temp item: Product A exists -> update qty 2+3=5 (use gomock.Any() for input to avoid pointer comparison failure)
				orderItemMock.EXPECT().
					GetOrderItemByProductID(10, 7).
					Return(&model.OrderItem{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime}, nil)
				orderItemMock.EXPECT().
					UpdateOrderItemByID(gomock.Any(), 1, 7, gomock.Any()).
					Return(&model.OrderItem{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 5, CreatedAt: fixedTime}, nil)
				// Second temp item: Product B does not exist -> create
				orderItemMock.EXPECT().
					GetOrderItemByProductID(20, 7).
					Return(nil, nil)
				orderItemMock.EXPECT().
					CreateOrderItem(gomock.Any(), 7, 20, 1).
					Return(&model.OrderItem{ID: 2, OrderID: 7, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime}, nil)
				orderItemMock.EXPECT().
					GetOrderItemsByOrderID(7).
					Return([]model.OrderItem{
						{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 5, CreatedAt: fixedTime},
						{ID: 2, OrderID: 7, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
					}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want: &response.OrderData{
				ID:           7,
				CustomerName: "John Doe",
				TotalPrice:   5500, // 1000*5 + 500*1
				Status:       constant.OrderStatusInProgress,
				OrderItems: []response.OrderItemData{
					{ID: 1, ProductName: "Product A", Price: 1000, Qty: 5, CreatedAt: fixedTime},
					{ID: 2, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:          "returns error when temp order not found",
			tempOrderID:   999,
			shopID:        1,
			activeOrderID: 7,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(999, 1).
					Return(nil, errors.New("temp order not found"))
				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				return orderMock, orderItemMock, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:          "returns error when active order not found",
			tempOrderID:   10,
			shopID:        1,
			activeOrderID: 99,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					GetOrderByID(99, 1).
					Return(nil, errors.New("active order not found"))
				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{}, nil)
				return orderMock, orderItemMock, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:          "returns error when GetTempOrderByID fails",
			tempOrderID:   10,
			shopID:        1,
			activeOrderID: 7,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(nil, errors.New("database error"))
				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				return orderMock, orderItemMock, nil
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:          "returns error when UpdateTempOrderStatus fails",
			tempOrderID:   10,
			shopID:        1,
			activeOrderID: 7,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					GetOrderByID(7, 1).
					Return(&model.Order{
						ID: 7, ShopID: 1, CustomerName: "John", TotalPrice: 0, Status: constant.OrderStatusInProgress, CreatedAt: fixedTime,
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderStatus(gomock.Any(), 10, constant.TempOrderStatusAccepted).
					Return(errors.New("update status failed"))

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{}, nil)
				orderItemMock.EXPECT().
					GetOrderItemsByOrderID(7).
					Return([]model.OrderItem{}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			orderMock, orderItemMock, mockDB := tt.mockSetup(ctrl)
			oldOrderStore := orderStore
			oldOrderItemStore := orderItemStore
			oldDBGetter := dbGetter
			defer func() {
				orderStore = oldOrderStore
				orderItemStore = oldOrderItemStore
				dbGetter = oldDBGetter
			}()
			orderStore = orderMock
			orderItemStore = orderItemMock
			if mockDB != nil {
				dbGetter = func() database.DB { return mockDB }
			}

			var o oservice
			got, gotErr := o.resolveActiveOrderConflict(tt.tempOrderID, tt.shopID, tt.activeOrderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("resolveActiveOrderConflict() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("resolveActiveOrderConflict() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolveActiveOrderConflict() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_oservice_MergeTempOrder(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name          string
		tempOrderID   int
		customerID    int
		shopID        int
		activeOrderID *int
		mockSetup     func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB)
		want          *response.OrderData
		wantErr       bool
	}{
		{
			name:          "success when activeOrderID does not exist (creates new order from temp order)",
			tempOrderID:   10,
			customerID:    5,
			shopID:        1,
			activeOrderID: nil,
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					CreateOrder(gomock.Any(), 5, 1, nil, gomock.Any()).
					Return(&model.Order{
						ID: 1, CustomerName: "John Doe", TotalPrice: 0, Status: constant.OrderStatusCreated, CreatedAt: fixedTime,
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderStatus(gomock.Any(), 10, constant.TempOrderStatusAccepted).
					Return(nil)

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{
						{ID: 1, TempOrderID: 10, ProductID: 10, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					CreateOrderItem(gomock.Any(), 1, 10, 2).
					Return(&model.OrderItem{ID: 1, OrderID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want: &response.OrderData{
				ID: 1, CustomerName: "John Doe", TotalPrice: 0, Status: constant.OrderStatusCreated,
				OrderItems: []response.OrderItemData{
					{ID: 1, OrderID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
		{
			name:          "success when activeOrderID exists (merges temp order into active order)",
			tempOrderID:   10,
			customerID:    5,
			shopID:        1,
			activeOrderID: func() *int { n := 7; return &n }(),
			mockSetup: func(ctrl *gomock.Controller) (*mock_store.MockOrderStore, *mock_store.MockOrderItemStore, *mock_database.MockDB) {
				mockTx := mock_database.NewMockTx(ctrl)
				mockTx.EXPECT().Commit().Return(nil)
				mockTx.EXPECT().Rollback().Return(nil)
				mockDB := mock_database.NewMockDB(ctrl)
				mockDB.EXPECT().Begin().Return(mockTx, nil)

				orderMock := mock_store.NewMockOrderStore(ctrl)
				orderMock.EXPECT().
					GetTempOrderByID(10, 1).
					Return(&model.TempOrder{
						ID: 10, ShopID: 1, CustomerName: "Jane", CustomerPhone: "+62", TotalPrice: 0, Status: "pending", CreatedAt: fixedTime, UpdatedAt: sql.NullTime{},
					}, nil)
				orderMock.EXPECT().
					GetOrderByID(7, 1).
					Return(&model.Order{
						ID: 7, ShopID: 1, CustomerName: "John Doe", TotalPrice: 0, Status: constant.OrderStatusInProgress, CreatedAt: fixedTime,
					}, nil)
				orderMock.EXPECT().
					UpdateTempOrderStatus(gomock.Any(), 10, constant.TempOrderStatusAccepted).
					Return(nil)

				orderItemMock := mock_store.NewMockOrderItemStore(ctrl)
				orderItemMock.EXPECT().
					GetTempOrderItemsByTempOrderID(10).
					Return([]model.TempOrderItem{
						{ID: 1, TempOrderID: 10, ProductID: 20, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					GetOrderItemsByOrderID(7).
					Return([]model.OrderItem{
						{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					}, nil)
				orderItemMock.EXPECT().
					GetOrderItemByProductID(20, 7).
					Return(nil, nil)
				orderItemMock.EXPECT().
					CreateOrderItem(gomock.Any(), 7, 20, 1).
					Return(&model.OrderItem{ID: 2, OrderID: 7, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime}, nil)
				orderItemMock.EXPECT().
					GetOrderItemsByOrderID(7).
					Return([]model.OrderItem{
						{ID: 1, OrderID: 7, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
						{ID: 2, OrderID: 7, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
					}, nil)

				return orderMock, orderItemMock, mockDB
			},
			want: &response.OrderData{
				ID: 7, CustomerName: "John Doe", TotalPrice: 2500, Status: constant.OrderStatusInProgress,
				OrderItems: []response.OrderItemData{
					{ID: 1, ProductName: "Product A", Price: 1000, Qty: 2, CreatedAt: fixedTime},
					{ID: 2, ProductName: "Product B", Price: 500, Qty: 1, CreatedAt: fixedTime},
				},
				CreatedAt: fixedTime,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			orderMock, orderItemMock, mockDB := tt.mockSetup(ctrl)
			oldOrderStore := orderStore
			oldOrderItemStore := orderItemStore
			oldDBGetter := dbGetter
			defer func() {
				orderStore = oldOrderStore
				orderItemStore = oldOrderItemStore
				dbGetter = oldDBGetter
			}()
			orderStore = orderMock
			orderItemStore = orderItemMock
			if mockDB != nil {
				dbGetter = func() database.DB { return mockDB }
			}

			var o oservice
			got, gotErr := o.MergeTempOrder(tt.tempOrderID, tt.customerID, tt.shopID, tt.activeOrderID)

			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("MergeTempOrder() error = %v, wantErr %v", gotErr, tt.wantErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("MergeTempOrder() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeTempOrder() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
