package event

import (
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func ProvideEventBus(store EventStore) Bus {
	return NewBus(store)
}

func ProvideEventStore(db *mongo.Database) (EventStore, error) {
	return NewMongoEventStore(db)
}

var EventSet = wire.NewSet(
	ProvideEventBus,
	ProvideEventStore,
)
