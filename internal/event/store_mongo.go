package event

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	eventLogCollection       = "event_logs"
	eventProgressCollection = "event_subscribers"
)

type mongoEventStore struct {
	db *mongo.Database
}

func NewMongoEventStore(db *mongo.Database) (EventStore, error) {
	s := &mongoEventStore{db: db}
	if err := s.initIndexes(context.Background()); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *mongoEventStore) initIndexes(ctx context.Context) error {
	// 建立 TTL 索引，30 天後自動刪除
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "occurred_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(30 * 24 * 60 * 60),
	}
	_, err := s.db.Collection(eventLogCollection).Indexes().CreateOne(ctx, indexModel)
	return err
}

type eventDoc struct {
	ID         string    `bson:"_id"`
	Topic      string    `bson:"topic"`
	OccurredAt time.Time `bson:"occurred_at"`
	Data       []byte    `bson:"data"` // 優化: 直接儲存二進制數據
}

func (s *mongoEventStore) Save(ctx context.Context, e Event) error {
	doc := eventDoc{
		ID:         e.ID(),
		Topic:      e.Topic(),
		OccurredAt: e.OccurredAt(),
		Data:       e.Data(),
	}
	_, err := s.db.Collection(eventLogCollection).UpdateOne(
		ctx,
		bson.M{"_id": doc.ID},
		bson.M{"$set": doc},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (s *mongoEventStore) UpdateProgress(ctx context.Context, subscriberID string, eventID string) error {
	_, err := s.db.Collection(eventProgressCollection).UpdateOne(
		ctx,
		bson.M{"_id": subscriberID},
		bson.M{
			"$set": bson.M{
				"last_processed_event_id": eventID,
				"updated_at":              time.Now(),
			},
		},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (s *mongoEventStore) FindUnprocessedEvents(ctx context.Context, subscriberID string, topic string) ([]Event, error) {
	var progress struct {
		LastEventID string `bson:"last_processed_event_id"`
	}
	err := s.db.Collection(eventProgressCollection).FindOne(ctx, bson.M{"_id": subscriberID}).Decode(&progress)
	
	query := bson.M{"topic": topic}
	if err == nil && progress.LastEventID != "" {
		query["_id"] = bson.M{"$gt": progress.LastEventID}
	} else if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	cursor, err := s.db.Collection(eventLogCollection).Find(ctx, query, options.Find().SetSort(bson.D{{Key: "occurred_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []Event
	for cursor.Next(ctx) {
		var doc eventDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		events = append(events, &genericEvent{
			id:         doc.ID,
			topic:      doc.Topic,
			occurredAt: doc.OccurredAt,
			data:       doc.Data,
		})
	}
	return events, nil
}

type genericEvent struct {
	id         string
	topic      string
	occurredAt time.Time
	data       []byte
}

func (e *genericEvent) ID() string          { return e.id }
func (e *genericEvent) Topic() string       { return e.topic }
func (e *genericEvent) OccurredAt() time.Time { return e.occurredAt }
func (e *genericEvent) Data() []byte        { return e.data }
