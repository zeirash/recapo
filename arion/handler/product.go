package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/service"
)

type (
	CreateProductRequest struct {
		Name          string  `json:"name"`
		Price         int     `json:"price"`
		Description   *string `json:"description"`
		OriginalPrice *int    `json:"original_price"`
	}

	UpdateProductRequest struct {
		Name          *string `json:"name"`
		Price         *int    `json:"price"`
		Description   *string `json:"description"`
		OriginalPrice *int    `json:"original_price"`
	}
)

// CreateProductHandler godoc
//
//	@Summary		Create product
//	@Description	Create a new product for the shop.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			product
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateProductRequest	true	"Product data"
//	@Success		200		{object}	response.ProductData
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON or validation)"
//	@Failure		500		{object}	ErrorApiResponse	"Internal server error"
//	@Router			/product [post]
func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	inp := CreateProductRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateCreateProduct(inp); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := productService.CreateProduct(shopID, inp.Name, inp.Description, inp.Price, inp.OriginalPrice)
	if err != nil {
		logger.WithError(err).Error("create_product_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "create_product")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetProductHandler godoc
//
//	@Summary		Get product by ID
//	@Description	Get a single product by ID.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			product
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			product_id	path		int	true	"Product ID"
//	@Success		200			{object}	response.ProductData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid product_id)"
//	@Failure		404	{object}	ErrorApiResponse	"Product not found"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/products/{product_id} [get]
func GetProductHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	params := mux.Vars(r)

	if valid, err := validateProductID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	productIDInt, _ := strconv.Atoi(params["product_id"])
	productID := productIDInt

	res, err := productService.GetProductByID(productID, shopID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("get_product_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_product")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetProductsHandler godoc
//
//	@Summary		List products
//	@Description	Get all products for the shop. Optional search query to filter by name or description.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			product
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search	query		string	false	"Search query"
//	@Success		200		{array}		response.ProductData
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/products [get]
func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	var searchQuery *string
	if q := r.URL.Query().Get("search"); q != "" {
		searchQuery = &q
	}

	res, err := productService.GetProductsByShopID(shopID, searchQuery)
	if err != nil {
		logger.WithError(err).Error("get_products_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_products")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// UpdateProductHandler godoc
//
//	@Summary		Update product
//	@Description	Update a product by ID. Only provided fields are updated (partial update).
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			product
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			product_id	path		int						true	"Product ID"
//	@Param			body		body		UpdateProductRequest	true	"Fields to update (name, description, price)"
//	@Success		200			{object}	response.ProductData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid product_id or JSON)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/products/{product_id} [patch]
func UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateProductID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	productIDInt, _ := strconv.Atoi(params["product_id"])
	productID := productIDInt

	inp := UpdateProductRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := productService.UpdateProduct(service.UpdateProductInput{
		ID:            productID,
		Name:          inp.Name,
		Description:   inp.Description,
		Price:         inp.Price,
		OriginalPrice: inp.OriginalPrice,
	})
	if err != nil {
		logger.WithError(err).Error("update_product_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "update_product")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// DeleteProductHandler godoc
//
//	@Summary		Delete product
//	@Description	Delete a product by ID.
//	@Description	Success Response envelope: { success, data, code, message }. data contains "OK" on success.
//	@Tags			product
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			product_id	path	int	true	"Product ID"
//	@Success		200			{string}	string	"Success. data contains \"OK\""
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid product_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/products/{product_id} [delete]
func DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateProductID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	productIDInt, _ := strconv.Atoi(params["product_id"])
	productID := productIDInt

	err := productService.DeleteProductByID(productID)
	if err != nil {
		logger.WithError(err).Error("delete_product_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "delete_product")
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
