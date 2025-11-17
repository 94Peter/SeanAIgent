package tool

import "seanAIgent/internal/service"

var trainingDateService service.TrainingDateService

func InitTool(svc service.TrainingDateService) {
	trainingDateService = svc
}
