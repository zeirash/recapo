package service

import (
	"errors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	CustomerService interface {
		CreateCustomer(name, email, password string, shopID int) (response.CustomerData, error)
		// GetCustomerByID(customerID int) (*response.CustomerData, error)
		GetCustomersByShopID(shopID int) ([]response.CustomerData, error)
		UpdateCustomer(input UpdateCustomerInput) (response.CustomerData, error)
		DeleteCustomerByID(id int) error
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
	//TODO: validate customer unique phone

	customer, err := customerStore.CreateCustomer(name, phone, address, shopID)
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

	if customer.UpdatedAt.Valid {
		res.UpdatedAt = &customer.UpdatedAt.Time
	}

	return res, nil
}

func (c *cservice) GetCustomersByShopID(shopID int) ([]response.CustomerData, error) {
	customers, err := customerStore.GetCustomersByShopID(shopID)
	if err != nil {
		return []response.CustomerData{}, err
	}

	var customersData []response.CustomerData
	for _, customer := range customers {
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

	//TODO: validate customer unique phone
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
