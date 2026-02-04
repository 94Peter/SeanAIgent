package appointment

import (
	"errors"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/infra/db/mongo/core"
	"seanAIgent/internal/util"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func NewApptRepository() repository.AppointmentRepository {
	return &migrateV1ToV2apptRepoImpl{
		AppointmentRepository: &apptRepoImpl{},
	}
}

type apptRepoImpl struct {
}

func getQueryByFilterAppt(filter repository.FilterAppointment) (bson.M, repository.RepoError) {
	var q bson.M
	switch f := filter.(type) {
	case repository.FilterApptByTrainID:
		oid, err := bson.ObjectIDFromHex(f.TrainingID)
		if err != nil {
			return nil, newInvalidDocumentIDError("getQueryByFilterAppt", err)
		}
		q = bson.M{"training_date_id": oid}
	case repository.FilterApptByIDs:
		oids := make([]bson.ObjectID, 0, len(f.ApptIDs))
		for _, id := range f.ApptIDs {
			oid, err := bson.ObjectIDFromHex(id)
			if err != nil {
				return nil, newInvalidDocumentIDError("getQueryByFilterAppt", err)
			}
			oids = append(oids, oid)
		}
		q = bson.M{"_id": bson.M{"$in": oids}}
	case repository.FilterAppointmentByUserID:
		q = bson.M{"user_id": f.UserID}
	default:
		// 處理未定義的 Filter 型別，避免靜默失敗
		filterName := util.GetTypeName(filter)
		return nil, newInternalError(
			"getQueryByFilterAppt", errors.New("Filter not implemented: "+filterName))
	}
	return q, nil
}

const repoName = "appointment"

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

func newInvalidCursorError(op string, err error) repository.RepoError {
	return core.NewInvalidCursorError(repoName, op, err)
}
