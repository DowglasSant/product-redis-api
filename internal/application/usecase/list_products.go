package usecase

import (
	"context"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/application/utils"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
	"go.uber.org/zap"
)

type ListProductsUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      *zap.Logger
}

func NewListProductsUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger *zap.Logger,
) *ListProductsUseCase {
	return &ListProductsUseCase{
		productRepo: productRepo,
		cacheRepo:   cacheRepo,
		cacheKeys:   cacheKeys,
		logger:      logger,
	}
}

func (uc *ListProductsUseCase) Execute(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
	uc.logger.Debug("listing products",
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	products, cacheHit := uc.getFromCache(ctx)
	if cacheHit && len(products) > 0 {
		return utils.PaginateProducts(products, limit, offset), nil
	}

	uc.logger.Debug("fetching products from database")
	products, err := uc.productRepo.FindAll(ctx, limit, offset)
	if err != nil {
		uc.logger.Error("failed to fetch products from database",
			zap.Error(err),
		)
		return nil, err
	}

	return products, nil
}

func (uc *ListProductsUseCase) getFromCache(ctx context.Context) ([]*entity.Product, bool) {
	productIDs, err := uc.cacheRepo.GetSet(ctx, uc.cacheKeys.AllProductsKey())
	if err != nil {
		uc.logger.Debug("failed to get all_products set",
			zap.Error(err),
		)
		return nil, false
	}

	if len(productIDs) == 0 {
		return nil, false
	}

	keys := make([]string, len(productIDs))
	for i, id := range productIDs {
		keys[i] = uc.cacheKeys.ProductKey(id)
	}

	products, err := uc.cacheRepo.GetMultiple(ctx, keys)
	if err != nil {
		uc.logger.Debug("failed to get products from cache",
			zap.Error(err),
		)
		return nil, false
	}

	if len(products) < len(productIDs) {
		uc.logger.Debug("partial cache miss",
			zap.Int("expected", len(productIDs)),
			zap.Int("got", len(products)),
		)
		return nil, false
	}

	uc.logger.Debug("cache hit for all products",
		zap.Int("count", len(products)),
	)

	return products, true
}
