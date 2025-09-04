package service

import (
	"errors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	ProductService interface {
		CreateProduct(shopID int, name string) (response.ProductData, error)
		// GetProductByID(id int) (*response.ProductData, error)
		GetProductsByShopID(shopID int) ([]response.ProductData, error)
		UpdateProduct(id int, name string) (response.ProductData, error)
		DeleteProductByID(id int) error
	}

	pservice struct {}
)

func NewProductService() ProductService {
	cfg = config.GetConfig()

	if productStore == nil {
		productStore = store.NewProductStore()
	}

	return &pservice{}
}

func (p *pservice) CreateProduct(shopID int, name string) (response.ProductData, error) {
	//TODO: validate product unique name

	product, err := productStore.CreateProduct(name, shopID)
	if err != nil {
		return response.ProductData{}, err
	}

	res := response.ProductData{
		ID:        product.ID,
		Name:      product.Name,
		CreatedAt: product.CreatedAt,
	}

	if product.UpdatedAt.Valid {
		res.UpdatedAt = &product.UpdatedAt.Time
	}

	return res, nil
}

func (p *pservice) GetProductsByShopID(shopID int) ([]response.ProductData, error) {
	products, err := productStore.GetProductsByShopID(shopID)
	if err != nil {
		return []response.ProductData{}, err
	}

	var productsData []response.ProductData
	for _, product := range products {
		res := response.ProductData{
			ID:        product.ID,
			Name:      product.Name,
			CreatedAt: product.CreatedAt,
		}

		if product.UpdatedAt.Valid {
			res.UpdatedAt = &product.UpdatedAt.Time
		}

		productsData = append(productsData, res)
	}

	return productsData, nil
}

func (p *pservice) UpdateProduct(id int, name string) (response.ProductData, error) {
	//TODO: validate product unique name

	product, err := productStore.GetProductByID(id)
	if err != nil {
		return response.ProductData{}, err
	}

	if product == nil {
		return response.ProductData{}, errors.New("product not found")
	}

	productData, err := productStore.UpdateProduct(id, name)
	if err != nil {
		return response.ProductData{}, err
	}

	res := response.ProductData{
		ID:        productData.ID,
		Name:      productData.Name,
		CreatedAt: productData.CreatedAt,
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
