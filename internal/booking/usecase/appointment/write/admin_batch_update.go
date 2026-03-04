package write

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type AttendanceUpdate struct {
	BookingID string `json:"bookingId"`
	Status    string `json:"status"` // "CheckedIn", "Leave", "Absent", "Pending"
}

type ReqAdminBatchUpdateAttendance struct {
	SessionID string             `json:"sessionId"`
	Updates   []AttendanceUpdate `json:"updates"`
}

type AdminBatchUpdateAttendanceUseCase core.WriteUseCase[ReqAdminBatchUpdateAttendance, int]

func NewAdminBatchUpdateAttendanceUseCase(repo adminCheckInUseCaseRepo, cw cacheWorker) AdminBatchUpdateAttendanceUseCase {
	return &adminBatchUpdateAttendanceUseCase{repo: repo, cw: cw}
}

type adminBatchUpdateAttendanceUseCase struct {
	repo adminCheckInUseCaseRepo
	cw   cacheWorker
}

func (uc *adminBatchUpdateAttendanceUseCase) Name() string {
	return "AdminBatchUpdateAttendance"
}

func (uc *adminBatchUpdateAttendanceUseCase) Execute(ctx context.Context, req ReqAdminBatchUpdateAttendance) (int, core.UseCaseError) {
	train, err := uc.repo.FindTrainDateByID(ctx, req.SessionID)
	if err != nil {
		return 0, ErrCheckInTrainNotFound.Wrap(err)
	}

	appts, err := uc.repo.FindApptsByFilter(ctx, repository.NewFilterApptByTrainID(req.SessionID))
	if err != nil {
		return 0, ErrCheckInFindApptFail.Wrap(err)
	}

	apptMap := make(map[string]*entity.Appointment)
	for _, a := range appts {
		apptMap[a.ID()] = a
	}

	var toUpdate []*entity.Appointment
	affectedUsers := make(map[string]struct{})

	startTime := train.Period().Start()

	for _, up := range req.Updates {
		appt, ok := apptMap[up.BookingID]
		if !ok {
			continue
		}

		var opErr error
		switch up.Status {
		case "CheckedIn":
			opErr = appt.AdminCheckIn(startTime)
		case "Leave":
			opErr = appt.AdminAppendLeave("教練批次標記請假", startTime)
		case "Absent":
			opErr = appt.AdminMarkAsAbsent(startTime)
		case "Pending":
			opErr = appt.AdminRestoreFromLeave(startTime)
		default:
			continue
		}

		if opErr != nil {
			return 0, core.NewUseCaseError("BATCH_UPDATE", "TIME_LOCK", "目前時間不允許修改此場次點名狀態", core.ErrInvalidInput).Wrap(opErr)
		}

		toUpdate = append(toUpdate, appt)
		affectedUsers[appt.User().UserID()] = struct{}{}
	}

	if len(toUpdate) > 0 {
		if err := uc.repo.UpdateManyAppts(ctx, toUpdate); err != nil {
			return 0, ErrCheckInUpdateApptFail.Wrap(err)
		}
		// 清理快取
		for uid := range affectedUsers {
			uc.cw.CleanSync(ctx, uid, req.SessionID, startTime)
		}
	}

	return len(toUpdate), nil
}
