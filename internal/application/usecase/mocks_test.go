package usecase

import (
	"context"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

type MockProductRepository struct {
	CreateFunc       func(ctx context.Context, product *entity.Product) error
	UpdateFunc       func(ctx context.Context, product *entity.Product, expectedVersion int) error
	DeleteFunc       func(ctx context.Context, id string) error
	FindByIDFunc     func(ctx context.Context, id string) (*entity.Product, error)
	FindAllFunc      func(ctx context.Context, limit, offset int) ([]*entity.Product, error)
	FindByCategoryFunc func(ctx context.Context, category string, limit, offset int) ([]*entity.Product, error)
	FindByNameFunc   func(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error)
	ExistsFunc       func(ctx context.Context, id string) (bool, error)
	HealthCheckFunc  func(ctx context.Context) error
}

func (m *MockProductRepository) Create(ctx context.Context, product *entity.Product) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, product)
	}
	return nil
}

func (m *MockProductRepository) Update(ctx context.Context, product *entity.Product, expectedVersion int) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, product, expectedVersion)
	}
	return nil
}

func (m *MockProductRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockProductRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, repository.ErrProductNotFound
}

func (m *MockProductRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx, limit, offset)
	}
	return []*entity.Product{}, nil
}

func (m *MockProductRepository) FindByCategory(ctx context.Context, category string, limit, offset int) ([]*entity.Product, error) {
	if m.FindByCategoryFunc != nil {
		return m.FindByCategoryFunc(ctx, category, limit, offset)
	}
	return []*entity.Product{}, nil
}

func (m *MockProductRepository) FindByName(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error) {
	if m.FindByNameFunc != nil {
		return m.FindByNameFunc(ctx, name, limit, offset)
	}
	return []*entity.Product{}, nil
}

func (m *MockProductRepository) Exists(ctx context.Context, id string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, id)
	}
	return false, nil
}

func (m *MockProductRepository) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return nil
}

type MockCacheRepository struct {
	GetFunc           func(ctx context.Context, key string) (*entity.Product, error)
	SetFunc           func(ctx context.Context, key string, product *entity.Product) error
	DeleteFunc        func(ctx context.Context, key string) error
	AddToSetFunc      func(ctx context.Context, setKey, productID string) error
	RemoveFromSetFunc func(ctx context.Context, setKey, productID string) error
	GetSetFunc        func(ctx context.Context, setKey string) ([]string, error)
	GetMultipleFunc   func(ctx context.Context, keys []string) ([]*entity.Product, error)
	ExistsFunc        func(ctx context.Context, key string) (bool, error)
	DeleteSetFunc     func(ctx context.Context, setKey string) error
	HealthCheckFunc   func(ctx context.Context) error
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (*entity.Product, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return nil, repository.ErrCacheNotFound
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, product *entity.Product) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, product)
	}
	return nil
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return nil
}

func (m *MockCacheRepository) AddToSet(ctx context.Context, setKey, productID string) error {
	if m.AddToSetFunc != nil {
		return m.AddToSetFunc(ctx, setKey, productID)
	}
	return nil
}

func (m *MockCacheRepository) RemoveFromSet(ctx context.Context, setKey, productID string) error {
	if m.RemoveFromSetFunc != nil {
		return m.RemoveFromSetFunc(ctx, setKey, productID)
	}
	return nil
}

func (m *MockCacheRepository) GetSet(ctx context.Context, setKey string) ([]string, error) {
	if m.GetSetFunc != nil {
		return m.GetSetFunc(ctx, setKey)
	}
	return []string{}, nil
}

func (m *MockCacheRepository) GetMultiple(ctx context.Context, keys []string) ([]*entity.Product, error) {
	if m.GetMultipleFunc != nil {
		return m.GetMultipleFunc(ctx, keys)
	}
	return []*entity.Product{}, nil
}

func (m *MockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, key)
	}
	return false, nil
}

func (m *MockCacheRepository) DeleteSet(ctx context.Context, setKey string) error {
	if m.DeleteSetFunc != nil {
		return m.DeleteSetFunc(ctx, setKey)
	}
	return nil
}

func (m *MockCacheRepository) HealthCheck(ctx context.Context) error {
	if m.HealthCheckFunc != nil {
		return m.HealthCheckFunc(ctx)
	}
	return nil
}

type MockCacheKeyGenerator struct{}

func (m *MockCacheKeyGenerator) ProductKey(id string) string {
	return "product_" + id
}

func (m *MockCacheKeyGenerator) NameKey(name string) string {
	return "product_by_name_" + name
}

func (m *MockCacheKeyGenerator) CategoryKey(category string) string {
	return "product_by_category_" + category
}

func (m *MockCacheKeyGenerator) AllProductsKey() string {
	return "all_products"
}

func newTestProduct() *entity.Product {
	product, _ := entity.NewProduct(
		"Test Product",
		"REF-001",
		"Electronics",
		"A test product",
		"SKU-001",
		"TestBrand",
		100,
		[]string{"image1.jpg"},
		map[string]interface{}{"color": "black"},
	)
	return product
}

func newTestProductWithData(name, ref, category string) *entity.Product {
	product, _ := entity.NewProduct(
		name,
		ref,
		category,
		"Description",
		"SKU-001",
		"Brand",
		50,
		[]string{},
		map[string]interface{}{},
	)
	return product
}
