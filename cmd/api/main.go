package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/dowglassantana/product-redis-api/docs"
	"github.com/dowglassantana/product-redis-api/internal/application/usecase"
	"github.com/dowglassantana/product-redis-api/internal/infrastructure/cache"
	"github.com/dowglassantana/product-redis-api/internal/infrastructure/config"
	"github.com/dowglassantana/product-redis-api/internal/infrastructure/database"
	"github.com/dowglassantana/product-redis-api/internal/infrastructure/http/handler"
	"github.com/dowglassantana/product-redis-api/internal/infrastructure/http/middleware"
	"github.com/dowglassantana/product-redis-api/internal/infrastructure/http/router"
	"github.com/dowglassantana/product-redis-api/internal/infrastructure/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// @title Product API
// @version 1.0
// @description API de gerenciamento de produtos com cache Redis integrado
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log, atomicLevel, err := logger.NewLogger(cfg.App.LogLevel, cfg.App.Environment)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("starting product API",
		zap.String("environment", cfg.App.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	dbPool, err := initDatabase(cfg.Database)
	if err != nil {
		log.Fatal("failed to initialize database", zap.Error(err))
	}
	defer dbPool.Close()
	log.Info("database connection established")

	redisClient, err := initRedis(cfg.Redis)
	if err != nil {
		log.Fatal("failed to initialize redis", zap.Error(err))
	}
	defer redisClient.Close()
	log.Info("redis connection established")

	productRepo := database.NewPostgresProductRepository(dbPool)
	cacheRepo := cache.NewRedisRepository(redisClient)
	cacheKeys := cache.NewRedisCacheKeyGenerator()

	appLogger := logger.NewZapAdapter(log)

	createUseCase := usecase.NewCreateProductUseCase(productRepo, cacheRepo, cacheKeys, appLogger)
	updateUseCase := usecase.NewUpdateProductUseCase(productRepo, cacheRepo, cacheKeys, appLogger)
	deleteUseCase := usecase.NewDeleteProductUseCase(productRepo, cacheRepo, cacheKeys, appLogger)
	getUseCase := usecase.NewGetProductUseCase(productRepo, cacheRepo, cacheKeys, appLogger)
	listUseCase := usecase.NewListProductsUseCase(productRepo, cacheRepo, cacheKeys, appLogger)
	searchByNameUseCase := usecase.NewSearchProductsByNameUseCase(productRepo, cacheRepo, cacheKeys, appLogger)
	searchByCategoryUseCase := usecase.NewSearchProductsByCategoryUseCase(productRepo, cacheRepo, cacheKeys, appLogger)

	productHandler := handler.NewProductHandler(
		createUseCase,
		updateUseCase,
		deleteUseCase,
		getUseCase,
		listUseCase,
		searchByNameUseCase,
		searchByCategoryUseCase,
		log,
	)
	healthHandler := handler.NewHealthHandler(productRepo, cacheRepo, log)

	jwtAuth := middleware.NewJWTAuth(&cfg.Keycloak, log)

	r := router.SetupRouter(productHandler, healthHandler, jwtAuth, atomicLevel, log)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("server listening", zap.String("address", srv.Addr))
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatal("server error", zap.Error(err))

	case sig := <-shutdown:
		log.Info("shutdown signal received", zap.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("graceful shutdown failed", zap.Error(err))
			if err := srv.Close(); err != nil {
				log.Fatal("server close failed", zap.Error(err))
			}
		}

		log.Info("server stopped gracefully")
	}
}

func initDatabase(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func initRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
