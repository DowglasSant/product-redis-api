package entity

import (
	"crypto/sha256"
	"errors"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	ErrInvalidProduct   = errors.New("invalid product")
	ErrInvalidName      = errors.New("product name is required")
	ErrInvalidReference = errors.New("product reference is required")
	ErrInvalidCategory  = errors.New("product category is required")
	ErrInvalidStock     = errors.New("product stock cannot be negative")
	ErrVersionConflict  = errors.New("product version conflict - concurrent modification detected")
)

type Product struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	ReferenceNumber string                 `json:"reference_number"`
	Category        string                 `json:"category"`
	Description     string                 `json:"description"`
	SKU             string                 `json:"sku"`
	Brand           string                 `json:"brand"`
	Stock           int                    `json:"stock"`
	Images          []string               `json:"images"`
	Specifications  map[string]interface{} `json:"specifications"`
	Version         int                    `json:"version"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

func NewProduct(name, referenceNumber, category, description, sku, brand string, stock int, images []string, specs map[string]interface{}) (*Product, error) {
	p := &Product{
		Name:            strings.TrimSpace(name),
		ReferenceNumber: strings.TrimSpace(referenceNumber),
		Category:        strings.TrimSpace(category),
		Description:     strings.TrimSpace(description),
		SKU:             strings.TrimSpace(sku),
		Brand:           strings.TrimSpace(brand),
		Stock:           stock,
		Images:          images,
		Specifications:  specs,
		Version:         1,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	p.ID = GenerateProductID(p.Name, p.ReferenceNumber)

	return p, nil
}

func (p *Product) Validate() error {
	if p.Name == "" {
		return ErrInvalidName
	}
	if p.ReferenceNumber == "" {
		return ErrInvalidReference
	}
	if p.Category == "" {
		return ErrInvalidCategory
	}
	if p.Stock < 0 {
		return ErrInvalidStock
	}
	return nil
}

func (p *Product) Update(name, category, description, sku, brand string, stock int, images []string, specs map[string]interface{}) error {
	p.Name = strings.TrimSpace(name)
	p.Category = strings.TrimSpace(category)
	p.Description = strings.TrimSpace(description)
	p.SKU = strings.TrimSpace(sku)
	p.Brand = strings.TrimSpace(brand)
	p.Stock = stock
	p.Images = images
	p.Specifications = specs
	p.UpdatedAt = time.Now().UTC()
	p.Version++

	return p.Validate()
}

func (p *Product) Equals(other *Product) bool {
	if other == nil {
		return false
	}

	if p.Name != other.Name ||
		p.ReferenceNumber != other.ReferenceNumber ||
		p.Category != other.Category ||
		p.Description != other.Description ||
		p.SKU != other.SKU ||
		p.Brand != other.Brand ||
		p.Stock != other.Stock {
		return false
	}

	if len(p.Images) != len(other.Images) {
		return false
	}
	for i := range p.Images {
		if p.Images[i] != other.Images[i] {
			return false
		}
	}

	if len(p.Specifications) != len(other.Specifications) {
		return false
	}
	for key, val := range p.Specifications {
		otherVal, exists := other.Specifications[key]
		if !exists || val != otherVal {
			return false
		}
	}

	return true
}

func GenerateProductID(name, referenceNumber string) string {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	normalizedRef := strings.ToLower(strings.TrimSpace(referenceNumber))
	seed := normalizedName + "|" + normalizedRef
	hash := sha256.Sum256([]byte(seed))
	entropy := hash[:16]
	id := ulid.MustNew(0, &deterministicReader{data: entropy})
	return id.String()
}

type deterministicReader struct {
	data []byte
	pos  int
}

func (r *deterministicReader) Read(p []byte) (n int, err error) {
	n = copy(p, r.data[r.pos:])
	r.pos += n
	if r.pos >= len(r.data) {
		r.pos = 0
	}
	return n, nil
}

func (p *Product) HashID() string {
	if len(p.ID) >= 8 {
		return p.ID[:8]
	}
	return p.ID
}
