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

type cancelApptUseCaseRepo interface {
	repository.AppointmentRepository
	repository.TrainRepository
}

func NewCancelApptUseCase(repo cancelApptUseCaseRepo) core.WriteUseCase[ReqCancelAppt, *entity.Appointment] {
	return &cancelApptUseCase{repo: repo}
}

type cancelApptUseCase struct {
	repo cancelApptUseCaseRepo
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
	err = uc.repo.DeleteAppointment(ctx, appt)
	if err != nil {
		return nil, ErrCancelApptDeleteApptFail.Wrap(err)
	}
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
)
