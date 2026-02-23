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

type CancelLeaveUseCase core.WriteUseCase[ReqCancelLeave, *entity.Appointment]

type cancelLeaveUseCaseRepo interface {
	repository.TrainRepository
	repository.AppointmentRepository
	repository.StatsRepository
}

func NewCancelLeaveUseCase(repo cancelLeaveUseCaseRepo, cw cacheWorker) CancelLeaveUseCase {
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
	appt, repoErr := uc.repo.FindApptByID(ctx, req.ApptID)
	if repoErr != nil {
		if errors.Is(repoErr, repository.ErrNotFound) {
			return nil, ErrCancelLeaveApptNotFound
		}
		return nil, ErrCancelLeaveFindApptFail.Wrap(repoErr)
	}
	
	// 取得課程資訊以獲得 startTime
	trainDate, repoErr := uc.repo.FindTrainDateByID(ctx, appt.TrainingID())
	if repoErr != nil {
		return nil, ErrCancelLeaveTrainDateNotFound.Wrap(repoErr)
	}

	// 這裡會檢查 req.UserID 是否為預約本人
	err := appt.CancelLeave(req.UserID)
	if err != nil {
		return nil, ErrCancelLeaveCancelLeaveFail.Wrap(err)
	}
	
	// 2. 扣除名額
	repoErr = uc.repo.DeductCapacity(ctx, appt.TrainingID(), 1)
	if repoErr != nil {
		return nil, ErrCancelLeaveDeductCapacityFail.Wrap(repoErr)
	}
	
	// 3. 更新 appointment
	repoErr = uc.repo.UpdateAppt(ctx, appt)
	if repoErr != nil {
		_ = uc.repo.IncreaseCapacity(ctx, appt.TrainingID(), 1)
		return nil, ErrCancelLeaveUpdateApptFail.Wrap(repoErr)
	}

	// 使用同步清理，確保跳轉頁面後資料一致
	uc.cw.CleanSync(ctx, appt.User().UserID(), appt.TrainingID(), trainDate.Period().Start())

	return appt, nil
}
