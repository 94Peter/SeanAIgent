package write

import (
	"context"
	"time"

	"seanAIgent/internal/booking/domain"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase/core"
	"seanAIgent/internal/event"
)

type ReqAdminCreateLeave struct {
	BookingID string
	Reason    string
}

type AdminCreateLeaveUseCase core.WriteUseCase[ReqAdminCreateLeave, *entity.Appointment]

func NewAdminCreateLeaveUseCase(repo adminCheckInUseCaseRepo, bus event.Bus) AdminCreateLeaveUseCase {
	return &adminCreateLeaveUseCase{repo: repo, bus: bus}
}

type adminCreateLeaveUseCase struct {
	repo adminCheckInUseCaseRepo
	bus  event.Bus
}

func (uc *adminCreateLeaveUseCase) Name() string {
	return "AdminCreateLeave"
}

func (uc *adminCreateLeaveUseCase) Execute(ctx context.Context, req ReqAdminCreateLeave) (*entity.Appointment, core.UseCaseError) {
	appt, err := uc.repo.FindApptByID(ctx, req.BookingID)
	if err != nil {
		return nil, ErrCheckInApptNotFound.Wrap(err)
	}

	train, err := uc.repo.FindTrainDateByID(ctx, appt.TrainingID())
	if err != nil {
		return nil, ErrCheckInTrainNotFound.Wrap(err)
	}

	oldStatus := appt.Status().String()
	if err := appt.AdminAppendLeave(req.Reason, train.Period().Start()); err != nil {
		return nil, core.NewUseCaseError("ADMIN_LEAVE", "DOMAIN_FAIL", "無法執行請假操作", core.ErrInvalidInput).Wrap(err)
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

type ReqAdminRestoreFromLeave struct {
	BookingID string
}

type AdminRestoreFromLeaveUseCase core.WriteUseCase[ReqAdminRestoreFromLeave, *entity.Appointment]

func NewAdminRestoreFromLeaveUseCase(repo adminCheckInUseCaseRepo, bus event.Bus) AdminRestoreFromLeaveUseCase {
	return &adminRestoreFromLeaveUseCase{repo: repo, bus: bus}
}

type adminRestoreFromLeaveUseCase struct {
	repo adminCheckInUseCaseRepo
	bus  event.Bus
}

func (uc *adminRestoreFromLeaveUseCase) Name() string {
	return "AdminRestoreFromLeave"
}

func (uc *adminRestoreFromLeaveUseCase) Execute(ctx context.Context, req ReqAdminRestoreFromLeave) (*entity.Appointment, core.UseCaseError) {
	appt, err := uc.repo.FindApptByID(ctx, req.BookingID)
	if err != nil {
		return nil, ErrCheckInApptNotFound.Wrap(err)
	}

	train, err := uc.repo.FindTrainDateByID(ctx, appt.TrainingID())
	if err != nil {
		return nil, ErrCheckInTrainNotFound.Wrap(err)
	}

	oldStatus := appt.Status().String()
	if err := appt.AdminRestoreFromLeave(train.Period().Start()); err != nil {
		return nil, core.NewUseCaseError("ADMIN_RESTORE", "DOMAIN_FAIL", "無法還原預約狀態", core.ErrInvalidInput).Wrap(err)
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
