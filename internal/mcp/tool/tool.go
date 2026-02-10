package tool

import (
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	"seanAIgent/internal/booking/usecase/core"
	writeTrain "seanAIgent/internal/booking/usecase/traindate/write"
	"seanAIgent/internal/service"
)

var trainingDateService service.TrainingDateService
var batchCreateTrainDateUC core.WriteUseCase[[]writeTrain.ReqCreateTrainDate, []*entity.TrainDate]
var cancelLeaveUC core.WriteUseCase[writeAppt.ReqCancelLeave, *entity.Appointment]

func InitTool(svc service.TrainingDateService, registry *usecase.Registry) {
	trainingDateService = svc
	batchCreateTrainDateUC = registry.BatchCreateTrainDate
	cancelLeaveUC = registry.CancelLeave
}
