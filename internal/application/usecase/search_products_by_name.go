package usecase

import (
	"context"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/application/utils"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

type SearchProductsByNameUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      port.Logger
}

func NewSearchProductsByNameUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger port.Logger,
) *SearchProductsByNameUseCase {
	return &SearchProductsByNameUseCase{
		productRepo: productRepo,
		cacheRepo:   cacheRepo,
		cacheKeys:   cacheKeys,
		logger:      logger,
	}
}

func (uc *SearchProductsByNameUseCase) Execute(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error) {
	uc.logger.Debug("searching products by name",
		"name", name,
		"limit", limit,
		"offset", offset,
	)

	products := uc.searchInCache(ctx, name)
	if len(products) > 0 {
		return utils.PaginateProducts(products, limit, offset), nil
	}

	uc.logger.Debug("cache miss - searching in database",
		"name", name,
	)

	products, err := uc.productRepo.FindByName(ctx, name, limit, offset)
	if err != nil {
		uc.logger.Error("failed to search products by name in database",
			"error", err,
			"name", name,
		)
		return nil, err
	}

	return products, nil
}

func (uc *SearchProductsByNameUseCase) searchInCache(ctx context.Context, name string) []*entity.Product {
	nameKey := uc.cacheKeys.NameKey(name)

	productIDs, err := uc.cacheRepo.GetSet(ctx, nameKey)
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

	uc.logger.Debug("cache hit for name search",
		"name", name,
		"count", len(products),
	)

	return products
}
