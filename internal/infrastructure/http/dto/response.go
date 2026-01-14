package dto

import (
	"time"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
)

type ProductResponse struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	ReferenceNumber string                 `json:"reference_number"`
	Category        string                 `json:"category"`
	Description     string                 `json:"description"`
	SKU             string                 `json:"sku"`
	Brand           string                 `json:"brand"`
	Stock           int                    `json:"stock"`
	Images          []string               `json:"images"`
	Specifications  map[string]interface{} `json:"specifications"`
	Version         int                    `json:"version"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

func ToProductResponse(product *entity.Product) *ProductResponse {
	return &ProductResponse{
		ID:              product.ID,
		Name:            product.Name,
		ReferenceNumber: product.ReferenceNumber,
		Category:        product.Category,
		Description:     product.Description,
		SKU:             product.SKU,
		Brand:           product.Brand,
		Stock:           product.Stock,
		Images:          product.Images,
		Specifications:  product.Specifications,
		Version:         product.Version,
		CreatedAt:       product.CreatedAt,
		UpdatedAt:       product.UpdatedAt,
	}
}

func ToProductResponseList(products []*entity.Product) []*ProductResponse {
	responses := make([]*ProductResponse, len(products))
	for i, product := range products {
		responses[i] = ToProductResponse(product)
	}
	return responses
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
