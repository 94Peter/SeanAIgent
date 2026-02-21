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
	appt, repoErr := uc.repo.FindApptByID(ctx, req.ApptID)
	if repoErr != nil {
		if errors.Is(repoErr, repository.ErrNotFound) {
			return nil, ErrCancelApptApptNotFound
		}
		return nil, ErrCancelApptFindApptFail.Wrap(repoErr)
	}
	err := appt.CancelAsMistake(req.UserID)
	if err != nil {
		return nil, ErrCancelApptCancelApptFail.Wrap(err)
	}

	// 取得課程資訊以獲得 startTime
	trainDate, repoErr := uc.repo.FindTrainDateByID(ctx, appt.TrainingID())
	if repoErr != nil {
		return nil, ErrCancelApptTrainDateNotFound.Wrap(repoErr)
	}

	// 2. 增加名額
	repoErr = uc.repo.IncreaseCapacity(ctx, appt.TrainingID(), 1)
	if repoErr != nil {
		return nil, ErrCancelApptIncreaseCapacityFail.Wrap(repoErr)
	}
	// 3. 刪除 appointment
	repoErr = uc.repo.DeleteAppointment(ctx, appt)
	if repoErr != nil {
		// rollback
		_ = uc.repo.DeductCapacity(ctx, appt.TrainingID(), 1)
		return nil, ErrCancelApptDeleteApptFail.Wrap(repoErr)
	}

	// 使用同步清理，確保跳轉頁面後資料一致
	uc.cw.CleanSync(ctx, req.UserID, appt.TrainingID(), trainDate.Period().Start())

	return appt, nil
}

var (
	ErrCancelApptApptNotFound = core.NewDBError(
		"CANCEL_APPT", "APPOINTMENT_NOT_FOUND", "appointment not found", core.ErrNotFound)
	ErrCancelApptFindApptFail = core.NewDBError(
		"CANCEL_APPT", "FIND_APPOINTMENT_FAIL", "find appointment fail", core.ErrInternal)
	ErrCancelApptTrainDateNotFound = core.NewDBError(
		"CANCEL_APPT", "TRAIN_DATE_NOT_FOUND", "train date not found", core.ErrNotFound)
	ErrCancelApptCancelApptFail = core.NewDomainError(
		"CANCEL_APPT", "CANCEL_APPOINTMENT_FAIL", "cancel appointment fail", core.ErrInternal)
	ErrCancelApptDeleteApptFail = core.NewDBError(
		"CANCEL_APPT", "DELETE_APPOINTMENT_FAIL", "delete appointment fail", core.ErrInternal)
	ErrCancelApptIncreaseCapacityFail = core.NewDBError(
		"CANCEL_APPT", "INCREASE_CAPACITY_FAIL", "increase capacity fail", core.ErrConflict)
)
