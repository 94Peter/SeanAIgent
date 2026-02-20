package usecase

import (
	"seanAIgent/internal/booking/domain/entity"
	readAppt "seanAIgent/internal/booking/usecase/appointment/read"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	"seanAIgent/internal/booking/usecase/core"
	migrationv1tov2 "seanAIgent/internal/booking/usecase/migration/v1tov2"
	readStats "seanAIgent/internal/booking/usecase/stats/read"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	writeTrain "seanAIgent/internal/booking/usecase/traindate/write"
)

type Registry struct {
	CreateTrainDate      core.WriteUseCase[writeTrain.ReqCreateTrainDate, *entity.TrainDate]
	BatchCreateTrainDate core.WriteUseCase[[]writeTrain.ReqCreateTrainDate, []*entity.TrainDate]
	DeleteTrainDate      core.WriteUseCase[writeTrain.ReqDeleteTrainDate, *entity.TrainDate]

	QueryFutureTrain       core.ReadUseCase[readTrain.ReqQueryFutureTrain, []*entity.TrainDateHasApptState]
	FindNearestTrainByTime core.ReadUseCase[readTrain.ReqFindNearestTrainByTime, *entity.TrainDateHasApptState]
	FindTrainHasApptsById  core.ReadUseCase[readTrain.ReqFindTrainHasApptsById, *entity.TrainDateHasApptState]
	UserQueryFutureTrain   core.ReadUseCase[readTrain.ReqUserQueryFutureTrain, []*entity.TrainDateHasUserApptState]
	UserQueryTrainByID     core.ReadUseCase[readTrain.ReqUserQueryTrainByID, *entity.TrainDateHasUserApptState]
	AdminQueryTrainRange   core.ReadUseCase[readTrain.ReqAdminQueryTrainRange, []*entity.TrainDateHasApptState]

	CreateAppt  core.WriteUseCase[writeAppt.ReqCreateAppt, []*entity.Appointment]
	CheckIn     core.WriteUseCase[writeAppt.ReqCheckIn, []*entity.Appointment]
	CancelAppt  core.WriteUseCase[writeAppt.ReqCancelAppt, *entity.Appointment]
	CreateLeave core.WriteUseCase[writeAppt.ReqCreateLeave, *entity.Appointment]
	CancelLeave core.WriteUseCase[writeAppt.ReqCancelLeave, *entity.Appointment]

	QueryUserBookings core.ReadUseCase[readAppt.ReqQueryUserBookings, *readAppt.RespQueryUserBookings]
	// FindBooking       core.ReadUseCase[readAppt.ReqFindBooking, *entity.AppointmentWithTrainDate]
	GetUserMonthlyStats   readStats.GetUserMonthlyStatsUseCase
	QueryTwoWeeksSchedule readTrain.QueryTwoWeeksScheduleUseCase
}

type MigrationRegistry struct {
	TrainDataMigrationV1ToV2 migrationv1tov2.TrainDataMigrationUseCase
	ApptMigrationV1ToV2      migrationv1tov2.ApptMigrationUseCase
}
