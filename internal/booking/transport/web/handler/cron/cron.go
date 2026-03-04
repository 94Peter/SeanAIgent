package cron

import (
	"net/http"
	"sync"
	"time"

	"seanAIgent/internal/booking/transport/web/handler"
	"seanAIgent/internal/booking/usecase"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	writeStats "seanAIgent/internal/booking/usecase/stats/write"

	"github.com/94peter/vulpes/ezapi"
	"github.com/gin-gonic/gin"
)

func NewCronApi(registry *usecase.Registry) handler.WebAPI {
	return &cronAPI{
		autoMarkAbsentUC:         registry.AutoMarkAbsent,
		batchSyncMonthlyStatsUC: registry.BatchSyncMonthlyStats,
	}
}

type cronAPI struct {
	autoMarkAbsentUC         writeAppt.AutoMarkAbsentUseCase
	batchSyncMonthlyStatsUC writeStats.BatchSyncMonthlyStatsUseCase
	once                     sync.Once
}

func (api *cronAPI) InitRouter(r ezapi.Router) {
	api.once.Do(func() {
		// 統一以 /cron 開頭
		r.POST("/cron/mark-absent", api.triggerAutoAbsent)
		r.POST("/cron/sync-all-stats", api.triggerSyncStats)
	})
}

func (api *cronAPI) triggerAutoAbsent(c *gin.Context) {
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

func (api *cronAPI) triggerSyncStats(c *gin.Context) {
	if api.batchSyncMonthlyStatsUC == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "batch sync stats use case is not initialized"})
		return
	}

	var req struct {
		Year  int `json:"year"`
		Month int `json:"month"`
	}

	// 如果沒有傳入，預設為當前月
	if err := c.ShouldBindJSON(&req); err != nil {
		now := time.Now()
		req.Year = now.Year()
		req.Month = int(now.Month())
	}

	resp, err := api.batchSyncMonthlyStatsUC.Execute(c.Request.Context(), writeStats.ReqBatchSyncMonthlyStats{
		Year:  req.Year,
		Month: req.Month,
	})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"processed_records": resp.TotalProcessed,
		"target_period":     time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.Local).Format("2006-01"),
		"executed_at":       time.Now().Format(time.RFC3339),
	})
}
