package port

type CacheKeyGenerator interface {
	ProductKey(id string) string
	NameKey(name string) string
	CategoryKey(category string) string
	AllProductsKey() string
}
