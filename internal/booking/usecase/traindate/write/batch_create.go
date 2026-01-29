package write

import (
	"context"
	"errors"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/service"
	"seanAIgent/internal/booking/usecase/core"
	"sort"
)

type batchCreateSlotUseCase struct {
	repo     createTrainRepo
	trainSvc service.TrainDateService // 依賴 Domain Service 介面
}

func NewBatchCreateTrainDateUseCase(
	repo createTrainRepo,
	trainSvc service.TrainDateService,
) core.WriteUseCase[[]ReqCreateTrainDate, []*entity.TrainDate] {
	return &batchCreateSlotUseCase{
		repo:     repo,
		trainSvc: trainSvc,
	}
}

func (uc *batchCreateSlotUseCase) Name() string {
	return "BatchCreateTrainDate"
}

func (uc *batchCreateSlotUseCase) Execute(
	ctx context.Context, req []ReqCreateTrainDate,
) (result []*entity.TrainDate, resultErr core.UseCaseError) {
	result = []*entity.TrainDate{}
	for _, r := range req {
		if err := r.Validate(); err != nil {
			resultErr = ErrCreateTrainDateInvalidInput.Wrap(err)
			return
		}
	}

	if checkRequestsHasOverlap(req) {
		resultErr = ErrCreateTrainDateInvalidInput.Wrap(errors.New("requests has overlap"))
		return
	}

	dates := make([]entity.TimeRange, 0, len(req))
	trainings := make([]*entity.TrainDate, 0, len(req))

	// 1. 準備所有時段並進行初步驗證
	for _, r := range req {
		// 利用 Value Object 檢查時間合法性 (StartTime < EndTime)
		tr, err := entity.NewTimeRange(r.StartTime, r.EndTime)
		if err != nil {
			resultErr = ErrCreateTrainDateNewDomainEntityFail.Wrap(err)
			return
		}
		dates = append(dates, tr)
		// 2. 取得新 ID (由 Repo 生成，避免污染)
		newID := uc.repo.GenerateID()

		// 3. 建立 Domain 物件
		slot, err := entity.NewTrainDate(
			entity.WithBasicTrainDate(
				newID,
				r.CoachID,
				r.Location,
				r.Capacity,
				tr,
			))
		if err != nil {
			resultErr = ErrCreateTrainDateNewDomainEntityFail.Wrap(err)
			return
		}
		trainings = append(trainings, slot)
	}
	// 4. 衝突檢查 (進階實作：一次性檢查該批次是否與資料庫現有資料重疊)
	// 為了效能，不要在迴圈裡一個一個查 DB
	err := uc.trainSvc.CheckAnyOverlap(ctx, req[0].CoachID, dates)
	if err != nil {
		resultErr = ErrCreateTrainDateCoachBusy.Wrap(err)
		return
	}
	// 5. 批次存檔
	err = uc.repo.SaveManyTrainDates(ctx, trainings)
	if err != nil {
		resultErr = ErrCreateTrainDateSaveToDBFail.Wrap(err)
		return
	}
	result = trainings
	return
}

func checkRequestsHasOverlap(req []ReqCreateTrainDate) bool {
	if len(req) <= 1 {
		return false
	}

	// 1. 先複製一份，避免修改到原始 Slice 的順序（副作用）
	sortedReq := make([]ReqCreateTrainDate, len(req))
	copy(sortedReq, req)

	// 2. 根據開始時間進行排序 - O(n log n)
	sort.Slice(sortedReq, func(i, j int) bool {
		return sortedReq[i].StartTime.Before(sortedReq[j].StartTime)
	})

	// 3. 線性掃描檢查相鄰區間 - O(n)
	for i := 0; i < len(sortedReq)-1; i++ {
		current := sortedReq[i]
		next := sortedReq[i+1]

		// 如果當前的結束時間晚於下一個的開始時間，代表重疊
		if current.EndTime.After(next.StartTime) {
			return true
		}
	}

	return false
}
