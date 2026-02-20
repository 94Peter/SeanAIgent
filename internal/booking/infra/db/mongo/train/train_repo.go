package train

import (
	"fmt"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/infra/db/mongo/core"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func NewTrainRepository() repository.TrainRepository {
	return &trainRepoImpl{}
}

type trainRepoImpl struct{}

func getQueryByFilterTrainDate(filter repository.FilterTrainDate) (bson.M, repository.RepoError) {
	var q bson.M
	switch f := filter.(type) {
	case repository.FilterTrainDateByAfterTime:
		q = bson.M{"start_date": bson.M{"$gte": f.Start}}
	case repository.FilterTrainingDateByTimeRange:
		q = bson.M{"start_date": bson.M{"$gte": f.StartTime}, "end_date": bson.M{"$lte": f.EndTime}}
	case repository.FilterTrainingDateByIDs:
		oids := make([]bson.ObjectID, 0, len(f.TrainingDateIDs))
		for _, id := range f.TrainingDateIDs {
			oid, err := bson.ObjectIDFromHex(id)
			if err != nil {
				return nil,
					newInvalidDocumentIDError("getQueryByFilterTrainDate", err)
			}
			oids = append(oids, oid)
		}
		q = bson.M{"_id": bson.M{"$in": oids}}
	case repository.FilterTrainDateByEndTime:
		q = bson.M{"end_date": bson.M{"$gt": f.Start}}
	default:
		// 處理未定義的 Filter 型別，避免靜默失敗
		return nil, newInternalError("getQueryByFilterTrainDate",
			fmt.Errorf("Filter not implemented: %T", f))
	}
	return q, nil
}

const repoName = "train"

func newInternalError(op string, err error) repository.RepoError {
	return core.NewInternalError(repoName, op, err)
}

func newNotFoundError(op string, err error) repository.RepoError {
	return core.NewNotFoundError(repoName, op, err)
}

func newConflictError(op string, err error) repository.RepoError {
	return core.NewConflictError(repoName, op, err)
}

func newInvalidDocumentIDError(op string, err error) repository.RepoError {
	return core.NewInvalidDocumentIDError(repoName, op, err)
}
