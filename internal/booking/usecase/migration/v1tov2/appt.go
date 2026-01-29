package migration

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"seanAIgent/internal/db/model"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const defaultLimit = 100

type apptUseCase struct {
	repo repository.AppointmentRepository
}

func NewApptUseCase(repo repository.AppointmentRepository) core.WriteUseCase[core.Empty, core.Empty] {
	return &apptUseCase{
		repo: repo,
	}
}

func (uc *apptUseCase) Name() string {
	return "ApptMigrationV1ToV2"
}

var empty = core.Empty{}

func (uc *apptUseCase) Execute(
	ctx context.Context, _ core.Empty,
) (core.Empty, core.UseCaseError) {
	// 1. 找出所有資料
	appts := model.NewAggUserAppointment()
	v1Datas, err := mgo.PipeFind(ctx, appts, bson.M{}, defaultLimit)
	if err != nil {
		return empty, core.NewDBError("migration_v1tov2", "appointment", "find v1 data fail", core.ErrInternal).Wrap(err)
	}
	// 2. 轉換資料
	v2Datas := make([]*entity.Appointment, len(v1Datas))
	for i, v := range v1Datas {
		leaveInfo := entity.EmptyLeaveInfo
		if v.IsOnLeave {
			leaveInfo = entity.NewLeaveInfo(
				v.LeaveInfo.Reason,
				entity.LeaveStatusApproved,
				v.LeaveInfo.CreatedAt,
			)
		}
		user, err := entity.NewUser(v.UserID, v.UserName)
		if err != nil {
			return empty, core.NewDomainError("migration_v1tov2", "new_user_fail", "new user fail", core.ErrInternal).Wrap(err)
		}
		var verifiedAt *time.Time
		if v.IsCheckedIn {
			t := time.Now()
			verifiedAt = &t
		}
		v2Datas[i], err = entity.NewAppointment(
			entity.WithApptID(v.ID.Hex()),
			entity.WithUser(user),
			entity.WithChildName(v.ChildName),
			entity.WithTrainingID(v.TrainingDateId.Hex()),
			entity.WithStatus(entity.StatusConfirmed),
			entity.WithCreatedAt(v.CreatedAt),
			entity.WithUpdatedAt(v.CreatedAt),
			entity.WithVerifiedAt(verifiedAt),
			entity.WithLeaveInfo(leaveInfo),
		)
	}
	// 3. 存檔
	err = uc.repo.UpdateManyAppts(ctx, v2Datas)
	if err != nil {
		return empty, core.NewDBError("migration_v1tov2", "appointment", "save v2 data fail", core.ErrInternal).Wrap(err)
	}
	return empty, nil
}
