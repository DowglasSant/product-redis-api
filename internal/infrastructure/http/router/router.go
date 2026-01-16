package router

import (
	"net/http"

	"github.com/dowglassantana/product-redis-api/internal/infrastructure/http/handler"
	"github.com/dowglassantana/product-redis-api/internal/infrastructure/http/middleware"
	customlogger "github.com/dowglassantana/product-redis-api/internal/infrastructure/logger"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

func SetupRouter(
	productHandler *handler.ProductHandler,
	healthHandler *handler.HealthHandler,
	jwtAuth *middleware.JWTAuth,
	atomicLevel *zap.AtomicLevel,
	logger *zap.Logger,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))
	r.Use(chimiddleware.Compress(5))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health/live", healthHandler.Liveness)
	r.Get("/health/ready", healthHandler.Readiness)
	r.Handle("/metrics", promhttp.Handler())

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	logLevelHandler := customlogger.NewAtomicLevelServer(atomicLevel)
	r.HandleFunc("/log/level", logLevelHandler.ServeHTTP)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(jwtAuth.Middleware)

		r.Route("/products", func(r chi.Router) {
			r.Get("/", productHandler.List)
			r.Post("/", productHandler.Create)
			r.Get("/{id}", productHandler.Get)
			r.Put("/{id}", productHandler.Update)
			r.Delete("/{id}", productHandler.Delete)

			r.Get("/search/name", productHandler.SearchByName)
			r.Get("/search/category", productHandler.SearchByCategory)
		})
	})

	return r
}
