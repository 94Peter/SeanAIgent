package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrAppointmentState() *AggrAppointmentState {
	return &AggrAppointmentState{
		Index: trainingDateCollection,
	}
}

type AggrAppointmentState struct {
	mgo.Index        `bson:"-"`
	UserId           string       `bson:"user_id"`
	UserName         string       `bson:"user_name"`
	CheckedInCount   int          `bson:"checked_in_count"`
	OnLeaveCount     int          `bson:"on_leave_count"`
	TotalAppointment int          `bson:"total_appointment"`
	ChildState       []childState `bson:"child_state"`
}

type childState struct {
	ChildName      string            `bson:"child_name"`
	CheckedInCount int               `bson:"checked_in_count"`
	OnLeaveCount   int               `bson:"on_leave_count"`
	Appointments   []AppointmentInfo `bson:"appointments"`
}

type AppointmentInfo struct {
	AppointmentDate time.Time `bson:"appointment_date"`
	Location        string    `bson:"location"`
	Capacity        int       `bson:"capacity"`
	StartDate       time.Time `bson:"start_date"`
	EndDate         time.Time `bson:"end_date"`
	Timezone        string    `bson:"timezone"`
	IsCheckedIn     bool      `bson:"is_checked_in"`
	IsOnLeave       bool      `bson:"is_on_leave"`
}

func (aggr *AggrAppointmentState) GetPipeline(q bson.M) mongo.Pipeline {
	pipe := mongo.Pipeline{
		{{"$match", q}},
		{{"$lookup", bson.D{
			{"from", "appointment"},
			{"localField", "_id"},
			{"foreignField", "training_date_id"},
			{"as", "appointments"},
		}}},
		{{"$unwind", "$appointments"}},
		{{"$sort", bson.D{{"start_date", 1}}}},
		{{"$group", bson.D{
			{"_id", bson.M{
				"child_name": "$appointments.child_name",
				"user_name":  "$appointments.user_name",
				"user_id":    "$appointments.user_id",
			}},
			{"total_appointment", bson.D{{"$sum", 1}}},
			{"checked_in_count", bson.D{
				{"$sum", bson.D{
					{"$cond", bson.D{
						{"if", "$appointments.is_checked_in"},
						{"then", 1},
						{"else", 0},
					}},
				}},
			}},
			{"on_leave_count", bson.D{
				{"$sum", bson.D{
					{"$cond", bson.D{
						{"if", "$appointments.is_on_leave"},
						{"then", 1},
						{"else", 0},
					}},
				}},
			}},
			{"appointments", bson.D{{"$push", bson.D{
				{"appointment_date", "$start_date"},
				{"location", "$location"},
				{"capacity", "$capacity"},
				{"start_date", "$start_date"},
				{"end_date", "$end_date"},
				{"timezone", "$timezone"},
				{"is_checked_in", "$appointments.is_checked_in"},
				{"is_on_leave", "$appointments.is_on_leave"},
			}}}},
		}}},
		{{"$sort", bson.D{{"_id.user_name", 1}, {"_id.child_name", 1}}}},
		{{"$group", bson.D{
			{"_id", "$_id.user_id"},
			{"user_name", bson.D{{"$first", "$_id.user_name"}}},
			{"total_appointment", bson.D{{"$sum", "$total_appointment"}}},
			{"checked_in_count", bson.D{{"$sum", "$checked_in_count"}}},
			{"on_leave_count", bson.D{{"$sum", "$on_leave_count"}}},
			{"child_state", bson.D{{"$push", bson.D{
				{"child_name", "$_id.child_name"},
				{"checked_in_count", "$checked_in_count"},
				{"on_leave_count", "$on_leave_count"},
				{"appointments", "$appointments"},
			}}}},
		}}},
		{{"$project", bson.D{
			{"_id", 0},
			{"user_id", "$_id"},
			{"user_name", "$user_name"},
			{"checked_in_count", "$checked_in_count"},
			{"on_leave_count", "$on_leave_count"},
			{"total_appointment", "$total_appointment"},
			{"child_state", "$child_state"},
		}}},
	}
	return pipe
}
