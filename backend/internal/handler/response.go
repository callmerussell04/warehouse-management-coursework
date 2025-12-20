package handler

import (
	"errors"
	"log/slog"
	"net/http"

	customErrors "warehouse-management-system/internal/errors"

	"github.com/gin-gonic/gin"
)

func RespondWithError(c *gin.Context, logger *slog.Logger, err error) {
	var status int
	var message string

	switch {
	case errors.Is(err, customErrors.ErrNotFound):
		status = http.StatusNotFound
		message = err.Error()
	case errors.Is(err, customErrors.ErrAlreadyExists):
		status = http.StatusConflict
		message = err.Error()
	case errors.Is(err, customErrors.ErrInvalidInput):
		status = http.StatusBadRequest
		message = err.Error()
	case errors.Is(err, customErrors.ErrInsufficientStock):
		status = http.StatusConflict
		message = err.Error()
	case errors.Is(err, customErrors.ErrConflict):
		status = http.StatusConflict
		message = err.Error()
	case errors.Is(err, customErrors.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = err.Error()
	default:
		logger.Error("internal server error", "error", err)
		status = http.StatusInternalServerError
		message = "internal server error"
	}

	c.AbortWithStatusJSON(status, gin.H{"error": message})
}
