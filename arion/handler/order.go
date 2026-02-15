package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
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
		CustomerID *int    `json:"customer_id"`
		TotalPrice *int    `json:"total_price"`
		Status     *string `json:"status"`
		Notes      *string `json:"notes"`
	}

	CreateOrderItemRequest struct {
		ProductID  int `json:"product_id"`
		Qty        int `json:"qty"`
	}

	UpdateOrderItemRequest struct {
		ProductID *int `json:"product_id"`
		Qty       *int `json:"qty"`
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
	orderID := orderIDInt

	res, err := orderService.GetOrderByID(orderID, shopID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("get_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
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
//	@Success		200		{array}		response.OrderData
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/orders [get]
func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	opts := model.OrderFilterOptions{}
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
	orderID := orderIDInt

	inp := UpdateOrderRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := orderService.UpdateOrderByID(service.UpdateOrderInput{
		ID:         orderID,
		TotalPrice: inp.TotalPrice,
		Status:     inp.Status,
		Notes:      inp.Notes,
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
	orderID := orderIDInt

	err := orderService.DeleteOrderByID(orderID)
	if err != nil {
		logger.WithError(err).Error("delete_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "delete_order")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

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
	orderID := orderIDInt

	res, err := orderService.CreateOrderItem(orderID, inp.ProductID, inp.Qty)
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
	orderID := orderIDInt

	orderItemIDInt, _ := strconv.Atoi(params["item_id"])
	orderItemID := orderItemIDInt

	res, err := orderService.UpdateOrderItemByID(service.UpdateOrderItemInput{
		OrderID:     orderID,
		OrderItemID: orderItemID,
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
	orderID := orderIDInt

	orderItemIDInt, _ := strconv.Atoi(params["item_id"])
	orderItemID := orderItemIDInt

	err := orderService.DeleteOrderItemByID(orderItemID, orderID)
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
	orderID := orderIDInt

	orderItemIDInt, _ := strconv.Atoi(params["item_id"])
	orderItemID := orderItemIDInt

	res, err := orderService.GetOrderItemByID(orderItemID, orderID)
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
	orderID := orderIDInt

	res, err := orderService.GetOrderItemsByOrderID(orderID)
	if err != nil {
		logger.WithError(err).Error("get_order_items_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_order_items")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func GetTempOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	opts := model.OrderFilterOptions{}
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

	res, err := orderService.GetTempOrdersByShopID(shopID, opts)
	if err != nil {
		logger.WithError(err).Error("get_temp_orders_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_temp_orders")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func validateCreateOrderItem(inp CreateOrderItemRequest) (bool, error) {
	if inp.ProductID <= 0 {
		return false, errors.New("product_id is required")
	}

	if inp.Qty <= 0 {
		return false, errors.New("qty is required")
	}

	return true, nil
}

func validateCreateOrder(inp CreateOrderRequest) (bool, error) {
	if inp.CustomerID <= 0 {
		return false, errors.New("customer_id is required")
	}

	return true, nil
}

func validateOrderID(params map[string]string) (bool, error) {
	if params["order_id"] == "" {
		return false, errors.New("order_id is required")
	}

	return true, nil
}

func validateOrderItemID(params map[string]string) (bool, error) {
	if params["item_id"] == "" {
		return false, errors.New("item_id is required")
	}

	return true, nil
}

// parseDate parses a date string in YYYY-MM-DD format.
func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.UTC)
}
