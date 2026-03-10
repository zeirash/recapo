package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/constant"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/service"
)

type (
	CreateOrderRequest struct {
		CustomerID int     `json:"customer_id"`
		Notes      *string `json:"notes"`
	}

	UpdateOrderRequest struct {
		CustomerID    *int    `json:"customer_id"`
		TotalPrice    *int    `json:"total_price"`
		Status        *string `json:"status"`
		PaymentStatus *string `json:"payment_status"`
		Notes         *string `json:"notes"`
	}

	CreateOrderItemRequest struct {
		ProductID int `json:"product_id"`
		Qty       int `json:"qty"`
	}

	UpdateOrderItemRequest struct {
		ProductID *int `json:"product_id"`
		Qty       *int `json:"qty"`
	}

	MergeOrderRequest struct {
		TempOrderID   int  `json:"temp_order_id"`
		CustomerID    int  `json:"customer_id"`
		ActiveOrderID *int `json:"active_order_id"`
	}

	OrderPaymentRequest struct {
		Amount int `json:"amount"`
	}
)

// CreateOrderHandler godoc
//
//	@Summary		Create order
//	@Description	Create a new order for the shop.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateOrderRequest	true	"Order data"
//	@Success		200		{object}	response.OrderData
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON or validation)"
//	@Failure		500		{object}	ErrorApiResponse	"Internal server error"
//	@Router			/order [post]
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	inp := CreateOrderRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateCreateOrder(inp); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := orderService.CreateOrder(inp.CustomerID, shopID, inp.Notes)
	if err != nil {
		if err.Error() == apierr.ErrActiveOrderExists {
			WriteErrorJson(w, r, http.StatusConflict, err, "duplicate_customer_order")
			return
		}
		logger.WithError(err).Error("create_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "create_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetOrderHandler godoc
//
//	@Summary		Get order by ID
//	@Description	Get a single order by ID.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path		int	true	"Order ID"
//	@Success		200			{object}	response.OrderData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid order_id)"
//	@Failure		404	{object}	ErrorApiResponse	"Order not found"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id} [get]
func GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])

	res, err := orderService.GetOrderByID(orderIDInt, shopID)
	if err != nil {
		if err.Error() == apierr.ErrOrderNotFound {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("get_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// ExportOrderRequest is the request body for ExportOrderHandler.
type ExportOrderRequest struct {
	Message string `json:"message"`
}

// ExportOrderHandler godoc
//
//	@Summary		Export order as PDF invoice
//	@Description	Generate and download a PDF invoice for a given order. Optional message in body appended as footer; falls back to order notes.
//	@Tags			order
//	@Accept			json
//	@Produce		application/pdf
//	@Security		BearerAuth
//	@Param			order_id	path		int					true	"Order ID"
//	@Param			body		body		ExportOrderRequest	false	"Optional closing message (supports newlines)"
//	@Success		200			{file}		binary
//	@Failure		400			{object}	ErrorApiResponse	"Bad request (invalid order_id)"
//	@Failure		404			{object}	ErrorApiResponse	"Order not found"
//	@Failure		500			{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id}/export [post]
func ExportOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderID, _ := strconv.Atoi(params["order_id"])

	inp := ExportOrderRequest{}
	if r.Body != nil && r.ContentLength != 0 {
		if err := ParseJson(r.Body, &inp); err != nil {
			WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
			return
		}
	}

	pdfBytes, err := orderService.GenerateOrderInvoice(orderID, shopID, inp.Message)
	if err != nil {
		if err.Error() == apierr.ErrOrderNotFound {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("export_order_invoice_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "export_order_invoice")
		return
	}

	filename := fmt.Sprintf("invoice-%d.pdf", orderID)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

// GetOrdersHandler godoc
//
//	@Summary		List orders
//	@Description	Get all orders for the shop. Optional search query to filter.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search		query		string	false	"Search query"
//	@Param			date_from	query		string	false	"Filter orders from date (YYYY-MM-DD)"
//	@Param			date_to		query		string	false	"Filter orders to date (YYYY-MM-DD)"
//	@Param			status		query		string	false	"Filter by status (e.g. pending,accepted,rejected)"
//	@Param			sort		  query		string	false	"Sort by column and order (e.g. created_at,desc)"
//	@Success		200		{array}		response.OrderData
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders [get]
func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	opts := model.OrderFilterOptions{}
	if s := r.URL.Query().Get("status"); s != "" && s != constant.FilterStatusAll {
		statuses := strings.Split(s, ",")
		opts.Status = statuses
	}
	if q := r.URL.Query().Get("search"); q != "" {
		opts.SearchQuery = &q
	}
	if df := r.URL.Query().Get("date_from"); df != "" {
		if t, err := parseDate(df); err == nil {
			opts.DateFrom = &t
		}
	}
	if dt := r.URL.Query().Get("date_to"); dt != "" {
		if t, err := parseDate(dt); err == nil {
			endOfDay := t.Add(24 * time.Hour)
			opts.DateTo = &endOfDay
		}
	}
	if sort := r.URL.Query().Get("sort"); sort != "" {
		opts.Sort = &sort
	}

	res, err := orderService.GetOrdersByShopID(shopID, opts)
	if err != nil {
		logger.WithError(err).Error("get_orders_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_orders")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// UpdateOrderHandler godoc
//
//	@Summary		Update order
//	@Description	Update an existing order. Only provided fields are updated.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path		int					true	"Order ID"
//	@Param			body		body		UpdateOrderRequest	true	"Fields to update"
//	@Success		200			{object}	response.OrderData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid JSON or order_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id} [patch]
func UpdateOrderHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])

	inp := UpdateOrderRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := orderService.UpdateOrderByID(service.UpdateOrderInput{
		ID:            orderIDInt,
		TotalPrice:    inp.TotalPrice,
		Status:        inp.Status,
		PaymentStatus: inp.PaymentStatus,
		Notes:         inp.Notes,
	})
	if err != nil {
		logger.WithError(err).Error("update_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "update_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// DeleteOrderHandler godoc
//
//	@Summary		Delete order
//	@Description	Delete an order by ID.
//	@Description	Success Response envelope: { success, data, code, message }. data contains "OK" on success.
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path	int	true	"Order ID"
//	@Success		200		{string}	string	"Success. data contains \"OK\""
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid order_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id} [delete]
func DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])

	err := orderService.DeleteOrderByID(orderIDInt)
	if err != nil {
		logger.WithError(err).Error("delete_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "delete_order")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

// MergeTempOrderHandler godoc
//
//	@Summary		Merge temp order
//	@Description	Accept a temp order by merging it into a new order or into an existing active order. Requires temp_order_id and customer_id; active_order_id is optional. When omitted, a new order is created from the temp order. When provided, temp order items are merged into that order.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		MergeOrderRequest	true	"temp_order_id, customer_id required; active_order_id optional"
//	@Success		200		{object}	response.OrderData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid JSON or validation)"
//	@Failure		404	{object}	ErrorApiResponse	"Temp order or order not found"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/temp_orders/merge [post]
func MergeTempOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	inp := MergeOrderRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateMergeTempOrder(inp); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := orderService.MergeTempOrder(inp.TempOrderID, inp.CustomerID, shopID, inp.ActiveOrderID)
	if err != nil {
		if err.Error() == apierr.ErrOrderNotFound {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("merge_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "merge_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// CreateOrderItemHandler godoc
//
//	@Summary		Create order item
//	@Description	Add an item to an order. Requires order_id (path), product_id and qty (body).
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path		int						true	"Order ID"
//	@Param			body		body		CreateOrderItemRequest	true	"product_id, qty"
//	@Success		200			{object}	response.OrderItemData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid JSON, order_id, or validation)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id}/item [post]
func CreateOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	inp := CreateOrderItemRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateCreateOrderItem(inp); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])

	res, err := orderService.CreateOrderItem(orderIDInt, inp.ProductID, inp.Qty)
	if err != nil {
		logger.WithError(err).Error("create_order_item_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "create_order_item")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// UpdateOrderItemHandler godoc
//
//	@Summary		Update order item
//	@Description	Update an order item. Only provided fields are updated.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path		int						true	"Order ID"
//	@Param			item_id	path		int						true	"Order item ID"
//	@Param			body		body		UpdateOrderItemRequest	true	"Fields to update"
//	@Success		200			{object}	response.OrderItemData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid JSON, order_id, or item_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id}/items/{item_id} [patch]
func UpdateOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	if valid, err := validateOrderItemID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	inp := UpdateOrderItemRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderItemIDInt, _ := strconv.Atoi(params["item_id"])

	res, err := orderService.UpdateOrderItemByID(service.UpdateOrderItemInput{
		OrderID:     orderIDInt,
		OrderItemID: orderItemIDInt,
		ProductID:   inp.ProductID,
		Qty:         inp.Qty,
	})
	if err != nil {
		logger.WithError(err).Error("update_order_item_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "update_order_item")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// DeleteOrderItemHandler godoc
//
//	@Summary		Delete order item
//	@Description	Delete an order item by ID.
//	@Description	Success Response envelope: { success, data, code, message }. data contains "OK" on success.
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path	int	true	"Order ID"
//	@Param			item_id	path	int	true	"Order item ID"
//	@Success		200		{string}	string	"Success. data contains \"OK\""
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid order_id or item_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id}/items/{item_id} [delete]
func DeleteOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	if valid, err := validateOrderItemID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderItemIDInt, _ := strconv.Atoi(params["item_id"])

	err := orderService.DeleteOrderItemByID(orderItemIDInt, orderIDInt)
	if err != nil {
		logger.WithError(err).Error("delete_order_item_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "delete_order_item")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

// GetOrderItemHandler godoc
//
//	@Summary		Get order item by ID
//	@Description	Get a single order item by ID.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path	int	true	"Order ID"
//	@Param			item_id	path	int	true	"Order item ID"
//	@Success		200		{object}	response.OrderItemData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid order_id or item_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id}/items/{item_id} [get]
func GetOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	if valid, err := validateOrderItemID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderItemIDInt, _ := strconv.Atoi(params["item_id"])

	res, err := orderService.GetOrderItemByID(orderItemIDInt, orderIDInt)
	if err != nil {
		logger.WithError(err).Error("get_order_item_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_order_item")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetOrderItemsHandler godoc
//
//	@Summary		List order items
//	@Description	Get all items for an order.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path	int	true	"Order ID"
//	@Success		200		{array}	response.OrderItemData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid order_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id}/items [get]
func GetOrderItemsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])

	res, err := orderService.GetOrderItemsByOrderID(orderIDInt)
	if err != nil {
		logger.WithError(err).Error("get_order_items_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_order_items")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetTempOrdersHandler godoc
//
//	@Summary		List temp orders
//	@Description	Get all temp orders for the shop. Optional query params: search (customer name or phone), date_from, date_to (YYYY-MM-DD).
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search		query		string	false	"Search by customer name or phone"
//	@Param			date_from	query		string	false	"Filter from date (YYYY-MM-DD)"
//	@Param			date_to		query		string	false	"Filter to date (YYYY-MM-DD)"
//	@Param			status		query		string	false	"Filter by status (e.g. created,in_progress,in_delivery,done,cancelled)"
//	@Param			sort		  query		string	false	"Sort by column and order (e.g. created_at,desc)"
//	@Success		200			{array}		response.TempOrderData
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/temp_orders [get]
func GetTempOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	opts := model.OrderFilterOptions{}
	if s := r.URL.Query().Get("status"); s != "" && s != constant.FilterStatusAll {
		statuses := strings.Split(s, ",")
		opts.Status = statuses
	}
	if q := r.URL.Query().Get("search"); q != "" {
		opts.SearchQuery = &q
	}
	if df := r.URL.Query().Get("date_from"); df != "" {
		if t, err := parseDate(df); err == nil {
			opts.DateFrom = &t
		}
	}
	if dt := r.URL.Query().Get("date_to"); dt != "" {
		if t, err := parseDate(dt); err == nil {
			endOfDay := t.Add(24 * time.Hour)
			opts.DateTo = &endOfDay
		}
	}
	if sort := r.URL.Query().Get("sort"); sort != "" {
		opts.Sort = &sort
	}

	res, err := orderService.GetTempOrdersByShopID(shopID, opts)
	if err != nil {
		logger.WithError(err).Error("get_temp_orders_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_temp_orders")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetTempOrderHandler godoc
//
//	@Summary		Get temp order by ID
//	@Description	Get a single temp order by ID, including its items.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			temp_order_id	path		int	true	"Temp order ID"
//	@Success		200				{object}	response.TempOrderData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid temp_order_id)"
//	@Failure		404	{object}	ErrorApiResponse	"Temp order not found"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/temp_orders/{temp_order_id} [get]
func GetTempOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	params := mux.Vars(r)

	if valid, err := validateTempOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	tempOrderIDInt, _ := strconv.Atoi(params["temp_order_id"])

	res, err := orderService.GetTempOrderByID(tempOrderIDInt, shopID)
	if err != nil {
		if err.Error() == apierr.ErrTempOrderNotFound {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("get_temp_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_temp_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// RejectTempOrderHandler godoc
//
//	@Summary		Reject temp order
//	@Description	Reject a temp order by ID. Sets the temp order status to rejected.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			temp_order_id	path		int	true	"Temp order ID"
//	@Success		200				{object}	response.TempOrderData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid or missing temp_order_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/temp_orders/{temp_order_id}/reject [patch]
func RejectTempOrderHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateTempOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	tempOrderIDInt, _ := strconv.Atoi(params["temp_order_id"])

	res, err := orderService.RejectTempOrderByID(tempOrderIDInt)
	if err != nil {
		logger.WithError(err).Error("reject_temp_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "reject_temp_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// CreateOrderPaymentHandler godoc
//
//	@Summary		Create order payment
//	@Description	Create a new order payment for the order.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//  @Param			order_id	path		int	true	"Order ID"
//  @Param			body		body		OrderPaymentRequest	true	"Order payment data"
//  @Success		200		{object}	response.OrderPaymentData
//  @Failure		400	{object}	ErrorApiResponse	"Bad request (invalid JSON or order_id)"
//  @Failure		500	{object}	ErrorApiResponse	"Internal server error"
//  @Router			/orders/{order_id}/payment [post]
func CreateOrderPaymentHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	inp := OrderPaymentRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])

	res, err := orderService.CreateOrderPayment(orderIDInt, inp.Amount)
	if err != nil {
		logger.WithError(err).Error("create_order_payment_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "create_order_payment")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// UpdateOrderPaymentAmountHandler godoc
//
//	@Summary		Update order payment amount
//	@Description	Update the amount of an order payment by ID.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path		int	true	"Order ID"
//	@Param			body		body		OrderPaymentRequest	true	"Order payment data"
//	@Success		200		{object}	response.OrderPaymentData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid JSON or order_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id}/payments/{payment_id} [patch]
func UpdateOrderPaymentAmountHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	if valid, err := validateOrderPaymentID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	inp := OrderPaymentRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	paymentIDInt, _ := strconv.Atoi(params["payment_id"])

	res, err := orderService.UpdateOrderPaymentAmountByID(paymentIDInt, orderIDInt, inp.Amount)
	if err != nil {
		logger.WithError(err).Error("update_order_payment_amount_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "update_order_payment_amount")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetOrderPaymentsHandler godoc
//
//	@Summary		List order payments
//	@Description	Get all payments for an order.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path		int	true	"Order ID"
//	@Success		200		{array}	response.OrderPaymentData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid order_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id}/payments [get]
func GetOrderPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])

	res, err := orderService.GetOrderPaymentsByOrderID(orderIDInt)
	if err != nil {
		logger.WithError(err).Error("get_order_payments_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_order_payments")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// DeleteOrderPaymentHandler godoc
//
//	@Summary		Delete order payment
//	@Description	Delete an order payment by ID.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			order
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			order_id	path		int	true	"Order ID"
//	@Param			payment_id	path		int	true	"Payment ID"
//	@Success		200		{string}	string	"Success. data contains \"OK\""
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid JSON or order_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders/{order_id}/payments/{payment_id} [delete]
func DeleteOrderPaymentHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	if valid, err := validateOrderPaymentID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	paymentIDInt, _ := strconv.Atoi(params["payment_id"])

	err := orderService.DeleteOrderPaymentByID(paymentIDInt, orderIDInt)
	if err != nil {
		logger.WithError(err).Error("delete_order_payment_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "delete_order_payment")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

func validateCreateOrderItem(inp CreateOrderItemRequest) (bool, error) {
	if inp.ProductID <= 0 {
		return false, errors.New(apierr.ErrProductIDRequired)
	}

	if inp.Qty <= 0 {
		return false, errors.New(apierr.ErrQtyRequired)
	}

	return true, nil
}

func validateCreateOrder(inp CreateOrderRequest) (bool, error) {
	if inp.CustomerID <= 0 {
		return false, errors.New(apierr.ErrCustomerIDRequired)
	}

	return true, nil
}

func validateOrderID(params map[string]string) (bool, error) {
	if params["order_id"] == "" {
		return false, errors.New(apierr.ErrOrderIDRequired)
	}

	return true, nil
}

func validateOrderItemID(params map[string]string) (bool, error) {
	if params["item_id"] == "" {
		return false, errors.New(apierr.ErrOrderItemIDRequired)
	}

	return true, nil
}

func validateOrderPaymentID(params map[string]string) (bool, error) {
	if params["payment_id"] == "" {
		return false, errors.New(apierr.ErrOrderPaymentIDRequired)
	}

	return true, nil
}

func validateTempOrderID(params map[string]string) (bool, error) {
	if params["temp_order_id"] == "" {
		return false, errors.New(apierr.ErrTempOrderIDRequired)
	}

	return true, nil
}

func validateMergeTempOrder(inp MergeOrderRequest) (bool, error) {
	if inp.TempOrderID <= 0 {
		return false, errors.New(apierr.ErrTempOrderIDRequired)
	}

	if inp.CustomerID <= 0 {
		return false, errors.New(apierr.ErrCustomerIDRequired)
	}

	return true, nil
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.UTC)
}
