package dto

import (
	"time"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
)

// ProductResponse representa a resposta de um produto
// @Description Dados completos de um produto
type ProductResponse struct {
	ID              string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name            string                 `json:"name" example:"iPhone 15 Pro"`
	ReferenceNumber string                 `json:"reference_number" example:"REF-12345"`
	Category        string                 `json:"category" example:"electronics"`
	Description     string                 `json:"description" example:"Smartphone Apple com chip A17 Pro"`
	SKU             string                 `json:"sku" example:"SKU-IP15P-256"`
	Brand           string                 `json:"brand" example:"Apple"`
	Stock           int                    `json:"stock" example:"100"`
	Images          []string               `json:"images" example:"https://example.com/image1.jpg"`
	Specifications  map[string]interface{} `json:"specifications"`
	Version         int                    `json:"version" example:"1"`
	CreatedAt       time.Time              `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt       time.Time              `json:"updated_at" example:"2024-01-15T10:30:00Z"`
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

// ErrorResponse representa uma resposta de erro
// @Description Estrutura de resposta de erro da API
type ErrorResponse struct {
	Error   string `json:"error" example:"validation_error"`
	Message string `json:"message,omitempty" example:"Invalid request body"`
	Code    string `json:"code,omitempty" example:"400"`
}

// SuccessResponse representa uma resposta de sucesso genérica
// @Description Estrutura de resposta de sucesso da API
type SuccessResponse struct {
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}
