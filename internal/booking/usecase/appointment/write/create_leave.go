package write

import (
	"context"
	"errors"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqCreateLeave struct {
	AppointmentID string
	User          entity.User
	Reason        string
}

type createLeaveUseCaseRepo interface {
	repository.AppointmentRepository
	repository.TrainRepository
}

type createLeaveUseCase struct {
	repo createLeaveUseCaseRepo
}

func NewCreateLeaveUseCase(repo createLeaveUseCaseRepo) core.WriteUseCase[ReqCreateLeave, *entity.Appointment] {
	return &createLeaveUseCase{repo: repo}
}

func (uc *createLeaveUseCase) Name() string {
	return "CreateLeave"
}

func (uc *createLeaveUseCase) Execute(
	ctx context.Context, req ReqCreateLeave,
) (*entity.Appointment, core.UseCaseError) {
	var err error
	// 1. 找 appointment
	appt, err := uc.repo.FindApptByID(ctx, req.AppointmentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCreateLeaveApptNotFound
		}
		return nil, ErrCreateLeaveFindApptFail.Wrap(err)
	}
	// 2. verify user
	if appt.User().UserID() != req.User.UserID() {
		return nil, ErrCreateLeavePermissionDenied
	}

	// 3. 找 training date
	trainDate, err := uc.repo.FindTrainDateByID(ctx, appt.TrainingID())
	if err != nil {
		return nil, ErrCreateLeaveTrainDateNotFound.Wrap(err)
	}

	err = appt.AppendLeaveRecord(req.Reason, trainDate.Period().Start())
	if err != nil {
		return nil, ErrCreateLeaveFail.Wrap(err)
	}
	//

	// increase training date capacity
	err = uc.repo.IncreaseCapacity(ctx, trainDate.ID(), 1)
	if err != nil {
		return nil, ErrCreateLeaveIncreaseCapacityFail.Wrap(err)
	}
	// update appt leave info
	err = uc.repo.UpdateAppt(ctx, appt)
	if err != nil {
		// rollback
		_ = uc.repo.DeductCapacity(ctx, trainDate.ID(), 1)
		return nil, ErrCreateLeaveSaveLeaveFail.Wrap(err)
	}
	return appt, nil
}

var (
	ErrCreateLeaveApptNotFound = core.NewDBError(
		"CREATE_LEAVE", "APPOINTMENT_NOT_FOUND", "appointment not found", core.ErrNotFound)
	ErrCreateLeaveFindApptFail = core.NewDBError(
		"CREATE_LEAVE", "FIND_APPOINTMENT_FAIL", "find appointment fail", core.ErrInternal)
	ErrCreateLeaveTrainDateNotFound = core.NewDBError(
		"CREATE_LEAVE", "TRAIN_DATE_NOT_FOUND", "train date not found", core.ErrNotFound)
	ErrCreateLeaveFindTrainDateFail = core.NewDBError(
		"CREATE_LEAVE", "FIND_TRAIN_DATE_FAIL", "find train date fail", core.ErrInternal)
	ErrCreateLeaveFail = core.NewDomainError(
		"CREATE_LEAVE", "APPEND_LEAVE_FAIL", "append leave fail", core.ErrInternal)
	ErrCreateLeaveSaveLeaveFail = core.NewDBError(
		"CREATE_LEAVE", "SAVE_LEAVE_FAIL", "save leave fail", core.ErrInternal)
	ErrCreateLeaveIncreaseCapacityFail = core.NewDBError(
		"CREATE_LEAVE", "INCREASE_CAPACITY_FAIL", "increase capacity fail", core.ErrConflict)
	ErrCreateLeaveUpdateApptFail = core.NewDBError(
		"CREATE_LEAVE", "UPDATE_APPOINTMENT_FAIL", "update appointment fail", core.ErrInternal)
	ErrCreateLeavePermissionDenied = core.NewUseCaseError(
		"CREATE_LEAVE", "PERMISSION_DENIED", "permission denied", core.ErrPermissionDenied)
)
