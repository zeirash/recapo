package service

import (
	"errors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	CustomerService interface {
		CreateCustomer(name, phone, address string, shopID int) (response.CustomerData, error)
		GetCustomerByID(customerID int, shopID ...int) (*response.CustomerData, error)
		GetCustomersByShopID(shopID int, searchQuery *string) ([]response.CustomerData, error)
		UpdateCustomer(input UpdateCustomerInput) (response.CustomerData, error)
		DeleteCustomerByID(id int) error
		HasActiveOrders(customerID int, shopID int) (response.CustomerHasActiveOrdersData, error)
		CheckActiveOrderByPhone(phone, name string, shopID int) (response.CustomerCheckActiveOrderByPhone, error)
	}

	cservice struct{}

	UpdateCustomerInput struct {
		ID      int
		Name    *string
		Phone   *string
		Address *string
	}
)

func NewCustomerService() CustomerService {
	cfg = config.GetConfig()

	if customerStore == nil {
		customerStore = store.NewCustomerStore()
	}

	return &cservice{}
}

func (c *cservice) CreateCustomer(name, phone, address string, shopID int) (response.CustomerData, error) {
	customer, err := customerStore.CreateCustomer(store.CreateCustomerInput{
		Name:    name,
		Phone:   phone,
		Address: &address,
		ShopID:  shopID,
	})
	if err != nil {
		return response.CustomerData{}, err
	}

	res := response.CustomerData{
		ID:        customer.ID,
		Name:      customer.Name,
		Phone:     customer.Phone,
		Address:   customer.Address,
		CreatedAt: customer.CreatedAt,
	}

	return res, nil
}

func (c *cservice) GetCustomerByID(customerID int, shopID ...int) (*response.CustomerData, error) {
	customer, err := customerStore.GetCustomerByID(customerID, shopID...)
	if err != nil {
		return nil, err
	}

	if customer == nil {
		return nil, errors.New("customer not found")
	}

	res := response.CustomerData{
		ID:        customer.ID,
		Name:      customer.Name,
		Phone:     customer.Phone,
		Address:   customer.Address,
		CreatedAt: customer.CreatedAt,
	}

	if customer.UpdatedAt.Valid {
		res.UpdatedAt = &customer.UpdatedAt.Time
	}

	return &res, nil
}

func (c *cservice) GetCustomersByShopID(shopID int, searchQuery *string) ([]response.CustomerData, error) {
	customers, err := customerStore.GetCustomersByShopID(shopID, searchQuery)
	if err != nil {
		return []response.CustomerData{}, err
	}

	customersData := []response.CustomerData{}
	for _, customer := range customers {
		res := response.CustomerData{
			ID:        customer.ID,
			Name:      customer.Name,
			Phone:     customer.Phone,
			Address:   customer.Address,
			CreatedAt: customer.CreatedAt,
		}

		if customer.UpdatedAt.Valid {
			t := customer.UpdatedAt.Time
			res.UpdatedAt = &t
		}

		customersData = append(customersData, res)
	}

	return customersData, nil
}

func (c *cservice) UpdateCustomer(input UpdateCustomerInput) (response.CustomerData, error) {
	customer, err := customerStore.GetCustomerByID(input.ID)
	if err != nil {
		return response.CustomerData{}, err
	}

	if customer == nil {
		return response.CustomerData{}, errors.New("customer not found")
	}

	updateData := store.UpdateCustomerInput{
		Name:    input.Name,
		Phone:   input.Phone,
		Address: input.Address,
	}

	customerData, err := customerStore.UpdateCustomer(input.ID, updateData)
	if err != nil {
		return response.CustomerData{}, err
	}

	res := response.CustomerData{
		ID:        customerData.ID,
		Name:      customerData.Name,
		Phone:     customerData.Phone,
		Address:   customerData.Address,
		CreatedAt: customerData.CreatedAt,
	}

	if customerData.UpdatedAt.Valid {
		res.UpdatedAt = &customerData.UpdatedAt.Time
	}

	return res, nil
}

func (c *cservice) DeleteCustomerByID(id int) error {
	err := customerStore.DeleteCustomerByID(id)
	if err != nil {
		return err
	}

	return nil
}

func (c *cservice) HasActiveOrders(customerID int, shopID int) (response.CustomerHasActiveOrdersData, error) {
	hasActiveOrders, err := orderStore.HasActiveOrdersByCustomerID(customerID, shopID)
	if err != nil {
		return response.CustomerHasActiveOrdersData{}, err
	}
	return response.CustomerHasActiveOrdersData{HasActiveOrders: hasActiveOrders}, nil
}

func (c *cservice) CheckActiveOrderByPhone(phone, name string, shopID int) (response.CustomerCheckActiveOrderByPhone, error) {
	customer, err := customerStore.GetCustomerByPhone(phone, shopID)
	if err != nil {
		return response.CustomerCheckActiveOrderByPhone{}, err
	}

	if customer == nil {
		customer, err = customerStore.CreateCustomer(store.CreateCustomerInput{
			Name:    name,
			Phone:   phone,
			ShopID:  shopID,
		})
		if err != nil {
			return response.CustomerCheckActiveOrderByPhone{}, err
		}
	}

	hasActiveOrders, err := orderStore.HasActiveOrdersByCustomerID(customer.ID, shopID)
	if err != nil {
		return response.CustomerCheckActiveOrderByPhone{}, err
	}

	return response.CustomerCheckActiveOrderByPhone{
		CustomerID:      customer.ID,
		HasActiveOrders: hasActiveOrders,
	}, nil
}
