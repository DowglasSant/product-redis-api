package usecase

import (
	"context"
	"errors"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
	"go.uber.org/zap"
)

type GetProductUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      *zap.Logger
}

func NewGetProductUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger *zap.Logger,
) *GetProductUseCase {
	return &GetProductUseCase{
		productRepo: productRepo,
		cacheRepo:   cacheRepo,
		cacheKeys:   cacheKeys,
		logger:      logger,
	}
}

func (uc *GetProductUseCase) Execute(ctx context.Context, id string) (*entity.Product, error) {
	uc.logger.Debug("fetching product",
		zap.String("product_id", id[:min(8, len(id))]),
	)

	cacheKey := uc.cacheKeys.ProductKey(id)
	product, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err == nil {
		uc.logger.Debug("cache hit",
			zap.String("product_id", id[:min(8, len(id))]),
		)
		return product, nil
	}

	uc.logger.Debug("cache miss or error",
		zap.Error(err),
		zap.String("product_id", id[:min(8, len(id))]),
	)

	product, err = uc.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			uc.logger.Debug("product not found",
				zap.String("product_id", id[:min(8, len(id))]),
			)
			return nil, err
		}

		uc.logger.Error("failed to fetch product from database",
			zap.Error(err),
			zap.String("product_id", id[:min(8, len(id))]),
		)
		return nil, err
	}

	return product, nil
}
