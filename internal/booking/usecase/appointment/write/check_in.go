package write

import (
	"context"
	"errors"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqCheckIn struct {
	TrainDateID         string
	CheckedInBookingIDs []string
}

type checkInUseCaseRepo interface {
	repository.AppointmentRepository
	repository.TrainRepository
}

func NewCheckInUseCase(repo checkInUseCaseRepo) core.WriteUseCase[ReqCheckIn, []*entity.Appointment] {
	return &checkInUseCase{repo: repo}
}

type checkInUseCase struct {
	repo checkInUseCaseRepo
}

func (uc *checkInUseCase) Name() string {
	return "CheckIn"
}

func (uc *checkInUseCase) Execute(
	ctx context.Context, req ReqCheckIn,
) ([]*entity.Appointment, core.UseCaseError) {
	if len(req.CheckedInBookingIDs) == 0 {
		return nil, ErrCheckInNoApptCheckedIn
	}
	var err error
	// find trainDateID all appointment
	appts, err := uc.repo.FindApptsByFilter(ctx, repository.NewFilterApptByTrainID(req.TrainDateID))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCheckInApptNotFound
		}
		return nil, ErrCheckInFindApptFail.Wrap(err)
	}
	if len(appts) == 0 {
		return nil, ErrCheckInApptNotFound
	}
	// find training date
	trainDate, err := uc.repo.FindTrainDateByID(ctx, req.TrainDateID)
	if err != nil {
		return nil, ErrCheckInTrainDateNotFound.Wrap(err)
	}
	// check if checkedInBookingIDs is in appts
	apptMap := make(map[string]*entity.Appointment, len(appts))
	for _, appt := range appts {
		apptMap[appt.ID()] = appt
	}
	updatedAppts := make([]*entity.Appointment, 0, len(req.CheckedInBookingIDs))
	// update appts
	for _, id := range req.CheckedInBookingIDs {
		appt, ok := apptMap[id]
		if !ok {
			continue
		}
		err = appt.MarkAsAttended(trainDate.Period().Start())
		if err != nil {
			return nil, ErrCheckInMarkAsAttendedFail.Wrap(err)
		}
		updatedAppts = append(updatedAppts, appt)
	}
	err = uc.repo.UpdateManyAppts(ctx, updatedAppts)
	if err != nil {
		return nil, ErrCheckInUpdateApptFail.Wrap(err)
	}
	return updatedAppts, nil
}

var (
	ErrCheckInNoApptCheckedIn = core.NewUseCaseError(
		"CHECK_IN", "NO_APPT_CHECKED_IN", "no appointment checked in", core.ErrInvalidInput)
	ErrCheckInApptNotFound = core.NewDBError(
		"CHECK_IN", "APPOINTMENT_NOT_FOUND", "appointment not found", core.ErrNotFound)
	ErrCheckInFindApptFail = core.NewDBError(
		"CHECK_IN", "FIND_APPOINTMENT_FAIL", "find appointment fail", core.ErrInternal)
	ErrCheckInTrainDateNotFound = core.NewDBError(
		"CHECK_IN", "TRAIN_DATE_NOT_FOUND", "train date not found", core.ErrNotFound)
	ErrCheckInFindTrainDateFail = core.NewDBError(
		"CHECK_IN", "FIND_TRAIN_DATE_FAIL", "find train date fail", core.ErrInternal)
	ErrCheckInUpdateApptFail = core.NewDBError(
		"CHECK_IN", "UPDATE_APPOINTMENT_FAIL", "update appointment fail", core.ErrInternal)
	ErrCheckInMarkAsAttendedFail = core.NewDomainError(
		"CHECK_IN", "MARK_AS_ATTENDED_FAIL", "mark as attended fail", core.ErrInternal)
)
