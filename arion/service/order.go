package service

import (
	"errors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/store"
)

type (
	OrderService interface {
		CreateOrder(customerID int, shopID int, notes *string) (response.OrderData, error)
		GetOrderByID(id int, shopID ...int) (*response.OrderData, error)
		GetOrdersByShopID(shopID int, opts model.OrderFilterOptions) ([]response.OrderData, error)
		UpdateOrderByID(input UpdateOrderInput) (response.OrderData, error)
		DeleteOrderByID(id int) error

		CreateOrderItem(orderID, productID, qty int) (response.OrderItemData, error)
		UpdateOrderItemByID(input UpdateOrderItemInput) (response.OrderItemData, error)
		DeleteOrderItemByID(orderItemID, orderID int) error
		GetOrderItemByID(orderItemID, orderID int) (response.OrderItemData, error)
		GetOrderItemsByOrderID(orderID int) ([]response.OrderItemData, error)

		CreateTempOrder(customerName, customerPhone, shareToken string, items []CreateTempOrderItemInput) (response.TempOrderData, error)
		GetTempOrderByID(id int, shopID ...int) (*response.TempOrderData, error)
		GetTempOrdersByShopID(shopID int, opts model.OrderFilterOptions) ([]response.TempOrderData, error)
		// UpdateOrderTempByID(input UpdateOrderTempInput) (response.OrderTempData, error)
		// DeleteOrderTempByID(id int) error

		// UpdateOrderTempItemByID(input UpdateOrderTempItemInput) (response.OrderTempItemData, error)
		// DeleteOrderTempItemByID(orderTempItemID, orderTempID int) error
	}

	oservice struct{}

	UpdateOrderInput struct {
		ID         int
		TotalPrice *int
		Status     *string
		Notes      *string
	}

	UpdateOrderItemInput struct {
		OrderID     int
		OrderItemID int
		ProductID   *int
		Qty         *int
	}

	CreateTempOrderItemInput struct {
		ProductID int
		Qty       int
	}
)

func NewOrderService() OrderService {
	cfg = config.GetConfig()

	if orderStore == nil {
		orderStore = store.NewOrderStore()
	}

	if orderItemStore == nil {
		orderItemStore = store.NewOrderItemStore()
	}

	return &oservice{}
}

func (o *oservice) CreateOrder(customerID int, shopID int, notes *string) (response.OrderData, error) {
	order, err := orderStore.CreateOrder(customerID, shopID, notes)
	if err != nil {
		return response.OrderData{}, err
	}

	res := response.OrderData{
		ID:           order.ID,
		CustomerName: order.CustomerName,
		TotalPrice:   order.TotalPrice,
		Status:       order.Status,
		Notes:        order.Notes,
		CreatedAt:    order.CreatedAt,
	}

	return res, nil
}

func (o *oservice) GetOrderByID(id int, shopID ...int) (*response.OrderData, error) {
	order, err := orderStore.GetOrderByID(id, shopID...)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, errors.New("order not found")
	}

	orderItems, err := orderItemStore.GetOrderItemsByOrderID(order.ID)
	if err != nil {
		return nil, err
	}

	orderItemsData := []response.OrderItemData{}
	for _, orderItem := range orderItems {
		orderItemsData = append(orderItemsData, response.OrderItemData{
			ID:          orderItem.ID,
			ProductName: orderItem.ProductName,
			Price:       orderItem.Price,
			Qty:         orderItem.Qty,
			CreatedAt:   orderItem.CreatedAt,
		})
	}

	res := response.OrderData{
		ID:           order.ID,
		CustomerName: order.CustomerName,
		TotalPrice:   order.TotalPrice,
		Status:       order.Status,
		Notes:        order.Notes,
		OrderItems:   orderItemsData,
		CreatedAt:    order.CreatedAt,
	}

	if order.UpdatedAt.Valid {
		t := order.UpdatedAt.Time
		res.UpdatedAt = &t
	}

	return &res, nil
}

func (o *oservice) GetOrdersByShopID(shopID int, opts model.OrderFilterOptions) ([]response.OrderData, error) {
	orders, err := orderStore.GetOrdersByShopID(shopID, opts)
	if err != nil {
		return []response.OrderData{}, err
	}

	ordersData := []response.OrderData{}
	for _, order := range orders {
		res := response.OrderData{
			ID:           order.ID,
			CustomerName: order.CustomerName,
			TotalPrice:   order.TotalPrice,
			Status:       order.Status,
			Notes:        order.Notes,
			CreatedAt:    order.CreatedAt,
		}

		if order.UpdatedAt.Valid {
			t := order.UpdatedAt.Time
			res.UpdatedAt = &t
		}

		ordersData = append(ordersData, res)
	}

	return ordersData, nil
}

