package usecase

import (
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/domain/service"
	readAppt "seanAIgent/internal/booking/usecase/appointment/read"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	"seanAIgent/internal/booking/usecase/core"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	writeTrain "seanAIgent/internal/booking/usecase/traindate/write"

	"github.com/google/wire"
)

type Repository interface {
	repository.IdentityGenerator
	repository.TrainRepository
	repository.AppointmentRepository
}

type ServiceAggregator struct {
	service.TrainDateService
}

// 為每個 UseCase 定義一個包裝過的 Provider
func ProvideCreateTrainDateUC(
	repo Repository, svc ServiceAggregator,
) core.WriteUseCase[writeTrain.ReqCreateTrainDate, *entity.TrainDate] {
	return core.WithWriteOTel(writeTrain.NewCreateTrainDateUseCase(repo, svc))
}

func ProvideBatchCreateTrainDateUC(
	repo Repository, svc ServiceAggregator,
) core.WriteUseCase[[]writeTrain.ReqCreateTrainDate, []*entity.TrainDate] {
	return core.WithWriteOTel(writeTrain.NewBatchCreateTrainDateUseCase(repo, svc))
}

func ProvideDeleteTrainDateUC(
	repo Repository,
) core.WriteUseCase[writeTrain.ReqDeleteTrainDate, *entity.TrainDate] {
	return core.WithWriteOTel(writeTrain.NewDeleteTrainDateUseCase(repo))
}

func ProvideQueryFutureTrainUC(
	repo Repository,
) core.ReadUseCase[readTrain.ReqQueryFutureTrain, []*entity.TrainDateHasApptState] {
	return core.WithReadOTel(readTrain.NewQueryFutureTrainUseCase(repo))
}

func ProvideUserQueryFutureTrainUC(
	repo Repository,
) core.ReadUseCase[readTrain.ReqUserQueryFutureTrain, []*entity.TrainDateHasUserApptState] {
	return core.WithReadOTel(readTrain.NewUserQueryFutureTrainUseCase(repo))
}

func ProvideUserQueryTrainByIDUC(
	repo Repository,
) core.ReadUseCase[readTrain.ReqUserQueryTrainByID, *entity.TrainDateHasUserApptState] {
	return core.WithReadOTel(readTrain.NewUserQueryTrainByIDUseCase(repo))
}

// Appointment UseCase

func ProvideCreateApptUC(
	repo Repository,
) core.WriteUseCase[writeAppt.ReqCreateAppt, []*entity.Appointment] {
	return core.WithWriteOTel(writeAppt.NewCreateApptUseCase(repo))
}

func ProvideCancelApptUC(
	repo Repository,
) core.WriteUseCase[writeAppt.ReqCancelAppt, *entity.Appointment] {
	return core.WithWriteOTel(writeAppt.NewCancelApptUseCase(repo))
}

func ProvideCheckInUC(
	repo Repository,
) core.WriteUseCase[writeAppt.ReqCheckIn, []*entity.Appointment] {
	return core.WithWriteOTel(writeAppt.NewCheckInUseCase(repo))
}

func ProvideQueryUserBookingsUC(
	repo Repository,
) core.ReadUseCase[readAppt.ReqQueryUserBookings, *readAppt.RespQueryUserBookings] {
	return core.WithReadOTel(readAppt.NewQueryUserBookingsUseCase(repo))
}

func ProvideCreateLeaveUC(
	repo Repository,
) core.WriteUseCase[writeAppt.ReqCreateLeave, *entity.Appointment] {
	return core.WithWriteOTel(writeAppt.NewCreateLeaveUseCase(repo))
}

func ProvideCancelLeaveUC(
	repo Repository,
) core.WriteUseCase[writeAppt.ReqCancelLeave, *entity.Appointment] {
	return core.WithWriteOTel(writeAppt.NewCancelLeaveUseCase(repo))
}

func ProvideFindNearestTrainByTimeUC(
	repo Repository,
) core.ReadUseCase[readTrain.ReqFindNearestTrainByTime, *entity.TrainDateHasApptState] {
	return core.WithReadOTel(readTrain.NewFindNearestTrainByTimeUseCase(repo))
}

var UseCaseSet = wire.NewSet(
	ProvideCreateTrainDateUC,
	ProvideBatchCreateTrainDateUC,
	ProvideDeleteTrainDateUC,
	ProvideQueryFutureTrainUC,
	ProvideUserQueryFutureTrainUC,
	ProvideUserQueryTrainByIDUC,
	ProvideFindNearestTrainByTimeUC,

	ProvideCreateApptUC,
	ProvideCancelApptUC,
	ProvideCheckInUC,
	ProvideQueryUserBookingsUC,
	ProvideCancelLeaveUC,
	ProvideCreateLeaveUC,

	wire.Struct(new(ServiceAggregator), "*"),
	wire.Struct(new(Registry), "*"),
)
