package handler

import (
	"net/http"
	uccore "seanAIgent/internal/booking/usecase/core"

	"github.com/gin-gonic/gin"
)

func ErrorHandler(c *gin.Context, err uccore.UseCaseError) {
	if err == nil {
		return
	}

	c.JSON(GetStatus(err.Type()), gin.H{
		"error":    err.Error(),
		"category": err.Category(),
		"code":     err.Code(),
		"kind":     err.Kind(),
		"message":  err.Message(),
	})
}

func GetStatus(typ uccore.ErrorType) int {
	switch typ {
	case uccore.ErrInternal:
		return http.StatusInternalServerError
	case uccore.ErrInvalidInput:
		return http.StatusBadRequest
	case uccore.ErrNotFound:
		return http.StatusNotFound
	case uccore.ErrConflict:
		return http.StatusConflict
	case uccore.ErrForbidden:
		return http.StatusForbidden
	case uccore.ErrPermissionDenied:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
