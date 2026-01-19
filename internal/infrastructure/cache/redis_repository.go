package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/redis/go-redis/v9"
)

var (
	ErrCacheNotFound = errors.New("cache entry not found")
	ErrCacheMiss     = errors.New("cache miss")
)

type RedisRepository struct {
	client     *redis.Client
	serializer Serializer
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client:     client,
		serializer: NewMsgpackSerializer(),
	}
}

func NewRedisRepositoryWithSerializer(client *redis.Client, serializer Serializer) *RedisRepository {
	return &RedisRepository{
		client:     client,
		serializer: serializer,
	}
}

func (r *RedisRepository) Get(ctx context.Context, key string) (*entity.Product, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheNotFound
		}
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	var product entity.Product
	if err := r.serializer.Unmarshal(data, &product); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	return &product, nil
}

func (r *RedisRepository) Set(ctx context.Context, key string, product *entity.Product) error {
	data, err := r.serializer.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product: %w", err)
	}

	err = r.client.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func (r *RedisRepository) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

func (r *RedisRepository) AddToSet(ctx context.Context, setKey, productID string) error {
	err := r.client.SAdd(ctx, setKey, productID).Err()
	if err != nil {
		return fmt.Errorf("failed to add to set: %w", err)
	}
	return nil
}

func (r *RedisRepository) RemoveFromSet(ctx context.Context, setKey, productID string) error {
	err := r.client.SRem(ctx, setKey, productID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove from set: %w", err)
	}
	return nil
}

func (r *RedisRepository) GetSet(ctx context.Context, setKey string) ([]string, error) {
	members, err := r.client.SMembers(ctx, setKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to get set members: %w", err)
	}
	return members, nil
}

func (r *RedisRepository) GetMultiple(ctx context.Context, keys []string) ([]*entity.Product, error) {
	if len(keys) == 0 {
		return []*entity.Product{}, nil
	}

	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	products := make([]*entity.Product, 0, len(keys))
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			return nil, fmt.Errorf("failed to get command result: %w", err)
		}

		var product entity.Product
		if err := r.serializer.Unmarshal(data, &product); err != nil {
			return nil, fmt.Errorf("failed to unmarshal product: %w", err)
		}

		products = append(products, &product)
	}

	return products, nil
}

func (r *RedisRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return count > 0, nil
}

func (r *RedisRepository) DeleteSet(ctx context.Context, setKey string) error {
	err := r.client.Del(ctx, setKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete set: %w", err)
	}
	return nil
}

func (r *RedisRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := r.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}

func (r *RedisRepository) GetClient() *redis.Client {
	return r.client
}

func (r *RedisRepository) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}
