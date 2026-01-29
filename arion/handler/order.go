package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/service"
)

type (
	CreateOrderRequest struct {
		CustomerID int `json:"customer_id"`
	}

	UpdateOrderRequest struct {
		CustomerID *int `json:"customer_id"`
		TotalPrice *int `json:"total_price"`
		Status     *string `json:"status"`
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

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	inp := CreateOrderRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateCreateOrder(inp); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := orderService.CreateOrder(inp.CustomerID, shopID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "create_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderID := orderIDInt

	res, err := orderService.GetOrderByID(orderID, shopID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_order")
		return
	}

	if res == nil {
		WriteErrorJson(w, http.StatusNotFound, err, "get_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	res, err := orderService.GetOrdersByShopID(shopID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_orders")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func UpdateOrderHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderID := orderIDInt

	inp := UpdateOrderRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := orderService.UpdateOrderByID(service.UpdateOrderInput{
		ID:         orderID,
		TotalPrice: inp.TotalPrice,
		Status:     inp.Status,
	})
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "update_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderID := orderIDInt

	err := orderService.DeleteOrderByID(orderID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "delete_order")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

func CreateOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	inp := CreateOrderItemRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateCreateOrderItem(inp); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderID := orderIDInt

	res, err := orderService.CreateOrderItem(orderID, inp.ProductID, inp.Qty)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "create_order_item")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func UpdateOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	if valid, err := validateOrderItemID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	inp := UpdateOrderItemRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
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
		WriteErrorJson(w, http.StatusInternalServerError, err, "update_order_item")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func DeleteOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	if valid, err := validateOrderItemID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderID := orderIDInt

	orderItemIDInt, _ := strconv.Atoi(params["item_id"])
	orderItemID := orderItemIDInt

	err := orderService.DeleteOrderItemByID(orderItemID, orderID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "delete_order_item")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

func GetOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	if valid, err := validateOrderItemID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderID := orderIDInt

	orderItemIDInt, _ := strconv.Atoi(params["item_id"])
	orderItemID := orderItemIDInt

	res, err := orderService.GetOrderItemByID(orderItemID, orderID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_order_item")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func GetOrderItemsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if valid, err := validateOrderID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	orderIDInt, _ := strconv.Atoi(params["order_id"])
	orderID := orderIDInt

	res, err := orderService.GetOrderItemsByOrderID(orderID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_order_items")
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
