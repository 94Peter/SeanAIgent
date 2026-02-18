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
}
