package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

type DeleteProductUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      port.Logger
}

func NewDeleteProductUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger port.Logger,
) *DeleteProductUseCase {
	return &DeleteProductUseCase{
		productRepo: productRepo,
		cacheRepo:   cacheRepo,
		cacheKeys:   cacheKeys,
		logger:      logger,
	}
}

func (uc *DeleteProductUseCase) Execute(ctx context.Context, id string) error {
	uc.logger.Info("deleting product",
		"product_id", id[:min(8, len(id))],
	)

	product, _ := uc.cacheRepo.Get(ctx, uc.cacheKeys.ProductKey(id))

	if err := uc.productRepo.Delete(ctx, id); err != nil {
		uc.logger.Error("failed to delete product from database",
			"error", err,
			"product_id", id[:min(8, len(id))],
		)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	uc.logger.Info("product deleted from database",
		"product_id", id[:min(8, len(id))],
	)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		uc.cleanupCache(ctx, id, product)
	}()

	return nil
}

func (uc *DeleteProductUseCase) cleanupCache(ctx context.Context, id string, product *entity.Product) {
	productKey := uc.cacheKeys.ProductKey(id)

	if err := uc.cacheRepo.Delete(ctx, productKey); err != nil {
		uc.logger.Debug("failed to delete product key from cache",
			"error", err,
			"product_id", id[:min(8, len(id))],
		)
	}

	if err := uc.cacheRepo.RemoveFromSet(ctx, uc.cacheKeys.AllProductsKey(), id); err != nil {
		uc.logger.Debug("failed to remove from all_products index",
			"error", err,
			"product_id", id[:min(8, len(id))],
		)
	}

	if product != nil {
		if err := uc.cacheRepo.RemoveFromSet(ctx, uc.cacheKeys.NameKey(product.Name), id); err != nil {
			uc.logger.Debug("failed to remove from name index",
				"error", err,
				"product_id", id[:min(8, len(id))],
			)
		}

		if err := uc.cacheRepo.RemoveFromSet(ctx, uc.cacheKeys.CategoryKey(product.Category), id); err != nil {
			uc.logger.Debug("failed to remove from category index",
				"error", err,
				"product_id", id[:min(8, len(id))],
			)
		}
	}

	uc.logger.Info("cache cleanup completed",
		"product_id", id[:min(8, len(id))],
	)
}
