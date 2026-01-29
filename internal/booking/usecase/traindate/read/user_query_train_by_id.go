package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqUserQueryTrainByID struct {
	UserID      string
	TrainDateID string
}

type userQueryTrainByIDUseCase struct {
	repo repository.TrainRepository
}

func NewUserQueryTrainByIDUseCase(
	repo repository.TrainRepository,
) core.ReadUseCase[ReqUserQueryTrainByID, *entity.TrainDateHasUserApptState] {
	return &userQueryTrainByIDUseCase{repo: repo}
}

func (uc *userQueryTrainByIDUseCase) Name() string {
	return "UserQueryTrainByID"
}

func (uc *userQueryTrainByIDUseCase) Execute(ctx context.Context, req ReqUserQueryTrainByID) (
	*entity.TrainDateHasUserApptState, core.UseCaseError,
) {
	trainDate, err := uc.repo.UserQueryTrainDateHasApptState(
		ctx, req.UserID, repository.NewFilterTrainDateByIds(req.TrainDateID),
	)
	if err != nil {
		return nil, ErrUserQueryTrainByIDFail.Wrap(err)
	}
	if len(trainDate) == 0 {
		return nil, ErrUserQueryTrainByIDNotFound
	}
	return trainDate[0], nil
}

var (
	ErrUserQueryTrainByIDFail = core.NewDBError(
		"USER_QUERY_TRAIN_BY_ID", "QUERY_FAIL", "user query train by id fail", core.ErrInternal)
	ErrUserQueryTrainByIDNotFound = core.NewDBError(
		"USER_QUERY_TRAIN_BY_ID", "NOT_FOUND", "train date not found", core.ErrNotFound)
)
