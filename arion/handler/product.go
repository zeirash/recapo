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
	CreateProductRequest struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
	}

	UpdateProductRequest struct {
		Name  *string `json:"name"`
		Price *int    `json:"price"`
	}
)

func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	inp := CreateProductRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateCreateProduct(inp); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := productService.CreateProduct(shopID, inp.Name, inp.Price)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "create_product")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func GetProductHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	params := mux.Vars(r)

	if valid, err := validateProductID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	productIDInt, _ := strconv.Atoi(params["product_id"])
	productID := productIDInt

	res, err := productService.GetProductByID(productID, shopID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_product")
		return
	}

	if res == nil {
		WriteErrorJson(w, http.StatusNotFound, err, "get_product")
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
	if valid, err := validateProductID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	productIDInt, _ := strconv.Atoi(params["product_id"])
	productID := productIDInt

	inp := UpdateProductRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := productService.UpdateProduct(service.UpdateProductInput{
		ID:    productID,
		Name:  inp.Name,
		Price: inp.Price,
	})
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "update_product")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateProductID(params); !valid {
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

func validateCreateProduct(inp CreateProductRequest) (bool, error) {
	if inp.Name == "" {
		return false, errors.New("name is required")
	}

	if inp.Price < 0 {
		return false, errors.New("price must be greater than or equal to 0")
	}

	return true, nil
}

func validateProductID(params map[string]string) (bool, error) {
	if params["product_id"] == "" {
		return false, errors.New("product_id is required")
	}

	return true, nil
}

func validateUpdateDeleteProductPrice(params map[string]string) (bool, error) {
	if params["product_id"] == "" {
		return false, errors.New("product_id is required")
	}

	if params["price_id"] == "" {
		return false, errors.New("price_id is required")
	}

	return true, nil
}
