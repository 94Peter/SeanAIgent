package usecase

import (
	"time"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/domain/service"
	readAppt "seanAIgent/internal/booking/usecase/appointment/read"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	"seanAIgent/internal/booking/usecase/core"
	readStats "seanAIgent/internal/booking/usecase/stats/read"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	writeTrain "seanAIgent/internal/booking/usecase/traindate/write"

	"github.com/google/wire"
)

type Repository interface {
	repository.IdentityGenerator
	repository.TrainRepository
	repository.AppointmentRepository
	repository.StatsRepository
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
	repo Repository, cw CacheWorker,
) writeAppt.CreateApptUseCase {
	return core.WithWriteOTel(writeAppt.NewCreateApptUseCase(repo, cw))
}

func ProvideCancelApptUC(
	repo Repository, cw CacheWorker,
) writeAppt.CancelApptUseCase {
	return core.WithWriteOTel(writeAppt.NewCancelApptUseCase(repo, cw))
}

func ProvideCheckInUC(
	adminUC writeAppt.AdminCheckInUseCase,
) writeAppt.CheckInUseCase {
	return core.WithWriteOTel(writeAppt.NewCheckInUseCase(adminUC))
}

func ProvideAdminCheckInUC(
	repo Repository, cw CacheWorker,
) writeAppt.AdminCheckInUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminCheckInUseCase(repo, cw))
}

func ProvideAdminToggleCheckInUC(
	repo Repository, cw CacheWorker,
) writeAppt.AdminToggleCheckInUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminToggleCheckInUseCase(repo, cw))
}

func ProvideAdminCreateLeaveUC(
	repo Repository, cw CacheWorker,
) writeAppt.AdminCreateLeaveUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminCreateLeaveUseCase(repo, cw))
}

func ProvideAdminRestoreFromLeaveUC(
	repo Repository, cw CacheWorker,
) writeAppt.AdminRestoreFromLeaveUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminRestoreFromLeaveUseCase(repo, cw))
}

func ProvideAdminCreateWalkInUC(
	repo Repository, cw CacheWorker,
) writeAppt.AdminCreateWalkInUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminCreateWalkInUseCase(repo, cw))
}

func ProvideAdminQueryStudentsUC(
	repo Repository,
) readStats.AdminQueryStudentsUseCase {
	return core.WithReadOTel(readStats.NewAdminQueryStudentsUseCase(repo))
}

func ProvideAutoMarkAbsentUC(
	repo Repository,
) writeAppt.AutoMarkAbsentUseCase {
	return core.WithWriteOTel(writeAppt.NewAutoMarkAbsentUseCase(repo))
}

func ProvideQueryUserBookingsUC(
	repo Repository,
) core.ReadUseCase[readAppt.ReqQueryUserBookings, *readAppt.RespQueryUserBookings] {
	return core.WithReadOTel(readAppt.NewQueryUserBookingsUseCase(repo))
}

func ProvideCreateLeaveUC(
	repo Repository, cw CacheWorker,
) writeAppt.CreateLeaveUseCase {
	return core.WithWriteOTel(writeAppt.NewCreateLeaveUseCase(repo, cw))
}

func ProvideCancelLeaveUC(
	repo Repository, cw CacheWorker,
) writeAppt.CancelLeaveUseCase {
	return core.WithWriteOTel(writeAppt.NewCancelLeaveUseCase(repo, cw))
}

func ProvideFindNearestTrainByTimeUC(
	repo Repository,
) core.ReadUseCase[readTrain.ReqFindNearestTrainByTime, *entity.TrainDateHasApptState] {
	return core.WithReadOTel(readTrain.NewFindNearestTrainByTimeUseCase(repo))
}

func ProvideFindTrainHasApptsByIdUC(
	repo Repository,
) core.ReadUseCase[readTrain.ReqFindTrainHasApptsById, *entity.TrainDateHasApptState] {
	return core.WithReadOTel(readTrain.NewFindTrainHasApptsByIdUseCase(repo))
}

func ProvideAdminQueryTrainRangeUC(
	repo Repository,
) core.ReadUseCase[readTrain.ReqAdminQueryTrainRange, []*entity.TrainDateHasApptState] {
	return core.WithReadOTel(readTrain.NewAdminQueryTrainRangeUseCase(repo))
}

func ProvideGetUserMonthlyStatsUC(
	repo Repository,
) readStats.GetUserMonthlyStatsUseCase {
	return core.WithReadOTel(readStats.NewGetUserMonthlyStatsUseCase(repo))
}

func ProvideQueryTwoWeeksScheduleUC(
	repo Repository,
) readTrain.QueryTwoWeeksScheduleUseCase {
	return core.WithReadOTel(readTrain.NewQueryTwoWeeksScheduleUseCase(repo))
}

func ProvideAdminQueryRecentTrainUC(
	repo Repository,
) core.ReadUseCase[readTrain.ReqAdminQueryRecentTrain, []*entity.TrainDate] {
	return core.WithReadOTel(readTrain.NewAdminQueryRecentTrainUseCase(repo))
}

func ProvideQueryAllUserApptStatsUC(
	repo Repository,
) core.ReadUseCase[readStats.ReqQueryAllUserApptStats, []*entity.UserApptStats] {
	return core.WithReadOTel(readStats.NewQueryAllUserApptStatsUseCase(repo))
}

var UseCaseSet = wire.NewSet(
	ProvideCreateTrainDateUC,
	ProvideBatchCreateTrainDateUC,
	ProvideDeleteTrainDateUC,
	ProvideQueryFutureTrainUC,
	ProvideUserQueryFutureTrainUC,
	ProvideUserQueryTrainByIDUC,
	ProvideFindNearestTrainByTimeUC,
	ProvideFindTrainHasApptsByIdUC,
	ProvideAdminQueryTrainRangeUC,
	ProvideAdminQueryRecentTrainUC,

	ProvideCreateApptUC,
	ProvideCancelApptUC,
	ProvideCheckInUC,
	ProvideAdminCheckInUC,
	ProvideAdminToggleCheckInUC,
	ProvideAdminCreateWalkInUC,
	ProvideAdminQueryStudentsUC,
	ProvideAutoMarkAbsentUC,
	ProvideQueryUserBookingsUC,
	ProvideCancelLeaveUC,
	ProvideCreateLeaveUC,
	ProvideAdminCreateLeaveUC,
	ProvideAdminRestoreFromLeaveUC,

	ProvideGetUserMonthlyStatsUC,
	ProvideQueryTwoWeeksScheduleUC,
	ProvideQueryAllUserApptStatsUC,

	NewCacheWorker,
	ProvideIdempotencyManager,

	wire.Struct(new(ServiceAggregator), "*"),
	wire.Struct(new(Registry), "*"),
)

func ProvideIdempotencyManager() IdempotencyManager {
	return NewIdempotencyManager(30 * time.Minute)
}
