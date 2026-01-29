package train

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/infra/db/mongo/core"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getPipelineTrainDateHasApptState(q bson.M) mongo.Pipeline {
	pipeline := mongo.Pipeline{
		{{"$match", q}},
		{{"$lookup", bson.D{
			{"from", "appointment"},
			{"localField", "_id"},
			{"foreignField", "training_date_id"},
			{"as", "appointments"},
		}}},
		{{"$project", bson.D{
			{"id", bson.D{{"$toString", "$_id"}}},
			{"date", "$date"},
			{"location", "$location"},
			{"capacity", "$capacity"},
			{"availableCapacity", "$available_capacity"},
			{"startDate", "$start_date"},
			{"endDate", "$end_date"},
			{"timezone", "$timezone"},
			{"userAppointments", bson.D{
				{"$map", bson.D{
					{"input", "$appointments"},
					{"as", "appt"},
					{"in", bson.D{
						// 這裡根據 UserAppointment 結構體的定義來對應
						{"id", bson.D{{"$toString", "$$appt._id"}}},
						{"userId", "$$appt.user_id"},
						{"userName", "$$appt.user_name"},
						{"childName", "$$appt.child_name"},
						{"isOnLeave", "$$appt.is_on_leave"},
						{"isCheckedIn", "$$appt.is_checked_in"},
						{"createdAt", "$$appt.created_at"},
						// 如果還有其他不一致的欄位，在這裡進行轉換
					}},
				}},
			}},
		}}},
	}
	return pipeline
}

func (*trainRepoImpl) QueryTrainDateHasAppointmentState(
	ctx context.Context, filter repository.FilterTrainDate,
) ([]*entity.TrainDateHasApptState, repository.RepoError) {
	q, repoErr := getQueryByFilterTrainDate(filter)
	if repoErr != nil {
		return nil, repoErr
	}
	result, mgoErr := mgo.PipeFindByPipeline[*entity.TrainDateHasApptState](
		ctx, TrainDateCollectionName, getPipelineTrainDateHasApptState(q), core.DefaultLimit,
	)
	if mgoErr != nil {
		return nil, newInternalError("query_train_date_has_appointment_state", mgoErr)
	}
	return result, nil
}
