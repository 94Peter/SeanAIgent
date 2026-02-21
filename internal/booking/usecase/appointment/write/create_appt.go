package write

import (
	"context"
	"time"

	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqCreateAppt struct {
	TrainDateID string
	User        entity.User
	ChildNames  []string
}

type CreateApptUseCase core.WriteUseCase[ReqCreateAppt, []*entity.Appointment]

type createApptUseCaseRepo interface {
	repository.IdentityGenerator
	repository.TrainRepository
	repository.AppointmentRepository
	repository.StatsRepository
}

// 加入 CacheWorker 介面定義或直接使用
type cacheWorker interface {
	Clean(uid string, tid string)
	CleanSync(ctx context.Context, uid string, tid string, startTime time.Time)
}

func NewCreateApptUseCase(repo createApptUseCaseRepo, cw cacheWorker) CreateApptUseCase {
	return &createApptUseCase{
		repo: repo,
		cw:   cw,
	}
}

type createApptUseCase struct {
	repo createApptUseCaseRepo
	cw   cacheWorker
}

func (uc *createApptUseCase) Name() string {
	return "CreateAppt"
}

func (uc *createApptUseCase) Execute(
	ctx context.Context, req ReqCreateAppt,
) ([]*entity.Appointment, core.UseCaseError) {
	trainDate, err := uc.repo.FindTrainDateByID(ctx, req.TrainDateID)
	if err != nil {
		return nil, ErrCreateApptTrainDateNotFound.Wrap(err)
	}

	apptCount := len(req.ChildNames)
	// new appointments
	appointments := make([]*entity.Appointment, 0, apptCount)
	for _, childName := range req.ChildNames {
		apptID := uc.repo.GenerateID()
		appt, err := entity.NewAppointment(
			entity.WithCreateAppt(
				apptID, req.TrainDateID, req.User, childName,
			))
		if err != nil {
			return nil, ErrCreateApptNewDomainEntityFail.Wrap(err)
		}
		appointments = append(appointments, appt)
	}
	// find training date deduct capacity, if failed do not save appointments
	err = uc.repo.DeductCapacity(ctx, req.TrainDateID, apptCount)
	if err != nil {
		return nil, ErrCreateApptDeductCapacityFail.Wrap(err)
	}
	// save appointments
	err = uc.repo.SaveManyAppointments(ctx, appointments)
	if err != nil {
		return nil, ErrCreateApptSaveApptFail.Wrap(err)
	}

	// 使用同步清理，確保跳轉頁面後資料一致
	uc.cw.CleanSync(ctx, req.User.UserID(), req.TrainDateID, trainDate.Period().Start())

	return appointments, nil
}

var (
	ErrCreateApptTrainDateNotFound = core.NewDBError(
		"CREATE_APPT", "TRAIN_DATE_NOT_FOUND", "train date not found", core.ErrNotFound)
	ErrCreateApptNewDomainEntityFail = core.NewDomainError(
		"CREATE_APPT", "DOMAIN_ERROR", "new domain entity failed", core.ErrInvalidInput)
	ErrCreateApptDeductCapacityFail = core.NewDBError(
		"CREATE_APPT", "DEDUCT_CAPACITY_FAIL", "deduct capacity fail", core.ErrConflict)
	ErrCreateApptSaveApptFail = core.NewDBError(
		"CREATE_APPT", "SAVE_APPOINTMENT_FAIL", "save appointment fail", core.ErrInternal)
)
