package usecase

import (
	"context"
	"fmt"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
	"go.uber.org/zap"
)

type DeleteProductUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      *zap.Logger
}

func NewDeleteProductUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger *zap.Logger,
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
		zap.String("product_id", id[:min(8, len(id))]),
	)

	product, _ := uc.cacheRepo.Get(ctx, uc.cacheKeys.ProductKey(id))

	if err := uc.productRepo.Delete(ctx, id); err != nil {
		uc.logger.Error("failed to delete product from database",
			zap.Error(err),
			zap.String("product_id", id[:min(8, len(id))]),
		)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	uc.logger.Info("product deleted from database",
		zap.String("product_id", id[:min(8, len(id))]),
	)

	go uc.cleanupCache(context.Background(), id, product)

	return nil
}

func (uc *DeleteProductUseCase) cleanupCache(ctx context.Context, id string, product *entity.Product) {
	productKey := uc.cacheKeys.ProductKey(id)

	if err := uc.cacheRepo.Delete(ctx, productKey); err != nil {
		uc.logger.Debug("failed to delete product key from cache",
			zap.Error(err),
			zap.String("product_id", id[:min(8, len(id))]),
		)
	}

	if err := uc.cacheRepo.RemoveFromSet(ctx, uc.cacheKeys.AllProductsKey(), id); err != nil {
		uc.logger.Debug("failed to remove from all_products index",
			zap.Error(err),
			zap.String("product_id", id[:min(8, len(id))]),
		)
	}

	if product != nil {
		if err := uc.cacheRepo.RemoveFromSet(ctx, uc.cacheKeys.NameKey(product.Name), id); err != nil {
			uc.logger.Debug("failed to remove from name index",
				zap.Error(err),
				zap.String("product_id", id[:min(8, len(id))]),
			)
		}

		if err := uc.cacheRepo.RemoveFromSet(ctx, uc.cacheKeys.CategoryKey(product.Category), id); err != nil {
			uc.logger.Debug("failed to remove from category index",
				zap.Error(err),
				zap.String("product_id", id[:min(8, len(id))]),
			)
		}
	}

	uc.logger.Info("cache cleanup completed",
		zap.String("product_id", id[:min(8, len(id))]),
	)
}
