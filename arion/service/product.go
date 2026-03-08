package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/response"
	"github.com/zeirash/recapo/arion/store"
)

type (
	ProductService interface {
		CreateProduct(shopID int, name string, description *string, price int, originalPrice *int, imageURL *string) (response.ProductData, error)
		GetProductByID(productID int, shopID ...int) (*response.ProductData, error)
		GetProductsByShopID(shopID int, searchQuery *string) ([]response.ProductData, error)
		UpdateProduct(input UpdateProductInput) (response.ProductData, error)
		DeleteProductByID(id int) error
		GetPurchaseListProducts(shopID int) ([]response.PurchaseListProductData, error)
		UploadProductImage(file io.Reader) (string, error)
		DeleteProductImage(imageURL string) error
	}

	pservice struct{}

	UpdateProductInput struct {
		ID            int
		Name          *string
		Description   *string
		Price         *int
		OriginalPrice *int
		ImageURL      *string
	}
)

// r2UploadFunc uploads an object to R2; overridable in tests.
var r2UploadFunc = func(key string, body io.Reader, contentType string) error {
	client, err := newR2Client(cfg)
	if err != nil {
		return err
	}
	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.R2BucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	return err
}

// r2DeleteFunc deletes an object from R2; overridable in tests.
var r2DeleteFunc = func(key string) error {
	client, err := newR2Client(cfg)
	if err != nil {
		return err
	}
	_, err = client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(cfg.R2BucketName),
		Key:    aws.String(key),
	})
	return err
}

func newR2Client(c config.Config) (*s3.Client, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", c.R2AccountID)
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			c.R2AccessKeyID, c.R2SecretAccessKey, "",
		)),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	}), nil
}

func NewProductService() ProductService {
	cfg = config.GetConfig()

	if productStore == nil {
		productStore = store.NewProductStore()
	}

	return &pservice{}
}

func (p *pservice) CreateProduct(shopID int, name string, description *string, price int, originalPrice *int, imageURL *string) (response.ProductData, error) {
	product, err := productStore.CreateProduct(name, description, price, shopID, originalPrice, imageURL)
	if err != nil {
		return response.ProductData{}, err
	}

	res := response.ProductData{
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		OriginalPrice: product.OriginalPrice,
		ImageURL:      product.ImageURL,
		CreatedAt:     product.CreatedAt,
	}

	return res, nil
}

func (p *pservice) GetProductByID(productID int, shopID ...int) (*response.ProductData, error) {
	product, err := productStore.GetProductByID(productID, shopID...)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, errors.New(apierr.ErrProductNotFound)
	}

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
		res.UpdatedAt = &product.UpdatedAt.Time
	}

	return &res, nil
}

func (p *pservice) GetProductsByShopID(shopID int, searchQuery *string) ([]response.ProductData, error) {
	products, err := productStore.GetProductsByShopID(shopID, searchQuery)
	if err != nil {
		return []response.ProductData{}, err
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

func (p *pservice) UpdateProduct(input UpdateProductInput) (response.ProductData, error) {
	updateData := store.UpdateProductInput{
		Name:          input.Name,
		Description:   input.Description,
		Price:         input.Price,
		OriginalPrice: input.OriginalPrice,
		ImageURL:      input.ImageURL,
	}
	productData, err := productStore.UpdateProduct(input.ID, updateData)
	if err != nil {
		return response.ProductData{}, err
	}

	res := response.ProductData{
		ID:            productData.ID,
		Name:          productData.Name,
		Description:   productData.Description,
		Price:         productData.Price,
		OriginalPrice: productData.OriginalPrice,
		ImageURL:      productData.ImageURL,
		CreatedAt:     productData.CreatedAt,
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

func (p *pservice) GetPurchaseListProducts(shopID int) ([]response.PurchaseListProductData, error) {
	products, err := productStore.GetProductsListByActiveOrders(shopID)
	if err != nil {
		return []response.PurchaseListProductData{}, err
	}

	productsData := []response.PurchaseListProductData{}
	for _, product := range products {
		productsData = append(productsData, response.PurchaseListProductData{
			ProductName: product.ProductName,
			Price:       product.Price,
			Qty:         product.Qty,
		})
	}

	return productsData, nil
}

func (p *pservice) UploadProductImage(file io.Reader) (string, error) {
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	contentType := http.DetectContentType(buf[:n])

	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return "", errors.New(apierr.ErrUnsupportedImageType)
	}

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%x%s", randBytes, ext)

	// Reconstruct the full file content from the already-read header bytes plus remainder.
	fullReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// Cloud path: upload to Cloudflare R2.
	if cfg.R2BucketName != "" {
		objectKey := "products/" + filename
		if err := r2UploadFunc(objectKey, fullReader, contentType); err != nil {
			return "", err
		}
		return cfg.R2PublicURL + "/" + objectKey, nil
	}

	// Local filesystem path.
	uploadDir := filepath.Join(cfg.UploadDir, "products")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, fullReader); err != nil {
		os.Remove(filePath)
		return "", err
	}

	return "/uploads/products/" + filename, nil
}

func (p *pservice) DeleteProductImage(imageURL string) error {
	// Cloud path: delete from Cloudflare R2.
	if cfg.R2BucketName != "" && cfg.R2PublicURL != "" && strings.HasPrefix(imageURL, cfg.R2PublicURL) {
		objectKey := strings.TrimPrefix(imageURL, cfg.R2PublicURL+"/")
		return r2DeleteFunc(objectKey)
	}

	// Local filesystem path.
	const urlPrefix = "/uploads/products/"
	if !strings.HasPrefix(imageURL, urlPrefix) {
		return errors.New(apierr.ErrInvalidImageURL)
	}

	filename := strings.TrimPrefix(imageURL, urlPrefix)
	filePath := filepath.Join(cfg.UploadDir, "products", filename)
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return errors.New(apierr.ErrImageNotFound)
		}
		return err
	}

	return nil
}
