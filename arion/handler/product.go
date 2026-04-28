package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/service"
)

type (
	CreateProductRequest struct {
		Name          string  `json:"name"`
		Price         int     `json:"price"`
		Description   *string `json:"description"`
		OriginalPrice *int    `json:"original_price"`
		ImageURL      *string `json:"image_url"`
	}

	UpdateProductRequest struct {
		Name          *string `json:"name"`
		Price         *int    `json:"price"`
		Description   *string `json:"description"`
		OriginalPrice *int    `json:"original_price"`
		ImageURL      *string `json:"image_url"`
		IsActive      *bool   `json:"is_active"`
	}

	DeleteProductImageRequest struct {
		ImageURL string `json:"image_url"`
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

	res, err := productService.CreateProduct(ctx, shopID, inp.Name, inp.Description, inp.Price, inp.OriginalPrice, inp.ImageURL)
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

	res, err := productService.GetProductByID(ctx, productID, shopID)
	if err != nil {
		if err.Error() == apierr.ErrProductNotFound {
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
//	@Param			sort  	query		string	false	"Sort by column and order (e.g. name,desc)"
//	@Success		200		{array}		response.ProductData
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/products [get]
func GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	filter := model.FilterOptions{}
	if q := r.URL.Query().Get("search"); q != "" {
		filter.SearchQuery = &q
	}
	if sort := r.URL.Query().Get("sort"); sort != "" {
		filter.Sort = &sort
	}
	if isActive := r.URL.Query().Get("is_active"); isActive != "" {
		v := isActive == "true"
		filter.IsActive = &v
	}

	res, err := productService.GetProductsByShopID(ctx, shopID, filter)
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
	ctx := r.Context()
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

	res, err := productService.UpdateProduct(ctx, service.UpdateProductInput{
		ID:            productID,
		Name:          inp.Name,
		Description:   inp.Description,
		Price:         inp.Price,
		OriginalPrice: inp.OriginalPrice,
		ImageURL:      inp.ImageURL,
		IsActive:      inp.IsActive,
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
	ctx := r.Context()
	params := mux.Vars(r)
	if valid, err := validateProductID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	productIDInt, _ := strconv.Atoi(params["product_id"])
	productID := productIDInt

	err := productService.DeleteProductByID(ctx, productID)
	if err != nil {
		logger.WithError(err).Error("delete_product_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "delete_product")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

// PurchaseListProductHandler godoc
//
//	@Summary		List purchase list products
//	@Description	Get all products that needs to be purchased (for the shop that have active orders).
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			product
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{array}	response.PurchaseListProductData
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/products/purchase_list [get]
func PurchaseListProductHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	res, err := productService.GetPurchaseListProducts(ctx, shopID)
	if err != nil {
		logger.WithError(err).Error("get_purchase_list_products_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_purchase_list_products")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// UploadProductImageHandler godoc
//
//	@Summary		Upload product image
//	@Description	Upload an image file (jpeg, png, webp, max 5MB). Returns the image_url to include in create/update product requests.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			product
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			image	formData	file	true	"Image file (jpeg/png/webp, max 5MB)"
//	@Success		200		{object}	response.UploadImageData
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (missing file, invalid type)"
//	@Failure		500		{object}	ErrorApiResponse	"Internal server error"
//	@Router			/products/image [post]
func UploadProductImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrImageTooLarge), "validation")
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrImageFieldRequired), "validation")
		return
	}
	defer file.Close()

	imageURL, err := productService.UploadProductImage(ctx, file)
	if err != nil {
		if err.Error() == apierr.ErrUnsupportedImageType {
			WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
			return
		}
		logger.WithError(err).Error("upload_product_image_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "upload_product_image")
		return
	}

	WriteJson(w, http.StatusOK, response.UploadImageData{ImageURL: imageURL})
}

// DeleteProductImageHandler godoc
//
//	@Summary		Delete product image
//	@Description	Delete an uploaded product image by its URL. Call this if product creation fails after a successful image upload.
//	@Description	Success Response envelope: { success, data, code, message }. data contains "OK" on success.
//	@Tags			product
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		DeleteProductImageRequest	true	"Image URL to delete"
//	@Success		200		{string}	string						"Success. data contains \"OK\""
//	@Failure		400		{object}	ErrorApiResponse			"Bad request (missing or invalid image_url)"
//	@Failure		404		{object}	ErrorApiResponse			"Image not found"
//	@Failure		500		{object}	ErrorApiResponse			"Internal server error"
//	@Router			/products/image [delete]
func DeleteProductImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	inp := DeleteProductImageRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if inp.ImageURL == "" {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrImageURLRequired), "validation")
		return
	}

	if err := productService.DeleteProductImage(ctx, inp.ImageURL); err != nil {
		if err.Error() == apierr.ErrImageNotFound {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
			return
		}
		logger.WithError(err).Error("delete_product_image_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "delete_product_image")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

func validateCreateProduct(inp CreateProductRequest) (bool, error) {
	if inp.Name == "" {
		return false, errors.New(apierr.ErrNameRequired)
	}

	if inp.Price < 0 {
		return false, errors.New(apierr.ErrPriceInvalid)
	}

	return true, nil
}

func validateProductID(params map[string]string) (bool, error) {
	if params["product_id"] == "" {
		return false, errors.New(apierr.ErrProductIDRequired)
	}

	return true, nil
}
