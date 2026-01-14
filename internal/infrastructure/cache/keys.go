package cache

import "strings"

type RedisCacheKeyGenerator struct{}

func NewRedisCacheKeyGenerator() *RedisCacheKeyGenerator {
	return &RedisCacheKeyGenerator{}
}

func (g *RedisCacheKeyGenerator) ProductKey(id string) string {
	return "product_" + id
}

func (g *RedisCacheKeyGenerator) NameKey(name string) string {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	return "product_by_name_" + normalizedName
}

func (g *RedisCacheKeyGenerator) CategoryKey(category string) string {
	normalizedCategory := strings.ToLower(strings.TrimSpace(category))
	return "product_by_category_" + normalizedCategory
}

func (g *RedisCacheKeyGenerator) AllProductsKey() string {
	return "all_products"
}
