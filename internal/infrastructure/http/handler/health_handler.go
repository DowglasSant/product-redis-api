package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
	"go.uber.org/zap"
)

type HealthHandler struct {
	productRepo repository.ProductRepository
	cacheRepo   repository.CacheRepository
	logger      *zap.Logger
}

func NewHealthHandler(
	productRepo repository.ProductRepository,
	cacheRepo repository.CacheRepository,
	logger *zap.Logger,
) *HealthHandler {
	return &HealthHandler{
		productRepo: productRepo,
		cacheRepo:   cacheRepo,
		logger:      logger,
	}
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	allHealthy := true

	if err := h.productRepo.HealthCheck(ctx); err != nil {
		services["database"] = "unhealthy"
		allHealthy = false
		h.logger.Warn("database health check failed", zap.Error(err))
	} else {
		services["database"] = "healthy"
	}

	if err := h.cacheRepo.HealthCheck(ctx); err != nil {
		services["cache"] = "unhealthy"
		allHealthy = false
		h.logger.Warn("cache health check failed", zap.Error(err))
	} else {
		services["cache"] = "healthy"
	}

	status := "healthy"
	statusCode := http.StatusOK
	if !allHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Services:  services,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
