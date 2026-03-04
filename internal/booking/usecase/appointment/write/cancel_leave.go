package write

import (
	"context"
	"errors"
	"time"

	"seanAIgent/internal/booking/domain"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"seanAIgent/internal/event"
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
	repository.IdentityGenerator
}

func NewCancelLeaveUseCase(repo cancelLeaveUseCaseRepo, bus event.Bus) CancelLeaveUseCase {
	return &cancelLeaveUseCase{
		repo: repo,
		bus:  bus,
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
	bus  event.Bus
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

	oldStatus := appt.Status().String()
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

	// 手動清理快取
	_ = uc.repo.CleanTrainCache(ctx, appt.User().UserID())
	_ = uc.repo.CleanStatsCache(ctx, appt.User().UserID(), trainDate.Period().Start().Year(), int(trainDate.Period().Start().Month()))

	// 發送領域事件
	evt := event.NewTypedEvent(uc.repo.GenerateID(), domain.TopicAppointmentStatusChanged, domain.AppointmentStatusChanged{
		BookingID:  appt.ID(),
		UserID:     appt.User().UserID(),
		TrainingID: appt.TrainingID(),
		OldStatus:  oldStatus,
		NewStatus:  appt.Status().String(),
		OccurredAt: time.Now(),
	})
	uc.bus.Publish(ctx, evt)

	return appt, nil
}
