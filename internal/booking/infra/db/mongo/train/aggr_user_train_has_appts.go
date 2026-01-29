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

func getUserPipelineTrainDateHasApptState(userID string, q bson.M) mongo.Pipeline {
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

			// B. UserAppointments: 先 Filter 該用戶，再 Map 格式化欄位
			{"userAppointments", bson.D{
				{"$map", bson.D{
					{"input", bson.D{
						{"$filter", bson.D{
							{"input", "$appointments"},
							{"as", "f"},
							{"cond", bson.D{{"$eq", bson.A{"$$f.user_id", userID}}}},
						}},
					}},
					{"as", "u"},
					{"in", bson.D{
						{"id", bson.D{{"$toString", "$$u._id"}}}, // 轉換 _id 為 id
						{"userId", "$$u.user_id"},
						{"userName", "$$u.user_name"},
						{"childName", "$$u.child_name"},
						{"isOnLeave", "$$u.is_on_leave"},
						{"isCheckIn", "$$u.is_check_in"},
						{"createdAt", "$$u.created_at"},
						// 在此處添加其他 UserAppointment 需要的欄位
					}},
				}},
			}},

			// C. AllUsers (other_users): 提取所有未請假的使用者名稱字串陣列
			{"allUsers", bson.D{
				{"$map", bson.D{
					{"input", bson.D{
						{"$filter", bson.D{
							{"input", "$appointments"},
							{"as", "a"},
							{"cond", bson.D{{"$eq", bson.A{"$$a.is_on_leave", false}}}},
						}},
					}},
					{"as", "u"},
					{"in", "$$u.child_name"},
				}},
			}},
		}}},
	}
	return pipeline
}

func (*trainRepoImpl) UserQueryTrainDateHasApptState(
	ctx context.Context, userID string, filter repository.FilterTrainDate,
) ([]*entity.TrainDateHasUserApptState, repository.RepoError) {

	q, repoErr := getQueryByFilterTrainDate(filter)
	if repoErr != nil {
		return nil, repoErr
	}
	result, mgoErr := mgo.PipeFindByPipeline[*entity.TrainDateHasUserApptState](
		ctx, TrainDateCollectionName, getUserPipelineTrainDateHasApptState(userID, q), core.DefaultLimit,
	)
	if mgoErr != nil {
		return nil, newInternalError("user_query_train_date_has_appt_state", mgoErr)
	}
	return result, nil
}
