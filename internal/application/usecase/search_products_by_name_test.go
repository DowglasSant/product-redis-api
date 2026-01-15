package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
)

func TestSearchProductsByNameUseCase_Execute_CacheHit(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("iPhone 15", "REF-001", "Smartphones"),
		newTestProductWithData("iPhone 14", "REF-002", "Smartphones"),
	}

	mockProductRepo := &MockProductRepository{}
	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			if setKey == "product_by_name_iPhone" {
				return []string{products[0].ID, products[1].ID}, nil
			}
			return []string{}, nil
		},
		GetMultipleFunc: func(ctx context.Context, keys []string) ([]*entity.Product, error) {
			return products, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewSearchProductsByNameUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), "iPhone", 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products, got %d", len(result))
	}
}

func TestSearchProductsByNameUseCase_Execute_CacheMiss_DatabaseSuccess(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("Samsung Galaxy", "REF-001", "Smartphones"),
	}

	dbCalled := false

	mockProductRepo := &MockProductRepository{
		FindByNameFunc: func(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error) {
			dbCalled = true
			if name == "Samsung" {
				return products, nil
			}
			return []*entity.Product{}, nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			return []string{}, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewSearchProductsByNameUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), "Samsung", 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !dbCalled {
		t.Error("Expected database to be called on cache miss")
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 product, got %d", len(result))
	}
}

func TestSearchProductsByNameUseCase_Execute_DatabaseError(t *testing.T) {
	dbError := errors.New("database error")

	mockProductRepo := &MockProductRepository{
		FindByNameFunc: func(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error) {
			return nil, dbError
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			return []string{}, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewSearchProductsByNameUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), "Product", 10, 0)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result on error")
	}
}

func TestSearchProductsByNameUseCase_Execute_CacheError_FallbackToDatabase(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("Product", "REF-001", "Category"),
	}

	dbCalled := false

	mockProductRepo := &MockProductRepository{
		FindByNameFunc: func(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error) {
			dbCalled = true
			return products, nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			return nil, errors.New("cache error")
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewSearchProductsByNameUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), "Product", 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !dbCalled {
		t.Error("Expected database to be called on cache error")
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 product, got %d", len(result))
	}
}

func TestSearchProductsByNameUseCase_Execute_PartialCacheMiss(t *testing.T) {
	product := newTestProductWithData("Product", "REF-001", "Category")

	dbCalled := false

	mockProductRepo := &MockProductRepository{
		FindByNameFunc: func(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error) {
			dbCalled = true
			return []*entity.Product{product}, nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			return []string{"id1", "id2", "id3"}, nil
		},
		GetMultipleFunc: func(ctx context.Context, keys []string) ([]*entity.Product, error) {
			return []*entity.Product{product}, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewSearchProductsByNameUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), "Product", 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !dbCalled {
		t.Error("Expected database to be called on partial cache miss")
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 product, got %d", len(result))
	}
}

func TestSearchProductsByNameUseCase_Execute_Pagination(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("Product 1", "REF-001", "Category"),
		newTestProductWithData("Product 2", "REF-002", "Category"),
		newTestProductWithData("Product 3", "REF-003", "Category"),
	}

	mockProductRepo := &MockProductRepository{}
	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			ids := make([]string, len(products))
			for i, p := range products {
				ids[i] = p.ID
			}
			return ids, nil
		},
		GetMultipleFunc: func(ctx context.Context, keys []string) ([]*entity.Product, error) {
			return products, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewSearchProductsByNameUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), "Product", 2, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products with limit=2, got %d", len(result))
	}

	result, err = uc.Execute(context.Background(), "Product", 2, 2)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 product with limit=2 offset=2, got %d", len(result))
	}
}

func TestSearchProductsByNameUseCase_Execute_EmptyResult(t *testing.T) {
	mockProductRepo := &MockProductRepository{
		FindByNameFunc: func(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error) {
			return []*entity.Product{}, nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			return []string{}, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewSearchProductsByNameUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), "NonExistent", 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 products, got %d", len(result))
	}
}

func TestSearchProductsByNameUseCase_Execute_GetMultipleError(t *testing.T) {
	product := newTestProductWithData("Product", "REF-001", "Category")

	dbCalled := false

	mockProductRepo := &MockProductRepository{
		FindByNameFunc: func(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error) {
			dbCalled = true
			return []*entity.Product{product}, nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			return []string{"id1"}, nil
		},
		GetMultipleFunc: func(ctx context.Context, keys []string) ([]*entity.Product, error) {
			return nil, errors.New("get multiple error")
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewSearchProductsByNameUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), "Product", 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !dbCalled {
		t.Error("Expected database to be called on GetMultiple error")
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 product, got %d", len(result))
	}
}

func TestSearchProductsByNameUseCase_Execute_CacheKeyGeneration(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("iPhone 15", "REF-001", "Smartphones"),
	}

	calledWithKey := ""

	mockProductRepo := &MockProductRepository{}
	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			calledWithKey = setKey
			return []string{products[0].ID}, nil
		},
		GetMultipleFunc: func(ctx context.Context, keys []string) ([]*entity.Product, error) {
			return products, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewSearchProductsByNameUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	_, err := uc.Execute(context.Background(), "IPHONE", 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if calledWithKey != "product_by_name_IPHONE" {
		t.Errorf("Expected key 'product_by_name_IPHONE', got '%s'", calledWithKey)
	}
}
