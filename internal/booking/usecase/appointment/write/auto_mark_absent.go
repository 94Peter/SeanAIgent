package write

import (
	"context"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"seanAIgent/internal/event"
	"time"
)

type AutoMarkAbsentUseCase core.WriteUseCase[struct{}, int64]

type autoMarkAbsentUseCaseRepo interface {
	repository.TrainRepository
	repository.AppointmentRepository
}

func NewAutoMarkAbsentUseCase(repo autoMarkAbsentUseCaseRepo, bus event.Bus) AutoMarkAbsentUseCase {
	return &autoMarkAbsentUseCase{
		repo: repo,
		bus:  bus,
	}
}

type autoMarkAbsentUseCase struct {
	repo autoMarkAbsentUseCaseRepo
	bus  event.Bus
}

func (uc *autoMarkAbsentUseCase) Name() string {
	return "AutoMarkAbsent"
}

const batchSize = uint16(300)

func (uc *autoMarkAbsentUseCase) Execute(ctx context.Context, _ struct{}) (int64, core.UseCaseError) {
	// 設定緩衝時間，例如課程結束 30 分鐘後才判定缺席
	cutoff := time.Now().Add(-30 * time.Minute)
	allCount := int64(0)
	var finalErr core.UseCaseError
	// 1. 從 TrainRepo 查出所有過期的課程 ID
	for {
		pastIDs, err := uc.repo.FindPastTrainDateIDs(ctx, cutoff, batchSize)
		if err != nil {
			finalErr = core.NewUseCaseError("AUTO_ABSENT", "FIND_TRAIN_FAIL", "查詢場次時中斷", core.ErrInternal).Wrap(err)
			break
		}

		if len(pastIDs) == 0 {
			break
		}

		// 2. 丟到 ApptRepo 更新為 ABSENT
		count, err := uc.repo.MarkAbsentByTrainIDs(ctx, pastIDs)
		if err != nil {
			finalErr = core.NewUseCaseError("AUTO_ABSENT", "BATCH_UPDATE_FAIL", "更新預約時中斷", core.ErrInternal).Wrap(err)
			break
		}
		allCount += count
		if len(pastIDs) < int(batchSize) {
			break
		}
	}
	return allCount, finalErr
}
