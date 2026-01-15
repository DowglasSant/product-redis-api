package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

func TestUpdateProductUseCase_Execute_Success(t *testing.T) {
	existingProduct := newTestProductWithData("Old Name", "REF-001", "Old Category")

	mockProductRepo := &MockProductRepository{
		UpdateFunc: func(ctx context.Context, product *entity.Product, expectedVersion int) error {
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewUpdateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.UpdateProductInput{
		Name:        "New Name",
		Category:    "New Category",
		Description: "Updated description",
		SKU:         "NEW-SKU",
		Brand:       "New Brand",
		Stock:       200,
	}

	product, err := uc.Execute(context.Background(), existingProduct.ID, input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if product == nil {
		t.Fatal("Expected product, got nil")
	}

	if product.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, product.Name)
	}

	if product.Category != input.Category {
		t.Errorf("Expected category %s, got %s", input.Category, product.Category)
	}

	if product.Version != existingProduct.Version+1 {
		t.Errorf("Expected version %d, got %d", existingProduct.Version+1, product.Version)
	}
}

func TestUpdateProductUseCase_Execute_ProductNotFound(t *testing.T) {
	mockProductRepo := &MockProductRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entity.Product, error) {
			return nil, repository.ErrProductNotFound
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return nil, repository.ErrCacheNotFound
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewUpdateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.UpdateProductInput{
		Name:     "New Name",
		Category: "Electronics",
	}

	product, err := uc.Execute(context.Background(), "non-existent-id", input)

	if !errors.Is(err, repository.ErrProductNotFound) {
		t.Errorf("Expected ErrProductNotFound, got %v", err)
	}

	if product != nil {
		t.Error("Expected nil product")
	}
}

func TestUpdateProductUseCase_Execute_NoChanges(t *testing.T) {
	existingProduct := newTestProductWithData("Same Name", "REF-001", "Same Category")
	updateCalled := false

	mockProductRepo := &MockProductRepository{
		UpdateFunc: func(ctx context.Context, product *entity.Product, expectedVersion int) error {
			updateCalled = true
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewUpdateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.UpdateProductInput{
		Name:           existingProduct.Name,
		Category:       existingProduct.Category,
		Description:    existingProduct.Description,
		SKU:            existingProduct.SKU,
		Brand:          existingProduct.Brand,
		Stock:          existingProduct.Stock,
		Images:         existingProduct.Images,
		Specifications: existingProduct.Specifications,
	}

	product, err := uc.Execute(context.Background(), existingProduct.ID, input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if product == nil {
		t.Fatal("Expected product, got nil")
	}

	if updateCalled {
		t.Error("Update should not be called when no changes detected")
	}
}

func TestUpdateProductUseCase_Execute_VersionConflict(t *testing.T) {
	existingProduct := newTestProductWithData("Old Name", "REF-001", "Category")

	mockProductRepo := &MockProductRepository{
		UpdateFunc: func(ctx context.Context, product *entity.Product, expectedVersion int) error {
			return repository.ErrVersionConflict
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewUpdateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.UpdateProductInput{
		Name:     "New Name",
		Category: "Category",
	}

	product, err := uc.Execute(context.Background(), existingProduct.ID, input)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, repository.ErrVersionConflict) {
		t.Errorf("Expected ErrVersionConflict, got %v", err)
	}

	if product != nil {
		t.Error("Expected nil product on version conflict")
	}
}

func TestUpdateProductUseCase_Execute_InvalidInput(t *testing.T) {
	existingProduct := newTestProductWithData("Old Name", "REF-001", "Category")

	mockProductRepo := &MockProductRepository{}
	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewUpdateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	tests := []struct {
		name  string
		input port.UpdateProductInput
	}{
		{
			name: "empty name",
			input: port.UpdateProductInput{
				Name:     "",
				Category: "Electronics",
			},
		},
		{
			name: "empty category",
			input: port.UpdateProductInput{
				Name:     "Product",
				Category: "",
			},
		},
		{
			name: "negative stock",
			input: port.UpdateProductInput{
				Name:     "Product",
				Category: "Electronics",
				Stock:    -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := uc.Execute(context.Background(), existingProduct.ID, tt.input)

			if err == nil {
				t.Error("Expected error, got nil")
			}

			if product != nil {
				t.Error("Expected nil product on validation error")
			}
		})
	}
}

func TestUpdateProductUseCase_Execute_DatabaseError(t *testing.T) {
	existingProduct := newTestProductWithData("Old Name", "REF-001", "Category")
	dbError := errors.New("database error")

	mockProductRepo := &MockProductRepository{
		UpdateFunc: func(ctx context.Context, product *entity.Product, expectedVersion int) error {
			return dbError
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewUpdateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.UpdateProductInput{
		Name:     "New Name",
		Category: "Category",
	}

	product, err := uc.Execute(context.Background(), existingProduct.ID, input)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if product != nil {
		t.Error("Expected nil product on database error")
	}
}

func TestUpdateProductUseCase_Execute_FetchFromDatabaseOnCacheMiss(t *testing.T) {
	existingProduct := newTestProductWithData("Old Name", "REF-001", "Category")
	dbFindCalled := false

	mockProductRepo := &MockProductRepository{
		FindByIDFunc: func(ctx context.Context, id string) (*entity.Product, error) {
			dbFindCalled = true
			return existingProduct, nil
		},
		UpdateFunc: func(ctx context.Context, product *entity.Product, expectedVersion int) error {
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return nil, repository.ErrCacheNotFound
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewUpdateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.UpdateProductInput{
		Name:     "New Name",
		Category: "Category",
	}

	product, err := uc.Execute(context.Background(), existingProduct.ID, input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !dbFindCalled {
		t.Error("Expected database to be queried on cache miss")
	}

	if product == nil {
		t.Fatal("Expected product, got nil")
	}
}

func TestUpdateProductUseCase_Execute_CategoryIndexUpdate(t *testing.T) {
	existingProduct := newTestProductWithData("Product", "REF-001", "OldCategory")
	oldCategoryRemoved := false
	newCategoryAdded := false

	mockProductRepo := &MockProductRepository{
		UpdateFunc: func(ctx context.Context, product *entity.Product, expectedVersion int) error {
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
		RemoveFromSetFunc: func(ctx context.Context, setKey, productID string) error {
			if setKey == "product_by_category_OldCategory" {
				oldCategoryRemoved = true
			}
			return nil
		},
		AddToSetFunc: func(ctx context.Context, setKey, productID string) error {
			if setKey == "product_by_category_NewCategory" {
				newCategoryAdded = true
			}
			return nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewUpdateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.UpdateProductInput{
		Name:     "Product",
		Category: "NewCategory",
	}

	_, err := uc.Execute(context.Background(), existingProduct.ID, input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !oldCategoryRemoved {
		t.Error("Expected old category index to be updated")
	}

	if !newCategoryAdded {
		t.Error("Expected new category index to be updated")
	}
}

func TestUpdateProductUseCase_Execute_NameIndexUpdate(t *testing.T) {
	existingProduct := newTestProductWithData("OldName", "REF-001", "Category")
	oldNameRemoved := false
	newNameAdded := false

	mockProductRepo := &MockProductRepository{
		UpdateFunc: func(ctx context.Context, product *entity.Product, expectedVersion int) error {
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
		RemoveFromSetFunc: func(ctx context.Context, setKey, productID string) error {
			if setKey == "product_by_name_OldName" {
				oldNameRemoved = true
			}
			return nil
		},
		AddToSetFunc: func(ctx context.Context, setKey, productID string) error {
			if setKey == "product_by_name_NewName" {
				newNameAdded = true
			}
			return nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewUpdateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.UpdateProductInput{
		Name:     "NewName",
		Category: "Category",
	}

	_, err := uc.Execute(context.Background(), existingProduct.ID, input)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !oldNameRemoved {
		t.Error("Expected old name index to be updated")
	}

	if !newNameAdded {
		t.Error("Expected new name index to be updated")
	}
}
