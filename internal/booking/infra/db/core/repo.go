package core

import (
	"seanAIgent/internal/booking/domain/repository"
)

type DbRepository interface {
	repository.AppointmentRepository
	repository.TrainRepository
	repository.IdentityGenerator
	repository.StatsRepository
}
