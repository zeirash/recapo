package service

import (
	"context"
	"errors"

	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/model"
	"github.com/zeirash/recapo/arion/store"
)

type (
	ShopService interface {
		GetShareTokenByID(ctx context.Context, shopID int) (string, error)
		GetPublicProducts(ctx context.Context, shareToken string) ([]response.ProductData, error)
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

func (s *shopService) GetShareTokenByID(ctx context.Context, shopID int) (string, error) {
	token, err := shopStore.GetShareTokenByID(ctx, shopID)
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", errors.New(apierr.ErrShopNotFound)
	}
	return token, nil
}

func (s *shopService) GetPublicProducts(ctx context.Context, shareToken string) ([]response.ProductData, error) {
	shop, err := shopStore.GetShopByShareToken(ctx, shareToken)
	if err != nil {
		return nil, err
	}

	if shop == nil {
		return nil, errors.New(apierr.ErrShopNotFound)
	}

	active := true
	products, err := productStore.GetProductsByShopID(ctx, shop.ID, model.FilterOptions{
		IsActive: &active,
	})
	if err != nil {
		return nil, err
	}

	productsData := []response.ProductData{}
	for _, product := range products {
		res := response.ProductData{
			ID:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			OriginalPrice: product.OriginalPrice,
			ImageURL:      product.ImageURL,
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