func (o *oservice) UpdateOrderByID(input UpdateOrderInput) (response.OrderData, error) {
	order, err := orderStore.GetOrderByID(input.ID)
	if err != nil {
		return response.OrderData{}, err
	}

	if order == nil {
		return response.OrderData{}, errors.New("order not found")
	}

	updateData := store.UpdateOrderInput{
		TotalPrice: input.TotalPrice,
		Status:     input.Status,
		Notes:      input.Notes,
	}

	orderData, err := orderStore.UpdateOrder(input.ID, updateData)
	if err != nil {
		return response.OrderData{}, err
	}

	res := response.OrderData{
		ID:           orderData.ID,
		CustomerName: orderData.CustomerName,
		TotalPrice:   orderData.TotalPrice,
		Status:       orderData.Status,
		Notes:        orderData.Notes,
		CreatedAt:    orderData.CreatedAt,
	}

	if orderData.UpdatedAt.Valid {
		res.UpdatedAt = &orderData.UpdatedAt.Time
	}

	return res, nil
}

func (o *oservice) DeleteOrderByID(id int) error {
	db := dbGetter()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = orderItemStore.DeleteOrderItemsByOrderID(tx, id)
	if err != nil {
		return err
	}

	err = orderStore.DeleteOrderByID(tx, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (o *oservice) CreateOrderItem(orderID, productID, qty int) (response.OrderItemData, error) {
	orderItem, err := orderItemStore.CreateOrderItem(orderID, productID, qty)
	if err != nil {
		return response.OrderItemData{}, err
	}

	res := response.OrderItemData{
		ID:          orderItem.ID,
		OrderID:     orderItem.OrderID,
		ProductName: orderItem.ProductName,
		Price:       orderItem.Price,
		Qty:         orderItem.Qty,
		CreatedAt:   orderItem.CreatedAt,
	}

	return res, nil
}

func (o *oservice) UpdateOrderItemByID(input UpdateOrderItemInput) (response.OrderItemData, error) {
	orderItem, err := orderItemStore.GetOrderItemByID(input.OrderItemID)
	if err != nil {
		return response.OrderItemData{}, err
	}

	if orderItem == nil {
		return response.OrderItemData{}, errors.New("order item not found")
	}

	updateData := store.UpdateOrderItemInput{
		ProductID: input.ProductID,
		Qty:       input.Qty,
	}

	orderItemData, err := orderItemStore.UpdateOrderItemByID(input.OrderItemID, input.OrderID, updateData)
	if err != nil {
		return response.OrderItemData{}, err
	}

	res := response.OrderItemData{
		ID:          orderItemData.ID,
		OrderID:     orderItemData.OrderID,
		ProductName: orderItemData.ProductName,
		Price:       orderItemData.Price,
		Qty:         orderItemData.Qty,
		CreatedAt:   orderItemData.CreatedAt,
	}

	if orderItemData.UpdatedAt.Valid {
		res.UpdatedAt = &orderItemData.UpdatedAt.Time
	}

	return res, nil
}

func (o *oservice) DeleteOrderItemByID(orderItemID, orderID int) error {
	err := orderItemStore.DeleteOrderItemByID(orderItemID, orderID)
	if err != nil {
		return err
	}

	return nil
}

func (o *oservice) GetOrderItemByID(orderItemID, orderID int) (response.OrderItemData, error) {
	orderItem, err := orderItemStore.GetOrderItemByID(orderItemID)
	if err != nil {
		return response.OrderItemData{}, err
	}

	if orderItem == nil {
		return response.OrderItemData{}, nil
	}

	res := response.OrderItemData{
		ID:          orderItem.ID,
		OrderID:     orderItem.OrderID,
		ProductName: orderItem.ProductName,
		Price:       orderItem.Price,
		Qty:         orderItem.Qty,
		CreatedAt:   orderItem.CreatedAt,
	}

	if orderItem.UpdatedAt.Valid {
		res.UpdatedAt = &orderItem.UpdatedAt.Time
	}

	return res, nil
}

func (o *oservice) GetOrderItemsByOrderID(orderID int) ([]response.OrderItemData, error) {
	orderItems, err := orderItemStore.GetOrderItemsByOrderID(orderID)
	if err != nil {
		return []response.OrderItemData{}, err
	}

	orderItemsData := []response.OrderItemData{}
	for _, orderItem := range orderItems {
		res := response.OrderItemData{
			ID:          orderItem.ID,
			ProductName: orderItem.ProductName,
			Price:       orderItem.Price,
			Qty:         orderItem.Qty,
			CreatedAt:   orderItem.CreatedAt,
		}

		if orderItem.UpdatedAt.Valid {
			res.UpdatedAt = &orderItem.UpdatedAt.Time
		}

		orderItemsData = append(orderItemsData, res)
	}

	return orderItemsData, nil
}

func (o *oservice) CreateTempOrder(customerName, customerPhone, shareToken string, items []CreateTempOrderItemInput) (response.TempOrderData, error) {
	shop, err := shopStore.GetShopByShareToken(shareToken)
	if err != nil {
		return response.TempOrderData{}, err
	}

	if shop == nil {
		return response.TempOrderData{}, errors.New("shop not found")
	}

	db := dbGetter()

	tx, err := db.Begin()
	if err != nil {
		return response.TempOrderData{}, err
	}
	defer tx.Rollback()

	tempOrder, err := orderStore.CreateTempOrder(tx, customerName, customerPhone, shop.ID)
	if err != nil {
		return response.TempOrderData{}, err
	}

	tempOrderItemsData := []response.TempOrderItemData{}
	totalPrice := 0
	for _, item := range items {
		orderItemTemp, err := orderItemStore.CreateTempOrderItem(tx, tempOrder.ID, item.ProductID, item.Qty)
		if err != nil {
			return response.TempOrderData{}, err
		}
		tempOrderItemsData = append(tempOrderItemsData, response.TempOrderItemData{
			ID:          orderItemTemp.ID,
			TempOrderID: orderItemTemp.TempOrderID,
			ProductName: orderItemTemp.ProductName,
			Price:       orderItemTemp.Price,
			Qty:         orderItemTemp.Qty,
			CreatedAt:   orderItemTemp.CreatedAt,
		})
		totalPrice += item.Qty * orderItemTemp.Price
	}

	err = orderStore.UpdateTempOrderTotalPrice(tx, tempOrder.ID, totalPrice)
	if err != nil {
		return response.TempOrderData{}, err
	}

	err = tx.Commit()
	if err != nil {
		return response.TempOrderData{}, err
	}

	res := response.TempOrderData{
		ID:             tempOrder.ID,
		CustomerName:   tempOrder.CustomerName,
		CustomerPhone:  tempOrder.CustomerPhone,
		TotalPrice:     tempOrder.TotalPrice,
		Status:         tempOrder.Status,
		TempOrderItems: tempOrderItemsData,
		CreatedAt:      tempOrder.CreatedAt,
	}
	if tempOrder.UpdatedAt.Valid {
		res.UpdatedAt = &tempOrder.UpdatedAt.Time
	}
	return res, nil
}

func (o *oservice) GetTempOrderByID(id int, shopID ...int) (*response.TempOrderData, error) {
	tempOrder, err := orderStore.GetTempOrderByID(id, shopID...)
	if err != nil {
		return nil, err
	}

	if tempOrder == nil {
		return nil, errors.New("temp order not found")
	}

	tempOrderItems, err := orderItemStore.GetTempOrderItemsByTempOrderID(tempOrder.ID)
	if err != nil {
		return nil, err
	}

	tempOrderItemsData := []response.TempOrderItemData{}
	for _, tempOrderItem := range tempOrderItems {
		tempOrderItemsData = append(tempOrderItemsData, response.TempOrderItemData{
			ID:          tempOrderItem.ID,
			TempOrderID: tempOrderItem.TempOrderID,
			ProductName: tempOrderItem.ProductName,
			Price:       tempOrderItem.Price,
			Qty:         tempOrderItem.Qty,
			CreatedAt:   tempOrderItem.CreatedAt,
		})
	}

	res := response.TempOrderData{
		ID:             tempOrder.ID,
		CustomerName:   tempOrder.CustomerName,
		CustomerPhone:  tempOrder.CustomerPhone,
		TotalPrice:     tempOrder.TotalPrice,
		Status:         tempOrder.Status,
		TempOrderItems: tempOrderItemsData,
		CreatedAt:      tempOrder.CreatedAt,
	}

	if tempOrder.UpdatedAt.Valid {
		t := tempOrder.UpdatedAt.Time
		res.UpdatedAt = &t
	}

	return &res, nil
}

func (o *oservice) GetTempOrdersByShopID(shopID int, opts model.OrderFilterOptions) ([]response.TempOrderData, error) {
	tempOrders, err := orderStore.GetTempOrdersByShopID(shopID, opts)
	if err != nil {
		return []response.TempOrderData{}, err
	}

	tempOrdersData := []response.TempOrderData{}
	for _, tempOrder := range tempOrders {
		res := response.TempOrderData{
			ID:             tempOrder.ID,
			CustomerName:   tempOrder.CustomerName,
			CustomerPhone:  tempOrder.CustomerPhone,
			TotalPrice:     tempOrder.TotalPrice,
			Status:         tempOrder.Status,
			CreatedAt:      tempOrder.CreatedAt,
		}

		if tempOrder.UpdatedAt.Valid {
			t := tempOrder.UpdatedAt.Time
			res.UpdatedAt = &t
		}

		tempOrdersData = append(tempOrdersData, res)
	}

	return tempOrdersData, nil
}
