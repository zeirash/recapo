package service

import (
	"errors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	ShopService interface {
		GetPublicProducts(shareToken string) ([]response.ProductData, error)
	}

	shopService struct{}
)

func NewShopService() ShopService {
	_ = config.GetConfig()

	if shopStore == nil {
		shopStore = store.NewShopStore()
	}
	if productStore == nil {
		productStore = store.NewProductStore()
	}

	return &shopService{}
}

func (s *shopService) GetPublicProducts(shareToken string) ([]response.ProductData, error) {
	shop, err := shopStore.GetShopByShareToken(shareToken)
	if err != nil {
		return nil, err
	}

	if shop == nil {
		return nil, errors.New("shop not found")
	}

	products, err := productStore.GetProductsByShopID(shop.ID, nil)
	if err != nil {
		return nil, err
	}

	var productsData []response.ProductData
	for _, product := range products {
		res := response.ProductData{
			ID:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			OriginalPrice: product.OriginalPrice,
			CreatedAt:     product.CreatedAt,
		}
		if product.UpdatedAt.Valid {
			t := product.UpdatedAt.Time
			res.UpdatedAt = &t
		}
		productsData = append(productsData, res)
	}

	return productsData, nil
}
