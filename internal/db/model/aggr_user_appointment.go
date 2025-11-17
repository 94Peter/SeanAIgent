package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggUserAppointment() *AggrUserAppointment {
	return &AggrUserAppointment{
		Index: appointmentCollection,
	}
}

// AggrUserAppointment represents the result of an aggregation that joins
// an appointment with its corresponding training date information.
// It's used to display a user's list of their own bookings.
type AggrUserAppointment struct {
	mgo.Index      `bson:"-"`
	ID             bson.ObjectID `bson:"_id"`
	UserID         string        `bson:"user_id"`
	UserName       string        `bson:"user_name"`
	ChildName      string        `bson:"child_name,omitempty"`
	CreatedAt      time.Time     `bson:"created_at"`
	TrainingDateId bson.ObjectID `bson:"training_date_id"`
	IsOnLeave      bool          `bson:"is_on_leave"`
	TrainingDate   *struct {     // Use a slice for the lookup result
		Date      string    `bson:"date"`
		Location  string    `bson:"location"`
		StartDate time.Time `bson:"start_date"`
		EndDate   time.Time `bson:"end_date"`
		Timezone  string    `bson:"timezone"`
	} `bson:"training_date_info"`
	LeaveInfo *struct {
		Reason    string    `bson:"reason"`
		Status    string    `bson:"status"`
		CreatedAt time.Time `bson:"created_at"`
	} `bson:"leave_info"`
}

func (aggr *AggrUserAppointment) GetPipeline(q bson.M) mongo.Pipeline {
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
		{{"$match", bson.D{
			{"training_date_info.start_date",
				bson.D{{"$gte", time.Now()}},
			},
		}}},
		// Sort by the start date
		{{"$sort", bson.D{{"training_date_info.start_date", 1}}}},
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
