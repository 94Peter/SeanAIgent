package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrTrainingDateHasAppoint() *AggrTrainingDateHasAppoint {
	return &AggrTrainingDateHasAppoint{
		Index: trainingDateCollection,
	}
}

type AggrTrainingDateHasAppoint struct {
	StartDate            time.Time `bson:"start_date"`
	EndDate              time.Time `bson:"end_date"`
	mgo.Index            `bson:"-"`
	Date                 string        `bson:"date"`
	Location             string        `bson:"location"`
	Timezone             string        `bson:"timezone"`
	targetUserID         string        `bson:"-"`
	AppointmentUserNames []string      `bson:"appointment_user_names"`
	OnLeaveUserNames     []string      `bson:"on_leave_user_names"`
	Capacity             int           `bson:"capacity"`
	TotalAppointments    int           `bson:"total_appointments"`
	ID                   bson.ObjectID `bson:"_id"`
}

func (aggr *AggrTrainingDateHasAppoint) GetPipeline(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		{{"$match", q}},
		// 左連接 appointments
		{{"$lookup", bson.D{
			{"from", "appointment"},
			{"localField", "_id"},
			{"foreignField", "training_date_id"},
			{"as", "appointments"},
		}}},

		// 分離請假與非請假的預約
		{{"$addFields", bson.D{
			{"on_leave_appointments", bson.D{
				{"$filter", bson.D{
					{"input", "$appointments"},
					{"as", "a"},
					{"cond", bson.D{{"$eq", bson.A{"$$a.is_on_leave", true}}}},
				}},
			}},
			{"regular_appointments", bson.D{
				{"$filter", bson.D{
					{"input", "$appointments"},
					{"as", "a"},
					{"cond", bson.D{
						{"$not", bson.D{
							{"$eq", bson.A{"$$a.is_on_leave", true}},
						}},
					}},
				}},
			}},
		}}},

		// 計算該時段目前預約人數
		{{"$addFields", bson.D{
			{"total_appointments", bson.D{{"$size", "$regular_appointments"}}},
		}}},

		// 取出所有預約人的名字清單
		{{"$addFields", bson.D{
			{"appointment_user_names", bson.D{
				{"$map", bson.D{
					{"input", "$regular_appointments"},
					{"as", "a"},
					{"in", "$$a.child_name"},
				}},
			}},
			{"on_leave_user_names", bson.D{
				{"$map", bson.D{
					{"input", "$on_leave_appointments"},
					{"as", "a"},
					{"in", "$$a.child_name"},
				}},
			}},
		}}},

		// 投影需要的欄位
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
			{"on_leave_user_names", 1},
		}}},
	}
	return pipeline
}
