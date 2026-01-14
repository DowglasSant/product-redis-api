package utils

import "github.com/dowglassantana/product-redis-api/internal/domain/entity"

func PaginateProducts(products []*entity.Product, limit, offset int) []*entity.Product {
	if offset >= len(products) {
		return []*entity.Product{}
	}

	end := offset + limit
	if end > len(products) {
		end = len(products)
	}

	return products[offset:end]
}
