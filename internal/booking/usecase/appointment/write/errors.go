package write

import "seanAIgent/internal/booking/usecase/core"

var (
	ErrCheckInNoApptCheckedIn = core.NewUseCaseError(
		"CHECK_IN", "NO_APPT_CHECKED_IN", "no appointment checked in", core.ErrInvalidInput)
	ErrCheckInApptNotFound = core.NewDBError(
		"CHECK_IN", "APPOINTMENT_NOT_FOUND", "appointment not found", core.ErrNotFound)
	ErrCheckInFindApptFail = core.NewDBError(
		"CHECK_IN", "FIND_APPOINTMENT_FAIL", "find appointment fail", core.ErrInternal)
	ErrCheckInTrainDateNotFound = core.NewDBError(
		"CHECK_IN", "TRAIN_DATE_NOT_FOUND", "train date not found", core.ErrNotFound)
	ErrCheckInFindTrainDateFail = core.NewDBError(
		"CHECK_IN", "FIND_TRAIN_DATE_FAIL", "find train date fail", core.ErrInternal)
	ErrCheckInUpdateApptFail = core.NewDBError(
		"CHECK_IN", "UPDATE_APPOINTMENT_FAIL", "update appointment fail", core.ErrInternal)
	ErrCheckInMarkAsAttendedFail = core.NewDomainError(
		"CHECK_IN", "MARK_AS_ATTENDED_FAIL", "mark as attended fail", core.ErrInternal)
	ErrCheckInTrainNotFound = core.NewUseCaseError(
		"CHECK_IN", "TRAIN_NOT_FOUND", "找不到此訓練場次", core.ErrNotFound)
	ErrCheckInDomainError = core.NewUseCaseError(
		"CHECK_IN", "DOMAIN_ERROR", "簽到失敗：時限或狀態不符", core.ErrInternal)

	// New consolidated errors
	ErrCheckInNotOpen = core.NewUseCaseError(
		"CHECK_IN", "CHECKIN_NOT_OPEN", "課程尚未開始，無法簽到", core.ErrForbidden)
	ErrCheckInTooLate = core.NewUseCaseError(
		"CHECK_IN", "CHECKIN_TOO_LATE", "已超過課程補簽時限 (7天)", core.ErrConflict)
	ErrWalkInNotOpen = core.NewUseCaseError(
		"CHECK_IN", "WALKIN_NOT_OPEN", "課程尚未開始，無法現場加人", core.ErrForbidden)
	ErrWalkInTooLate = core.NewUseCaseError(
		"CHECK_IN", "WALKIN_TOO_LATE", "已超過課程補登時限 (7天)", core.ErrConflict)
)
