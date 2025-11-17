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
	mgo.Index   `bson:"-"`
	ID          bson.ObjectID `bson:"_id,omitempty"`
	UserID      string        `bson:"user_id"`   // User who requested the leave
	ChildName   string        `bson:"childName"` // Name of the child (optional)
	Reason      string        `bson:"reason"`    // Reason for the leave (optional)
	Status      LeaveStatus   `bson:"status"`    // Status of the leave request (e.g., Pending, Approved, Rejected)
	CreatedAt   time.Time     `bson:"created_at"`
	UpdatedAt   time.Time     `bson:"updated_at"`
	BookingInfo *struct {
		ID        bson.ObjectID `bson:"_id,omitempty"`
		CreatedAt time.Time     `bson:"created_at"`
	} `bson:"booking_info"`
	TrainingInfo *struct {
		ID        bson.ObjectID `bson:"_id,omitempty"`
		Location  string        `bson:"location"`
		Capacity  int           `bson:"capacity"`
		StartDate time.Time     `bson:"start_date"`
		EndDate   time.Time     `bson:"end_date"`
		Timezone  string        `bson:"timezone"`
	} `bson:"training_info"`
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
