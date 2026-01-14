package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dowglassantana/product-redis-api/internal/application/port"
	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
	"github.com/dowglassantana/product-redis-api/internal/infrastructure/http/dto"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type ProductHandler struct {
	createUseCase           port.ProductCreator
	updateUseCase           port.ProductUpdater
	deleteUseCase           port.ProductDeleter
	getUseCase              port.ProductGetter
	listUseCase             port.ProductLister
	searchByNameUseCase     port.ProductSearcherByName
	searchByCategoryUseCase port.ProductSearcherByCategory
	logger                  *zap.Logger
}

func NewProductHandler(
	createUseCase port.ProductCreator,
	updateUseCase port.ProductUpdater,
	deleteUseCase port.ProductDeleter,
	getUseCase port.ProductGetter,
	listUseCase port.ProductLister,
	searchByNameUseCase port.ProductSearcherByName,
	searchByCategoryUseCase port.ProductSearcherByCategory,
	logger *zap.Logger,
) *ProductHandler {
	return &ProductHandler{
		createUseCase:           createUseCase,
		updateUseCase:           updateUseCase,
		deleteUseCase:           deleteUseCase,
		getUseCase:              getUseCase,
		listUseCase:             listUseCase,
		searchByNameUseCase:     searchByNameUseCase,
		searchByCategoryUseCase: searchByCategoryUseCase,
		logger:                  logger,
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", err)
		return
	}

	input := port.CreateProductInput{
		Name:            req.Name,
		ReferenceNumber: req.ReferenceNumber,
		Category:        req.Category,
		Description:     req.Description,
		SKU:             req.SKU,
		Brand:           req.Brand,
		Stock:           req.Stock,
		Images:          req.Images,
		Specifications:  req.Specifications,
	}

	product, err := h.createUseCase.Execute(r.Context(), input)
	if err != nil {
		if errors.Is(err, repository.ErrProductAlreadyExists) {
			h.respondError(w, http.StatusConflict, "product_exists", "Product already exists", err)
			return
		}
		if errors.Is(err, entity.ErrInvalidName) || errors.Is(err, entity.ErrInvalidReference) ||
			errors.Is(err, entity.ErrInvalidCategory) || errors.Is(err, entity.ErrInvalidStock) {
			h.respondError(w, http.StatusBadRequest, "validation_error", err.Error(), err)
			return
		}
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to create product", err)
		return
	}

	h.respondJSON(w, http.StatusCreated, dto.ToProductResponse(product))
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Product ID is required", nil)
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", err)
		return
	}

	input := port.UpdateProductInput{
		Name:           req.Name,
		Category:       req.Category,
		Description:    req.Description,
		SKU:            req.SKU,
		Brand:          req.Brand,
		Stock:          req.Stock,
		Images:         req.Images,
		Specifications: req.Specifications,
	}

	product, err := h.updateUseCase.Execute(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			h.respondError(w, http.StatusNotFound, "product_not_found", "Product not found", err)
			return
		}
		if errors.Is(err, repository.ErrVersionConflict) {
			h.respondError(w, http.StatusConflict, "version_conflict", "Product was modified by another process", err)
			return
		}
		if errors.Is(err, entity.ErrInvalidName) || errors.Is(err, entity.ErrInvalidCategory) ||
			errors.Is(err, entity.ErrInvalidStock) {
			h.respondError(w, http.StatusBadRequest, "validation_error", err.Error(), err)
			return
		}
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to update product", err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToProductResponse(product))
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Product ID is required", nil)
		return
	}

	if err := h.deleteUseCase.Execute(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			h.respondError(w, http.StatusNotFound, "product_not_found", "Product not found", err)
			return
		}
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to delete product", err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.SuccessResponse{
		Message: "Product deleted successfully",
	})
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_id", "Product ID is required", nil)
		return
	}

	product, err := h.getUseCase.Execute(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			h.respondError(w, http.StatusNotFound, "product_not_found", "Product not found", err)
			return
		}
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to get product", err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToProductResponse(product))
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := h.getPagination(r)

	products, err := h.listUseCase.Execute(r.Context(), limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to list products", err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToProductResponseList(products))
}

func (h *ProductHandler) SearchByName(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("q")
	if name == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_query", "Search query is required", nil)
		return
	}

	limit, offset := h.getPagination(r)

	products, err := h.searchByNameUseCase.Execute(r.Context(), name, limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to search products", err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToProductResponseList(products))
}

func (h *ProductHandler) SearchByCategory(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("q")
	if category == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_query", "Category query is required", nil)
		return
	}

	limit, offset := h.getPagination(r)

	products, err := h.searchByCategoryUseCase.Execute(r.Context(), category, limit, offset)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "internal_error", "Failed to search products", err)
		return
	}

	h.respondJSON(w, http.StatusOK, dto.ToProductResponseList(products))
}

func (h *ProductHandler) getPagination(r *http.Request) (limit, offset int) {
	limit = 50 // default
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

func (h *ProductHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *ProductHandler) respondError(w http.ResponseWriter, status int, code, message string, err error) {
	if err != nil {
		h.logger.Error("request error",
			zap.String("code", code),
			zap.String("message", message),
			zap.Error(err),
		)
	}

	h.respondJSON(w, status, dto.ErrorResponse{
		Error:   code,
		Message: message,
	})
}
