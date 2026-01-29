package appointment

import (
	"context"
	"errors"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func getLookupAndUnwindTrainDate(q bson.M) []bson.D {
	return []bson.D{
		{{"$match", q}},
		{{"$lookup", bson.D{
			{"from", "training_date"},
			{"localField", "training_date_id"},
			{"foreignField", "_id"},
			{"as", "training_date_info"},
		}}},
		{{"$unwind", bson.D{{"path", "$training_date_info"}}}},
	}
}
func getProject() bson.D {
	return bson.D{
		{"$project", bson.D{
			{"id", bson.D{{"$toString", "$_id"}}}, // 將 MongoDB 的 _id 轉為結構體的 id
			{"userId", "$user_id"},
			{"userName", "$user_name"},
			{"childName", "$child_name"},
			{"createdAt", "$created_at"},
			{"trainingDateId", bson.D{{"$toString", "$training_date_id"}}},
			{"isOnLeave", "$is_on_leave"},

			// 映射 TrainDateUI 結構
			{"trainDate", bson.D{
				{"id", bson.D{{"$toString", "$training_date_info._id"}}},
				{"date", "$training_date_info.date"},
				{"location", "$training_date_info.location"},
				{"capacity", "$training_date_info.capacity"},
				{"availableCapacity", "$training_date_info.available_capacity"},
				{"timezone", "$training_date_info.timezone"},
				{"startDate", "$training_date_info.start_date"},
				{"endDate", "$training_date_info.end_date"},
				// 根據你的 TrainDateUI 定義繼續增加欄位...
			}},

			// 映射 LeaveInfoUI (假設來自原預約紀錄中的 leave_details)
			{"leaveInfo", bson.D{
				{"reason", "$leave.reason"},
				{"status", "$leave.status"},
				{"updatedAt", "$leave.updated_at"},
			}},
		}},
	}
}

func getTrainFilterPipeline(trainFilter repository.FilterTrainDate) bson.D {
	switch f := trainFilter.(type) {
	case repository.FilterTrainDateByAfterTime:
		return bson.D{{"$match", bson.D{
			{"training_date_info.start_date",
				bson.D{{"$gte", f.Start}},
			},
		}}}
	default:
		return bson.D{}
	}
}

func getPipelineApptWithTrainDateWithTrainFilter(
	q bson.M, trainFilter repository.FilterTrainDate,
) mongo.Pipeline {
	sort := bson.D{{"$sort", bson.D{{"training_date_info.start_date", 1}}}}
	pipeline := append(
		getLookupAndUnwindTrainDate(q),
		getTrainFilterPipeline(trainFilter),
		sort,
		getProject(),
	)
	return pipeline
}

func getPipelineWithCursor(
	q bson.M,
	trainFilter repository.FilterTrainDate,
	cursor *repository.FilterApptsWithTrainDateByCursor,
) mongo.Pipeline {
	// 基礎的 Lookup 與 Unwind
	pipeline := getLookupAndUnwindTrainDate(q)

	// 游標過濾：必須放在 $lookup 之後，因為我們要過濾的是關聯後的 start_date
	if cursor != nil && !cursor.LastStartDate.IsZero() && cursor.LastID != "" {
		oid, err := bson.ObjectIDFromHex(cursor.LastID)
		if err == nil {
			pipeline = append(pipeline, bson.D{{"$match", bson.D{
				{"$or", bson.A{
					// 情況 A: start_date 大於上一次最後一個 (升序)
					bson.D{{"training_date_info.start_date", bson.D{{"$gt", cursor.LastStartDate}}}},
					// 情況 B: start_date 相同，但 ID 較大 (確保不遺漏)
					bson.D{
						{"training_date_info.start_date", cursor.LastStartDate},
						{"_id", bson.D{{"$gt", oid}}},
					},
				}},
			}}})
		}
	}

	// 加入原本的 TrainFilter
	pipeline = append(pipeline, getTrainFilterPipeline(trainFilter))

	// 排序與限制數量 (Limit 是分頁核心)
	pipeline = append(pipeline, bson.D{{"$sort", bson.D{
		{"training_date_info.start_date", 1},
		{"_id", 1},
	}}})

	if cursor != nil {
		pipeline = append(pipeline, bson.D{{"$limit", cursor.PageSize}})
	}

	// 最後做 Project 減少傳輸量
	pipeline = append(pipeline, getProject())

	return pipeline
}

func (*apptRepoImpl) PageFindApptsWithTrainDateByFilterAndTrainFilter(
	ctx context.Context,
	apptFilter repository.FilterAppointment,
	trainFilter repository.FilterTrainDate,
	cursorStr string,
) ([]*entity.AppointmentWithTrainDate, string, repository.RepoError) {
	const op = "find_appt_with_train_date_by_id"
	if cursorStr == "" {
		return nil, "", newInvalidCursorError(op, errors.New("cursor is empty"))
	}
	const emptyStr = ""
	apptQ, repoErr := getQueryByFilterAppt(apptFilter)
	if repoErr != nil {
		return nil, emptyStr, repoErr
	}
	var cursor repository.FilterApptsWithTrainDateByCursor

	err := cursor.Decode(cursorStr)
	if err != nil {
		return nil, emptyStr, newInternalError(op, err)
	}

	result, mgoErr := mgo.PipeFindByPipeline[*entity.AppointmentWithTrainDate](
		ctx,
		appointmentCollectionName,
		getPipelineWithCursor(apptQ, trainFilter, &cursor),
		uint16(cursor.PageSize),
	)
	if mgoErr != nil {
		return nil, emptyStr, newInternalError(op, mgoErr)
	}
	if len(result) == 0 {
		return result, emptyStr, nil
	}
	last := result[len(result)-1]
	cursor.LastStartDate = last.TrainDate.StartDate
	cursor.LastID = last.ID
	cursor.PageSize = cursor.PageSize
	return result, cursor.Encode(), nil
}
