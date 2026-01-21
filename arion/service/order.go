package service

import (
	"errors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	OrderService interface {
		CreateOrder(customerID int, shopID int) (response.OrderData, error)
		GetOrderByID(id int, shopID ...int) (*response.OrderData, error)
		GetOrdersByShopID(shopID int) ([]response.OrderData, error)
		UpdateOrderByID(input UpdateOrderInput) (response.OrderData, error)
		DeleteOrderByID(id int) error
		CreateOrderItem(orderID, productID, qty int) (response.OrderItemData, error)
		UpdateOrderItemByID(input UpdateOrderItemInput) (response.OrderItemData, error)
		DeleteOrderItemByID(orderItemID, orderID int) error
		GetOrderItemByID(orderItemID, orderID int) (response.OrderItemData, error)
		GetOrderItemsByOrderID(orderID int) ([]response.OrderItemData, error)
	}

	oservice struct{}

	UpdateOrderInput struct {
		ID         int
		CustomerID *int
		TotalPrice *int
		Status     *string
	}

	UpdateOrderItemInput struct {
		OrderID     int
		OrderItemID int
		ProductID   *int
		Qty         *int
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

func (o *oservice) CreateOrder(customerID int, shopID int) (response.OrderData, error) {
	// TODO: better hardcoded status on service or store level?
	order, err := orderStore.CreateOrder(customerID, shopID, "created")
	if err != nil {
		return response.OrderData{}, err
	}

	res := response.OrderData{
		ID: order.ID,
		CustomerName: order.CustomerName,
		TotalPrice: order.TotalPrice,
		Status: order.Status,
		CreatedAt: order.CreatedAt,
	}

	if order.UpdatedAt.Valid {
		res.UpdatedAt = &order.UpdatedAt.Time
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
		ID: order.ID,
		CustomerName: order.CustomerName,
		TotalPrice: order.TotalPrice,
		Status: order.Status,
		OrderItems: orderItemsData,
	}

	return &res, nil
}

func (o *oservice) GetOrdersByShopID(shopID int) ([]response.OrderData, error) {
	orders, err := orderStore.GetOrdersByShopID(shopID)
	if err != nil {
		return []response.OrderData{}, err
	}

	var ordersData []response.OrderData
	for _, order := range orders {
		res := response.OrderData{
			ID:           order.ID,
			CustomerName: order.CustomerName,
			TotalPrice:   order.TotalPrice,
			Status:       order.Status,
			CreatedAt:    order.CreatedAt,
		}

		if order.UpdatedAt.Valid {
			res.UpdatedAt = &order.UpdatedAt.Time
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
		CustomerID: input.CustomerID,
		TotalPrice: input.TotalPrice,
		Status:     input.Status,
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
		CreatedAt:    orderData.CreatedAt,
	}

	if orderData.UpdatedAt.Valid {
		res.UpdatedAt = &orderData.UpdatedAt.Time
	}

	return res, nil
}

func (o *oservice) DeleteOrderByID(id int) error {
	db := database.GetDB()
	defer db.Close()

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
		Price:			 orderItem.Price,
		Qty:				 orderItem.Qty,
		CreatedAt:   orderItem.CreatedAt,
	}

	if orderItem.UpdatedAt.Valid {
		res.UpdatedAt = &orderItem.UpdatedAt.Time
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
		return response.OrderItemData{}, errors.New("order item not found")
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
