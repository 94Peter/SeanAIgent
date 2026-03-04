package event

import (
	"sync"
	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	busOnce sync.Once
	bus     Bus
)

func ProvideEventBus(store EventStore) Bus {
	busOnce.Do(func() {
		bus = NewBus(store)
	})
	return bus
}

func ProvideEventStore(db *mongo.Database) (EventStore, error) {
	return NewMongoEventStore(db)
}

var EventSet = wire.NewSet(
	ProvideEventBus,
	ProvideEventStore,
)
