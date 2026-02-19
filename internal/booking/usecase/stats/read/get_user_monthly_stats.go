package read

import (
	"context"
	"errors"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

type ReqGetUserMonthlyStats struct {
	UserID string
	Year   int
	Month  int
}

type UserMonthlyStatsVO struct {
	TotalSessions int            `json:"total_sessions"`
	TotalLeave    int            `json:"total_leave"`
	Children      []*ChildStatVO `json:"children"`
}

type ChildStatVO struct {
	Name      string  `json:"name"`
	Completed int     `json:"completed"`
	Leave     int     `json:"leave"`
	Absent    int     `json:"absent"`
	AvgWeek   float64 `json:"avg_week"`
}

type GetUserMonthlyStatsUseCase core.ReadUseCase[ReqGetUserMonthlyStats, *UserMonthlyStatsVO]

type getUserMonthlyStatsUseCase struct {
	repo repository.StatsRepository
}

func NewGetUserMonthlyStatsUseCase(repo repository.StatsRepository) GetUserMonthlyStatsUseCase {
	return &getUserMonthlyStatsUseCase{repo: repo}
}

func (uc *getUserMonthlyStatsUseCase) Name() string {
	return "GetUserMonthlyStats"
}

func (uc *getUserMonthlyStatsUseCase) Execute(ctx context.Context, req ReqGetUserMonthlyStats) (*UserMonthlyStatsVO, core.UseCaseError) {
	if req.UserID == "" {
		return nil, ErrGetUserMonthlyStatsFail.Wrap(errors.New("user id is empty"))
	}
	// 計算該月份的範圍
	startOfMonth := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.Local)
	endOfMonth := startOfMonth.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	filter := repository.NewFilterUserApptStatsByTrainTimeRange(startOfMonth, endOfMonth)
	stats, err := uc.repo.GetUserApptStats(ctx, req.UserID, filter)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return &UserMonthlyStatsVO{Children: []*ChildStatVO{}}, nil
		}
		return nil, ErrGetUserMonthlyStatsFail.Wrap(err)
	}

	return transformToVO(stats), nil
}

func transformToVO(stats *entity.UserApptStats) *UserMonthlyStatsVO {
	vo := &UserMonthlyStatsVO{
		TotalSessions: stats.TotalAppointment,
		TotalLeave:    stats.OnLeaveCount,
		Children:      make([]*ChildStatVO, 0),
	}

	for _, child := range stats.ChildState {
		// 假設一個月約 4 週來計算週均
		avg := float64(child.CheckedInCount) / 4.0
		vo.Children = append(vo.Children, &ChildStatVO{
			Name:      child.ChildName,
			Completed: child.CheckedInCount,
			Leave:     child.OnLeaveCount,
			Absent:    len(child.Appointments) - child.CheckedInCount - child.OnLeaveCount,
			AvgWeek:   avg,
		})
	}
	return vo
}

var ErrGetUserMonthlyStatsFail = core.NewDBError("GET_USER_MONTHLY_STATS", "QUERY_FAIL", "failed to get monthly stats", core.ErrInternal)
