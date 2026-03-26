package http

import (
	"errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/mzhryns/titik-nol-backend/internal/domain"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
)

func handleDomainError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound),
		errors.Is(err, domain.ErrTransactionNotFound),
		errors.Is(err, domain.ErrCategoryNotFound),
		errors.Is(err, domain.ErrAlreadyDeleted):
		response.NotFound(c, err.Error())
	case errors.Is(err, domain.ErrForbidden):
		response.NotFound(c, "resource not found")
	case errors.Is(err, domain.ErrInvalidCredentials):
		response.Error(c, 401, "Authentication failed", err.Error(), nil)
	case errors.Is(err, domain.ErrPasswordNotSet):
		response.Error(c, 401, "Authentication failed", err.Error(), nil)
	case errors.Is(err, domain.ErrInvalidTokenClaims):
		response.Error(c, 401, "Authentication failed", err.Error(), nil)
	case errors.Is(err, domain.ErrValidationFailed),
		errors.Is(err, domain.ErrNegativeBalance),
		errors.Is(err, domain.ErrEmptyBulkRequest),
		errors.Is(err, domain.ErrInvalidAccountType),
		errors.Is(err, domain.ErrInvalidTxType),
		errors.Is(err, domain.ErrInvalidCategoryType):
		response.BadRequest(c, "Validation failed", err.Error())
	default:
		slog.ErrorContext(c.Request.Context(), "Unhandled domain error", "error", err)
		response.InternalServerError(c, "Internal server error", err.Error())
	}
}
