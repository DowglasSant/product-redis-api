package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

func TestCreateProductUseCase_Execute_Success(t *testing.T) {
	mockProductRepo := &MockProductRepository{
		CreateFunc: func(ctx context.Context, product *entity.Product) error {
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
	uc := NewCreateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.CreateProductInput{
		Name:            "iPhone 15",
		ReferenceNumber: "APL-IP15-001",
		Category:        "Smartphones",
		Description:     "Latest iPhone",
		SKU:             "APPLE-IP15",
		Brand:           "Apple",
		Stock:           100,
		Images:          []string{"image1.jpg"},
		Specifications:  map[string]interface{}{"color": "black"},
	}

	product, err := uc.Execute(context.Background(), input)

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
}

func TestCreateProductUseCase_Execute_InvalidInput(t *testing.T) {
	mockProductRepo := &MockProductRepository{}
	mockCacheRepo := &MockCacheRepository{}
	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewCreateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	tests := []struct {
		name  string
		input port.CreateProductInput
	}{
		{
			name: "empty name",
			input: port.CreateProductInput{
				Name:            "",
				ReferenceNumber: "REF-001",
				Category:        "Electronics",
			},
		},
		{
			name: "empty reference",
			input: port.CreateProductInput{
				Name:            "Product",
				ReferenceNumber: "",
				Category:        "Electronics",
			},
		},
		{
			name: "empty category",
			input: port.CreateProductInput{
				Name:            "Product",
				ReferenceNumber: "REF-001",
				Category:        "",
			},
		},
		{
			name: "negative stock",
			input: port.CreateProductInput{
				Name:            "Product",
				ReferenceNumber: "REF-001",
				Category:        "Electronics",
				Stock:           -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := uc.Execute(context.Background(), tt.input)

			if err == nil {
				t.Error("Expected error, got nil")
			}

			if product != nil {
				t.Error("Expected nil product on error")
			}
		})
	}
}

func TestCreateProductUseCase_Execute_ProductAlreadyExistsInCache(t *testing.T) {
	existingProduct, _ := entity.NewProduct(
		"iPhone 15",
		"APL-IP15-001",
		"Smartphones",
		"Latest iPhone",
		"APPLE-IP15",
		"Apple",
		50,
		[]string{},
		map[string]interface{}{},
	)

	mockProductRepo := &MockProductRepository{}
	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewCreateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.CreateProductInput{
		Name:            "iPhone 15",
		ReferenceNumber: "APL-IP15-001",
		Category:        "Smartphones",
		Description:     "Latest iPhone",
		SKU:             "APPLE-IP15",
		Brand:           "Apple",
		Stock:           50,
		Images:          []string{},
		Specifications:  map[string]interface{}{},
	}

	product, err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("Expected no error for identical product, got %v", err)
	}

	if product == nil {
		t.Fatal("Expected existing product to be returned")
	}
}

func TestCreateProductUseCase_Execute_ProductExistsWithDifferentData(t *testing.T) {
	existingProduct, _ := entity.NewProduct(
		"iPhone 15",
		"APL-IP15-001",
		"Smartphones",
		"Original description",
		"ORIGINAL-SKU",
		"Apple",
		50,
		[]string{},
		map[string]interface{}{},
	)

	mockProductRepo := &MockProductRepository{}
	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return existingProduct, nil
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewCreateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.CreateProductInput{
		Name:            "iPhone 15",
		ReferenceNumber: "APL-IP15-001",
		Category:        "Tablets",
		Description:     "Different description",
		SKU:             "DIFFERENT-SKU",
		Brand:           "Apple",
		Stock:           200,
		Images:          []string{},
		Specifications:  map[string]interface{}{},
	}

	product, err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected error for product with different data")
	}

	if !errors.Is(err, repository.ErrProductAlreadyExists) {
		t.Errorf("Expected ErrProductAlreadyExists, got %v", err)
	}

	if product != nil {
		t.Error("Expected nil product on duplicate error")
	}
}

func TestCreateProductUseCase_Execute_DatabaseError(t *testing.T) {
	dbError := errors.New("database connection failed")

	mockProductRepo := &MockProductRepository{
		CreateFunc: func(ctx context.Context, product *entity.Product) error {
			return dbError
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return nil, repository.ErrCacheNotFound
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewCreateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.CreateProductInput{
		Name:            "Test Product",
		ReferenceNumber: "REF-001",
		Category:        "Electronics",
		Stock:           10,
	}

	product, err := uc.Execute(context.Background(), input)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if product != nil {
		t.Error("Expected nil product on database error")
	}
}

func TestCreateProductUseCase_Execute_ProductAlreadyExistsInDatabase(t *testing.T) {
	mockProductRepo := &MockProductRepository{
		CreateFunc: func(ctx context.Context, product *entity.Product) error {
			return repository.ErrProductAlreadyExists
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return nil, repository.ErrCacheNotFound
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewCreateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.CreateProductInput{
		Name:            "Test Product",
		ReferenceNumber: "REF-001",
		Category:        "Electronics",
		Stock:           10,
	}

	product, err := uc.Execute(context.Background(), input)

	if !errors.Is(err, repository.ErrProductAlreadyExists) {
		t.Errorf("Expected ErrProductAlreadyExists, got %v", err)
	}

	if product != nil {
		t.Error("Expected nil product")
	}
}

func TestCreateProductUseCase_Execute_CacheUpdateFailure(t *testing.T) {
	mockProductRepo := &MockProductRepository{
		CreateFunc: func(ctx context.Context, product *entity.Product) error {
			return nil
		},
	}

	mockCacheRepo := &MockCacheRepository{
		GetFunc: func(ctx context.Context, key string) (*entity.Product, error) {
			return nil, repository.ErrCacheNotFound
		},
		SetFunc: func(ctx context.Context, key string, product *entity.Product) error {
			return errors.New("cache set failed")
		},
		AddToSetFunc: func(ctx context.Context, setKey, productID string) error {
			return errors.New("cache add to set failed")
		},
	}

	mockCacheKeys := &MockCacheKeyGenerator{}
	logger := &MockLogger{}
	uc := NewCreateProductUseCase(mockProductRepo, mockCacheRepo, mockCacheKeys, logger)

	input := port.CreateProductInput{
		Name:            "Test Product",
		ReferenceNumber: "REF-001",
		Category:        "Electronics",
		Stock:           10,
	}

	product, err := uc.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("Cache failures should not cause use case to fail, got %v", err)
	}

	if product == nil {
		t.Error("Expected product even with cache failures")
	}
}
