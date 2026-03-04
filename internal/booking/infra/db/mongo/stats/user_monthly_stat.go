package stats

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"

	"github.com/94peter/vulpes/db/mgo"
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
