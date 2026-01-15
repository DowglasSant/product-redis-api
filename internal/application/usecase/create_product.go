package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

type CreateProductUseCase struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	cacheKeys   port.CacheKeyGenerator
	logger      port.Logger
}

func NewCreateProductUseCase(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	cacheKeys port.CacheKeyGenerator,
	logger port.Logger,
) *CreateProductUseCase {
	return &CreateProductUseCase{
		productRepo: productRepo,
		cacheRepo:   cacheRepo,
		cacheKeys:   cacheKeys,
		logger:      logger,
	}
}

func (uc *CreateProductUseCase) Execute(ctx context.Context, input port.CreateProductInput) (*entity.Product, error) {
	product, err := entity.NewProduct(
		input.Name,
		input.ReferenceNumber,
		input.Category,
		input.Description,
		input.SKU,
		input.Brand,
		input.Stock,
		input.Images,
		input.Specifications,
	)
	if err != nil {
		uc.logger.Error("failed to create product entity",
			"error", err,
			"name", input.Name,
			"reference", input.ReferenceNumber,
		)
		return nil, fmt.Errorf("invalid product data: %w", err)
	}

	uc.logger.Info("attempting to create product",
		"product_id", product.HashID(),
		"name", product.Name,
		"reference", product.ReferenceNumber,
	)

	cacheKey := uc.cacheKeys.ProductKey(product.ID)
	cachedProduct, cacheErr := uc.cacheRepo.Get(ctx, cacheKey)

	if cacheErr == nil && cachedProduct != nil {
		if product.Equals(cachedProduct) {
			uc.logger.Info("product already exists with identical data - ignoring",
				"product_id", product.HashID(),
			)
			return cachedProduct, nil
		}

		uc.logger.Warn("product exists but data has changed - treating as duplicate",
			"product_id", product.HashID(),
		)
		return nil, repository.ErrProductAlreadyExists
	}

	if cacheErr != nil {
		uc.logger.Warn("cache check failed - proceeding with database",
			"error", cacheErr,
			"product_id", product.HashID(),
		)
	}

	if err := uc.productRepo.Create(ctx, product); err != nil {
		if errors.Is(err, repository.ErrProductAlreadyExists) {
			uc.logger.Info("product already exists in database",
				"product_id", product.HashID(),
			)
			return nil, err
		}

		uc.logger.Error("failed to create product in database",
			"error", err,
			"product_id", product.HashID(),
		)
		return nil, fmt.Errorf("failed to save product: %w", err)
	}

	uc.logger.Info("product created successfully in database",
		"product_id", product.HashID(),
	)

	uc.updateCache(ctx, product)

	return product, nil
}

func (uc *CreateProductUseCase) updateCache(ctx context.Context, product *entity.Product) {
	if err := uc.cacheRepo.Set(ctx, uc.cacheKeys.ProductKey(product.ID), product); err != nil {
		uc.logger.Error("failed to cache product",
			"error", err,
			"product_id", product.HashID(),
		)
	}

	if err := uc.cacheRepo.AddToSet(ctx, uc.cacheKeys.AllProductsKey(), product.ID); err != nil {
		uc.logger.Error("failed to add to all_products set",
			"error", err,
			"product_id", product.HashID(),
		)
	}

	nameKey := uc.cacheKeys.NameKey(product.Name)
	if err := uc.cacheRepo.AddToSet(ctx, nameKey, product.ID); err != nil {
		uc.logger.Error("failed to add to name index",
			"error", err,
			"product_id", product.HashID(),
			"name", product.Name,
		)
	}

	categoryKey := uc.cacheKeys.CategoryKey(product.Category)
	if err := uc.cacheRepo.AddToSet(ctx, categoryKey, product.ID); err != nil {
		uc.logger.Error("failed to add to category index",
			"error", err,
			"product_id", product.HashID(),
			"category", product.Category,
		)
	}

	uc.logger.Info("cache and indices updated successfully",
		"product_id", product.HashID(),
	)
}
