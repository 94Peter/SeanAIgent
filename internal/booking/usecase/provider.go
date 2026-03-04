package usecase

import (
	"time"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/domain/service"
	"seanAIgent/internal/booking/infra"
	readAppt "seanAIgent/internal/booking/usecase/appointment/read"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	"seanAIgent/internal/booking/usecase/core"
	readStats "seanAIgent/internal/booking/usecase/stats/read"
	writeStats "seanAIgent/internal/booking/usecase/stats/write"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	writeTrain "seanAIgent/internal/booking/usecase/traindate/write"
	"seanAIgent/internal/event"

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

// 為每個 UseCase 定定義一個包裝過的 Provider
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
	repo Repository, bus event.Bus,
) writeAppt.CreateApptUseCase {
	return core.WithWriteOTel(writeAppt.NewCreateApptUseCase(repo, bus))
}

func ProvideCancelApptUC(
	repo Repository, bus event.Bus,
) writeAppt.CancelApptUseCase {
	return core.WithWriteOTel(writeAppt.NewCancelApptUseCase(repo, bus))
}

func ProvideCheckInUC(
	adminUC writeAppt.AdminCheckInUseCase,
) writeAppt.CheckInUseCase {
	return core.WithWriteOTel(writeAppt.NewCheckInUseCase(adminUC))
}

func ProvideAdminCheckInUC(
	repo Repository, bus event.Bus,
) writeAppt.AdminCheckInUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminCheckInUseCase(repo, bus))
}

func ProvideAdminToggleCheckInUC(
	repo Repository, bus event.Bus,
) writeAppt.AdminToggleCheckInUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminToggleCheckInUseCase(repo, bus))
}

func ProvideAdminCreateLeaveUC(
	repo Repository, bus event.Bus,
) writeAppt.AdminCreateLeaveUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminCreateLeaveUseCase(repo, bus))
}

func ProvideAdminRestoreFromLeaveUC(
	repo Repository, bus event.Bus,
) writeAppt.AdminRestoreFromLeaveUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminRestoreFromLeaveUseCase(repo, bus))
}

func ProvideAdminCreateWalkInUC(
	repo Repository, bus event.Bus,
) writeAppt.AdminCreateWalkInUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminCreateWalkInUseCase(repo, bus))
}

func ProvideAdminQueryStudentsUC(
	repo Repository,
) readStats.AdminQueryStudentsUseCase {
	return core.WithReadOTel(readStats.NewAdminQueryStudentsUseCase(repo))
}

func ProvideAutoMarkAbsentUC(
	repo Repository, bus event.Bus,
) writeAppt.AutoMarkAbsentUseCase {
	return core.WithWriteOTel(writeAppt.NewAutoMarkAbsentUseCase(repo, bus))
}

func ProvideAdminBatchUpdateAttendanceUC(
	repo Repository, bus event.Bus,
) writeAppt.AdminBatchUpdateAttendanceUseCase {
	return core.WithWriteOTel(writeAppt.NewAdminBatchUpdateAttendanceUseCase(repo, bus))
}

func ProvideQueryUserBookingsUC(
	repo Repository,
) core.ReadUseCase[readAppt.ReqQueryUserBookings, *readAppt.RespQueryUserBookings] {
	return core.WithReadOTel(readAppt.NewQueryUserBookingsUseCase(repo))
}

func ProvideCreateLeaveUC(
	repo Repository, bus event.Bus,
) writeAppt.CreateLeaveUseCase {
	return core.WithWriteOTel(writeAppt.NewCreateLeaveUseCase(repo, bus))
}

func ProvideCancelLeaveUC(
	repo Repository, bus event.Bus,
) writeAppt.CancelLeaveUseCase {
	return core.WithWriteOTel(writeAppt.NewCancelLeaveUseCase(repo, bus))
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

func ProvideBatchSyncMonthlyStatsUC(
	repo Repository,
) writeStats.BatchSyncMonthlyStatsUseCase {
	return core.WithWriteOTel(writeStats.NewBatchSyncMonthlyStatsUseCase(repo))
}

func ProvideQueryMonthlyUserReportsUC(
	repo Repository,
) readStats.QueryMonthlyUserReportsUseCase {
	return core.WithReadOTel(readStats.NewQueryMonthlyUserReportsUseCase(repo))
}

func ProvideGetBusinessAnalyticsUC(
	repo Repository,
) readStats.GetBusinessAnalyticsUseCase {
	return core.WithReadOTel(readStats.NewGetBusinessAnalyticsUseCase(repo))
}

func ProvideGetUserDetailUC(
	repo Repository,
) readStats.GetUserDetailUseCase {
	return core.WithReadOTel(readStats.NewGetUserDetailUseCase(repo))
}

func ProvideSubscribers(
	repo Repository,
) []event.Subscriber {
	subs := []event.Subscriber{
		infra.NewCacheSubscriber(repo, repo),
	}
	subs = append(subs, infra.NewUserMonthlyStatsSubscriber(repo, repo)...)
	return subs
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
	ProvideAdminBatchUpdateAttendanceUC,
	ProvideQueryUserBookingsUC,
	ProvideCancelLeaveUC,
	ProvideCreateLeaveUC,
	ProvideAdminCreateLeaveUC,
	ProvideAdminRestoreFromLeaveUC,

	ProvideGetUserMonthlyStatsUC,
	ProvideQueryTwoWeeksScheduleUC,
	ProvideQueryAllUserApptStatsUC,
	ProvideBatchSyncMonthlyStatsUC,
	ProvideQueryMonthlyUserReportsUC,
	ProvideGetBusinessAnalyticsUC,
	ProvideGetUserDetailUC,

	ProvideSubscribers,
	event.EventSet,
	ProvideIdempotencyManager,

	wire.Struct(new(ServiceAggregator), "*"),
	wire.Struct(new(Registry), "*"),
)

func ProvideIdempotencyManager() IdempotencyManager {
	return NewIdempotencyManager(30 * time.Minute)
}
