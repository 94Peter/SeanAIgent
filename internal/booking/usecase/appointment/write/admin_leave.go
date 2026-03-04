package write

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqAdminCreateLeave struct {
	BookingID string
	Reason    string
}

type AdminCreateLeaveUseCase core.WriteUseCase[ReqAdminCreateLeave, *entity.Appointment]

func NewAdminCreateLeaveUseCase(repo adminCheckInUseCaseRepo, cw cacheWorker) AdminCreateLeaveUseCase {
	return &adminCreateLeaveUseCase{repo: repo, cw: cw}
}

type adminCreateLeaveUseCase struct {
	repo adminCheckInUseCaseRepo
	cw   cacheWorker
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

	if err := appt.AdminAppendLeave(req.Reason, train.Period().Start()); err != nil {
		return nil, core.NewUseCaseError("ADMIN_LEAVE", "DOMAIN_FAIL", "無法執行請假操作", core.ErrInvalidInput).Wrap(err)
	}

	if err := uc.repo.UpdateAppt(ctx, appt); err != nil {
		return nil, ErrCheckInUpdateApptFail.Wrap(err)
	}

	uc.cw.CleanSync(ctx, appt.User().UserID(), appt.TrainingID(), train.Period().Start())
	return appt, nil
}

type ReqAdminRestoreFromLeave struct {
	BookingID string
}

type AdminRestoreFromLeaveUseCase core.WriteUseCase[ReqAdminRestoreFromLeave, *entity.Appointment]

func NewAdminRestoreFromLeaveUseCase(repo adminCheckInUseCaseRepo, cw cacheWorker) AdminRestoreFromLeaveUseCase {
	return &adminRestoreFromLeaveUseCase{repo: repo, cw: cw}
}

type adminRestoreFromLeaveUseCase struct {
	repo adminCheckInUseCaseRepo
	cw   cacheWorker
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

	if err := appt.AdminRestoreFromLeave(train.Period().Start()); err != nil {
		return nil, core.NewUseCaseError("ADMIN_RESTORE", "DOMAIN_FAIL", "無法還原預約狀態", core.ErrInvalidInput).Wrap(err)
	}

	if err := uc.repo.UpdateAppt(ctx, appt); err != nil {
		return nil, ErrCheckInUpdateApptFail.Wrap(err)
	}

	uc.cw.CleanSync(ctx, appt.User().UserID(), appt.TrainingID(), train.Period().Start())
	return appt, nil
}
