package repository

import (
	"context"
	"errors"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
)

var (
	ErrCacheNotFound = errors.New("cache entry not found")
	ErrCacheMiss     = errors.New("cache miss")
)

type CacheRepository interface {
	Get(ctx context.Context, key string) (*entity.Product, error)

	Set(ctx context.Context, key string, product *entity.Product) error

	Delete(ctx context.Context, key string) error

	AddToSet(ctx context.Context, setKey, productID string) error

	RemoveFromSet(ctx context.Context, setKey, productID string) error

	GetSet(ctx context.Context, setKey string) ([]string, error)

	GetMultiple(ctx context.Context, keys []string) ([]*entity.Product, error)

	Exists(ctx context.Context, key string) (bool, error)

	DeleteSet(ctx context.Context, setKey string) error

	HealthCheck(ctx context.Context) error
}
