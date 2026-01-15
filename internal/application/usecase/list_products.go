package usecase

import (
	"context"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/application/utils"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

type ListProductsUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      port.Logger
}

func NewListProductsUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger port.Logger,
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
		"limit", limit,
		"offset", offset,
	)

	products, cacheHit := uc.getFromCache(ctx)
	if cacheHit && len(products) > 0 {
		return utils.PaginateProducts(products, limit, offset), nil
	}

	uc.logger.Debug("fetching products from database")
	products, err := uc.productRepo.FindAll(ctx, limit, offset)
	if err != nil {
		uc.logger.Error("failed to fetch products from database",
			"error", err,
		)
		return nil, err
	}

	return products, nil
}

func (uc *ListProductsUseCase) getFromCache(ctx context.Context) ([]*entity.Product, bool) {
	productIDs, err := uc.cacheRepo.GetSet(ctx, uc.cacheKeys.AllProductsKey())
	if err != nil {
		uc.logger.Debug("failed to get all_products set",
			"error", err,
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
			"error", err,
		)
		return nil, false
	}

	if len(products) < len(productIDs) {
		uc.logger.Debug("partial cache miss",
			"expected", len(productIDs),
			"got", len(products),
		)
		return nil, false
	}

	uc.logger.Debug("cache hit for all products",
		"count", len(products),
	)

	return products, true
}
