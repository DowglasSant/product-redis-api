package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProductRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{
		pool: pool,
	}
}

func (r *PostgresProductRepository) Create(ctx context.Context, product *entity.Product) error {
	query := `
		INSERT INTO products (
			id, name, reference_number, category, description,
			sku, brand, stock, images, specifications,
			version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	specsJSON, err := json.Marshal(product.Specifications)
	if err != nil {
		return fmt.Errorf("failed to marshal specifications: %w", err)
	}

	_, err = r.pool.Exec(ctx, query,
		product.ID,
		product.Name,
		product.ReferenceNumber,
		product.Category,
		product.Description,
		product.SKU,
		product.Brand,
		product.Stock,
		imagesJSON,
		specsJSON,
		product.Version,
		product.CreatedAt,
		product.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return repository.ErrProductAlreadyExists
		}
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, product *entity.Product, expectedVersion int) error {
	query := `
		UPDATE products
		SET name = $1, category = $2, description = $3,
		    sku = $4, brand = $5, stock = $6,
		    images = $7, specifications = $8,
		    version = $9, updated_at = $10
		WHERE id = $11 AND version = $12
	`

	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}

	specsJSON, err := json.Marshal(product.Specifications)
	if err != nil {
		return fmt.Errorf("failed to marshal specifications: %w", err)
	}

	result, err := r.pool.Exec(ctx, query,
		product.Name,
		product.Category,
		product.Description,
		product.SKU,
		product.Brand,
		product.Stock,
		imagesJSON,
		specsJSON,
		product.Version,
		product.UpdatedAt,
		product.ID,
		expectedVersion,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	if result.RowsAffected() == 0 {
		exists, err := r.Exists(ctx, product.ID)
		if err != nil {
			return err
		}
		if !exists {
			return repository.ErrProductNotFound
		}
		return repository.ErrVersionConflict
	}

	return nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return repository.ErrProductNotFound
	}

	return nil
}

func (r *PostgresProductRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	query := `
		SELECT id, name, reference_number, category, description,
		       sku, brand, stock, images, specifications,
		       version, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product entity.Product
	var imagesJSON, specsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.ReferenceNumber,
		&product.Category,
		&product.Description,
		&product.SKU,
		&product.Brand,
		&product.Stock,
		&imagesJSON,
		&specsJSON,
		&product.Version,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
		return nil, fmt.Errorf("failed to unmarshal images: %w", err)
	}

	if err := json.Unmarshal(specsJSON, &product.Specifications); err != nil {
		return nil, fmt.Errorf("failed to unmarshal specifications: %w", err)
	}

	return &product, nil
}

func (r *PostgresProductRepository) FindAll(ctx context.Context, limit, offset int) ([]*entity.Product, error) {
	query := `
		SELECT id, name, reference_number, category, description,
		       sku, brand, stock, images, specifications,
		       version, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find all products: %w", err)
	}
	defer rows.Close()

	return r.scanProducts(rows)
}

func (r *PostgresProductRepository) FindByCategory(ctx context.Context, category string, limit, offset int) ([]*entity.Product, error) {
	query := `
		SELECT id, name, reference_number, category, description,
		       sku, brand, stock, images, specifications,
		       version, created_at, updated_at
		FROM products
		WHERE LOWER(category) = LOWER($1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find products by category: %w", err)
	}
	defer rows.Close()

	return r.scanProducts(rows)
}

func (r *PostgresProductRepository) FindByName(ctx context.Context, name string, limit, offset int) ([]*entity.Product, error) {
	query := `
		SELECT id, name, reference_number, category, description,
		       sku, brand, stock, images, specifications,
		       version, created_at, updated_at
		FROM products
		WHERE LOWER(name) LIKE LOWER($1)
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`

	searchPattern := "%" + name + "%"
	rows, err := r.pool.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find products by name: %w", err)
	}
	defer rows.Close()

	return r.scanProducts(rows)
}

func (r *PostgresProductRepository) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return exists, nil
}

func (r *PostgresProductRepository) HealthCheck(ctx context.Context) error {
	var result int
	err := r.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return repository.ErrDatabaseConnection
	}
	return nil
}

func (r *PostgresProductRepository) scanProducts(rows pgx.Rows) ([]*entity.Product, error) {
	var products []*entity.Product

	for rows.Next() {
		var product entity.Product
		var imagesJSON, specsJSON []byte

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.ReferenceNumber,
			&product.Category,
			&product.Description,
			&product.SKU,
			&product.Brand,
			&product.Stock,
			&imagesJSON,
			&specsJSON,
			&product.Version,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		if len(imagesJSON) > 0 {
			if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
				return nil, fmt.Errorf("failed to unmarshal images: %w", err)
			}
		}

		if len(specsJSON) > 0 {
			if err := json.Unmarshal(specsJSON, &product.Specifications); err != nil {
				return nil, fmt.Errorf("failed to unmarshal specifications: %w", err)
			}
		}

		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

func (r *PostgresProductRepository) GetPool() *pgxpool.Pool {
	return r.pool
}
