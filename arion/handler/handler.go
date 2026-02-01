package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/service"
)

type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
}

var (
	userService     service.UserService
	customerService service.CustomerService
	productService  service.ProductService
	orderService    service.OrderService
)

func Init() {
	if userService == nil {
		userService = service.NewUserService()
	}

	if customerService == nil {
		customerService = service.NewCustomerService()
	}

	if productService == nil {
		productService = service.NewProductService()
	}

	if orderService == nil {
		orderService = service.NewOrderService()
	}
}

// SetProductService sets the product service (for testing)
func SetProductService(s service.ProductService) {
	productService = s
}

// SetCustomerService sets the customer service (for testing)
func SetCustomerService(s service.CustomerService) {
	customerService = s
}

// SetUserService sets the user service (for testing)
func SetUserService(s service.UserService) {
	userService = s
}

// SetOrderService sets the order service (for testing)
func SetOrderService(s service.OrderService) {
	orderService = s
}

func WriteJson(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	res := ApiResponse{
		Success: true,
		Data:    body,
	}

	jsonResp, err := json.Marshal(res)
	if err != nil {
		logger.WithError(err).Error("error marshall body")
		return
	}
	w.Write(jsonResp)
}

func WriteErrorJson(w http.ResponseWriter, status int, err error, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	res := ApiResponse{
		Success: false,
		Code:    code,
		Message: err.Error(),
	}

	jsonResp, err := json.Marshal(res)
	if err != nil {
		logger.WithError(err).Error("error marshall body")
		return
	}
	w.Write(jsonResp)
}

func ParseJson(input io.ReadCloser, result interface{}) error {
	err := json.NewDecoder(input).Decode(result)
	return errors.Wrap(err, "Failed parsing json")
}
