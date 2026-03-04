package repository

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"time"
)

type FilterUserApptStats interface {
	isCriteria()
}

type FilterUserApptStatsByTrainTimeRange struct {
	TrainStart time.Time
	TrainEnd   time.Time
}

func (f FilterUserApptStatsByTrainTimeRange) isCriteria() {}

func NewFilterUserApptStatsByTrainTimeRange(trainStart, trainEnd time.Time) FilterUserApptStats {
	return FilterUserApptStatsByTrainTimeRange{
		TrainStart: trainStart,
		TrainEnd:   trainEnd,
	}
}

// 介面定義在 UseCase 內部，直接回傳 DTO
type StatsRepository interface {
	GetAllUserApptStats(
		ctx context.Context, filter FilterUserApptStats,
	) ([]*entity.UserApptStats, RepoError)
	GetUserApptStats(
		ctx context.Context, userID string, filter FilterUserApptStats,
	) (*entity.UserApptStats, RepoError)
	CleanStatsCache(ctx context.Context, userID string, year, month int) RepoError

	// Phase 1 & 2: 預聚合統計表相關操作
	UpsertUserMonthlyStats(ctx context.Context, stat *entity.UserMonthlyStat) RepoError
	UpsertManyUserMonthlyStats(ctx context.Context, stats []*entity.UserMonthlyStat) RepoError
	FindMonthlyStats(ctx context.Context, year, month int, skip, limit int64, search string) ([]*entity.UserMonthlyStat, int64, RepoError)
	AggregateUserMonthlyStats(ctx context.Context, userID string, year, month int) (*entity.UserMonthlyStat, RepoError)
	AggregateAllUsersMonthlyStats(ctx context.Context, year, month int) ([]*entity.UserMonthlyStat, RepoError)

	// GetHistoricalAnalytics 獲取全域經營趨勢數據
	GetHistoricalAnalytics(ctx context.Context, monthsLimit int) ([]*entity.MonthlyBusinessStat, RepoError)
}
