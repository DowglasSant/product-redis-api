package port

import (
	"context"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
)

type CreateProductInput struct {
	Name            string
	ReferenceNumber string
	Category        string
	Description     string
	SKU             string
	Brand           string
	Stock           int
	Images          []string
	Specifications  map[string]interface{}
}

type UpdateProductInput struct {
	Name           string
	Category       string
	Description    string
	SKU            string
	Brand          string
	Stock          int
	Images         []string
	Specifications map[string]interface{}
}

type ProductCreator interface {
	Execute(ctx context.Context, input CreateProductInput) (*entity.Product, error)
}

type ProductUpdater interface {
	Execute(ctx context.Context, id string, input UpdateProductInput) (*entity.Product, error)
}

type ProductDeleter interface {
	Execute(ctx context.Context, id string) error
}

type ProductGetter interface {
	Execute(ctx context.Context, id string) (*entity.Product, error)
}

type ProductLister interface {
	Execute(ctx context.Context, limit, offset int) ([]*entity.Product, error)
}

type ProductSearcherByName interface {
	Execute(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error)
}

type ProductSearcherByCategory interface {
	Execute(ctx context.Context, category string, limit, offset int) ([]*entity.Product, error)
}
