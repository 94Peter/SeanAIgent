package train

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/db/model"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type migrateV1ToV2trainRepoImpl struct {
	repository.TrainRepository
}

func (m *migrateV1ToV2trainRepoImpl) IncreaseCapacity(
	ctx context.Context, trainingID string, count int,
) repository.RepoError {
	// 檢查資料是不是v2
	const op = "migrateV1ToV2trainRepoImpl_increase_capacity"
	isV2, oid := isV2(ctx, trainingID)
	if !isV2 {
		err := m.migrateToV2(ctx, oid)
		if err != nil {
			return newInternalError(op, err)
		}
	}
	return m.TrainRepository.IncreaseCapacity(ctx, trainingID, count)
}

func (m *migrateV1ToV2trainRepoImpl) DeductCapacity(
	ctx context.Context, trainingID string, count int,
) repository.RepoError {
	const op = "migrateV1ToV2trainRepoImpl_deduct_capacity"
	isV2, oid := isV2(ctx, trainingID)
	if !isV2 {
		err := m.migrateToV2(ctx, oid)
		if err != nil {
			return newInternalError(op, err)
		}
	}
	return m.TrainRepository.DeductCapacity(ctx, trainingID, count)
}

func (m *migrateV1ToV2trainRepoImpl) migrateToV2(
	ctx context.Context, id bson.ObjectID,
) error {
	trainingDate := model.NewAggrTrainingHasAppointOnLeave()
	err := mgo.PipeFindOne(ctx, trainingDate, bson.M{"_id": id})
	if err != nil {
		return err
	}
	datePeriod, err := entity.NewTimeRange(trainingDate.StartDate, trainingDate.EndDate)
	if err != nil {
		return err
	}
	v2Data, err := entity.NewTrainDate(
		entity.WithTrainDateID(trainingDate.ID.Hex()),
		entity.WithTrainDateUserID(trainingDate.UserID),
		entity.WithTrainDateLocation(trainingDate.Location),
		entity.WithTrainDateMaxCapacity(trainingDate.Capacity),
		entity.WithTrainDateAvailableCapacity(trainingDate.Capacity-trainingDate.TotalAppointments),
		entity.WithTrainDatePeriod(datePeriod),
		entity.WithTrainDateTimezone(trainingDate.Timezone),
		entity.WithTrainDateStatus(entity.TrainDateStatusActive),
		entity.WithTrainDateCreatedAt(trainingDate.ID.Timestamp()),
		entity.WithTrainDateUpdatedAt(time.Now()),
	)
	if err != nil {
		return err
	}
	err = m.TrainRepository.UpdateManyTrainDates(ctx, []*entity.TrainDate{v2Data})
	if err != nil {
		return err
	}
	return nil
}

func isV2(ctx context.Context, id string) (bool, bson.ObjectID) {
	doc, err := newTrainDate(withTrainDateID(id))
	if err != nil {
		return false, bson.NilObjectID
	}
	err = mgo.FindById(ctx, doc)
	if err != nil {
		return false, bson.NilObjectID
	}
	return doc.Migration.Version == 2, doc.ID
}
