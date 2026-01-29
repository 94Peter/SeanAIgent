package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrLeaveHasAppointmentHasTraining() *AggrLeaveHasAppointmentHasTraining {
	return &AggrLeaveHasAppointmentHasTraining{
		Index: leaveCollection,
	}
}

type AggrLeaveHasAppointmentHasTraining struct {
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	mgo.Index   `bson:"-"`
	BookingInfo *struct {
		CreatedAt time.Time     `bson:"created_at"`
		ID        bson.ObjectID `bson:"_id,omitempty"`
	} `bson:"booking_info"`
	TrainingInfo *struct {
		StartDate time.Time     `bson:"start_date"`
		EndDate   time.Time     `bson:"end_date"`
		Location  string        `bson:"location"`
		Timezone  string        `bson:"timezone"`
		Capacity  int           `bson:"capacity"`
		ID        bson.ObjectID `bson:"_id,omitempty"`
	} `bson:"training_info"`
	UserID    string        `bson:"user_id"`
	ChildName string        `bson:"childName"`
	Reason    string        `bson:"reason"`
	Status    LeaveStatus   `bson:"status"`
	ID        bson.ObjectID `bson:"_id,omitempty"`
}

func (aggr *AggrLeaveHasAppointmentHasTraining) GetPipeline(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		{{"$match", q}},
		{{"$lookup", bson.M{"from": AppointmentCollectionName, "localField": "booking_id", "foreignField": "_id", "as": "booking_info"}}},
		{{"$unwind", bson.M{"path": "$booking_info"}}},
		{{"$lookup", bson.M{"from": TrainingDateCollectionName, "localField": "booking_info.training_date_id", "foreignField": "_id", "as": "training_info"}}},
		{{"$unwind", bson.M{"path": "$training_info"}}},
	}
	return pipeline
}
