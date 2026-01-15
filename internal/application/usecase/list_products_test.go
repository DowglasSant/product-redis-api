package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
)

func TestListProductsUseCase_Execute_CacheHit(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("Product 1", "REF-001", "Category"),
		newTestProductWithData("Product 2", "REF-002", "Category"),
		newTestProductWithData("Product 3", "REF-003", "Category"),
	}

	mockProductRepo := &MockProductRepository{}
	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			return []string{products[0].ID, products[1].ID, products[2].ID}, nil
		},
		GetMultipleFunc: func(ctx context.Context, keys []string) ([]*entity.Product, error) {
			return products, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewListProductsUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 products, got %d", len(result))
	}
}

func TestListProductsUseCase_Execute_CacheMiss_DatabaseSuccess(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("Product 1", "REF-001", "Category"),
		newTestProductWithData("Product 2", "REF-002", "Category"),
	}

	dbCalled := false

	mockProductRepo := &MockProductRepository{
		FindAllFunc: func(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
			dbCalled = true
			return products, nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			return []string{}, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewListProductsUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !dbCalled {
		t.Error("Expected database to be called on cache miss")
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products, got %d", len(result))
	}
}

func TestListProductsUseCase_Execute_DatabaseError(t *testing.T) {
	dbError := errors.New("database error")

	mockProductRepo := &MockProductRepository{
		FindAllFunc: func(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
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
	uc := NewListProductsUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), 10, 0)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if result != nil {
		t.Error("Expected nil result on error")
	}
}

func TestListProductsUseCase_Execute_CacheError_FallbackToDatabase(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("Product 1", "REF-001", "Category"),
	}

	dbCalled := false

	mockProductRepo := &MockProductRepository{
		FindAllFunc: func(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
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
	uc := NewListProductsUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), 10, 0)

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

func TestListProductsUseCase_Execute_PartialCacheMiss(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("Product 1", "REF-001", "Category"),
	}

	dbCalled := false

	mockProductRepo := &MockProductRepository{
		FindAllFunc: func(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
			dbCalled = true
			return products, nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetSetFunc: func(ctx context.Context, setKey string) ([]string, error) {
			return []string{"id1", "id2", "id3"}, nil
		},
		GetMultipleFunc: func(ctx context.Context, keys []string) ([]*entity.Product, error) {
			return []*entity.Product{products[0]}, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewListProductsUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), 10, 0)

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

func TestListProductsUseCase_Execute_Pagination(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("Product 1", "REF-001", "Category"),
		newTestProductWithData("Product 2", "REF-002", "Category"),
		newTestProductWithData("Product 3", "REF-003", "Category"),
		newTestProductWithData("Product 4", "REF-004", "Category"),
		newTestProductWithData("Product 5", "REF-005", "Category"),
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
	uc := NewListProductsUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), 2, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products with limit=2, got %d", len(result))
	}

	result, err = uc.Execute(context.Background(), 2, 2)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 products with limit=2 offset=2, got %d", len(result))
	}
}

func TestListProductsUseCase_Execute_EmptyResult(t *testing.T) {
	mockProductRepo := &MockProductRepository{
		FindAllFunc: func(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
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
	uc := NewListProductsUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), 10, 0)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 products, got %d", len(result))
	}
}

func TestListProductsUseCase_Execute_GetMultipleError(t *testing.T) {
	products := []*entity.Product{
		newTestProductWithData("Product 1", "REF-001", "Category"),
	}

	dbCalled := false

	mockProductRepo := &MockProductRepository{
		FindAllFunc: func(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
			dbCalled = true
			return products, nil
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
	uc := NewListProductsUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	result, err := uc.Execute(context.Background(), 10, 0)

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
