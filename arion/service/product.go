package service

import (
	"errors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	ProductService interface {
		CreateProduct(shopID int, name string, price int) (response.ProductData, error)
		// GetProductByID(id int) (*response.ProductData, error)
		GetProductsByShopID(shopID int) ([]response.ProductData, error)
		UpdateProduct(productID int, name string) (response.ProductData, error)
		DeleteProductByID(productID int) error
		CreateProductPrice(productID int, price int) (response.PriceData, error)
		UpdateProductPrice(productID, priceID, price int) (response.PriceData, error)
		DeleteProductPrice(productID, priceID int) error
	}

	pservice struct{}
)

func NewProductService() ProductService {
	cfg = config.GetConfig()

	if productStore == nil {
		productStore = store.NewProductStore()
	}

	return &pservice{}
}

func (p *pservice) CreateProduct(shopID int, name string, price int) (response.ProductData, error) {
	//TODO: validate product unique name

	product, err := productStore.CreateProduct(name, shopID)
	if err != nil {
		return response.ProductData{}, err
	}

	priceData, err := productStore.CreatePrice(product.ID, price)
	if err != nil {
		return response.ProductData{}, err
	}

	res := response.ProductData{
		ID:   product.ID,
		Name: product.Name,
		Prices: []response.PriceData{
			{
				ID:        priceData.ID,
				Price:     priceData.Price,
				CreatedAt: priceData.CreatedAt,
			},
		},
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

	// TODO: improve query performance
	var productsData []response.ProductData
	for _, product := range products {
		prices, err := productStore.GetPricesByProductID(product.ID)
		if err != nil {
			return []response.ProductData{}, err
		}

		var pricesData []response.PriceData
		for _, price := range prices {
			priceRes := response.PriceData{
				ID:        price.ID,
				Price:     price.Price,
				CreatedAt: price.CreatedAt,
			}

			if price.UpdatedAt.Valid {
				priceRes.UpdatedAt = &price.UpdatedAt.Time
			}

			pricesData = append(pricesData, priceRes)
		}

		res := response.ProductData{
			ID:        product.ID,
			Name:      product.Name,
			CreatedAt: product.CreatedAt,
			Prices:    pricesData,
		}

		if product.UpdatedAt.Valid {
			res.UpdatedAt = &product.UpdatedAt.Time
		}

		productsData = append(productsData, res)
	}

	return productsData, nil
}

func (p *pservice) UpdateProduct(productID int, name string) (response.ProductData, error) {
	//TODO: validate product unique name

	productData, err := productStore.UpdateProduct(productID, name)
	if err != nil {
		return response.ProductData{}, err
	}

	if productData == nil {
		return response.ProductData{}, errors.New("product not found")
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

func (p *pservice) DeleteProductByID(productID int) error {
	err := productStore.DeleteProductByID(productID)
	if err != nil {
		return err
	}

	err = productStore.DeletePricesByProductID(productID)
	if err != nil {
		return err
	}

	return nil
}

func (p *pservice) CreateProductPrice(productID int, price int) (response.PriceData, error) {
	product, err := productStore.GetProductByID(productID)
	if err != nil {
		return response.PriceData{}, err
	}

	if product == nil {
		return response.PriceData{}, errors.New("product not found")
	}

	priceData, err := productStore.CreatePrice(productID, price)
	if err != nil {
		return response.PriceData{}, err
	}

	if priceData == nil {
		return response.PriceData{}, errors.New("price not found")
	}

	res := response.PriceData{
		ID:        priceData.ID,
		Price:     priceData.Price,
		CreatedAt: priceData.CreatedAt,
	}

	if priceData.UpdatedAt.Valid {
		res.UpdatedAt = &priceData.UpdatedAt.Time
	}

	return res, nil
}

func (p *pservice) UpdateProductPrice(productID, priceID, price int) (response.PriceData, error) {
	priceData, err := productStore.UpdatePrice(productID, priceID, price)
	if err != nil {
		return response.PriceData{}, err
	}

	if priceData == nil {
		return response.PriceData{}, errors.New("price not found")
	}

	res := response.PriceData{
		ID:        priceData.ID,
		Price:     priceData.Price,
		CreatedAt: priceData.CreatedAt,
	}

	if priceData.UpdatedAt.Valid {
		res.UpdatedAt = &priceData.UpdatedAt.Time
	}

	return res, nil
}

func (p *pservice) DeleteProductPrice(productID, priceID int) error {
	err := productStore.DeletePrice(productID, priceID)
	if err != nil {
		return err
	}

	return nil
}
