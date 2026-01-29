package handler

import (
	"net/http"
	uccore "seanAIgent/internal/booking/usecase/core"

	"github.com/gin-gonic/gin"
)

func errorHandler(c *gin.Context, err uccore.UseCaseError) {
	if err == nil {
		return
	}

	appErr := err.(uccore.UseCaseError)
	c.JSON(getStatus(appErr.Type()), gin.H{
		"error":    appErr.Error(),
		"category": appErr.Category(),
		"code":     appErr.Code(),
		"kind":     appErr.Kind(),
		"message":  appErr.Message(),
	})
}

func getStatus(typ uccore.ErrorType) int {
	switch typ {
	case uccore.ErrInternal:
		return http.StatusInternalServerError
	case uccore.ErrInvalidInput:
		return http.StatusBadRequest
	case uccore.ErrNotFound:
		return http.StatusNotFound
	case uccore.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
