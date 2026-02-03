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

func CreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	inp := CreateCustomerRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	if valid, err := validateCreateCustomer(inp); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	res, err := customerService.CreateCustomer(inp.Name, inp.Phone, inp.Address, shopID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "create_customer")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func GetCustomerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)
	params := mux.Vars(r)

	if valid, err := validateCustomerID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	customerIDInt, _ := strconv.Atoi(params["customer_id"])
	customerID := customerIDInt

	res, err := customerService.GetCustomerByID(customerID, shopID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_customer")
		return
	}

	if res == nil {
		WriteErrorJson(w, http.StatusNotFound, errors.New("customer not found"), "get_customer")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func GetCustomersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	var searchQuery *string
	if q := r.URL.Query().Get("search"); q != "" {
		searchQuery = &q
	}

	res, err := customerService.GetCustomersByShopID(int(shopID), searchQuery)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "get_customers")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func UpdateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateCustomerID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	customerIDInt, _ := strconv.Atoi(params["customer_id"])
	customerID := customerIDInt

	inp := UpdateCustomerRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, http.StatusBadRequest, err, "parse_json")
		return
	}

	res, err := customerService.UpdateCustomer(service.UpdateCustomerInput{
		ID:      customerID,
		Name:    inp.Name,
		Phone:   inp.Phone,
		Address: inp.Address,
	})
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "update_customer")
		return
	}

	WriteJson(w, http.StatusOK, res)
}

func DeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if valid, err := validateCustomerID(params); !valid {
		WriteErrorJson(w, http.StatusBadRequest, err, "validation")
		return
	}

	customerIDInt, _ := strconv.Atoi(params["customer_id"])
	customerID := customerIDInt

	err := customerService.DeleteCustomerByID(customerID)
	if err != nil {
		WriteErrorJson(w, http.StatusInternalServerError, err, "delete_customer")
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
