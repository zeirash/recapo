package service

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/go-pdf/fpdf"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/constant"
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
		GetOrderItemByID(orderItemID, orderID int) (*response.OrderItemData, error)
		GetOrderItemsByOrderID(orderID int) ([]response.OrderItemData, error)

		MergeTempOrder(tempOrderID, customerID, shopID int, activeOrderID *int) (*response.OrderData, error)

		GenerateOrderInvoice(orderID, shopID int, message string) ([]byte, error)

		CreateTempOrder(customerName, customerPhone, shareToken string, items []CreateTempOrderItemInput) (response.TempOrderData, error)
		GetTempOrderByID(id int, shopID ...int) (*response.TempOrderData, error)
		GetTempOrdersByShopID(shopID int, opts model.OrderFilterOptions) ([]response.TempOrderData, error)
		RejectTempOrderByID(id int) (response.TempOrderData, error)
	}

	oservice struct{}

	UpdateOrderInput struct {
		ID            int
		TotalPrice    *int
		Status        *string
		PaymentStatus *string
		Notes         *string
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
	activeOrder, err := orderStore.GetActiveOrderByCustomerID(customerID, shopID)
	if err != nil {
		return response.OrderData{}, err
	}
	if activeOrder != nil {
		return response.OrderData{}, errors.New(apierr.ErrActiveOrderExists)
	}

	order, err := orderStore.CreateOrder(nil, customerID, shopID, notes, nil)
	if err != nil {
		return response.OrderData{}, err
	}

	res := response.OrderData{
		ID:            order.ID,
		CustomerName:  order.CustomerName,
		TotalPrice:    order.TotalPrice,
		Status:        order.Status,
		PaymentStatus: order.PaymentStatus,
		Notes:         order.Notes,
		CreatedAt:     order.CreatedAt,
	}

	return res, nil
}

func (o *oservice) GetOrderByID(id int, shopID ...int) (*response.OrderData, error) {
	order, err := orderStore.GetOrderByID(id, shopID...)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, errors.New(apierr.ErrOrderNotFound)
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
		ID:            order.ID,
		CustomerName:  order.CustomerName,
		TotalPrice:    order.TotalPrice,
		Status:        order.Status,
		PaymentStatus: order.PaymentStatus,
		Notes:         order.Notes,
		OrderItems:    orderItemsData,
		CreatedAt:     order.CreatedAt,
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
			ID:            order.ID,
			CustomerName:  order.CustomerName,
			TotalPrice:    order.TotalPrice,
			Status:        order.Status,
			PaymentStatus: order.PaymentStatus,
			Notes:         order.Notes,
			CreatedAt:     order.CreatedAt,
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
		return response.OrderData{}, errors.New(apierr.ErrOrderNotFound)
	}

	updateData := store.UpdateOrderInput{
		TotalPrice:    input.TotalPrice,
		Status:        input.Status,
		PaymentStatus: input.PaymentStatus,
		Notes:         input.Notes,
	}

	orderData, err := orderStore.UpdateOrder(nil, input.ID, updateData)
	if err != nil {
		return response.OrderData{}, err
	}

	res := response.OrderData{
		ID:            orderData.ID,
		CustomerName:  orderData.CustomerName,
		TotalPrice:    orderData.TotalPrice,
		Status:        orderData.Status,
		PaymentStatus: orderData.PaymentStatus,
		Notes:         orderData.Notes,
		CreatedAt:     orderData.CreatedAt,
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
	orderItem, err := orderItemStore.CreateOrderItem(nil, orderID, productID, qty)
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
		return response.OrderItemData{}, errors.New(apierr.ErrOrderItemNotFound)
	}

	updateData := store.UpdateOrderItemInput{
		ProductID: input.ProductID,
		Qty:       input.Qty,
	}

	orderItemData, err := orderItemStore.UpdateOrderItemByID(nil, input.OrderItemID, input.OrderID, updateData)
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

