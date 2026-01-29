package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

type ReqQueryFutureTrain struct {
	TimeAfter time.Time
}

type queryFutureTrainUseCase struct {
	repo repository.TrainRepository
}

func NewQueryFutureTrainUseCase(
	repo repository.TrainRepository,
) core.ReadUseCase[ReqQueryFutureTrain, []*entity.TrainDateHasApptState] {
	return &queryFutureTrainUseCase{repo: repo}
}

func (uc *queryFutureTrainUseCase) Name() string {
	return "QueryFutureTrain"
}

func (uc *queryFutureTrainUseCase) Execute(ctx context.Context, req ReqQueryFutureTrain) (
	[]*entity.TrainDateHasApptState, core.UseCaseError,
) {
	trainDates, err := uc.repo.QueryTrainDateHasAppointmentState(
		ctx, repository.NewFilterTrainDateByAfterTime(req.TimeAfter),
	)
	if err != nil {
		return nil, ErrQueryFutureTrainFail.Wrap(err)
	}
	return trainDates, nil
}

var (
	ErrQueryFutureTrainFail = core.NewDBError(
		"QUERY_FUTURE_TRAIN", "QUERY_FAIL", "query future train fail", core.ErrInternal)
)
