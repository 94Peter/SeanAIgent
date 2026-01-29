package model

import (
	"sync"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const TrainingDateCollectionName = "training_date"

var trainingDateCollection = mgo.NewCollectDef(TrainingDateCollectionName, func() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "start_date", Value: 1}, {Key: "end_date", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
})

func init() {
	mgo.RegisterIndex(trainingDateCollection)
}

func NewTrainingDate() *TrainingDate {
	return &TrainingDate{
		Index: trainingDateCollection,
		ID:    bson.NewObjectID(),
	}
}

type TrainingDate struct {
	StartDate time.Time `bson:"start_date"`
	EndDate   time.Time `bson:"end_date"`
	mgo.Index `bson:"-"`
	UserID    string        `bson:"user_id"`
	Date      string        `bson:"date"`
	Location  string        `bson:"location"`
	Timezone  string        `bson:"timezone"`
	Capacity  int           `bson:"capacity"`
	ID        bson.ObjectID `bson:"_id"`
}

func (s *TrainingDate) GetId() any {
	if s.ID.IsZero() {
		return nil
	}
	return s.ID
}

func (s *TrainingDate) SetId(id any) {
	oid, ok := id.(bson.ObjectID)
	if !ok {
		return
	}
	s.ID = oid
}

func (p *TrainingDate) Validate() error {
	return nil
}

var locationCache sync.Map

func ToTime(t time.Time, timezone string) time.Time {
	if loc, ok := locationCache.Load(timezone); ok {
		return t.In(loc.(*time.Location))
	}

	// 快取沒中，才載入並存入快取
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// 建議不要 panic，在生產環境改為記錄日誌並回傳 UTC
		return t.UTC()
	}

	locationCache.Store(timezone, loc)
	return t.In(loc)
}
