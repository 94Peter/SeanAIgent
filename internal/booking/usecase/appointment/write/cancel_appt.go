package write

import (
	"context"
	"errors"

	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqCancelAppt struct {
	ApptID string
	UserID string
}

type CancelApptUseCase core.WriteUseCase[ReqCancelAppt, *entity.Appointment]

type cancelApptUseCaseRepo interface {
	repository.AppointmentRepository
	repository.TrainRepository
	repository.StatsRepository
}

func NewCancelApptUseCase(repo cancelApptUseCaseRepo, cw cacheWorker) core.WriteUseCase[ReqCancelAppt, *entity.Appointment] {
	return &cancelApptUseCase{
		repo: repo,
		cw:   cw,
	}
}

type cancelApptUseCase struct {
	repo cancelApptUseCaseRepo
	cw   cacheWorker
}

func (uc *cancelApptUseCase) Name() string {
	return "CancelAppt"
}

func (uc *cancelApptUseCase) Execute(
	ctx context.Context, req ReqCancelAppt,
) (*entity.Appointment, core.UseCaseError) {
	var err error
	appt, err := uc.repo.FindApptByID(ctx, req.ApptID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCancelApptApptNotFound
		}
		return nil, ErrCancelApptFindApptFail.Wrap(err)
	}
	err = appt.CancelAsMistake(req.UserID)
	if err != nil {
		return nil, ErrCancelApptCancelApptFail.Wrap(err)
	}
	// 2. 增加名額
	err = uc.repo.IncreaseCapacity(ctx, appt.TrainingID(), 1)
	if err != nil {
		return nil, ErrCancelApptIncreaseCapacityFail.Wrap(err)
	}
	// 3. 刪除 appointment
	err = uc.repo.DeleteAppointment(ctx, appt)
	if err != nil {
		// rollback
		_ = uc.repo.DeductCapacity(ctx, appt.TrainingID(), 1)
		return nil, ErrCancelApptDeleteApptFail.Wrap(err)
	}

	// 使用背景 Worker 進行非同步清理
	uc.cw.Clean(req.UserID, appt.TrainingID())

	return appt, nil
}

var (
	ErrCancelApptApptNotFound = core.NewDBError(
		"CANCEL_APPT", "APPOINTMENT_NOT_FOUND", "appointment not found", core.ErrNotFound)
	ErrCancelApptFindApptFail = core.NewDBError(
		"CANCEL_APPT", "FIND_APPOINTMENT_FAIL", "find appointment fail", core.ErrInternal)
	ErrCancelApptCancelApptFail = core.NewDomainError(
		"CANCEL_APPT", "CANCEL_APPOINTMENT_FAIL", "cancel appointment fail", core.ErrInternal)
	ErrCancelApptDeleteApptFail = core.NewDBError(
		"CANCEL_APPT", "DELETE_APPOINTMENT_FAIL", "delete appointment fail", core.ErrInternal)
	ErrCancelApptIncreaseCapacityFail = core.NewDBError(
		"CANCEL_APPT", "INCREASE_CAPACITY_FAIL", "increase capacity fail", core.ErrConflict)
)
