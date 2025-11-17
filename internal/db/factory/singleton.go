package factory

import (
	"context"

	"seanAIgent/internal/db"
	"seanAIgent/internal/db/mongodb"
)

type Stores struct {
	TrainingDateStore db.TrainingDateStore
	AppointmentStore  db.AppointmentStore
}

var (
	store *Stores

	closeDbFunc db.CloseDbFunc
)

func InitializeDb(ctx context.Context, opts ...option) error {
	for _, opt := range opts {
		opt(defaultConfig)
	}
	var err error
	closeDbFunc, err = mongodb.IniMongodb(ctx, defaultConfig.mongo.URI, defaultConfig.mongo.DB)
	if err != nil {
		return err
	}
	store = &Stores{
		TrainingDateStore: mongodb.NewTrainingDateStore(),
		AppointmentStore:  mongodb.NewAppointmentStore(),
	}
	return nil
}

func Close(ctx context.Context) error {
	return closeDbFunc(ctx)
}

func InjectStore(f func(*Stores)) {
	f(store)
}
