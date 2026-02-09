package service

import (
	"database/sql"
	"errors"
	"reflect"
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
	}{
		{
			name:       "successfully create order",
			customerID: 1,
			shopID:     1,
			notes:      nil,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					CreateOrder(1, 1, constant.OrderStatusCreated, nil).
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
			name:       "create order returns error on store failure",
			customerID: 1,
			shopID:     1,
			notes:      nil,
			mockSetup: func(ctrl *gomock.Controller) *mock_store.MockOrderStore {
				mock := mock_store.NewMockOrderStore(ctrl)
				mock.EXPECT().
					CreateOrder(1, 1, constant.OrderStatusCreated, nil).
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
			name:   "get order by ID not found returns nil result",
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
			wantErr:    false,
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
			wantResult: nil,
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
			name: "get orders by shop ID with date_from and date_to returns filtered orders",
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
					UpdateOrder(1, store.UpdateOrderInput{Status: strPtr(constant.OrderStatusDone)}).
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
					UpdateOrder(1, store.UpdateOrderInput{TotalPrice: intPtr(500), Status: strPtr(constant.OrderStatusDone)}).
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
					UpdateOrder(1, store.UpdateOrderInput{Status: strPtr(constant.OrderStatusDone)}).
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
					CreateOrderItem(1, 10, 2).
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
					CreateOrderItem(1, 10, 2).
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
					UpdateOrderItemByID(1, 1, store.UpdateOrderItemInput{Qty: intPtr(5)}).
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
					UpdateOrderItemByID(1, 1, store.UpdateOrderItemInput{ProductID: intPtr(20), Qty: intPtr(3)}).
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
					UpdateOrderItemByID(1, 1, store.UpdateOrderItemInput{Qty: intPtr(5)}).
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
