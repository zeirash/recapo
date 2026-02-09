package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/service"
)

type (
	CreateCustomerRequest struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}

	UpdateCustomerRequest struct {
		Name    *string `json:"name"`
		Phone   *string `json:"phone"`
		Address *string `json:"address"`
	}
)

// CreateCustomerHandler godoc
//
//	@Summary		Create customer
//	@Description	Create a new customer for the shop.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			customer
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateCustomerRequest	true	"Customer data"
//	@Success		200		{object}	response.CustomerData
//	@Failure		400		{object}	ErrorApiResponse	"Bad request (invalid JSON or validation)"
//	@Failure		500		{object}	ErrorApiResponse	"Internal server error"
//	@Router			/customer [post]
func CreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	inp := CreateCustomerRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateCreateCustomer(inp); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := customerService.CreateCustomer(inp.Name, inp.Phone, inp.Address, shopID)
	if err != nil {
		logger.WithError(err).Error("create_customer_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "create_customer")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetCustomerHandler godoc
//
//	@Summary		Get customer by ID
//	@Description	Get a single customer by ID.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			customer
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			customer_id	path		int	true	"Customer ID"
//	@Success		200			{object}	response.CustomerData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid customer_id)"
//	@Failure		404	{object}	ErrorApiResponse	"Customer not found"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/customers/{customer_id} [get]
func GetCustomerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	params := mux.Vars(r)

	if valid, err := validateCustomerID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	customerIDInt, _ := strconv.Atoi(params["customer_id"])
	customerID := customerIDInt

	res, err := customerService.GetCustomerByID(customerID, shopID)
	if err != nil {
		logger.WithError(err).Error("get_customer_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_customer")
		return
	}

	if res == nil {
		WriteErrorJson(w, r, http.StatusNotFound, errors.New("customer not found"), "not_found")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// GetCustomersHandler godoc
//
//	@Summary		List customers
//	@Description	Get all customers for the shop. Optional search query to filter by name, phone, or address.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			customer
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search	query		string	false	"Search query"
//	@Success		200		{array}		response.CustomerData
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/customers [get]
func GetCustomersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	var searchQuery *string
	if q := r.URL.Query().Get("search"); q != "" {
		searchQuery = &q
	}

	res, err := customerService.GetCustomersByShopID(int(shopID), searchQuery)
	if err != nil {
		logger.WithError(err).Error("get_customers_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_customers")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// UpdateCustomerHandler godoc
//
//	@Summary		Update customer
//	@Description	Update an existing customer. Only provided fields are updated.
//	@Description	Success Response envelope: { success, data, code, message }. Schema below shows the data field (inner payload).
//	@Tags			customer
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			customer_id	path		int						true	"Customer ID"
//	@Param			body		body		UpdateCustomerRequest	true	"Fields to update"
//	@Success		200			{object}	response.CustomerData
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid JSON or customer_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/customers/{customer_id} [patch]
func UpdateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateCustomerID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	customerIDInt, _ := strconv.Atoi(params["customer_id"])
	customerID := customerIDInt

	inp := UpdateCustomerRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := customerService.UpdateCustomer(service.UpdateCustomerInput{
		ID:      customerID,
		Name:    inp.Name,
		Phone:   inp.Phone,
		Address: inp.Address,
	})
	if err != nil {
		logger.WithError(err).Error("update_customer_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "update_customer")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

// DeleteCustomerHandler godoc
//
//	@Summary		Delete customer
//	@Description	Delete a customer by ID.
//	@Description	Success Response envelope: { success, data, code, message }. data contains "OK" on success.
//	@Tags			customer
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			customer_id	path	int	true	"Customer ID"
//	@Success		200			{string}	string	"Success. data contains \"OK\""
//	@Failure		400	{object}	ErrorApiResponse	"Bad request (invalid customer_id)"
//	@Failure		500	{object}	ErrorApiResponse	"Internal server error"
//	@Router			/customers/{customer_id} [delete]
func DeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateCustomerID(params); !valid {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "validation")
		return
	}

	customerIDInt, _ := strconv.Atoi(params["customer_id"])
	customerID := customerIDInt

	err := customerService.DeleteCustomerByID(customerID)
	if err != nil {
		logger.WithError(err).Error("delete_customer_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "delete_customer")
		return
	}

	WriteJson(w, http.StatusOK, "OK")
}

func validateCreateCustomer(inp CreateCustomerRequest) (bool, error) {
	if inp.Name == "" {
		return false, errors.New("name is required")
	}

	if inp.Phone == "" {
		return false, errors.New("phone is required")
	}

	if inp.Address == "" {
		return false, errors.New("address is required")
	}

	return true, nil
}

func validateCustomerID(params map[string]string) (bool, error) {
	if params["customer_id"] == "" {
		return false, errors.New("customer_id is required")
	}

	return true, nil
}
