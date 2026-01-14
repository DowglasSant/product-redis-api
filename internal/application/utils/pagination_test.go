package utils

import (
	"testing"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
)

func createTestProducts(count int) []*entity.Product {
	products := make([]*entity.Product, count)
	for i := 0; i < count; i++ {
		products[i] = &entity.Product{
			ID:   "id-" + string(rune('a'+i)),
			Name: "Product " + string(rune('A'+i)),
		}
	}
	return products
}

func TestPaginateProducts_BasicPagination(t *testing.T) {
	products := createTestProducts(10)

	result := PaginateProducts(products, 5, 0)

	if len(result) != 5 {
		t.Errorf("Expected 5 products, got %d", len(result))
	}
}

func TestPaginateProducts_WithOffset(t *testing.T) {
	products := createTestProducts(10)

	result := PaginateProducts(products, 5, 5)

	if len(result) != 5 {
		t.Errorf("Expected 5 products, got %d", len(result))
	}

	if result[0].ID != products[5].ID {
		t.Errorf("Expected first product to be %s, got %s", products[5].ID, result[0].ID)
	}
}

func TestPaginateProducts_OffsetBeyondLength(t *testing.T) {
	products := createTestProducts(5)

	result := PaginateProducts(products, 10, 10)

	if len(result) != 0 {
		t.Errorf("Expected 0 products when offset beyond length, got %d", len(result))
	}
}

func TestPaginateProducts_LimitBeyondRemaining(t *testing.T) {
	products := createTestProducts(10)

	result := PaginateProducts(products, 100, 5)

	if len(result) != 5 {
		t.Errorf("Expected 5 products (remaining after offset), got %d", len(result))
	}
}

func TestPaginateProducts_EmptySlice(t *testing.T) {
	products := []*entity.Product{}

	result := PaginateProducts(products, 10, 0)

	if len(result) != 0 {
		t.Errorf("Expected 0 products from empty slice, got %d", len(result))
	}
}

func TestPaginateProducts_ZeroLimit(t *testing.T) {
	products := createTestProducts(5)

	result := PaginateProducts(products, 0, 0)

	if len(result) != 0 {
		t.Errorf("Expected 0 products with limit 0, got %d", len(result))
	}
}

func TestPaginateProducts_ZeroOffset(t *testing.T) {
	products := createTestProducts(5)

	result := PaginateProducts(products, 3, 0)

	if len(result) != 3 {
		t.Errorf("Expected 3 products, got %d", len(result))
	}

	if result[0].ID != products[0].ID {
		t.Errorf("Expected first product to be %s, got %s", products[0].ID, result[0].ID)
	}
}

func TestPaginateProducts_ExactLength(t *testing.T) {
	products := createTestProducts(5)

	result := PaginateProducts(products, 5, 0)

	if len(result) != 5 {
		t.Errorf("Expected 5 products, got %d", len(result))
	}
}

func TestPaginateProducts_LastPage(t *testing.T) {
	products := createTestProducts(10)

	result := PaginateProducts(products, 3, 9)

	if len(result) != 1 {
		t.Errorf("Expected 1 product on last page, got %d", len(result))
	}

	if result[0].ID != products[9].ID {
		t.Errorf("Expected last product to be %s, got %s", products[9].ID, result[0].ID)
	}
}

func TestPaginateProducts_OffsetEqualsLength(t *testing.T) {
	products := createTestProducts(5)

	result := PaginateProducts(products, 10, 5)

	if len(result) != 0 {
		t.Errorf("Expected 0 products when offset equals length, got %d", len(result))
	}
}

func TestPaginateProducts_PreservesOrder(t *testing.T) {
	products := createTestProducts(5)

	result := PaginateProducts(products, 3, 1)

	for i := 0; i < len(result); i++ {
		if result[i].ID != products[i+1].ID {
			t.Errorf("Order not preserved at index %d: expected %s, got %s", i, products[i+1].ID, result[i].ID)
		}
	}
}
