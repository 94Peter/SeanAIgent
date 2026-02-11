package migration

import (
	"context"
	"errors"
	"fmt"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"seanAIgent/internal/db/model"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type trainDataUseCase struct {
	repo repository.TrainRepository
}

type TrainDataMigrationUseCase core.WriteUseCase[core.Empty, core.Empty]

func NewTrainDataUseCase(repo repository.TrainRepository) TrainDataMigrationUseCase {
	return &trainDataUseCase{
		repo: repo,
	}
}

func (uc *trainDataUseCase) Name() string {
	return "TrainDataMigrationV1ToV2"
}

func (uc *trainDataUseCase) Execute(
	ctx context.Context, _ core.Empty,
) (core.Empty, core.UseCaseError) {
	// 1. 找出所有資料
	trainingDates := model.NewAggrTrainingHasAppointOnLeave()
	v1Datas, err := mgo.PipeFind(ctx, trainingDates, bson.M{"_migration.version": bson.M{"$exists": false}}, defaultLimit)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			fmt.Println("TrainDataMigrationV1ToV2 done, no v1 data")
			return empty, nil
		}
		return empty, core.NewDBError("migration_v1tov2", "training_date", "find v1 data fail", core.ErrInternal).Wrap(err)
	}
	if len(v1Datas) == 0 {
		fmt.Println("TrainDataMigrationV1ToV2 done, no v1 data")
		return empty, nil
	}
	// 2. 轉換資料
	v2Datas := make([]*entity.TrainDate, len(v1Datas))

	for i, v := range v1Datas {
		datePeriod, err := entity.NewTimeRange(v.StartDate, v.EndDate)
		if err != nil {
			return empty, core.NewDomainError("migration_v1tov2", "new_time_range_fail", "new time range fail", core.ErrInternal).Wrap(err)
		}

		v2Datas[i], err = entity.NewTrainDate(
			entity.WithTrainDateID(v.ID.Hex()),
			entity.WithTrainDateUserID(v.UserID),
			entity.WithTrainDateLocation(v.Location),
			entity.WithTrainDateMaxCapacity(v.Capacity),
			entity.WithTrainDateAvailableCapacity(v.Capacity-v.TotalAppointments),
			entity.WithTrainDatePeriod(datePeriod),
			entity.WithTrainDateTimezone(v.Timezone),
			entity.WithTrainDateStatus(entity.TrainDateStatusActive),
			entity.WithTrainDateCreatedAt(v.ID.Timestamp()),
			entity.WithTrainDateUpdatedAt(time.Now()),
		)
	}
	// 3. 存檔
	err = uc.repo.UpdateManyTrainDates(ctx, v2Datas)
	if err != nil {
		return empty, core.NewDBError("migration_v1tov2", "training_date", "save v2 data fail", core.ErrInternal).Wrap(err)
	}
	fmt.Println("TrainDataMigrationV1ToV2 done", len(v2Datas))
	return empty, nil
}
