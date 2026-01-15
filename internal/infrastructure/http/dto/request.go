package dto

// CreateProductRequest representa a requisição para criar um produto
// @Description Dados para criação de um novo produto
type CreateProductRequest struct {
	Name            string                 `json:"name" example:"iPhone 15 Pro"`
	ReferenceNumber string                 `json:"reference_number" example:"REF-12345"`
	Category        string                 `json:"category" example:"electronics"`
	Description     string                 `json:"description" example:"Smartphone Apple com chip A17 Pro"`
	SKU             string                 `json:"sku" example:"SKU-IP15P-256"`
	Brand           string                 `json:"brand" example:"Apple"`
	Stock           int                    `json:"stock" example:"100"`
	Images          []string               `json:"images" example:"https://example.com/image1.jpg,https://example.com/image2.jpg"`
	Specifications  map[string]interface{} `json:"specifications"`
}

// UpdateProductRequest representa a requisição para atualizar um produto
// @Description Dados para atualização de um produto existente
type UpdateProductRequest struct {
	Name           string                 `json:"name" example:"iPhone 15 Pro Max"`
	Category       string                 `json:"category" example:"electronics"`
	Description    string                 `json:"description" example:"Smartphone Apple com chip A17 Pro"`
	SKU            string                 `json:"sku" example:"SKU-IP15PM-256"`
	Brand          string                 `json:"brand" example:"Apple"`
	Stock          int                    `json:"stock" example:"50"`
	Images         []string               `json:"images" example:"https://example.com/image1.jpg"`
	Specifications map[string]interface{} `json:"specifications"`
}
