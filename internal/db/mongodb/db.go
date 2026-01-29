package mongodb

import (
	"context"
	"fmt"

	"github.com/94peter/vulpes/db/mgo"
	"go.opentelemetry.io/otel/trace"

	"seanAIgent/internal/db"
)

const defaultLimit = 100

func IniMongodb(ctx context.Context, uri string, dbName string, tracer trace.Tracer) (db.CloseDbFunc, error) {
	err := mgo.InitConnection(ctx, dbName, tracer, mgo.WithURI(uri), mgo.WithMinPoolSize(50), mgo.WithMaxPoolSize(100))
	if err != nil {
		return nil, err
	}
	err = mgo.SyncIndexes(ctx)
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context) error {
		return mgo.Close(ctx)
	}, nil
}

func errorWrap(dbErr error, err error) error {
	return fmt.Errorf("%w: %w", dbErr, err)
}
