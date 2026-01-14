package entity

import (
	"testing"
)

func TestNewProduct(t *testing.T) {
	tests := []struct {
		name            string
		productName     string
		referenceNumber string
		category        string
		description     string
		sku             string
		brand           string
		stock           int
		wantErr         bool
		expectedErr     error
	}{
		{
			name:            "valid product",
			productName:     "iPhone 15 Pro",
			referenceNumber: "APL-IP15P-001",
			category:        "Smartphones",
			description:     "Latest iPhone",
			sku:             "APPLE-IP15P",
			brand:           "Apple",
			stock:           50,
			wantErr:         false,
		},
		{
			name:            "missing name",
			productName:     "",
			referenceNumber: "APL-IP15P-001",
			category:        "Smartphones",
			description:     "Latest iPhone",
			sku:             "APPLE-IP15P",
			brand:           "Apple",
			stock:           50,
			wantErr:         true,
			expectedErr:     ErrInvalidName,
		},
		{
			name:            "missing reference number",
			productName:     "iPhone 15 Pro",
			referenceNumber: "",
			category:        "Smartphones",
			description:     "Latest iPhone",
			sku:             "APPLE-IP15P",
			brand:           "Apple",
			stock:           50,
			wantErr:         true,
			expectedErr:     ErrInvalidReference,
		},
		{
			name:            "missing category",
			productName:     "iPhone 15 Pro",
			referenceNumber: "APL-IP15P-001",
			category:        "",
			description:     "Latest iPhone",
			sku:             "APPLE-IP15P",
			brand:           "Apple",
			stock:           50,
			wantErr:         true,
			expectedErr:     ErrInvalidCategory,
		},
		{
			name:            "negative stock",
			productName:     "iPhone 15 Pro",
			referenceNumber: "APL-IP15P-001",
			category:        "Smartphones",
			description:     "Latest iPhone",
			sku:             "APPLE-IP15P",
			brand:           "Apple",
			stock:           -10,
			wantErr:         true,
			expectedErr:     ErrInvalidStock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := NewProduct(
				tt.productName,
				tt.referenceNumber,
				tt.category,
				tt.description,
				tt.sku,
				tt.brand,
				tt.stock,
				[]string{},
				map[string]interface{}{},
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewProduct() expected error but got none")
					return
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("NewProduct() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewProduct() unexpected error = %v", err)
				return
			}

			if product == nil {
				t.Error("NewProduct() returned nil product")
				return
			}

			// Verify ID was generated
			if product.ID == "" {
				t.Error("NewProduct() ID was not generated")
			}

			// Verify version is 1
			if product.Version != 1 {
				t.Errorf("NewProduct() version = %d, want 1", product.Version)
			}

			// Verify fields
			if product.Name != tt.productName {
				t.Errorf("NewProduct() name = %s, want %s", product.Name, tt.productName)
			}
		})
	}
}

func TestGenerateProductID(t *testing.T) {
	tests := []struct {
		name            string
		productName     string
		referenceNumber string
		wantSameID      bool
	}{
		{
			name:            "same inputs generate same ID",
			productName:     "iPhone 15 Pro",
			referenceNumber: "APL-IP15P-001",
			wantSameID:      true,
		},
		{
			name:            "case insensitive",
			productName:     "IPHONE 15 PRO",
			referenceNumber: "apl-ip15p-001",
			wantSameID:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id1 := GenerateProductID(tt.productName, tt.referenceNumber)
			id2 := GenerateProductID("iPhone 15 Pro", "APL-IP15P-001")

			if tt.wantSameID && id1 != id2 {
				t.Errorf("GenerateProductID() expected same IDs but got different: %s vs %s", id1, id2)
			}

			// Verify ID is valid ULID format (26 characters)
			if len(id1) != 26 {
				t.Errorf("GenerateProductID() ID length = %d, want 26", len(id1))
			}
		})
	}
}

func TestProductEquals(t *testing.T) {
	product1, _ := NewProduct(
		"iPhone 15 Pro",
		"APL-IP15P-001",
		"Smartphones",
		"Latest iPhone",
		"APPLE-IP15P",
		"Apple",
		50,
		[]string{"img1.jpg"},
		map[string]interface{}{"storage": "256GB"},
	)

	product2, _ := NewProduct(
		"iPhone 15 Pro",
		"APL-IP15P-001",
		"Smartphones",
		"Latest iPhone",
		"APPLE-IP15P",
		"Apple",
		50,
		[]string{"img1.jpg"},
		map[string]interface{}{"storage": "256GB"},
	)

	product3, _ := NewProduct(
		"iPhone 15 Pro",
		"APL-IP15P-001",
		"Smartphones",
		"Different description",
		"APPLE-IP15P",
		"Apple",
		50,
		[]string{"img1.jpg"},
		map[string]interface{}{"storage": "256GB"},
	)

	tests := []struct {
		name     string
		p1       *Product
		p2       *Product
		expected bool
	}{
		{
			name:     "identical products",
			p1:       product1,
			p2:       product2,
			expected: true,
		},
		{
			name:     "different description",
			p1:       product1,
			p2:       product3,
			expected: false,
		},
		{
			name:     "nil comparison",
			p1:       product1,
			p2:       nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.p1.Equals(tt.p2)
			if result != tt.expected {
				t.Errorf("Product.Equals() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProductUpdate(t *testing.T) {
	product, _ := NewProduct(
		"iPhone 15 Pro",
		"APL-IP15P-001",
		"Smartphones",
		"Latest iPhone",
		"APPLE-IP15P",
		"Apple",
		50,
		[]string{"img1.jpg"},
		map[string]interface{}{"storage": "256GB"},
	)

	oldVersion := product.Version

	err := product.Update(
		"iPhone 15 Pro",
		"Smartphones",
		"Updated description",
		"APPLE-IP15P",
		"Apple",
		45,
		[]string{"img1.jpg", "img2.jpg"},
		map[string]interface{}{"storage": "256GB", "color": "Titanium"},
	)

	if err != nil {
		t.Errorf("Product.Update() unexpected error = %v", err)
	}

	if product.Version != oldVersion+1 {
		t.Errorf("Product.Update() version = %d, want %d", product.Version, oldVersion+1)
	}

	if product.Description != "Updated description" {
		t.Errorf("Product.Update() description not updated")
	}

	if product.Stock != 45 {
		t.Errorf("Product.Update() stock = %d, want 45", product.Stock)
	}
}
