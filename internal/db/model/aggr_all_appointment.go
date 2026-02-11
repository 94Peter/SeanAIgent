package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggAllAppointment() *AggrAllAppointment {
	return &AggrAllAppointment{
		Index: appointmentCollection,
	}
}

// AggrUserAppointment represents the result of an aggregation that joins
// an appointment with its corresponding training date information.
// It's used to display a user's list of their own bookings.
type AggrAllAppointment struct {
	CreatedAt    time.Time `bson:"created_at"`
	UpdateAt     time.Time `bson:"update_at"`
	mgo.Index    `bson:"-"`
	TrainingDate *struct {
		Date      string    `bson:"date"`
		Location  string    `bson:"location"`
		StartDate time.Time `bson:"start_date"`
		EndDate   time.Time `bson:"end_date"`
		Timezone  string    `bson:"timezone"`
	} `bson:"training_date_info"`
	LeaveInfo *struct {
		CreatedAt time.Time `bson:"created_at"`
		Reason    string    `bson:"reason"`
		Status    string    `bson:"status"`
	} `bson:"leave_info"`
	UserID         string        `bson:"user_id"`
	UserName       string        `bson:"user_name"`
	ChildName      string        `bson:"child_name,omitempty"`
	ID             bson.ObjectID `bson:"_id"`
	TrainingDateId bson.ObjectID `bson:"training_date_id"`
	IsCheckedIn    bool          `bson:"is_checked_in"`
	IsOnLeave      bool          `bson:"is_on_leave"`
}

func (aggr *AggrAllAppointment) GetPipeline(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		// Match appointments for the given user
		{{"$match", q}},
		// Lookup training date information
		{{"$lookup", bson.D{
			{"from", "training_date"},
			{"localField", "training_date_id"},
			{"foreignField", "_id"},
			{"as", "training_date_info"},
		}}},
		// Unwind the training date information
		{{"$unwind", bson.D{{"path", "$training_date_info"}}}},
		// Filter for appointments where the training date is in the future
		// Lookup leave information
		{{"$lookup", bson.D{
			{"from", "leave"},
			{"localField", "_id"},
			{"foreignField", "booking_id"},
			{"as", "leave_info"},
		}}},
		// Unwind the leave information
		{{"$unwind", bson.D{
			{"path", "$leave_info"},
			{"preserveNullAndEmptyArrays", true},
		}}},
	}
	return pipeline
}
