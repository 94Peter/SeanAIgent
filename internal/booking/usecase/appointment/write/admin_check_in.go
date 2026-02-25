package write

import (
	"context"
	"errors"

	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

// AdminCheckInUseCase handles bulk check-in from coach management dashboard.
type ReqAdminCheckIn struct {
	TrainDateID         string
	CheckedInBookingIDs []string
}

type AdminCheckInUseCase core.WriteUseCase[ReqAdminCheckIn, []*entity.Appointment]

type adminCheckInUseCaseRepo interface {
	repository.IdentityGenerator
	repository.AppointmentRepository
	repository.TrainRepository
	repository.StatsRepository
}

func NewAdminCheckInUseCase(repo adminCheckInUseCaseRepo, cw cacheWorker) AdminCheckInUseCase {
	return &adminCheckInUseCase{
		repo: repo,
		cw:   cw,
	}
}

type adminCheckInUseCase struct {
	repo adminCheckInUseCaseRepo
	cw   cacheWorker
}

func (uc *adminCheckInUseCase) Name() string {
	return "AdminCheckIn"
}

func (uc *adminCheckInUseCase) Execute(
	ctx context.Context, req ReqAdminCheckIn,
) ([]*entity.Appointment, core.UseCaseError) {
	train, err := uc.repo.FindTrainDateByID(ctx, req.TrainDateID)
	if err != nil {
		return nil, ErrCheckInTrainNotFound.Wrap(err)
	}

	appts, err := uc.repo.FindApptsByFilter(ctx, repository.NewFilterApptByTrainID(req.TrainDateID))
	if err != nil {
		return nil, ErrCheckInFindApptFail.Wrap(err)
	}

	apptMap := make(map[string]*entity.Appointment)
	for _, a := range appts {
		apptMap[a.ID()] = a
	}

	var updated []*entity.Appointment
	affectedUserIDs := make(map[string]struct{})

	for _, id := range req.CheckedInBookingIDs {
		if a, ok := apptMap[id]; ok {
			if err := a.AdminCheckIn(train.Period().Start()); err != nil {
				return nil, ErrCheckInDomainError.Wrap(err)
			}
			updated = append(updated, a)
			affectedUserIDs[a.User().UserID()] = struct{}{}
		}
	}

	if len(updated) > 0 {
		if err := uc.repo.UpdateManyAppts(ctx, updated); err != nil {
			return nil, ErrCheckInUpdateApptFail.Wrap(err)
		}
		// Use CleanSync to ensure consistency across V1 and V2
		for uid := range affectedUserIDs {
			uc.cw.CleanSync(ctx, uid, req.TrainDateID, train.Period().Start())
		}
	}

	return updated, nil
}

// AdminToggleCheckInUseCase allows toggling a single student's attendance via HTMX/Dashboard
type ReqAdminToggleCheckIn struct {
	BookingID string
}

type AdminToggleCheckInUseCase core.WriteUseCase[ReqAdminToggleCheckIn, *entity.Appointment]

func NewAdminToggleCheckInUseCase(repo adminCheckInUseCaseRepo, cw cacheWorker) AdminToggleCheckInUseCase {
	return &adminToggleCheckInUseCase{repo: repo, cw: cw}
}

type adminToggleCheckInUseCase struct {
	repo adminCheckInUseCaseRepo
	cw   cacheWorker
}

func (uc *adminToggleCheckInUseCase) Name() string {
	return "AdminToggleCheckIn"
}

func (uc *adminToggleCheckInUseCase) Execute(ctx context.Context, req ReqAdminToggleCheckIn) (*entity.Appointment, core.UseCaseError) {
	appt, err := uc.repo.FindApptByID(ctx, req.BookingID)
	if err != nil {
		return nil, ErrCheckInApptNotFound.Wrap(err)
	}

	train, err := uc.repo.FindTrainDateByID(ctx, appt.TrainingID())
	if err != nil {
		return nil, ErrCheckInTrainNotFound.Wrap(err)
	}

	if appt.Status() == entity.StatusAttended {
		appt.AdminRestoreFromLeave() // Reverts to StatusConfirmed
	} else {
		if err := appt.AdminCheckIn(train.Period().Start()); err != nil {
			if errors.Is(err, entity.ErrAppointmentCheckInNotOpen) {
				return nil, ErrCheckInNotOpen.Wrap(err)
			}
			if errors.Is(err, entity.ErrAppointmentCheckInTooLate) {
				return nil, ErrCheckInTooLate.Wrap(err)
			}
			return nil, ErrCheckInDomainError.Wrap(err)
		}
	}

	if err := uc.repo.UpdateAppt(ctx, appt); err != nil {
		return nil, ErrCheckInUpdateApptFail.Wrap(err)
	}

	// Use CleanSync for immediate feedback in V1/V2
	uc.cw.CleanSync(ctx, appt.User().UserID(), appt.TrainingID(), train.Period().Start())
	return appt, nil
}
