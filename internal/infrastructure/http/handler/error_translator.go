package handler

import (
	"errors"
	"net/http"

	"github.com/dowglassantana/product-redis-api/internal/domain/entity"
	"github.com/dowglassantana/product-redis-api/internal/domain/repository"
)

// HTTPError representa um erro HTTP traduzido do domínio.
type HTTPError struct {
	StatusCode int
	Code       string
	Message    string
}

// TranslateDomainError traduz erros de domínio para erros HTTP.
// Isso centraliza a lógica de mapeamento e desacopla o handler
// de conhecer detalhes específicos dos erros de domínio.
func TranslateDomainError(err error) *HTTPError {
	if err == nil {
		return nil
	}

	// Erros de repositório
	if errors.Is(err, repository.ErrProductNotFound) {
		return &HTTPError{
			StatusCode: http.StatusNotFound,
			Code:       "product_not_found",
			Message:    "Product not found",
		}
	}

	if errors.Is(err, repository.ErrProductAlreadyExists) {
		return &HTTPError{
			StatusCode: http.StatusConflict,
			Code:       "product_exists",
			Message:    "Product already exists",
		}
	}

	if errors.Is(err, repository.ErrVersionConflict) {
		return &HTTPError{
			StatusCode: http.StatusConflict,
			Code:       "version_conflict",
			Message:    "Product was modified by another process",
		}
	}

	// Erros de validação de entidade
	if errors.Is(err, entity.ErrInvalidName) {
		return &HTTPError{
			StatusCode: http.StatusBadRequest,
			Code:       "validation_error",
			Message:    "Invalid product name",
		}
	}

	if errors.Is(err, entity.ErrInvalidReference) {
		return &HTTPError{
			StatusCode: http.StatusBadRequest,
			Code:       "validation_error",
			Message:    "Invalid reference number",
		}
	}

	if errors.Is(err, entity.ErrInvalidCategory) {
		return &HTTPError{
			StatusCode: http.StatusBadRequest,
			Code:       "validation_error",
			Message:    "Invalid category",
		}
	}

	if errors.Is(err, entity.ErrInvalidStock) {
		return &HTTPError{
			StatusCode: http.StatusBadRequest,
			Code:       "validation_error",
			Message:    "Invalid stock value",
		}
	}

	// Erro desconhecido - retorna nil para que o handler trate como erro interno
	return nil
}

// IsValidationError verifica se o erro é um erro de validação de entidade.
func IsValidationError(err error) bool {
	return errors.Is(err, entity.ErrInvalidName) ||
		errors.Is(err, entity.ErrInvalidReference) ||
		errors.Is(err, entity.ErrInvalidCategory) ||
		errors.Is(err, entity.ErrInvalidStock)
}

// IsNotFoundError verifica se o erro é um erro de não encontrado.
func IsNotFoundError(err error) bool {
	return errors.Is(err, repository.ErrProductNotFound)
}

// IsConflictError verifica se o erro é um erro de conflito.
func IsConflictError(err error) bool {
	return errors.Is(err, repository.ErrProductAlreadyExists) ||
		errors.Is(err, repository.ErrVersionConflict)
}