func (o *oservice) GetOrderItemByID(orderItemID, orderID int) (*response.OrderItemData, error) {
	orderItem, err := orderItemStore.GetOrderItemByID(orderItemID)
	if err != nil {
		return nil, err
	}

	if orderItem == nil {
		return nil, errors.New(apierr.ErrOrderItemNotFound)
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
		t := orderItem.UpdatedAt.Time
		res.UpdatedAt = &t
	}

	return &res, nil
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
			t := orderItem.UpdatedAt.Time
			res.UpdatedAt = &t
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
		return response.TempOrderData{}, errors.New(apierr.ErrShopNotFound)
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
		return nil, errors.New(apierr.ErrTempOrderNotFound)
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
			ProductID:   tempOrderItem.ProductID,
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

func (o *oservice) MergeTempOrder(tempOrderID, customerID, shopID int, activeOrderID *int) (*response.OrderData, error) {
	if activeOrderID == nil {
		return o.createOrderFromTempOrder(tempOrderID, customerID, shopID)
	}

	return o.resolveActiveOrderConflict(tempOrderID, shopID, *activeOrderID)
}

func (o *oservice) RejectTempOrderByID(id int) (response.TempOrderData, error) {
	tempOrder, err := o.GetTempOrderByID(id)
	if err != nil {
		return response.TempOrderData{}, err
	}

	err = orderStore.UpdateTempOrderStatus(nil, id, constant.TempOrderStatusRejected)
	if err != nil {
		return response.TempOrderData{}, err
	}

	tempOrder.Status = constant.TempOrderStatusRejected

	return *tempOrder, nil
}

func (o *oservice) GenerateOrderInvoice(orderID, shopID int, message string) ([]byte, error) {
	order, err := o.GetOrderByID(orderID, shopID)
	if err != nil {
		return nil, err
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(false)
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(190, 10, "INVOICE", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Invoice meta
	pdf.SetFont("Arial", "", 11)
	pdf.CellFormat(40, 7, "Invoice #:", "", 0, "L", false, 0, "")
	pdf.CellFormat(150, 7, strconv.Itoa(order.ID), "", 1, "L", false, 0, "")
	pdf.CellFormat(40, 7, "Date:", "", 0, "L", false, 0, "")
	pdf.CellFormat(150, 7, order.CreatedAt.Format("02 January 2006"), "", 1, "L", false, 0, "")
	pdf.CellFormat(40, 7, "Customer:", "", 0, "L", false, 0, "")
	pdf.CellFormat(150, 7, order.CustomerName, "", 1, "L", false, 0, "")
	pdf.Ln(6)

	// Items table header
	pdf.SetFont("Arial", "B", 11)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(80, 8, "Product", "1", 0, "L", true, 0, "")
	pdf.CellFormat(25, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(42, 8, "Price (Rp)", "1", 0, "R", true, 0, "")
	pdf.CellFormat(43, 8, "Subtotal (Rp)", "1", 1, "R", true, 0, "")

	// Items rows
	pdf.SetFont("Arial", "", 10)
	for _, item := range order.OrderItems {
		pdf.CellFormat(80, 7, item.ProductName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(25, 7, strconv.Itoa(item.Qty), "1", 0, "C", false, 0, "")
		pdf.CellFormat(42, 7, formatRupiah(item.Price), "1", 0, "R", false, 0, "")
		pdf.CellFormat(43, 7, formatRupiah(item.Price*item.Qty), "1", 1, "R", false, 0, "")
	}

	// Total row
	pdf.Ln(2)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(147, 8, "Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(43, 8, formatRupiah(order.TotalPrice), "1", 1, "R", false, 0, "")

	// Custom message footer
	if message != "" {
		pdf.Ln(8)
		pdf.SetFont("Arial", "I", 10)
		pdf.MultiCell(190, 6, message, "", "L", false)
	}

	var buf bytes.Buffer
	if err = pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// formatRupiah formats integer price with period thousands separator (e.g. 1500000 → "1.500.000")
func formatRupiah(price int) string {
	s := strconv.Itoa(price)
	n := len(s)
	if n <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func (o *oservice) createOrderFromTempOrder(tempOrderID, customerID, shopID int) (*response.OrderData, error) {
	tempOrder, err := o.GetTempOrderByID(tempOrderID, shopID)
	if err != nil {
		return nil, err
	}

	db := dbGetter()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	order, err := orderStore.CreateOrder(tx, customerID, shopID, nil, &tempOrder.TotalPrice)
	if err != nil {
		return nil, err
	}

	orderItems := []response.OrderItemData{}
	for _, tempOrderItem := range tempOrder.TempOrderItems {
		orderItem, err := orderItemStore.CreateOrderItem(tx, order.ID, tempOrderItem.ProductID, tempOrderItem.Qty)
		if err != nil {
			return nil, err
		}
		orderItems = append(orderItems, response.OrderItemData{
			ID:          orderItem.ID,
			OrderID:     orderItem.OrderID,
			ProductName: orderItem.ProductName,
			Price:       orderItem.Price,
			Qty:         orderItem.Qty,
			CreatedAt:   orderItem.CreatedAt,
		})
	}

	err = orderStore.UpdateTempOrderStatus(tx, tempOrder.ID, constant.TempOrderStatusAccepted)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &response.OrderData{
		ID:            order.ID,
		CustomerName:  order.CustomerName,
		TotalPrice:    order.TotalPrice,
		Status:        order.Status,
		PaymentStatus: order.PaymentStatus,
		Notes:         order.Notes,
		OrderItems:    orderItems,
		CreatedAt:     order.CreatedAt,
	}, nil
}

func (o *oservice) resolveActiveOrderConflict(tempOrderID, shopID, activeOrderID int) (*response.OrderData, error) {
	tempOrder, err := o.GetTempOrderByID(tempOrderID, shopID)
	if err != nil {
		return nil, err
	}

	// TODO: try to make merge logic more efficient
	activeOrder, err := o.GetOrderByID(activeOrderID, shopID)
	if err != nil {
		return nil, err
	}

	db := dbGetter()
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, tempOrderItem := range tempOrder.TempOrderItems {
		// check if order item in temp order already exists in active order
		existsOrderItem, err := orderItemStore.GetOrderItemByProductID(tempOrderItem.ProductID, activeOrder.ID)
		if err != nil {
			return nil, err
		}
		if existsOrderItem != nil {
			// update order item quantity
			qty := existsOrderItem.Qty + tempOrderItem.Qty
			_, err := orderItemStore.UpdateOrderItemByID(tx, existsOrderItem.ID, activeOrder.ID, store.UpdateOrderItemInput{
				Qty: &qty,
			})
			if err != nil {
				return nil, err
			}
		} else {
			// create order item
			_, err := orderItemStore.CreateOrderItem(tx, activeOrder.ID, tempOrderItem.ProductID, tempOrderItem.Qty)
			if err != nil {
				return nil, err
			}
		}
	}

	err = orderStore.UpdateTempOrderStatus(tx, tempOrder.ID, constant.TempOrderStatusAccepted)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	orderItems, err := o.GetOrderItemsByOrderID(activeOrder.ID)
	if err != nil {
		return nil, err
	}

	var totalPrice int
	for _, orderItem := range orderItems {
		totalPrice += orderItem.Price * orderItem.Qty
	}

	return &response.OrderData{
		ID:            activeOrder.ID,
		CustomerName:  activeOrder.CustomerName,
		TotalPrice:    totalPrice,
		Status:        activeOrder.Status,
		PaymentStatus: activeOrder.PaymentStatus,
		Notes:         activeOrder.Notes,
		OrderItems:    orderItems,
		CreatedAt:     activeOrder.CreatedAt,
	}, nil
}
