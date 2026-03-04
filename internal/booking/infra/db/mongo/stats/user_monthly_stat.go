package stats

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/94peter/vulpes/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	userMonthlyStatsCol = "user_monthly_stats"
)

func init() {
	mgo.RegisterIndex(mgo.NewCollectDef(userMonthlyStatsCol, func() []mongo.IndexModel {
		return []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "year", Value: 1},
					{Key: "month", Value: 1},
					{Key: "user_id", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{
					{Key: "user_name", Value: "text"},
				},
			},
		}
	}))
}

func (s *statsRepoImpl) UpsertUserMonthlyStats(ctx context.Context, stat *entity.UserMonthlyStat) repository.RepoError {
	const op = "upsert_user_monthly_stats"
	filter := bson.M{
		"year":    stat.Year,
		"month":   stat.Month,
		"user_id": stat.UserID,
	}
	update := bson.M{
		"$set": stat,
	}
	
	_, err := mgo.GetDatabase().Collection(userMonthlyStatsCol).UpdateOne(
		ctx, filter, update, options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return newInternalError(op, err)
	}
	return nil
}

func (s *statsRepoImpl) UpsertManyUserMonthlyStats(ctx context.Context, stats []*entity.UserMonthlyStat) repository.RepoError {
	if len(stats) == 0 {
		return nil
	}
	const op = "upsert_many_user_monthly_stats"
	bulk, err := mgo.NewBulkOperation(userMonthlyStatsCol)
	if err != nil {
		return newInternalError(op, err)
	}

	validCount := 0
	for _, stat := range stats {
		// 1. 執行實體級別的驗證
		if err := stat.Validate(); err != nil {
			// 紀錄錯誤並跳過該筆，不影響其他資料
			log.Errorf("%s: skipping invalid stat for user %s (%d/%d): %v", op, stat.UserID, stat.Year, stat.Month, err)
			continue
		}

		filter := bson.M{
			"year":    stat.Year,
			"month":   stat.Month,
			"user_id": stat.UserID,
		}
		update := bson.M{
			"$set": stat,
		}
		bulk.UpsertOne(filter, update)
		validCount++
	}

	if validCount == 0 {
		return nil
	}

	// 2. 執行批次寫入
	_, err = bulk.Execute(ctx)
	if err != nil {
		// 這裡如果是 MongoDB 網路錯誤，會整批失敗，這是合理的 (需要重試)
		// 如果是資料約束錯誤 (如重複 Key)，因為我們用的是 Upsert，風險較低
		return newInternalError(op, err)
	}
	return nil
}

func (s *statsRepoImpl) FindMonthlyStats(
	ctx context.Context, year, month int, skip, limit int64, search string,
) ([]*entity.UserMonthlyStat, int64, repository.RepoError) {
	const op = "find_monthly_stats"
	filter := bson.M{
		"year":  year,
		"month": month,
	}
	if search != "" {
		filter["user_name"] = bson.M{"$regex": search, "$options": "i"}
	}

	coll := mgo.GetDatabase().Collection(userMonthlyStatsCol)
	
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, newInternalError(op, err)
	}

	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "user_name", Value: 1}})

	cursor, err := coll.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, newInternalError(op, err)
	}
	defer cursor.Close(ctx)

	var results []*entity.UserMonthlyStat
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, newInternalError(op, err)
	}

	return results, total, nil
}

func (s *statsRepoImpl) GetHistoricalAnalytics(ctx context.Context, monthsLimit int) ([]*entity.MonthlyBusinessStat, repository.RepoError) {
	const op = "get_historical_analytics"

	pipe := mongo.Pipeline{
		{{"$group", bson.D{
			{"_id", bson.M{"year": "$year", "month": "$month"}},
			{"total_bookings", bson.D{{"$sum", "$total_bookings"}}},
			{"attended_count", bson.D{{"$sum", "$attended_count"}}},
			{"leave_count", bson.D{{"$sum", "$leave_count"}}},
			{"active_users", bson.D{{"$sum", 1}}},
		}}},
		{{"$sort", bson.D{
			{"_id.year", -1},
			{"_id.month", -1},
		}}},
		{{"$limit", monthsLimit}},
		{{"$project", bson.D{
			{"_id", 0},
			{"year", "$_id.year"},
			{"month", "$_id.month"},
			{"total_bookings", "$total_bookings"},
			{"attended_count", "$attended_count"},
			{"leave_count", "$leave_count"},
			{"active_users", "$active_users"},
		}}},
	}

	cursor, err := mgo.GetDatabase().Collection(userMonthlyStatsCol).Aggregate(ctx, pipe)
	if err != nil {
		return nil, newInternalError(op, err)
	}
	defer cursor.Close(ctx)

	var results []*entity.MonthlyBusinessStat
	if err := cursor.All(ctx, &results); err != nil {
		return nil, newInternalError(op, err)
	}

	return results, nil
}

func (s *statsRepoImpl) FindUserMonthlyStats(ctx context.Context, userID string) ([]*entity.UserMonthlyStat, repository.RepoError) {
	const op = "find_user_monthly_stats"
	filter := bson.M{"user_id": userID}
	
	opts := options.Find().SetSort(bson.D{{Key: "year", Value: -1}, {Key: "month", Value: -1}})
	cursor, err := mgo.GetDatabase().Collection(userMonthlyStatsCol).Find(ctx, filter, opts)
	if err != nil {
		return nil, newInternalError(op, err)
	}
	defer cursor.Close(ctx)

	var results []*entity.UserMonthlyStat
	if err := cursor.All(ctx, &results); err != nil {
		return nil, newInternalError(op, err)
	}
	return results, nil
}
