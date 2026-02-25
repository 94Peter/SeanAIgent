package write

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase/core"
)

// ReqCheckIn is Deprecated: Use ReqAdminCheckIn instead.
type ReqCheckIn struct {
	TrainDateID         string
	CheckedInBookingIDs []string
}

// CheckInUseCase is Deprecated: Use AdminCheckInUseCase instead.
// This is kept for V1 compatibility but delegates logic to AdminCheckIn logic.
type CheckInUseCase core.WriteUseCase[ReqCheckIn, []*entity.Appointment]

func NewCheckInUseCase(adminUC AdminCheckInUseCase) CheckInUseCase {
	return &checkInUseCaseLegacy{
		adminUC: adminUC,
	}
}

type checkInUseCaseLegacy struct {
	adminUC AdminCheckInUseCase
}

func (uc *checkInUseCaseLegacy) Name() string {
	return "CheckIn (Deprecated)"
}

func (uc *checkInUseCaseLegacy) Execute(
	ctx context.Context, req ReqCheckIn,
) ([]*entity.Appointment, core.UseCaseError) {
	// Delegation to the new consolidated logic
	return uc.adminUC.Execute(ctx, ReqAdminCheckIn{
		TrainDateID:         req.TrainDateID,
		CheckedInBookingIDs: req.CheckedInBookingIDs,
	})
}
