package tool

import (
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	"seanAIgent/internal/booking/usecase/core"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	writeTrain "seanAIgent/internal/booking/usecase/traindate/write"
)

var batchCreateTrainDateUC core.WriteUseCase[[]writeTrain.ReqCreateTrainDate, []*entity.TrainDate]
var cancelLeaveUC core.WriteUseCase[writeAppt.ReqCancelLeave, *entity.Appointment]
var adminQueryTrainRangeUC core.ReadUseCase[readTrain.ReqAdminQueryTrainRange, []*entity.TrainDateHasApptState]

func InitTool(registry *usecase.Registry) {
	batchCreateTrainDateUC = registry.BatchCreateTrainDate
	cancelLeaveUC = registry.CancelLeave
	adminQueryTrainRangeUC = registry.AdminQueryTrainRange
}
