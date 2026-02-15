package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/service"
)

type (
	CreateShopTempOrderRequest struct {
		CustomerName  string                           `json:"customer_name"`
		CustomerPhone string                           `json:"customer_phone"`
		Items         []CreateShopTempOrderItemRequest `json:"order_items"`
	}

	CreateShopTempOrderItemRequest struct {
		ProductID int `json:"product_id"`
		Qty       int `json:"qty"`
	}
)

// GetShopShareTokenHandler godoc
//
//	@Summary		Get shop share token
//	@Description	Get the authenticated shop's share_token. Used for generating share links.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			shop
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	object{share_token=string}
//	@Failure		404	{object}	ErrorApiResponse	"Shop not found"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/shop/share_token [get]
func GetShopShareTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	token, err := shopService.GetShareTokenByID(shopID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("get_shop_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_shop")
		return
	}

	WriteJson(w, http.StatusOK, map[string]string{"share_token": token})
}

// GetShopProductsHandler godoc
//
//	@Summary		List shop products (public)
//	@Description	Get all products for a shop by its share token. No authentication required. Used for public product catalog share links.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			shop
//	@Produce		json
//	@Param			share_token	path		string	true	"Shop share token"
//	@Success		200			{array}		response.ProductData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (share_token required)"
//	@Failure		404	{object}	ErrorApiResponse	"Shop not found"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/public/shops/{share_token}/products [get]
func GetShopProductsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shareToken := params["share_token"]

	if shareToken == "" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New("share_token is required"), "validation")
		return
	}

	products, err := shopService.GetPublicProducts(shareToken)
	if err != nil {
		if err.Error() == "shop not found" {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("get_shop_products_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_shop_products")
		return
	}

	WriteJson(w, http.StatusOK, products)
}

// CreateShopOrderTempHandler godoc
//
//	@Summary		Create order temp (public)
//	@Description	Create a temporary order for a shop by share token. No authentication required. Used for public share-page checkout.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			shop
//	@Accept			json
//	@Produce		json
//	@Param			share_token	path		string						true	"Shop share token"
//	@Param			body		body		CreateShopOrderTempRequest	true	"Customer name, phone, and order items (product_id, qty)"
//	@Success		200			{object}	response.OrderTempData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (missing share_token, invalid JSON, or validation: customer_name/customer_phone required)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/public/shops/{share_token}/orders [post]
func CreateShopTempOrderHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shareToken := params["share_token"]

	if shareToken == "" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New("share_token is required"), "validation")
		return
	}

	inp := CreateShopTempOrderRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateCreateShopTempOrder(inp); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	items := []service.CreateTempOrderItemInput{}
	for _, item := range inp.Items {
		items = append(items, service.CreateTempOrderItemInput{
			ProductID: item.ProductID,
			Qty:       item.Qty,
		})
	}
	res, err := orderService.CreateTempOrder(inp.CustomerName, inp.CustomerPhone, shareToken, items)
	if err != nil {
		logger.WithError(err).Error("create_shop_temp_order_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "create_shop_temp_order")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func validateCreateShopTempOrder(inp CreateShopTempOrderRequest) (bool, error) {
	if inp.CustomerName == "" {
		return false, errors.New("customer_name is required")
	}

	if inp.CustomerPhone == "" {
		return false, errors.New("customer_phone is required")
	}

	if len(inp.Items) == 0 {
		return false, errors.New("order_items is required")
	}

	for _, item := range inp.Items {
		if item.ProductID <= 0 {
			return false, errors.New("product_id is required")
		}
	}

	for _, item := range inp.Items {
		if item.Qty <= 0 {
			return false, errors.New("qty is required")
		}
	}

	return true, nil
}
