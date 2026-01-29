package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

type ReqFindNearestTrainByTime struct {
	TimeAfter time.Time
}

type findNearestTrainByTimeUseCase struct {
	repo repository.TrainRepository
}

func NewFindNearestTrainByTimeUseCase(
	repo repository.TrainRepository,
) core.ReadUseCase[ReqFindNearestTrainByTime, *entity.TrainDateHasApptState] {
	return &findNearestTrainByTimeUseCase{repo: repo}
}

func (uc *findNearestTrainByTimeUseCase) Name() string {
	return "FindNearestTrainByTime"
}

func (uc *findNearestTrainByTimeUseCase) Execute(ctx context.Context, req ReqFindNearestTrainByTime) (
	*entity.TrainDateHasApptState, core.UseCaseError,
) {
	trainDates, err := uc.repo.QueryTrainDateHasAppointmentState(
		ctx, repository.NewFilterTrainDateByAfterTime(req.TimeAfter),
	)
	if err != nil {
		return nil, ErrFindNearestTrainByTimeFail.Wrap(err)
	}
	if len(trainDates) == 0 {
		return nil, ErrFindNearestTrainByTimeNotFound
	}
	return trainDates[0], nil
}

var (
	ErrFindNearestTrainByTimeNotFound = core.NewDBError(
		"FindNearestTrainByTime", "NOT_FOUND", "train date not found", core.ErrNotFound)
	ErrFindNearestTrainByTimeFail = core.NewDBError(
		"FindNearestTrainByTime", "QUERY_FAIL", "find nearest train by time fail", core.ErrInternal)
)
