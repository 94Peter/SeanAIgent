package write

import (
	"context"
	"errors"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/domain/service"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

var (
	ErrCreateTrainDateInvalidInput = core.NewUseCaseError(
		"CREATE_TRAIN_DATE", "INVALID_INPUT", "invalid input", core.ErrInvalidInput)
	ErrCreateTrainDateCoachBusy = core.NewDomainError(
		"CREATE_TRAIN_DATE", "COACH_BUSY", "coach is busy", core.ErrConflict)
	ErrCreateTrainDateNewDomainEntityFail = core.NewDomainError(
		"CREATE_TRAIN_DATE", "DOMAIN_ERROR", "new domain entity failed", core.ErrInvalidInput)
	ErrCreateTrainDateSaveToDBFail = core.NewDBError(
		"CREATE_TRAIN_DATE", "DB_ERROR", "save to db failed", core.ErrInternal)
)

type createTrainRepo interface {
	repository.IdentityGenerator
	repository.TrainRepository
}

type createSlotUseCase struct {
	repo     createTrainRepo
	trainSvc service.TrainDateService // 依賴 Domain Service 介面
}

// NewCreateSlotUseCase 負責組裝並回傳泛型介面
func NewCreateTrainDateUseCase(
	repo createTrainRepo,
	trainSvc service.TrainDateService,
) core.WriteUseCase[ReqCreateTrainDate, *entity.TrainDate] {
	return &createSlotUseCase{
		repo:     repo,
		trainSvc: trainSvc,
	}
}

func (uc *createSlotUseCase) Name() string {
	return "CreateTrainDate"
}

func (uc *createSlotUseCase) Execute(
	ctx context.Context, req ReqCreateTrainDate,
) (result *entity.TrainDate, returnErr core.UseCaseError) {
	if err := req.Validate(); err != nil {
		returnErr = ErrCreateTrainDateInvalidInput.Wrap(err)
		return
	}
	// 1. 建立 Domain Entity (內部會驗證基本規則，如時間先後)
	period, err := entity.NewTimeRange(req.StartTime, req.EndTime)
	if err != nil {
		returnErr = ErrCreateTrainDateNewDomainEntityFail.Wrap(err)
		return
	}
	trainingID := uc.repo.GenerateID()
	// 3. 建立 Domain Entity
	training, err := entity.NewTrainDate(
		entity.WithBasicTrainDate(
			trainingID,
			req.CoachID,
			req.Location,
			req.Capacity,
			period,
		))
	if err != nil {
		returnErr = ErrCreateTrainDateNewDomainEntityFail.Wrap(err)
		return
	}

	// 2. 業務規則檢查 (Domain Service)：檢查教練在該時間段是否已有排程
	// 這需要查詢資料庫，所以放在 Service 實作
	if err := uc.trainSvc.CheckTrainerAvailability(ctx, training); err != nil {
		returnErr = ErrCreateTrainDateCoachBusy.Wrap(err)
	}

	// 3. 存檔
	err = uc.repo.SaveTrainDate(ctx, training)
	if err != nil {
		returnErr = ErrCreateTrainDateSaveToDBFail.Wrap(err)
		return
	}
	result = training
	return
}

// Request DTO
type ReqCreateTrainDate struct {
	StartTime time.Time
	EndTime   time.Time
	CoachID   string
	Location  string
	Capacity  int
}

func (r *ReqCreateTrainDate) Validate() error {
	if r.CoachID == "" {
		return errors.New("coach id is empty")
	}
	if r.Capacity <= 0 {
		return errors.New("capacity is invalid")
	}
	if r.Location == "" {
		return errors.New("location is empty")
	}
	if r.StartTime.IsZero() {
		return errors.New("start time is empty")
	}
	if r.EndTime.IsZero() {
		return errors.New("end time is empty")
	}
	if r.StartTime.After(r.EndTime) {
		return errors.New("start time is after end time")
	}
	return nil
}
