package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

type ReqUserQueryFutureTrain struct {
	TimeAfter time.Time
	UserID    string
}

type userQueryFutureTrainUseCase struct {
	repo repository.TrainRepository
}

func NewUserQueryFutureTrainUseCase(
	repo repository.TrainRepository,
) core.ReadUseCase[ReqUserQueryFutureTrain, []*entity.TrainDateHasUserApptState] {
	return &userQueryFutureTrainUseCase{repo: repo}
}

func (uc *userQueryFutureTrainUseCase) Name() string {
	return "UserQueryFutureTrain"
}

func (uc *userQueryFutureTrainUseCase) Execute(ctx context.Context, req ReqUserQueryFutureTrain) (
	[]*entity.TrainDateHasUserApptState, core.UseCaseError,
) {
	trainDates, err := uc.repo.UserQueryTrainDateHasApptState(
		ctx, req.UserID, repository.NewFilterTrainDateByAfterTime(req.TimeAfter),
	)
	if err != nil {
		return nil, ErrUserQueryFutureTrainFail.Wrap(err)
	}
	return trainDates, nil
}

var (
	ErrUserQueryFutureTrainFail = core.NewDBError(
		"USER_QUERY_FUTURE_TRAIN", "QUERY_FAIL", "user query future train fail", core.ErrInternal)
)
