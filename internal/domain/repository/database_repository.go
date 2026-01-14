package repository

import (
	"context"
	"errors"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
)

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrDatabaseConnection   = errors.New("database connection error")
	ErrVersionConflict      = entity.ErrVersionConflict
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error

	Update(ctx context.Context, product *entity.Product, expectedVersion int) error

	Delete(ctx context.Context, id string) error

	FindByID(ctx context.Context, id string) (*entity.Product, error)

	FindAll(ctx context.Context, limit, offset int) ([]*entity.Product, error)

	FindByCategory(ctx context.Context, category string, limit, offset int) ([]*entity.Product, error)

	FindByName(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error)

	Exists(ctx context.Context, id string) (bool, error)

	HealthCheck(ctx context.Context) error
}
