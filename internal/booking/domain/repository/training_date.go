package repository

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"time"
)

// Repository
type TrainRepository interface {
	SaveTrainDate(ctx context.Context, training *entity.TrainDate) RepoError
	SaveManyTrainDates(ctx context.Context, trainings []*entity.TrainDate) RepoError // 批次寫入提高效能
	UpdateManyTrainDates(ctx context.Context, trainings []*entity.TrainDate) RepoError
	DeleteTrainingDate(ctx context.Context, training *entity.TrainDate) RepoError

	FindTrainDateByID(ctx context.Context, id string) (*entity.TrainDate, RepoError)
	FindTrainDates(ctx context.Context, filter FilterTrainDate) ([]*entity.TrainDate, RepoError)

	QueryTrainDateHasAppointmentState(
		ctx context.Context, filter FilterTrainDate,
	) ([]*entity.TrainDateHasApptState, RepoError)

	UserQueryTrainDateHasApptState(
		ctx context.Context, userID string, filter FilterTrainDate,
	) ([]*entity.TrainDateHasUserApptState, RepoError)

	// 檢查是否有重疊時段
	// 邏輯：Find slots WHERE coach_id = ? AND start_time < req.end AND end_time > req.start
	CheckOverlap(ctx context.Context, coachID string, tr entity.TimeRange) (bool, RepoError)
	HasAnyOverlap(ctx context.Context, coachID string, tr []entity.TimeRange) (bool, RepoError)

	// 我們之前討論過的原子扣名額
	DeductCapacity(ctx context.Context, trainingID string, count int) RepoError
	// 原子增加名額
	IncreaseCapacity(ctx context.Context, trainingID string, count int) RepoError
}

// Filter
type FilterTrainDate interface {
	isCriteria() // 標記用介面
}

// 條件 A：按時間範圍
func NewFilterTrainDataByTimeRange(start, end time.Time) FilterTrainDate {
	return FilterTrainingDateByTimeRange{StartTime: start, EndTime: end}
}

type FilterTrainingDateByTimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

func (f FilterTrainingDateByTimeRange) isCriteria() {}

// 條件 B：按 ID 列表
func NewFilterTrainDateByIds(ids ...string) FilterTrainDate {
	return FilterTrainingDateByIDs{
		TrainingDateIDs: ids,
	}
}

type FilterTrainingDateByIDs struct {
	TrainingDateIDs []string
}

func (f FilterTrainingDateByIDs) isCriteria() {}

// 條件 C：取得未來時段
func NewFilterTrainDateByAfterTime(start time.Time) FilterTrainDate {
	return FilterTrainDateByAfterTime{
		Start: start,
	}
}

type FilterTrainDateByAfterTime struct {
	Start time.Time
}

func (f FilterTrainDateByAfterTime) isCriteria() {}

// 條件 D：取得最近的時段, 依訓練時間為準，超過就不再顯示
func NewFilterTrainDateByEndTime(start time.Time) FilterTrainDate {
	return FilterTrainDateByEndTime{
		Start: start,
	}
}

type FilterTrainDateByEndTime struct {
	Start time.Time
}

func (f FilterTrainDateByEndTime) isCriteria() {}
