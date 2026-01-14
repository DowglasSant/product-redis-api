package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
	"go.uber.org/zap"
)

func TestDeleteProductUseCase_Execute_Success(t *testing.T) {
	existingProduct := newTestProductWithData("Product", "REF-001", "Category")
	deleteCalled := false

	mockProductRepo := &MockProductRepository{
		DeleteFunc: func(ctx context.Context, id string) error {
			deleteCalled = true
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := zap.NewNop()
	uc := NewDeleteProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	err := uc.Execute(context.Background(), existingProduct.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !deleteCalled {
		t.Error("Expected database delete to be called")
	}
}

func TestDeleteProductUseCase_Execute_DatabaseError(t *testing.T) {
	dbError := errors.New("database error")

	mockProductRepo := &MockProductRepository{
		DeleteFunc: func(ctx context.Context, id string) error {
			return dbError
		},
	}

	mockCacheRepo := &MockCacheRepository{}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := zap.NewNop()
	uc := NewDeleteProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	err := uc.Execute(context.Background(), "some-id")

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDeleteProductUseCase_Execute_CacheCleanupOnSuccess(t *testing.T) {
	existingProduct := newTestProductWithData("Product", "REF-001", "Category")

	var mu sync.Mutex
	deletedKeys := make([]string, 0)
	removedFromSets := make([]string, 0)

	mockProductRepo := &MockProductRepository{
		DeleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
		DeleteFunc: func(ctx context.Context, key string) error {
			mu.Lock()
			deletedKeys = append(deletedKeys, key)
			mu.Unlock()
			return nil
		},
		RemoveFromSetFunc: func(ctx context.Context, setKey, productID string) error {
			mu.Lock()
			removedFromSets = append(removedFromSets, setKey)
			mu.Unlock()
			return nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := zap.NewNop()
	uc := NewDeleteProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	err := uc.Execute(context.Background(), existingProduct.ID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(deletedKeys) == 0 {
		t.Error("Expected product key to be deleted from cache")
	}

	if len(removedFromSets) < 3 {
		t.Errorf("Expected at least 3 sets to be updated (all_products, name, category), got %d", len(removedFromSets))
	}
}

func TestDeleteProductUseCase_Execute_CacheCleanupWithoutProductInfo(t *testing.T) {
	var mu sync.Mutex
	deletedKeys := make([]string, 0)
	removedFromSets := make([]string, 0)

	mockProductRepo := &MockProductRepository{
		DeleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return nil, repository.ErrCacheNotFound
		},
		DeleteFunc: func(ctx context.Context, key string) error {
			mu.Lock()
			deletedKeys = append(deletedKeys, key)
			mu.Unlock()
			return nil
		},
		RemoveFromSetFunc: func(ctx context.Context, setKey, productID string) error {
			mu.Lock()
			removedFromSets = append(removedFromSets, setKey)
			mu.Unlock()
			return nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := zap.NewNop()
	uc := NewDeleteProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	err := uc.Execute(context.Background(), "some-product-id")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(deletedKeys) == 0 {
		t.Error("Expected product key to be deleted from cache")
	}

	hasAllProductsSet := false
	for _, setKey := range removedFromSets {
		if setKey == "all_products" {
			hasAllProductsSet = true
			break
		}
	}

	if !hasAllProductsSet {
		t.Error("Expected all_products set to be updated even without product info")
	}
}

func TestDeleteProductUseCase_Execute_CacheErrorsDoNotFail(t *testing.T) {
	existingProduct := newTestProductWithData("Product", "REF-001", "Category")

	mockProductRepo := &MockProductRepository{
		DeleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
		DeleteFunc: func(ctx context.Context, key string) error {
			return errors.New("cache delete error")
		},
		RemoveFromSetFunc: func(ctx context.Context, setKey, productID string) error {
			return errors.New("cache remove from set error")
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := zap.NewNop()
	uc := NewDeleteProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	err := uc.Execute(context.Background(), existingProduct.ID)

	if err != nil {
		t.Errorf("Cache errors should not cause use case to fail, got %v", err)
	}
}

func TestDeleteProductUseCase_Execute_ShortProductID(t *testing.T) {
	mockProductRepo := &MockProductRepository{
		DeleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := zap.NewNop()
	uc := NewDeleteProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	err := uc.Execute(context.Background(), "abc")

	if err != nil {
		t.Errorf("Should handle short IDs gracefully, got %v", err)
	}
}
