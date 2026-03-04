package usecase

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	readAppt "seanAIgent/internal/booking/usecase/appointment/read"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	"seanAIgent/internal/booking/usecase/core"
	readStats "seanAIgent/internal/booking/usecase/stats/read"
	writeStats "seanAIgent/internal/booking/usecase/stats/write"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	writeTrain "seanAIgent/internal/booking/usecase/traindate/write"
	"seanAIgent/internal/event"
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
	AdminQueryRecentTrain  core.ReadUseCase[readTrain.ReqAdminQueryRecentTrain, []*entity.TrainDate]

	CreateAppt  writeAppt.CreateApptUseCase
	CheckIn     writeAppt.CheckInUseCase
	CancelAppt  writeAppt.CancelApptUseCase
	CreateLeave writeAppt.CreateLeaveUseCase
	CancelLeave writeAppt.CancelLeaveUseCase

	AdminCheckIn          writeAppt.AdminCheckInUseCase
	AdminToggleCheckIn    writeAppt.AdminToggleCheckInUseCase
	AdminCreateLeave      writeAppt.AdminCreateLeaveUseCase
	AdminRestoreFromLeave writeAppt.AdminRestoreFromLeaveUseCase
	AdminCreateWalkIn     writeAppt.AdminCreateWalkInUseCase
	AdminQueryStudents    readStats.AdminQueryStudentsUseCase
	AutoMarkAbsent        writeAppt.AutoMarkAbsentUseCase
	AdminBatchUpdateAttendance writeAppt.AdminBatchUpdateAttendanceUseCase

	QueryUserBookings core.ReadUseCase[readAppt.ReqQueryUserBookings, *readAppt.RespQueryUserBookings]
	// FindBooking       core.ReadUseCase[readAppt.ReqFindBooking, *entity.AppointmentWithTrainDate]
	GetUserMonthlyStats   readStats.GetUserMonthlyStatsUseCase
	QueryTwoWeeksSchedule readTrain.QueryTwoWeeksScheduleUseCase
	QueryAllUserApptStats core.ReadUseCase[readStats.ReqQueryAllUserApptStats, []*entity.UserApptStats]

	BatchSyncMonthlyStats writeStats.BatchSyncMonthlyStatsUseCase

	Bus                event.Bus
	Subscribers        []event.Subscriber
	IdempotencyManager IdempotencyManager
}

func (r *Registry) Start(ctx context.Context) {
	if r.Bus != nil {
		for _, s := range r.Subscribers {
			r.Bus.Subscribe(s.Topic(), s)
		}
	}
}
