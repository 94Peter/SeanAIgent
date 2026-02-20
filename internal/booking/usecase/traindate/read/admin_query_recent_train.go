package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

type ReqAdminQueryRecentTrain struct {
	StartTime time.Time
	EndTime   time.Time
}

type adminQueryRecentTrainUseCase struct {
	repo repository.TrainRepository
}

func NewAdminQueryRecentTrainUseCase(
	repo repository.TrainRepository,
) core.ReadUseCase[ReqAdminQueryRecentTrain, []*entity.TrainDate] {
	return &adminQueryRecentTrainUseCase{repo: repo}
}

func (uc *adminQueryRecentTrainUseCase) Name() string {
	return "AdminQueryRecentTrain"
}

func (uc *adminQueryRecentTrainUseCase) Execute(ctx context.Context, req ReqAdminQueryRecentTrain) (
	[]*entity.TrainDate, core.UseCaseError,
) {
	trainDates, err := uc.repo.FindTrainDates(
		ctx, repository.NewFilterTrainDataByTimeRange(req.StartTime, req.EndTime),
	)
	if err != nil {
		return nil, ErrAdminQueryRecentTrainFail.Wrap(err)
	}
	return trainDates, nil
}

var (
	ErrAdminQueryRecentTrainFail = core.NewDBError(
		"ADMIN_QUERY_RECENT_TRAIN", "QUERY_FAIL", "admin query recent train fail", core.ErrInternal)
)
