package write

import (
	"context"
	"errors"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqCancelLeave struct {
	ApptID string
	UserID string
}

type cancelLeaveUseCaseRepo interface {
	repository.TrainRepository
	repository.AppointmentRepository
	repository.StatsRepository
}

func NewCancelLeaveUseCase(repo cancelLeaveUseCaseRepo, cw cacheWorker) core.WriteUseCase[ReqCancelLeave, *entity.Appointment] {
	return &cancelLeaveUseCase{
		repo: repo,
		cw:   cw,
	}
}

var (
	ErrCancelLeaveApptNotFound = core.NewDBError(
		"CANCEL_LEAVE", "APPOINTMENT_NOT_FOUND", "appointment not found", core.ErrNotFound)
	ErrCancelLeaveFindApptFail = core.NewDBError(
		"CANCEL_LEAVE", "FIND_APPOINTMENT_FAIL", "find appointment fail", core.ErrInternal)
	ErrCancelLeaveCancelLeaveFail = core.NewDomainError(
		"CANCEL_LEAVE", "CANCEL_LEAVE_FAIL", "cancel leave fail", core.ErrInternal)
	ErrCancelLeaveDeductCapacityFail = core.NewDBError(
		"CANCEL_LEAVE", "DEDUCT_CAPACITY_FAIL", "deduct capacity fail", core.ErrConflict)
	ErrCancelLeaveUpdateApptFail = core.NewDBError(
		"CANCEL_LEAVE", "UPDATE_APPOINTMENT_FAIL", "update appointment fail", core.ErrInternal)
	ErrCancelLeaveTrainDateNotFound = core.NewDBError(
		"CANCEL_LEAVE", "TRAIN_DATE_NOT_FOUND", "train date not found", core.ErrNotFound)
)

type cancelLeaveUseCase struct {
	repo cancelLeaveUseCaseRepo
	cw   cacheWorker
}

func (uc *cancelLeaveUseCase) Name() string {
	return "CancelLeave"
}

func (uc *cancelLeaveUseCase) Execute(
	ctx context.Context, req ReqCancelLeave) (*entity.Appointment, core.UseCaseError) {
	var err error
	appt, err := uc.repo.FindApptByID(ctx, req.ApptID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCancelLeaveApptNotFound
		}
		return nil, ErrCancelLeaveFindApptFail.Wrap(err)
	}
	err = appt.CancelLeave(req.UserID)
	if err != nil {
		return nil, ErrCancelLeaveCancelLeaveFail.Wrap(err)
	}
	// 2. 扣除名額, 若失敗則回傳錯誤，不更新 appointment
	err = uc.repo.DeductCapacity(ctx, appt.TrainingID(), 1)
	if err != nil {
		return nil, ErrCancelLeaveDeductCapacityFail.Wrap(err)
	}
	// 3. 更新 appointment
	err = uc.repo.UpdateAppt(ctx, appt)
	if err != nil {
		// rollback
		_ = uc.repo.IncreaseCapacity(ctx, appt.TrainingID(), 1)
		return nil, ErrCancelLeaveUpdateApptFail.Wrap(err)
	}

	// 使用背景 Worker 進行非同步清理
	uc.cw.Clean(appt.User().UserID(), appt.TrainingID())

	return appt, nil
}
