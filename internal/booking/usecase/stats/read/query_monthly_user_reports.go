package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqQueryMonthlyUserReports struct {
	Year   int
	Month  int
	Page   int64
	Limit  int64
	Search string
}

type RespQueryMonthlyUserReports struct {
	UserStats []*entity.UserMonthlyStat
	Total     int64
}

type QueryMonthlyUserReportsUseCase core.ReadUseCase[ReqQueryMonthlyUserReports, *RespQueryMonthlyUserReports]

type queryMonthlyUserReportsUseCase struct {
	repo repository.StatsRepository
}

func NewQueryMonthlyUserReportsUseCase(repo repository.StatsRepository) QueryMonthlyUserReportsUseCase {
	return &queryMonthlyUserReportsUseCase{repo: repo}
}

func (uc *queryMonthlyUserReportsUseCase) Name() string {
	return "QueryMonthlyUserReports"
}

func (uc *queryMonthlyUserReportsUseCase) Execute(ctx context.Context, req ReqQueryMonthlyUserReports) (*RespQueryMonthlyUserReports, core.UseCaseError) {
	skip := (req.Page - 1) * req.Limit
	if skip < 0 {
		skip = 0
	}

	stats, total, err := uc.repo.FindMonthlyStats(ctx, req.Year, req.Month, skip, req.Limit, req.Search)
	if err != nil {
		return nil, core.NewDBError("QUERY_USER_REPORT", "FETCH_FAIL", "failed to fetch user reports", core.ErrInternal).Wrap(err)
	}

	return &RespQueryMonthlyUserReports{
		UserStats: stats,
		Total:     total,
	}, nil
}
