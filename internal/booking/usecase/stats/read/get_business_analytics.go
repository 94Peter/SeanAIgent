package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqGetBusinessAnalytics struct {
	MonthsLimit int
}

type RespGetBusinessAnalytics struct {
	HistoricalStats []*entity.MonthlyBusinessStat
	Metrics         BusinessMetrics
}

type BusinessMetrics struct {
	AvgAttendanceRate float64
	RetentionRate     float64
	ActiveStudents    int
	RevenueGrowth     float64
}

type GetBusinessAnalyticsUseCase core.ReadUseCase[ReqGetBusinessAnalytics, *RespGetBusinessAnalytics]

type getBusinessAnalyticsUseCase struct {
	repo repository.StatsRepository
}

func NewGetBusinessAnalyticsUseCase(repo repository.StatsRepository) GetBusinessAnalyticsUseCase {
	return &getBusinessAnalyticsUseCase{repo: repo}
}

func (uc *getBusinessAnalyticsUseCase) Name() string {
	return "GetBusinessAnalytics"
}

func (uc *getBusinessAnalyticsUseCase) Execute(ctx context.Context, req ReqGetBusinessAnalytics) (*RespGetBusinessAnalytics, core.UseCaseError) {
	if req.MonthsLimit <= 0 {
		req.MonthsLimit = 12
	}

	stats, err := uc.repo.GetHistoricalAnalytics(ctx, req.MonthsLimit)
	if err != nil {
		return nil, core.NewDBError("GET_ANALYTICS", "FETCH_FAIL", "failed to fetch analytics", core.ErrInternal).Wrap(err)
	}

	metrics := BusinessMetrics{}
	if len(stats) > 0 {
		latest := stats[0]
		
		// 1. 平均出席率 (本月)
		if latest.TotalBookings > 0 {
			metrics.AvgAttendanceRate = float64(latest.AttendedCount) / float64(latest.TotalBookings)
		}
		
		// 2. 活躍學員 (本月)
		metrics.ActiveStudents = latest.ActiveUsers

		// 3. 留存率 (簡單定義：本月學員 / 上月學員，實際需更複雜，此處為示意)
		if len(stats) > 1 && stats[1].ActiveUsers > 0 {
			metrics.RetentionRate = float64(latest.ActiveUsers) / float64(stats[1].ActiveUsers)
			if metrics.RetentionRate > 1.0 {
				metrics.RetentionRate = 1.0
			}
		} else {
			metrics.RetentionRate = 1.0
		}

		// 4. 營收增長 (示意：本月出席數 / 上月出席數 - 1)
		if len(stats) > 1 && stats[1].AttendedCount > 0 {
			metrics.RevenueGrowth = (float64(latest.AttendedCount)/float64(stats[1].AttendedCount) - 1) * 100
		}
	}

	return &RespGetBusinessAnalytics{
		HistoricalStats: stats,
		Metrics:         metrics,
	}, nil
}
