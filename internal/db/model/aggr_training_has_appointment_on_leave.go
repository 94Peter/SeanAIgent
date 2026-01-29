package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewAggrTrainingHasAppointOnLeave() *AggrTrainingHasAppointOnLeave {
	return &AggrTrainingHasAppointOnLeave{
		Index: trainingDateCollection,
	}
}

type AggrTrainingHasAppointOnLeave struct {
	StartDate      time.Time `bson:"start_date"`
	EndDate        time.Time `bson:"end_date"`
	mgo.Index      `bson:"-"`
	UserID         string `bson:"user_id"`
	Date           string `bson:"date"`
	Location       string `bson:"location"`
	Timezone       string `bson:"timezone"`
	OnLeaveApplies []*struct {
		CreatedAt time.Time     `bson:"created_at"`
		UpdatedAt time.Time     `bson:"updated_at"`
		UserID    string        `bson:"user_id"`
		ChildName string        `bson:"childName"`
		Reason    string        `bson:"reason"`
		Status    LeaveStatus   `bson:"status"`
		ID        bson.ObjectID `bson:"_id,omitempty"`
	} `bson:"on_leave_applies"`
	Capacity          int           `bson:"capacity"`
	TotalAppointments int           `bson:"total_appointments"`
	ID                bson.ObjectID `bson:"_id"` // User who requested the leave
	// Status of the leave request (e.g., Pending, Approved, Rejected)
}

func (aggr *AggrTrainingHasAppointOnLeave) GetPipeline(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		// 1. $match: 過濾出特定日期的訓練場次
		{{"$match", q}},

		// 2. $lookup (非請假預約): 查詢並過濾出實際出席的預約
		{{"$lookup", bson.D{
			{"from", "appointment"},
			{"localField", "_id"},
			{"foreignField", "training_date_id"},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{{"is_on_leave", false}}}},
			}},
			{"as", "appointments"},
		}}},

		// 3. $addFields: 計算實際出席的預約總數
		{{"$addFields", bson.D{
			{"total_appointments", bson.D{{"$size", "$appointments"}}},
		}}},

		// 4. $lookup (請假預約): 查詢請假的預約
		{{"$lookup", bson.D{
			{"from", "appointment"},
			{"localField", "_id"},
			{"foreignField", "training_date_id"},
			{"pipeline", bson.A{
				bson.D{{"$match", bson.D{{"is_on_leave", true}}}}, // 過濾請假
				bson.D{{"$lookup", bson.D{ // 連接 leave 集合
					{"from", "leave"},
					{"localField", "_id"},
					{"foreignField", "booking_id"},
					{"as", "leave_info"},
				}}},
				bson.D{{"$unwind", "$leave_info"}}, // 展開 leave_info
			}},
			{"as", "on_leave_appointments"},
		}}},

		// 5. $project: 投影需要的欄位並重命名
		{{"$project", bson.D{
			{"user_id", "$user_id"},
			{"date", "$date"},
			{"location", "$location"},
			{"capacity", "$capacity"},
			{"start_date", "$start_date"},
			{"end_date", "$end_date"},
			{"timezone", "$timezone"},
			{"total_appointments", "$total_appointments"},
			// 從 on_leave_appointments 陣列中提取 leave_info 子文件
			{"on_leave_applies", "$on_leave_appointments.leave_info"},
		}}},
	}
	return pipeline
}
