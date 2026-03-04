package write

import (
	"context"
	"seanAIgent/internal/booking/domain"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"seanAIgent/internal/event"
	"time"
)

type AutoMarkAbsentUseCase core.WriteUseCase[struct{}, int64]

type autoMarkAbsentUseCaseRepo interface {
	repository.TrainRepository
	repository.AppointmentRepository
	repository.IdentityGenerator
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

	affectedUserIDs := make(map[string]struct{})

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

		// 2. 丟到 ApptRepo 更新為 ABSENT，並取得受影響用戶
		uids, err := uc.repo.MarkAbsentByTrainIDs(ctx, pastIDs)
		if err != nil {
			finalErr = core.NewUseCaseError("AUTO_ABSENT", "BATCH_UPDATE_FAIL", "更新預約時中斷", core.ErrInternal).Wrap(err)
			break
		}
		
		for _, uid := range uids {
			affectedUserIDs[uid] = struct{}{}
		}
		allCount += int64(len(uids))

		if len(pastIDs) < int(batchSize) {
			break
		}
	}

	// 3. 發送重新聚合請求事件
	now := time.Now()
	for uid := range affectedUserIDs {
		// 這裡為了簡化，目前假設是重新計算當月
		// TODO: 更好的做法是從受影響的 pastIDs 算出對應的年月
		evt := event.NewTypedEvent(uc.repo.GenerateID(), domain.TopicUserStatsRefreshRequested, domain.UserStatsRefreshRequested{
			UserID:     uid,
			Year:       now.Year(),
			Month:      int(now.Month()),
			Reason:     "AutoMarkAbsent",
			OccurredAt: now,
		})
		uc.bus.Publish(ctx, evt)
	}

	return allCount, finalErr
}
