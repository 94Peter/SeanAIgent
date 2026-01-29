package stats

import (
	"fmt"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/infra/db/mongo/core"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func NewStatsRepository() repository.StatsRepository {
	return &statsRepoImpl{}
}

type statsRepoImpl struct {
}

func getQueryByFilterUserApptStats(
	filter repository.FilterUserApptStats,
) (bson.M, repository.RepoError) {
	var q bson.M
	switch f := filter.(type) {
	case repository.FilterUserApptStatsByTrainTimeRange:
		q = bson.M{"start_date": bson.M{"$gte": f.TrainStart, "$lte": f.TrainEnd}}
	default:
		// 處理未定義的 Filter 型別，避免靜默失敗
		return nil, newInternalError(
			"getQueryByFilterUserApptStats",
			fmt.Errorf("Filter not implemented: %T", f),
		)
	}
	return q, nil
}

const repo = "stats"

func newInternalError(op string, err error) repository.RepoError {
	return core.NewInternalError(repo, op, err)
}

func newNotFoundError(op string, err error) repository.RepoError {
	return core.NewNotFoundError(repo, op, err)
}

func newConflictError(op string, err error) repository.RepoError {
	return core.NewConflictError(repo, op, err)
}

func newInvalidDocumentIDError(op string, err error) repository.RepoError {
	return core.NewInvalidDocumentIDError(repo, op, err)
}
