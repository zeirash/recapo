package handler

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/logger"
)

// GetShopShareTokenHandler godoc
//
//	@Summary		Get shop share token
//	@Description	Get the authenticated shop's share_token. Used for generating share links.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			shop
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.ShareTokenData
//	@Failure		404	{object}	ErrorApiResponse	"Shop not found"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/shop/share_token [get]
func GetShopShareTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	// TODO: generalized other handler error handling (follow this pattern)
	token, err := shopService.GetShareTokenByID(shopID)
	if err != nil {
		if err.Error() == "shop not found" {
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
