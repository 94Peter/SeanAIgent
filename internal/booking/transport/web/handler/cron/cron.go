package cron

import (
	"net/http"
	"sync"
	"time"

	"seanAIgent/internal/booking/transport/web/handler"
	"seanAIgent/internal/booking/usecase"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"

	"github.com/94peter/vulpes/ezapi"
	"github.com/gin-gonic/gin"
)

func NewCronApi(registry *usecase.Registry) handler.WebAPI {
	return &cronAPI{
		autoMarkAbsentUC: registry.AutoMarkAbsent,
	}
}

type cronAPI struct {
	autoMarkAbsentUC writeAppt.AutoMarkAbsentUseCase
	once             sync.Once
}

func (api *cronAPI) InitRouter(r ezapi.Router) {
	api.once.Do(func() {
		// 統一以 /cron 開頭
		r.POST("/cron/mark-absent", api.triggerAutoAbsent)
	})
}

func (api *cronAPI) triggerAutoAbsent(c *gin.Context) {
	// 這裡可以加上特定的 Header 驗證，例如 X-Cron-Secret
	if api.autoMarkAbsentUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auto mark absent use case is not initialized"})
		return
	}
	count, err := api.autoMarkAbsentUC.Execute(c.Request.Context(), struct{}{})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"updated_count": count,
		"executed_at":   time.Now().Format(time.RFC3339),
	})
}
