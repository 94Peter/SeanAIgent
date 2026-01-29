package mongo

import (
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/infra/db/core"
	"seanAIgent/internal/booking/infra/db/mongo/appointment"
	"seanAIgent/internal/booking/infra/db/mongo/stats"
	"seanAIgent/internal/booking/infra/db/mongo/train"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func NewRepoAndIdGenerate() core.DbRepository {
	repoImpl := &dbRepoImpl{
		AppointmentRepository: appointment.NewApptRepository(),
		TrainRepository:       train.NewTrainRepository(),
		StatsRepository:       stats.NewStatsRepository(),
	}
	return repoImpl
}

type dbRepoImpl struct {
	repository.AppointmentRepository
	repository.TrainRepository
	repository.StatsRepository
}

func (dbRepoImpl) GenerateID() string {
	return bson.NewObjectID().Hex()
}
