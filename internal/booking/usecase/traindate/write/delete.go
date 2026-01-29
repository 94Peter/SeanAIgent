package write

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqDeleteTrainDate struct {
	TrainDateID string
	UserID      string
}

type deleteTrainDateUseCaseRepo interface {
	repository.TrainRepository
}

func NewDeleteTrainDateUseCase(repo deleteTrainDateUseCaseRepo) core.WriteUseCase[
	ReqDeleteTrainDate, *entity.TrainDate] {
	return &deleteTrainDateUseCase{repo: repo}
}

type deleteTrainDateUseCase struct {
	repo deleteTrainDateUseCaseRepo
}

func (uc *deleteTrainDateUseCase) Name() string {
	return "DeleteTrainDate"
}

func (uc *deleteTrainDateUseCase) Execute(
	ctx context.Context, req ReqDeleteTrainDate,
) (result *entity.TrainDate, returnErr core.UseCaseError) {
	var err error

	trainDate, err := uc.repo.FindTrainDateByID(ctx, req.TrainDateID)
	if err != nil {
		returnErr = ErrDeleteTrainDateFindTrainDateFail.Wrap(err)
		return
	}
	err = trainDate.Delete()
	if err != nil {
		returnErr = ErrDeleteTrainDateDeleteFail.Wrap(err)
		return
	}
	err = uc.repo.DeleteTrainingDate(ctx, trainDate)
	if err != nil {
		returnErr = ErrDeleteTrainDateDeleteFail.Wrap(err)
		return
	}
	result = trainDate
	return
}

var (
	ErrDeleteTrainDateFindTrainDateFail = core.NewDBError(
		"DELETE_TRAIN_DATE", "FIND_TRAIN_DATE_FAIL", "find train date fail", core.ErrInternal)
	ErrDeleteTrainDateDeleteFail = core.NewDomainError(
		"DELETE_TRAIN_DATE", "DELETE_TRAIN_DATE_FAIL", "delete train date fail", core.ErrInternal)
)
