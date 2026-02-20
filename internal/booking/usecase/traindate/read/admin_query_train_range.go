package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

type ReqAdminQueryTrainRange struct {
	StartTime time.Time
	EndTime   time.Time
}

type adminQueryTrainRangeUseCase struct {
	repo repository.TrainRepository
}

func NewAdminQueryTrainRangeUseCase(
	repo repository.TrainRepository,
) core.ReadUseCase[ReqAdminQueryTrainRange, []*entity.TrainDateHasApptState] {
	return &adminQueryTrainRangeUseCase{repo: repo}
}

func (uc *adminQueryTrainRangeUseCase) Name() string {
	return "AdminQueryTrainRange"
}

func (uc *adminQueryTrainRangeUseCase) Execute(ctx context.Context, req ReqAdminQueryTrainRange) (
	[]*entity.TrainDateHasApptState, core.UseCaseError,
) {
	trainDates, err := uc.repo.QueryTrainDateHasAppointmentState(
		ctx, repository.NewFilterTrainDataByTimeRange(req.StartTime, req.EndTime),
	)
	if err != nil {
		return nil, ErrAdminQueryTrainRangeFail.Wrap(err)
	}
	return trainDates, nil
}

var (
	ErrAdminQueryTrainRangeFail = core.NewDBError(
		"ADMIN_QUERY_TRAIN_RANGE", "QUERY_FAIL", "admin query train range fail", core.ErrInternal)
)
