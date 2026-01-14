package dto

type CreateProductRequest struct {
	Name            string                 `json:"name"`
	ReferenceNumber string                 `json:"reference_number"`
	Category        string                 `json:"category"`
	Description     string                 `json:"description"`
	SKU             string                 `json:"sku"`
	Brand           string                 `json:"brand"`
	Stock           int                    `json:"stock"`
	Images          []string               `json:"images"`
	Specifications  map[string]interface{} `json:"specifications"`
}

type UpdateProductRequest struct {
	Name           string                 `json:"name"`
	Category       string                 `json:"category"`
	Description    string                 `json:"description"`
	SKU            string                 `json:"sku"`
	Brand          string                 `json:"brand"`
	Stock          int                    `json:"stock"`
	Images         []string               `json:"images"`
	Specifications map[string]interface{} `json:"specifications"`
}
