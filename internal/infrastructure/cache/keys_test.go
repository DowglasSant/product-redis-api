package cache

import "testing"

func TestRedisCacheKeyGenerator_ProductKey(t *testing.T) {
	g := NewRedisCacheKeyGenerator()

	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "simple id",
			id:       "abc123",
			expected: "product_abc123",
		},
		{
			name:     "uuid-like id",
			id:       "550e8400-e29b-41d4-a716-446655440000",
			expected: "product_550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "empty id",
			id:       "",
			expected: "product_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.ProductKey(tt.id)
			if result != tt.expected {
				t.Errorf("ProductKey(%s) = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}
}

func TestRedisCacheKeyGenerator_NameKey(t *testing.T) {
	g := NewRedisCacheKeyGenerator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "iPhone",
			expected: "product_by_name_iphone",
		},
		{
			name:     "name with spaces",
			input:    "iPhone 15 Pro",
			expected: "product_by_name_iphone 15 pro",
		},
		{
			name:     "uppercase name",
			input:    "SAMSUNG GALAXY",
			expected: "product_by_name_samsung galaxy",
		},
		{
			name:     "mixed case name",
			input:    "MacBook Pro",
			expected: "product_by_name_macbook pro",
		},
		{
			name:     "name with leading/trailing spaces",
			input:    "  iPhone  ",
			expected: "product_by_name_iphone",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "product_by_name_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.NameKey(tt.input)
			if result != tt.expected {
				t.Errorf("NameKey(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRedisCacheKeyGenerator_CategoryKey(t *testing.T) {
	g := NewRedisCacheKeyGenerator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple category",
			input:    "Electronics",
			expected: "product_by_category_electronics",
		},
		{
			name:     "category with spaces",
			input:    "Home Appliances",
			expected: "product_by_category_home appliances",
		},
		{
			name:     "uppercase category",
			input:    "SMARTPHONES",
			expected: "product_by_category_smartphones",
		},
		{
			name:     "mixed case category",
			input:    "Gaming Accessories",
			expected: "product_by_category_gaming accessories",
		},
		{
			name:     "category with leading/trailing spaces",
			input:    "  Laptops  ",
			expected: "product_by_category_laptops",
		},
		{
			name:     "empty category",
			input:    "",
			expected: "product_by_category_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.CategoryKey(tt.input)
			if result != tt.expected {
				t.Errorf("CategoryKey(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRedisCacheKeyGenerator_AllProductsKey(t *testing.T) {
	g := NewRedisCacheKeyGenerator()

	result := g.AllProductsKey()
	expected := "all_products"

	if result != expected {
		t.Errorf("AllProductsKey() = %s, want %s", result, expected)
	}
}

func TestRedisCacheKeyGenerator_KeyConsistency(t *testing.T) {
	g := NewRedisCacheKeyGenerator()

	name1 := g.NameKey("iPhone")
	name2 := g.NameKey("iphone")
	name3 := g.NameKey("IPHONE")

	if name1 != name2 || name2 != name3 {
		t.Error("NameKey should produce consistent keys regardless of case")
	}

	cat1 := g.CategoryKey("Electronics")
	cat2 := g.CategoryKey("electronics")
	cat3 := g.CategoryKey("ELECTRONICS")

	if cat1 != cat2 || cat2 != cat3 {
		t.Error("CategoryKey should produce consistent keys regardless of case")
	}
}
