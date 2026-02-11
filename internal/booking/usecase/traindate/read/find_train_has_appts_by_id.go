package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqFindTrainHasApptsById struct {
	TrainID string
}

type findTrainHasApptsByIdUseCase struct {
	repo repository.TrainRepository
}

func NewFindTrainHasApptsByIdUseCase(
	repo repository.TrainRepository,
) core.ReadUseCase[ReqFindTrainHasApptsById, *entity.TrainDateHasApptState] {
	return &findTrainHasApptsByIdUseCase{repo: repo}
}

func (uc *findTrainHasApptsByIdUseCase) Name() string {
	return "FindTrainHasApptsById"
}

func (uc *findTrainHasApptsByIdUseCase) Execute(ctx context.Context, req ReqFindTrainHasApptsById) (
	*entity.TrainDateHasApptState, core.UseCaseError,
) {
	trainDates, err := uc.repo.QueryTrainDateHasAppointmentState(
		ctx, repository.NewFilterTrainDateByIds(req.TrainID),
	)
	if err != nil {
		return nil, ErrFindTrainHasApptsByIdFail.Wrap(err)
	}
	if len(trainDates) == 0 {
		return nil, ErrFindTrainHasApptsByIdNotFound
	}
	return trainDates[0], nil
}

var (
	ErrFindTrainHasApptsByIdNotFound = core.NewDBError(
		"FindTrainHasApptsById", "NOT_FOUND", "train date not found", core.ErrNotFound)
	ErrFindTrainHasApptsByIdFail = core.NewDBError(
		"FindTrainHasApptsById", "QUERY_FAIL", "find train has appts by id fail", core.ErrInternal)
)
