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

// TODO: fix error message translation
// ErrorApiResponse is the shape of error responses for Swagger. Data is an
// empty object {} rather than a string or null.
type ErrorApiResponse struct {
	Success bool     `json:"success"`
	Data    struct{} `json:"data"`
	Code    string   `json:"code"`
	Message string   `json:"message"`
}

var (
	userService     service.UserService
	customerService service.CustomerService
	productService  service.ProductService
	orderService    service.OrderService
	shopService     service.ShopService
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

	if shopService == nil {
		shopService = service.NewShopService()
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

// SetShopService sets the shop service (for testing)
func SetShopService(s service.ShopService) {
	shopService = s
}

// GetUserService returns the current user service (for testing).
func GetUserService() service.UserService {
	return userService
}

// GetCustomerService returns the current customer service (for testing).
func GetCustomerService() service.CustomerService {
	return customerService
}

// GetProductService returns the current product service (for testing).
func GetProductService() service.ProductService {
	return productService
}

// GetOrderService returns the current order service (for testing).
func GetOrderService() service.OrderService {
	return orderService
}

// GetShopService returns the current shop service (for testing).
func GetShopService() service.ShopService {
	return shopService
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

// WriteErrorJson writes a JSON error response. Message is translated using
// Accept-Language when r is non-nil, except for code "validation" which
// always uses err.Error() to preserve field-level validation details.
func WriteErrorJson(w http.ResponseWriter, r *http.Request, status int, err error, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	res := ErrorApiResponse{
		Success: false,
		Data:    struct{}{},
		Code:    code,
		Message: err.Error(),
	}

	jsonResp, errMarshal := json.Marshal(res)
	if errMarshal != nil {
		logger.WithError(errMarshal).Error("error marshall body")
		return
	}
	w.Write(jsonResp)
}

func ParseJson(input io.ReadCloser, result interface{}) error {
	err := json.NewDecoder(input).Decode(result)
	return errors.Wrap(err, "Failed parsing json")
}
