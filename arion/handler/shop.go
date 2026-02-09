package handler

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common/logger"
)

// GetShopProductsHandler returns products for a shop by its share token.
// No authentication required. Used for public product catalog share links.
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
			WriteErrorJson(w, r, http.StatusNotFound, err, "shop_not_found")
			return
		}
		logger.WithError(err).Error("get_shop_products_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_shop_products")
		return
	}

	WriteJson(w, http.StatusOK, products)
}
