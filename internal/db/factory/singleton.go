package factory

import (
	"context"
	"errors"

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
	if defaultConfig.tracer == nil {
		return errors.New("tracer not set")
	}
	var err error
	closeDbFunc, err = mongodb.IniMongodb(
		ctx,
		defaultConfig.mongo.URI, defaultConfig.mongo.DB,
		defaultConfig.mongo.MaxPoolSize,
		defaultConfig.mongo.MinPoolSize,
		defaultConfig.tracer)
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

// 依然保留原本的 option 邏輯
func ProvideStores(ctx context.Context, opts ...option) (*Stores, func(), error) {
	// 1. 執行原本的 option 邏輯
	cfg := defaultConfig // 假設你有個 default
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.tracer == nil {
		return nil, nil, errors.New("tracer not set")
	}

	// 2. 初始化底層驅動
	closeDbFunc, err := mongodb.IniMongodb(ctx,
		cfg.mongo.URI, cfg.mongo.DB,
		cfg.mongo.MaxPoolSize, cfg.mongo.MinPoolSize,
		cfg.tracer,
	)
	if err != nil {
		return nil, nil, err
	}

	s := &Stores{
		TrainingDateStore: mongodb.NewTrainingDateStore(),
		AppointmentStore:  mongodb.NewAppointmentStore(),
	}

	cleanup := func() { closeDbFunc(context.Background()) }
	return s, cleanup, nil
}
