package service

import (
	"errors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	ProductService interface {
		CreateProduct(shopID int, name string, description *string, price int) (response.ProductData, error)
		GetProductByID(productID int, shopID ...int) (*response.ProductData, error)
		GetProductsByShopID(shopID int, searchQuery *string) ([]response.ProductData, error)
		UpdateProduct(input UpdateProductInput) (response.ProductData, error)
		DeleteProductByID(id int) error
	}

	pservice struct{}

	UpdateProductInput struct {
		ID          int
		Name        *string
		Description *string
		Price       *int
	}
)

func NewProductService() ProductService {
	cfg = config.GetConfig()

	if productStore == nil {
		productStore = store.NewProductStore()
	}

	return &pservice{}
}

func (p *pservice) CreateProduct(shopID int, name string, description *string, price int) (response.ProductData, error) {
	product, err := productStore.CreateProduct(name, description, price, shopID)
	if err != nil {
		return response.ProductData{}, err
	}

	res := response.ProductData{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		CreatedAt:   product.CreatedAt,
	}

	return res, nil
}

func (p *pservice) GetProductByID(productID int, shopID ...int) (*response.ProductData, error) {
	product, err := productStore.GetProductByID(productID, shopID...)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, errors.New("product not found")
	}

	res := response.ProductData{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		CreatedAt:   product.CreatedAt,
		Price:       product.Price,
	}

	if product.UpdatedAt.Valid {
		res.UpdatedAt = &product.UpdatedAt.Time
	}

	return &res, nil
}

func (p *pservice) GetProductsByShopID(shopID int, searchQuery *string) ([]response.ProductData, error) {
	products, err := productStore.GetProductsByShopID(shopID, searchQuery)
	if err != nil {
		return []response.ProductData{}, err
	}

	var productsData []response.ProductData
	for _, product := range products {
		res := response.ProductData{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			CreatedAt:   product.CreatedAt,
			Price:       product.Price,
		}

		if product.UpdatedAt.Valid {
			t := product.UpdatedAt.Time
			res.UpdatedAt = &t
		}

		productsData = append(productsData, res)
	}

	return productsData, nil
}

func (p *pservice) UpdateProduct(input UpdateProductInput) (response.ProductData, error) {
	updateData := store.UpdateProductInput{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
	}
	productData, err := productStore.UpdateProduct(input.ID, updateData)
	if err != nil {
		return response.ProductData{}, err
	}

	if productData == nil {
		return response.ProductData{}, errors.New("product not found")
	}

	res := response.ProductData{
		ID:          productData.ID,
		Name:        productData.Name,
		Description: productData.Description,
		Price:       productData.Price,
		CreatedAt:   productData.CreatedAt,
	}

	if productData.UpdatedAt.Valid {
		res.UpdatedAt = &productData.UpdatedAt.Time
	}

	return res, nil
}

func (p *pservice) DeleteProductByID(id int) error {
	err := productStore.DeleteProductByID(id)
	if err != nil {
		return err
	}

	return nil
}
