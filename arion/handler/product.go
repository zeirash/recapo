package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
)

type (
	ProductRequest struct {
		Name string `json:"name"`
	}
)

func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	inp := ProductRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := productService.CreateProduct(shopID, inp.Name)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "create_product")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	res, err := productService.GetProductsByShopID(shopID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_products")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateUpdateDeleteProduct(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	productIDInt, _ := strconv.Atoi(params["product_id"])
	productID := productIDInt

	inp := ProductRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := productService.UpdateProduct(productID, inp.Name)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "update_product")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateUpdateDeleteProduct(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	productIDInt, _ := strconv.Atoi(params["product_id"])
	productID := productIDInt

	err := productService.DeleteProductByID(productID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "delete_product")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

func validateUpdateDeleteProduct(params map[string]string) (bool, error) {
	if params["product_id"] == "" {
		return false, errors.New("product_id is required")
	}

	return true, nil
}
