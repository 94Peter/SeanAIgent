package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

type ReqQueryAllUserApptStats struct {
	StartTime time.Time
	EndTime   time.Time
}

type queryAllUserApptStatsUseCase struct {
	repo repository.StatsRepository
}

func NewQueryAllUserApptStatsUseCase(
	repo repository.StatsRepository,
) core.ReadUseCase[ReqQueryAllUserApptStats, []*entity.UserApptStats] {
	return &queryAllUserApptStatsUseCase{repo: repo}
}

func (uc *queryAllUserApptStatsUseCase) Name() string {
	return "QueryAllUserApptStats"
}

func (uc *queryAllUserApptStatsUseCase) Execute(ctx context.Context, req ReqQueryAllUserApptStats) (
	[]*entity.UserApptStats, core.UseCaseError,
) {
	stats, err := uc.repo.GetAllUserApptStats(
		ctx, repository.NewFilterUserApptStatsByTrainTimeRange(req.StartTime, req.EndTime),
	)
	if err != nil {
		return nil, ErrQueryAllUserApptStatsFail.Wrap(err)
	}
	return stats, nil
}

var (
	ErrQueryAllUserApptStatsFail = core.NewDBError(
		"QUERY_ALL_USER_APPT_STATS", "QUERY_FAIL", "query all user appt stats fail", core.ErrInternal)
)
