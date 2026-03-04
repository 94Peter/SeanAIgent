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

func NewAdminCheckInUseCase(repo adminCheckInUseCaseRepo, bus event.Bus) AdminCheckInUseCase {
	return &adminCheckInUseCase{
		repo: repo,
		bus:  bus,
	}
}

type adminCheckInUseCase struct {
	repo adminCheckInUseCaseRepo
	bus  event.Bus
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
			oldStatus := a.Status().String()
			if err := a.AdminCheckIn(train.Period().Start()); err != nil {
				return nil, ErrCheckInDomainError.Wrap(err)
			}
			updated = append(updated, a)
			affectedUserIDs[a.User().UserID()] = struct{}{}

			// 發送領域事件
			evt := event.NewTypedEvent(uc.repo.GenerateID(), domain.TopicAppointmentStatusChanged, domain.AppointmentStatusChanged{
				BookingID:  a.ID(),
				UserID:     a.User().UserID(),
				TrainingID: a.TrainingID(),
				OldStatus:  oldStatus,
				NewStatus:  a.Status().String(),
				OccurredAt: time.Now(),
			})
			uc.bus.Publish(ctx, evt)
		}
	}

	if len(updated) > 0 {
		if err := uc.repo.UpdateManyAppts(ctx, updated); err != nil {
			return nil, ErrCheckInUpdateApptFail.Wrap(err)
		}
		// 手動清理快取
		for uid := range affectedUserIDs {
			_ = uc.repo.CleanTrainCache(ctx, uid)
			_ = uc.repo.CleanStatsCache(ctx, uid, train.Period().Start().Year(), int(train.Period().Start().Month()))
		}
	}

	return updated, nil
}

// AdminToggleCheckInUseCase allows toggling a single student's attendance via HTMX/Dashboard
type ReqAdminToggleCheckIn struct {
	BookingID string
}

type AdminToggleCheckInUseCase core.WriteUseCase[ReqAdminToggleCheckIn, *entity.Appointment]

func NewAdminToggleCheckInUseCase(repo adminCheckInUseCaseRepo, bus event.Bus) AdminToggleCheckInUseCase {
	return &adminToggleCheckInUseCase{repo: repo, bus: bus}
}

type adminToggleCheckInUseCase struct {
	repo adminCheckInUseCaseRepo
	bus  event.Bus
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

	oldStatus := appt.Status().String()
	if appt.Status() == entity.StatusAttended {
		if err := appt.AdminRestoreFromLeave(train.Period().Start()); err != nil {
			return nil, ErrCheckInDomainError.Wrap(err)
		}
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

	// 手動清理快取
	_ = uc.repo.CleanTrainCache(ctx, appt.User().UserID())
	_ = uc.repo.CleanStatsCache(ctx, appt.User().UserID(), train.Period().Start().Year(), int(train.Period().Start().Month()))

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
