package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
	"go.uber.org/zap"
)

type UpdateProductUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      *zap.Logger
}

func NewUpdateProductUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger *zap.Logger,
) *UpdateProductUseCase {
	return &UpdateProductUseCase{
		productRepo: productRepo,
		cacheRepo:   cacheRepo,
		cacheKeys:   cacheKeys,
		logger:      logger,
	}
}

func (uc *UpdateProductUseCase) Execute(ctx context.Context, id string, input port.UpdateProductInput) (*entity.Product, error) {
	uc.logger.Info("attempting to update product",
		zap.String("product_id", id[:min(8, len(id))]),
	)

	currentProduct, err := uc.getCurrentProduct(ctx, id)
	if err != nil {
		return nil, err
	}

	oldCategory := currentProduct.Category
	oldName := currentProduct.Name
	expectedVersion := currentProduct.Version

	updatedProduct := *currentProduct
	err = updatedProduct.Update(
		input.Name,
		input.Category,
		input.Description,
		input.SKU,
		input.Brand,
		input.Stock,
		input.Images,
		input.Specifications,
	)
	if err != nil {
		uc.logger.Error("failed to validate updated product",
			zap.Error(err),
			zap.String("product_id", id[:min(8, len(id))]),
		)
		return nil, fmt.Errorf("invalid product data: %w", err)
	}

	if currentProduct.Equals(&updatedProduct) {
		uc.logger.Info("no changes detected - ignoring update",
			zap.String("product_id", id[:min(8, len(id))]),
		)
		return currentProduct, nil
	}

	if err := uc.productRepo.Update(ctx, &updatedProduct, expectedVersion); err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			uc.logger.Warn("version conflict detected",
				zap.String("product_id", id[:min(8, len(id))]),
				zap.Int("expected_version", expectedVersion),
			)
			return nil, fmt.Errorf("product was modified by another process: %w", err)
		}

		uc.logger.Error("failed to update product in database",
			zap.Error(err),
			zap.String("product_id", id[:min(8, len(id))]),
		)
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	uc.logger.Info("product updated successfully in database",
		zap.String("product_id", id[:min(8, len(id))]),
		zap.Int("new_version", updatedProduct.Version),
	)

	uc.updateCache(ctx, &updatedProduct, oldCategory, oldName)

	return &updatedProduct, nil
}

func (uc *UpdateProductUseCase) getCurrentProduct(ctx context.Context, id string) (*entity.Product, error) {
	cacheKey := uc.cacheKeys.ProductKey(id)
	product, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err == nil {
		uc.logger.Debug("product found in cache",
			zap.String("product_id", id[:min(8, len(id))]),
		)
		return product, nil
	}

	uc.logger.Debug("cache miss - fetching from database",
		zap.String("product_id", id[:min(8, len(id))]),
	)

	product, err = uc.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, err
		}
		uc.logger.Error("failed to fetch product from database",
			zap.Error(err),
			zap.String("product_id", id[:min(8, len(id))]),
		)
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}

	return product, nil
}

func (uc *UpdateProductUseCase) updateCache(ctx context.Context, product *entity.Product, oldCategory, oldName string) {
	if err := uc.cacheRepo.Set(ctx, uc.cacheKeys.ProductKey(product.ID), product); err != nil {
		uc.logger.Error("failed to update cache",
			zap.Error(err),
			zap.String("product_id", product.HashID()),
		)
	}

	if oldCategory != product.Category {
		oldCategoryKey := uc.cacheKeys.CategoryKey(oldCategory)
		if err := uc.cacheRepo.RemoveFromSet(ctx, oldCategoryKey, product.ID); err != nil {
			uc.logger.Error("failed to remove from old category index",
				zap.Error(err),
				zap.String("product_id", product.HashID()),
				zap.String("old_category", oldCategory),
			)
		}

		newCategoryKey := uc.cacheKeys.CategoryKey(product.Category)
		if err := uc.cacheRepo.AddToSet(ctx, newCategoryKey, product.ID); err != nil {
			uc.logger.Error("failed to add to new category index",
				zap.Error(err),
				zap.String("product_id", product.HashID()),
				zap.String("new_category", product.Category),
			)
		}
	}

	if oldName != product.Name {
		oldNameKey := uc.cacheKeys.NameKey(oldName)
		if err := uc.cacheRepo.RemoveFromSet(ctx, oldNameKey, product.ID); err != nil {
			uc.logger.Error("failed to remove from old name index",
				zap.Error(err),
				zap.String("product_id", product.HashID()),
				zap.String("old_name", oldName),
			)
		}

		newNameKey := uc.cacheKeys.NameKey(product.Name)
		if err := uc.cacheRepo.AddToSet(ctx, newNameKey, product.ID); err != nil {
			uc.logger.Error("failed to add to new name index",
				zap.Error(err),
				zap.String("product_id", product.HashID()),
				zap.String("new_name", product.Name),
			)
		}
	}

	uc.logger.Info("cache and indices updated successfully",
		zap.String("product_id", product.HashID()),
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
