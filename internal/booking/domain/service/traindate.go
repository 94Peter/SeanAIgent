package service

import (
	"context"
	"errors"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
)

var (
	ErrTrainerTimeOverlap = errors.New("trainer already has a scheduled slot during this period")
)

type TrainDateService interface {
	// 檢查教練在指定時間是否已經有其他訓練時段存在 (避免重疊)
	CheckTrainerAvailability(ctx context.Context, slot *entity.TrainDate) error
	// 檢查教練在指定時間是否已經有其他訓練時段存在 (避免重疊)
	CheckAnyOverlap(ctx context.Context, coachID string, tr []entity.TimeRange) error
}

type trainDateService struct {
	repo repository.TrainRepository
}

// NewTrainDateService 建立實作 (供 Registry 呼叫)
func NewTrainDateService(repo repository.TrainRepository) TrainDateService {
	return &trainDateService{
		repo: repo,
	}
}

func (s *trainDateService) CheckTrainerAvailability(ctx context.Context, newSlot *entity.TrainDate) error {
	// 1. 從 Repository 查詢該教練在「新時段」範圍內是否已有任何存在的時段
	// 邏輯：現有時段的 Start < 新時段的 End  AND 現有時段的 End > 新時段的 Start
	existingSlots, err := s.repo.CheckOverlap(
		ctx,
		newSlot.UserID(),
		newSlot.Period(),
	)
	if err != nil {
		return err
	}

	if existingSlots {
		return ErrTrainerTimeOverlap
	}

	return nil
}

func (s *trainDateService) CheckAnyOverlap(ctx context.Context, coachID string, tr []entity.TimeRange) error {
	// 1. 從 Repository 查詢該教練在「新時段」範圍內是否已有任何存在的時段
	isOverlapped, err := s.repo.HasAnyOverlap(ctx, coachID, tr)
	if err != nil {
		return err
	}
	if isOverlapped {
		return ErrTrainerTimeOverlap
	}
	return nil
}
