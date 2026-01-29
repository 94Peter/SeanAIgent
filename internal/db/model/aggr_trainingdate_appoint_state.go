package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrTrainingDateAppointState(targetUserID string) *AggrTrainingDateAppointState {
	return &AggrTrainingDateAppointState{
		Index:        trainingDateCollection,
		targetUserID: targetUserID,
	}
}

type UserAppointment struct {
	CreatedAt time.Time     `bson:"created_at"`
	ChildName string        `bson:"child_name,omitempty"`
	ID        bson.ObjectID `bson:"_id"`
	IsOnLeave bool          `bson:"is_on_leave"`
}

type AggrTrainingDateAppointState struct {
	StartDate            time.Time `bson:"start_date"`
	EndDate              time.Time `bson:"end_date"`
	mgo.Index            `bson:"-"`
	Date                 string            `bson:"date"`
	Location             string            `bson:"location"`
	Timezone             string            `bson:"timezone"`
	targetUserID         string            `bson:"-"`
	UserAppointments     []UserAppointment `bson:"user_appointments"`
	AppointmentUserNames []string          `bson:"appointment_user_names"`
	Capacity             int               `bson:"capacity"`
	TotalAppointments    int               `bson:"total_appointments"`
	ID                   bson.ObjectID     `bson:"_id"`
}

func (aggr *AggrTrainingDateAppointState) GetPipeline(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		{{"$match", q}},
		// 左連接 appointments
		{{"$lookup", bson.D{
			{"from", "appointment"},
			{"localField", "_id"},
			{"foreignField", "training_date_id"},
			{"as", "appointments"},
		}}},

		// 過濾掉請假的預約
		{{"$addFields", bson.D{
			{"regular_appointments", bson.D{
				{"$filter", bson.D{
					{"input", "$appointments"},
					{"as", "a"},
					{"cond", bson.D{
						{"$eq", bson.A{bson.D{{"$ifNull", bson.A{"$$a.is_on_leave", false}}}, false}},
					}},
				}},
			}},
		}}},

		// 計算該時段目前預約人數
		{{"$addFields", bson.D{
			{"total_appointments", bson.D{{"$size", "$regular_appointments"}}},
		}}},

		// 投影需要的欄位，包含處理 user_appointments
		{{"$project", bson.D{
			{"_id", 1},
			{"date", 1},
			{"location", 1},
			{"capacity", 1},
			{"start_date", 1},
			{"timezone", 1},
			{"end_date", 1},
			{"total_appointments", 1},
			{"appointment_user_names", bson.D{
				{"$map", bson.D{
					{"input", "$regular_appointments"},
					{"as", "a"},
					{"in", bson.D{{"$ifNull", bson.A{"$$a.child_name", "$$a.user_name"}}}},
				}},
			}},
			{"user_appointments", bson.D{
				{"$filter", bson.D{
					{"input", "$appointments"},
					{"as", "a"},
					{"cond", bson.D{{"$eq", bson.A{"$$a.user_id", aggr.targetUserID}}}},
				}},
			}},
		}}},
		// 只保留 user_appointments 中的必要欄位
		{{"$project", bson.D{
			{"_id", 1},
			{"date", 1},
			{"location", 1},
			{"capacity", 1},
			{"start_date", 1},
			{"end_date", 1},
			{"timezone", 1},
			{"total_appointments", 1},
			{"appointment_user_names", 1},
			{"user_appointments._id", 1},
			{"user_appointments.child_name", 1},
			{"user_appointments.created_at", 1},
			{"user_appointments.is_on_leave", 1},
		}}},

		// order by start_date
		{{"$sort", bson.D{{"start_date", 1}}}},
	}
	return pipeline
}
