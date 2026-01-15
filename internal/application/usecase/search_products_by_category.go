package usecase

import (
	"context"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/application/utils"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

type SearchProductsByCategoryUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      port.Logger
}

func NewSearchProductsByCategoryUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger port.Logger,
) *SearchProductsByCategoryUseCase {
	return &SearchProductsByCategoryUseCase{
		productRepo: productRepo,
		cacheRepo:   cacheRepo,
		cacheKeys:   cacheKeys,
		logger:      logger,
	}
}

func (uc *SearchProductsByCategoryUseCase) Execute(ctx context.Context, category string, limit, offset int) ([]*entity.Product, error) {
	uc.logger.Debug("searching products by category",
		"category", category,
		"limit", limit,
		"offset", offset,
	)

	products := uc.searchInCache(ctx, category)
	if len(products) > 0 {
		return utils.PaginateProducts(products, limit, offset), nil
	}

	uc.logger.Debug("cache miss - searching in database",
		"category", category,
	)

	products, err := uc.productRepo.FindByCategory(ctx, category, limit, offset)
	if err != nil {
		uc.logger.Error("failed to search products by category in database",
			"error", err,
			"category", category,
		)
		return nil, err
	}

	return products, nil
}

func (uc *SearchProductsByCategoryUseCase) searchInCache(ctx context.Context, category string) []*entity.Product {
	categoryKey := uc.cacheKeys.CategoryKey(category)

	productIDs, err := uc.cacheRepo.GetSet(ctx, categoryKey)
	if err != nil || len(productIDs) == 0 {
		return nil
	}

	keys := make([]string, len(productIDs))
	for i, id := range productIDs {
		keys[i] = uc.cacheKeys.ProductKey(id)
	}

	products, err := uc.cacheRepo.GetMultiple(ctx, keys)
	if err != nil {
		uc.logger.Debug("failed to get products from cache",
			"error", err,
		)
		return nil
	}

	if len(products) < len(productIDs) {
		return nil
	}

	uc.logger.Debug("cache hit for category search",
		"category", category,
		"count", len(products),
	)

	return products
}
