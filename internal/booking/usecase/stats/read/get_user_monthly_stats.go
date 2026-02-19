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
	TotalUpcoming int            `json:"total_upcoming"`
	TotalSessions int            `json:"total_sessions"`
	TotalLeave    int            `json:"total_leave"`
	Children      []*ChildStatVO `json:"children"`
}

type ChildStatVO struct {
	Name      string  `json:"name"`
	Upcoming  int     `json:"upcoming"`
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
	taipeiLoc := time.FixedZone("Asia/Taipei", 8*60*60)
	startOfMonth := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, taipeiLoc)
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

	now := time.Now()
	totalUpcoming := 0
	for _, child := range stats.ChildState {
		upcoming := 0
		completed := 0
		leave := 0
		absent := 0

		for _, appt := range child.Appointments {
			if appt.IsOnLeave {
				leave++
			} else if appt.IsCheckedIn {
				completed++
			} else if now.Before(appt.EndDate) {
				upcoming++
			} else {
				absent++
			}
		}

		totalUpcoming += upcoming
		// 假設一個月約 4 週來計算週均
		avg := float64(completed) + float64(upcoming)/4.0
		vo.Children = append(vo.Children, &ChildStatVO{
			Name:      child.ChildName,
			Upcoming:  upcoming,
			Completed: completed,
			Leave:     leave,
			Absent:    absent,
			AvgWeek:   avg,
		})
	}
	vo.TotalUpcoming = totalUpcoming
	return vo
}

var ErrGetUserMonthlyStatsFail = core.NewDBError("GET_USER_MONTHLY_STATS", "QUERY_FAIL", "failed to get monthly stats", core.ErrInternal)
