package write

import (
	"context"
	"errors"
	"fmt"
	"seanAIgent/internal/util/validator"
	"time"

	"seanAIgent/internal/booking/domain"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase/core"
	"seanAIgent/internal/event"
)

type ReqAdminCreateWalkIn struct {
	TrainDateID string
	ChildName   string
	UserID      string // Can be empty for new experience students
	ParentName  string // For display/logging
	ContactInfo string // For new experience students
}

type AdminCreateWalkInUseCase core.WriteUseCase[ReqAdminCreateWalkIn, *entity.Appointment]

func NewAdminCreateWalkInUseCase(repo adminCheckInUseCaseRepo, bus event.Bus) AdminCreateWalkInUseCase {
	return &adminCreateWalkInUseCase{repo: repo, bus: bus}
}

type adminCreateWalkInUseCase struct {
	repo adminCheckInUseCaseRepo
	bus  event.Bus
}

func (uc *adminCreateWalkInUseCase) Name() string {
	return "AdminCreateWalkIn"
}

func (uc *adminCreateWalkInUseCase) Execute(ctx context.Context, req ReqAdminCreateWalkIn) (*entity.Appointment, core.UseCaseError) {
	// 1. Find Training Date
	train, err := uc.repo.FindTrainDateByID(ctx, req.TrainDateID)
	if err != nil {
		return nil, ErrCreateApptTrainDateNotFound.Wrap(err)
	}

	// 2. Prepare User Entity & Validate Contact Info for Guests
	userID := req.UserID
	userName := req.ParentName
	contactInfo := req.ContactInfo
	isGuest := false
	
	if userID == "" {
		isGuest = true
		// 驗證電話格式
		cleanedPhone, ok := validator.ValidatePhone(contactInfo)
		if !ok {
			return nil, ErrCreateApptNewDomainEntityFail.Wrap(fmt.Errorf("聯絡資訊必須是有效的 10 位手機號碼 (例如 0912345678)"))
		}
		contactInfo = cleanedPhone
		userID = "GUEST_" + contactInfo
		if userName == "" {
			userName = "體驗家長"
		}
	}

	user, _ := entity.NewUser(userID, userName)

	// 3. Create Appointment (Unrestricted)
	apptID := uc.repo.GenerateID()
	appt, domainErr := entity.NewAppointment(
		entity.WithCreateAppt(apptID, req.TrainDateID, user, req.ChildName),
		entity.WithWalkIn(true),
		entity.WithGuest(isGuest, contactInfo),
	)
	if domainErr != nil {
		return nil, ErrCreateApptNewDomainEntityFail.Wrap(domainErr)
	}

	oldStatus := appt.Status().String()
	// 4. Admin auto-checkin (with consolidated constraints)
	if err := appt.AdminCheckIn(train.Period().Start()); err != nil {
		if errors.Is(err, entity.ErrAppointmentCheckInNotOpen) {
			return nil, ErrWalkInNotOpen.Wrap(err)
		}
		if errors.Is(err, entity.ErrAppointmentCheckInTooLate) {
			return nil, ErrWalkInTooLate.Wrap(err)
		}
		return nil, ErrCheckInDomainError.Wrap(err)
	}

	// 5. Save & Deduct Capacity (Admin version allows overbooking)
	if err := uc.repo.AdminDeductCapacity(ctx, req.TrainDateID, 1); err != nil {
		return nil, ErrCreateApptSaveApptFail.Wrap(err)
	}

	if err := uc.repo.SaveAppointment(ctx, appt); err != nil {
		return nil, ErrCreateApptSaveApptFail.Wrap(err)
	}

	// 手動清理快取
	_ = uc.repo.CleanTrainCache(ctx, userID)
	_ = uc.repo.CleanStatsCache(ctx, userID, train.Period().Start().Year(), int(train.Period().Start().Month()))

	// 發送領域事件
	evt := event.NewTypedEvent(uc.repo.GenerateID(), domain.TopicAppointmentStatusChanged, domain.AppointmentStatusChanged{
		BookingID:  appt.ID(),
		UserID:     userID,
		TrainingID: req.TrainDateID,
		OldStatus:  oldStatus,
		NewStatus:  appt.Status().String(),
		OccurredAt: time.Now(),
	})
	uc.bus.Publish(ctx, evt)

	return appt, nil
}
