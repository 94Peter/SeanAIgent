package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrTrainingdateHasCheckinItems() *AggrTrainingdateHasCheckinItems {
	return &AggrTrainingdateHasCheckinItems{
		Index: trainingDateCollection,
	}
}

// AggrTrainingdateHasCheckinItems represents a single TrainingDate document
type AggrTrainingdateHasCheckinItems struct {
	mgo.Index    `bson:"-"`
	ID           bson.ObjectID      `bson:"_id"` // This is the TrainingDate ID
	Date         string             `bson:"date"`
	Location     string             `bson:"location"`
	StartDate    time.Time          `bson:"start_date"`
	EndDate      time.Time          `bson:"end_date"`
	TimeZone     string             `bson:"timezone"`
	CheckinItems []*aggrCheckinItem `bson:"checkin_items"` // Slice of simplified AggrCheckinItem
	OnLeaveItems []*aggrCheckinItem `bson:"on_leave_items"`
}

type aggrCheckinItem struct {
	ID          bson.ObjectID `bson:"_id"` // This is the Appointment ID
	UserID      string        `bson:"user_id"`
	UserName    string        `bson:"user_name"`
	ChildName   string        `bson:"child_name,omitempty"`
	IsCheckedIn bool          `bson:"is_checked_in"`
	IsOnLeave   bool          `bson:"is_on_leave"`
}

func (aggr *AggrTrainingdateHasCheckinItems) GetPipeline(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		{{"$match", q}},
		{{"$lookup", bson.D{
			{"from", "appointment"},
			{"localField", "_id"},
			{"foreignField", "training_date_id"},
			{"as", "appointments"},
		}}},
		{{"$addFields", bson.D{
			{"checkin_items", bson.D{
				{"$filter", bson.D{
					{"input", "$appointments"},
					{"as", "item"},
					{"cond", bson.D{
						{"$eq", bson.A{bson.D{{"$ifNull", bson.A{"$$item.is_on_leave", false}}}, false}},
					}},
				}},
			}},
			{"on_leave_items", bson.D{
				{"$filter", bson.D{
					{"input", "$appointments"},
					{"as", "item"},
					{"cond", bson.D{
						{"$eq", bson.A{bson.D{{"$ifNull", bson.A{"$$item.is_on_leave", false}}}, true}},
					}},
				}},
			}},
		}}},
		{{"$project", bson.D{
			{"appointments", 0},
		}}},
	}
	return pipeline
}
