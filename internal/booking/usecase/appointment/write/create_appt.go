package write

import (
	"context"
	"time"

	"seanAIgent/internal/booking/domain"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"seanAIgent/internal/event"
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

func NewCreateApptUseCase(repo createApptUseCaseRepo, bus event.Bus) CreateApptUseCase {
	return &createApptUseCase{
		repo: repo,
		bus:  bus,
	}
}

type createApptUseCase struct {
	repo createApptUseCaseRepo
	bus  event.Bus
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

	// 暫時保留同步清理邏輯以確保即時性，改為手動呼叫 repo 清理
	_ = uc.repo.CleanTrainCache(ctx, req.User.UserID())
	_ = uc.repo.CleanStatsCache(ctx, req.User.UserID(), trainDate.Period().Start().Year(), int(trainDate.Period().Start().Month()))

	// 發送領域事件
	for _, appt := range appointments {
		evt := event.NewTypedEvent(uc.repo.GenerateID(), domain.TopicAppointmentStatusChanged, domain.AppointmentStatusChanged{
			BookingID:  appt.ID(),
			UserID:     appt.User().UserID(),
			TrainingID: appt.TrainingID(),
			OldStatus:  "",
			NewStatus:  appt.Status().String(),
			OccurredAt: time.Now(),
		})
		uc.bus.Publish(ctx, evt)
	}

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
