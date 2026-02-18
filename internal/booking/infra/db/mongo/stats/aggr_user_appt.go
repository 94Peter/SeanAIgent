package stats

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/infra/db/mongo/core"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getPipeline(q bson.M, userID string) mongo.Pipeline {
	matchAppt := bson.M{}
	if userID != "" {
		matchAppt["appointments.user_id"] = userID
	}

	pipe := mongo.Pipeline{
		{{"$match", q}},
		{{"$lookup", bson.D{
			{"from", "appointment"},
			{"localField", "_id"},
			{"foreignField", "training_date_id"},
			{"as", "appointments"},
		}}},
		{{"$unwind", "$appointments"}},
	}
	if userID != "" {
		pipe = append(pipe, bson.D{{"$match", matchAppt}})
	}
	pipe = append(pipe, mongo.Pipeline{
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
				{"appointmentDate", "$start_date"},
				{"location", "$location"},
				{"capacity", "$capacity"},
				{"startDate", "$start_date"},
				{"endDate", "$end_date"},
				{"timezone", "$timezone"},
				{"isCheckedIn", "$appointments.is_checked_in"},
				{"isOnLeave", "$appointments.is_on_leave"},
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
				{"childName", "$_id.child_name"},
				{"checkedInCount", "$checked_in_count"},
				{"onLeaveCount", "$on_leave_count"},
				{"appointments", "$appointments"},
			}}}},
		}}},
		{{"$project", bson.D{
			{"_id", 0},
			{"userId", "$_id"},
			{"userName", "$user_name"},
			{"checkedInCount", "$checked_in_count"},
			{"onLeaveCount", "$on_leave_count"},
			{"totalAppointment", "$total_appointment"},
			{"childState", "$child_state"},
		}}},
	}...)
	return pipe
}

func (*statsRepoImpl) GetAllUserApptStats(
	ctx context.Context, filter repository.FilterUserApptStats,
) ([]*entity.UserApptStats, repository.RepoError) {
	q, repoErr := getQueryByFilterUserApptStats(filter)
	if repoErr != nil {
		return nil, repoErr
	}
	const op = "get_all_user_appt_stats"
	result, err := mgo.PipeFindByPipeline[*entity.UserApptStats](
		ctx, "training_date", getPipeline(q, ""), core.DefaultLimit,
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, newNotFoundError(op, err)
		}
		return nil, newInternalError(op, err)
	}
	return result, nil
}

func (*statsRepoImpl) GetUserApptStats(
	ctx context.Context, userID string, filter repository.FilterUserApptStats,
) (*entity.UserApptStats, repository.RepoError) {
	q, repoErr := getQueryByFilterUserApptStats(filter)
	if repoErr != nil {
		return nil, repoErr
	}
	const op = "get_user_appt_stats"
	result, err := mgo.PipeFindByPipeline[*entity.UserApptStats](
		ctx, "training_date", getPipeline(q, userID), 1,
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, newNotFoundError(op, err)
		}
		return nil, newInternalError(op, err)
	}
	if len(result) == 0 {
		return nil, newNotFoundError(op, mongo.ErrNoDocuments)
	}
	return result[0], nil
}
