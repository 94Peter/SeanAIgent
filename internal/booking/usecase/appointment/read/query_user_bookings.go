package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

type ReqQueryUserBookings struct {
	TrainDateAfter time.Time
	UserID         string
	Cursor         string
	PageSize       int
}

type queryUserBookingsUseCase struct {
	repo repository.AppointmentRepository
}

type RespQueryUserBookings struct {
	Appts  []*entity.AppointmentWithTrainDate
	Cursor string
}

type QueryUserBookingsUseCase core.ReadUseCase[ReqQueryUserBookings, *RespQueryUserBookings]

func NewQueryUserBookingsUseCase(
	repo repository.AppointmentRepository,
) QueryUserBookingsUseCase {
	return &queryUserBookingsUseCase{repo: repo}
}

func (uc *queryUserBookingsUseCase) Name() string {
	return "QueryUserBookings"
}

const defaultPageSize = 20

func (uc *queryUserBookingsUseCase) Execute(ctx context.Context, req ReqQueryUserBookings) (
	*RespQueryUserBookings, core.UseCaseError,
) {
	if req.Cursor == "" {
		if req.PageSize == 0 {
			req.PageSize = defaultPageSize
		}
		req.Cursor = repository.NewFilterApptsWithTrainDateByCursor(req.PageSize)
	}
	appts, cursor, err := uc.repo.PageFindApptsWithTrainDateByFilterAndTrainFilter(
		ctx,
		repository.NewFilterApptByUserID(req.UserID),
		repository.NewFilterTrainDateByAfterTime(req.TrainDateAfter),
		req.Cursor,
	)
	if err != nil {
		return nil, ErrQueryUserBookingsFindApptsWithTrainDateFail.Wrap(err)
	}
	return &RespQueryUserBookings{
		Appts:  appts,
		Cursor: cursor,
	}, nil
}

var (
	ErrQueryUserBookingsFindApptsWithTrainDateFail = core.NewDBError(
		"QUERY_USER_BOOKINGS", "FIND_APPTS_WITH_TRAIN_DATE_FAIL", "find appts with train date fail", core.ErrInternal)
)
