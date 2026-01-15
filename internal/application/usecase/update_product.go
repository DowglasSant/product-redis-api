package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

type UpdateProductUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      port.Logger
}

func NewUpdateProductUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger port.Logger,
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
		"product_id", id[:min(8, len(id))],
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
			"error", err,
			"product_id", id[:min(8, len(id))],
		)
		return nil, fmt.Errorf("invalid product data: %w", err)
	}

	if currentProduct.Equals(&updatedProduct) {
		uc.logger.Info("no changes detected - ignoring update",
			"product_id", id[:min(8, len(id))],
		)
		return currentProduct, nil
	}

	if err := uc.productRepo.Update(ctx, &updatedProduct, expectedVersion); err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			uc.logger.Warn("version conflict detected",
				"product_id", id[:min(8, len(id))],
				"expected_version", expectedVersion,
			)
			return nil, fmt.Errorf("product was modified by another process: %w", err)
		}

		uc.logger.Error("failed to update product in database",
			"error", err,
			"product_id", id[:min(8, len(id))],
		)
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	uc.logger.Info("product updated successfully in database",
		"product_id", id[:min(8, len(id))],
		"new_version", updatedProduct.Version,
	)

	uc.updateCache(ctx, &updatedProduct, oldCategory, oldName)

	return &updatedProduct, nil
}

func (uc *UpdateProductUseCase) getCurrentProduct(ctx context.Context, id string) (*entity.Product, error) {
	cacheKey := uc.cacheKeys.ProductKey(id)
	product, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err == nil {
		uc.logger.Debug("product found in cache",
			"product_id", id[:min(8, len(id))],
		)
		return product, nil
	}

	uc.logger.Debug("cache miss - fetching from database",
		"product_id", id[:min(8, len(id))],
	)

	product, err = uc.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, err
		}
		uc.logger.Error("failed to fetch product from database",
			"error", err,
			"product_id", id[:min(8, len(id))],
		)
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}

	return product, nil
}

func (uc *UpdateProductUseCase) updateCache(ctx context.Context, product *entity.Product, oldCategory, oldName string) {
	if err := uc.cacheRepo.Set(ctx, uc.cacheKeys.ProductKey(product.ID), product); err != nil {
		uc.logger.Error("failed to update cache",
			"error", err,
			"product_id", product.HashID(),
		)
	}

	if oldCategory != product.Category {
		oldCategoryKey := uc.cacheKeys.CategoryKey(oldCategory)
		if err := uc.cacheRepo.RemoveFromSet(ctx, oldCategoryKey, product.ID); err != nil {
			uc.logger.Error("failed to remove from old category index",
				"error", err,
				"product_id", product.HashID(),
				"old_category", oldCategory,
			)
		}

		newCategoryKey := uc.cacheKeys.CategoryKey(product.Category)
		if err := uc.cacheRepo.AddToSet(ctx, newCategoryKey, product.ID); err != nil {
			uc.logger.Error("failed to add to new category index",
				"error", err,
				"product_id", product.HashID(),
				"new_category", product.Category,
			)
		}
	}

	if oldName != product.Name {
		oldNameKey := uc.cacheKeys.NameKey(oldName)
		if err := uc.cacheRepo.RemoveFromSet(ctx, oldNameKey, product.ID); err != nil {
			uc.logger.Error("failed to remove from old name index",
				"error", err,
				"product_id", product.HashID(),
				"old_name", oldName,
			)
		}

		newNameKey := uc.cacheKeys.NameKey(product.Name)
		if err := uc.cacheRepo.AddToSet(ctx, newNameKey, product.ID); err != nil {
			uc.logger.Error("failed to add to new name index",
				"error", err,
				"product_id", product.HashID(),
				"new_name", product.Name,
			)
		}
	}

	uc.logger.Info("cache and indices updated successfully",
		"product_id", product.HashID(),
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
