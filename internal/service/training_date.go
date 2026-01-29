package service

import (
	"context"
	"errors"
	"fmt"
	"seanAIgent/internal/db"
	"seanAIgent/internal/db/model"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func newTrainingDateService(
	trainingDateStore db.TrainingDateStore,
	appointmentStore db.AppointmentStore) TrainingDateService {
	return &trainingDateService{
		trainingDateStore: trainingDateStore,
		appointmentStore:  appointmentStore,
	}
}

type TrainingDateService interface {
	// Sample 會從 topics 中隨機選取 number 個問題
	FutureTrainingDate(ctx context.Context) ([]*model.AggrTrainingDateHasAppoint, error)
	AddTrainingDate(ctx context.Context, trainingDate *model.TrainingDate) ([]*model.AggrTrainingDateHasAppoint, error)
	AddTrainingDates(ctx context.Context, trainingDates []*model.TrainingDate) ([]*model.AggrTrainingDateHasAppoint, error)
	GetTrainingDateDetail(ctx context.Context, trainingDateID string) (*model.AggrTrainingDateHasAppoint, error)
	QueryDateTimeRangeTrainingDate(ctx context.Context, start, end time.Time) ([]*model.AggrTrainingDateHasAppoint, error)
	DeleteTrainingDate(ctx context.Context, trainingDateID string) error
	QueryFutureTrainingDateAppointmentState(ctx context.Context, userID string) ([]*model.AggrTrainingDateAppointState, error)
	AppointmentTrainingDates(ctx context.Context, slotID, userID, userName string, childNames []string) error
	// 取消預約，回傳已取消的 appointment
	AppointmentCancel(ctx context.Context, bookingId, userID string) (*model.Appointment, error)
	QueryUserBookings(ctx context.Context, userID string) ([]*model.AggrUserAppointment, error)
	QueryCheckinList(ctx context.Context, now time.Time) (*model.AggrTrainingdateHasCheckinItems, error)
	UpdateCheckinStatus(ctx context.Context, slotID string, checkedInBookingIDs []string) error
	CreateLeave(ctx context.Context, bookingId string, userID string, leaveReason string) (*model.Leave, error)
	GetLeave(ctx context.Context, leaveID string) (*model.AggrLeaveHasAppointmentHasTraining, error)
	QueryLeaveByTrainingDate(ctx context.Context, date time.Time) ([]*model.AggrTrainingHasAppointOnLeave, error)
	CancelLeave(ctx context.Context, leaveID string) error
	TrainingDateRangeFormat(start, end time.Time, timezone string) string
}

type trainingDateService struct {
	trainingDateStore db.TrainingDateStore
	appointmentStore  db.AppointmentStore
}

func (s *trainingDateService) QueryLeaveByTrainingDate(ctx context.Context, date time.Time) ([]*model.AggrTrainingHasAppointOnLeave, error) {
	queryStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	queryEnd := queryStart.Add(24 * time.Hour)
	data, err := s.appointmentStore.QueryLeaveByDate(ctx, bson.M{"start_date": bson.M{"$gte": queryStart, "$lt": queryEnd}})
	if err != nil {
		return nil, fmt.Errorf("query leave failed: %w", err)
	}
	for i := range data {
		data[i].StartDate = model.ToTime(data[i].StartDate, "Asia/Taipei")
		data[i].EndDate = model.ToTime(data[i].EndDate, "Asia/Taipei")
	}
	return data, nil
}

func (s *trainingDateService) CancelLeave(ctx context.Context, leaveID string) error {
	leave, err := s.appointmentStore.CancelLeave(ctx, leaveID)
	if err != nil {
		return fmt.Errorf("cancel leave failed: %w", err)
	}
	err = s.appointmentStore.UpdateIsOnLeave(ctx, leave.BookingID, false)
	if err != nil {
		return fmt.Errorf("update is_on_leave failed: %w", err)
	}
	return nil
}

func (s *trainingDateService) FutureTrainingDate(ctx context.Context) ([]*model.AggrTrainingDateHasAppoint, error) {
	return s.trainingDateStore.QueryTrainingDataHasAppointment(ctx, bson.M{"start_date": bson.M{"$gte": time.Now()}})
}

func (s *trainingDateService) QueryDateTimeRangeTrainingDate(ctx context.Context, start, end time.Time) ([]*model.AggrTrainingDateHasAppoint, error) {
	return s.trainingDateStore.QueryTrainingDataHasAppointment(ctx, bson.M{"start_date": bson.M{"$gte": start, "$lte": end}})
}

func (s *trainingDateService) GetTrainingDateDetail(ctx context.Context, trainingDateID string) (*model.AggrTrainingDateHasAppoint, error) {
	oid, err := bson.ObjectIDFromHex(trainingDateID)
	if err != nil {
		return nil, err
	}
	result, err := s.trainingDateStore.QueryTrainingDataHasAppointment(ctx, bson.M{"_id": oid})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, errors.New("not found")
	}
	return result[0], nil
}

func (s *trainingDateService) AddTrainingDate(ctx context.Context, trainingDate *model.TrainingDate) ([]*model.AggrTrainingDateHasAppoint, error) {
	err := s.trainingDateStore.Add(ctx, trainingDate)
	if err != nil {
		return nil, err
	}
	return s.FutureTrainingDate(ctx)
}

func (s *trainingDateService) AddTrainingDates(ctx context.Context, trainingDates []*model.TrainingDate) ([]*model.AggrTrainingDateHasAppoint, error) {
	err := s.trainingDateStore.AddMany(ctx, trainingDates)
	if err != nil {
		return nil, err
	}
	return s.FutureTrainingDate(ctx)
}

func (s *trainingDateService) DeleteTrainingDate(ctx context.Context, trainingDateID string) error {
	return s.trainingDateStore.Delete(ctx, trainingDateID)
}

func (s *trainingDateService) QueryFutureTrainingDateAppointmentState(ctx context.Context, userID string) ([]*model.AggrTrainingDateAppointState, error) {
	return s.trainingDateStore.QueryTrainingDateAppointmentState(ctx, userID, bson.M{"start_date": bson.M{"$gte": time.Now()}})
}

func (s *trainingDateService) AppointmentTrainingDates(ctx context.Context, slotID string, userID string, userName string, childNames []string) error {
	oid, err := bson.ObjectIDFromHex(slotID)
	if err != nil {
		return err
	}

	var appointments []*model.Appointment
	for _, childName := range childNames {
		appointment := model.NewAppointment()
		appointment.CreatedAt = time.Now()
		appointment.TrainingDateId = oid
		appointment.UserID = userID
		appointment.UserName = userName
		appointment.ChildName = childName
		appointments = append(appointments, appointment)
	}

	return s.appointmentStore.AddMany(ctx, appointments)
}

func (s *trainingDateService) AppointmentCancel(ctx context.Context, bookingId string, userID string) (*model.Appointment, error) {
	appoint, err := s.appointmentStore.Get(ctx, bookingId)
	if err != nil {
		return nil, err
	}
	if appoint.UserID != userID {
		return nil, fmt.Errorf("not your appointment")
	}
	return appoint, s.appointmentStore.Delete(ctx, bookingId)
}

func (s *trainingDateService) QueryUserBookings(ctx context.Context, userID string) ([]*model.AggrUserAppointment, error) {
	return s.appointmentStore.QueryUserAppointments(ctx, userID)
}

func (s *trainingDateService) QueryCheckinList(ctx context.Context, now time.Time) (*model.AggrTrainingdateHasCheckinItems, error) {
	return s.trainingDateStore.QueryTrainingDateHasCheckinList(ctx, now)
}

func (s *trainingDateService) UpdateCheckinStatus(ctx context.Context, slotID string, checkedInBookingIDs []string) error {
	return s.appointmentStore.UpdateCheckinStatus(ctx, slotID, checkedInBookingIDs)
}

func (s *trainingDateService) CreateLeave(ctx context.Context, bookingId string, userID string, leaveReason string) (*model.Leave, error) {
	appointment, err := s.appointmentStore.Get(ctx, bookingId)
	if err != nil {
		return nil, err
	}
	if appointment.UserID != userID {
		return nil, fmt.Errorf("not your appointment")
	}
	leave := model.NewLeave()
	leave.UserID = appointment.UserID
	leave.ChildName = appointment.ChildName
	leave.BookingID = appointment.ID
	leave.Reason = leaveReason
	leave.Status = model.LeaveStatusNone
	leave.CreatedAt = time.Now()
	leave.UpdatedAt = leave.CreatedAt
	createdLeave, err := s.appointmentStore.CreateLeave(ctx, leave)
	if err != nil {
		return nil, err
	}
	return createdLeave, s.appointmentStore.UpdateIsOnLeave(ctx, appointment.ID, true)
}

func (s *trainingDateService) GetLeave(ctx context.Context, leaveID string) (*model.AggrLeaveHasAppointmentHasTraining, error) {
	return s.appointmentStore.GetLeave(ctx, leaveID)
}

func (s *trainingDateService) TrainingDateRangeFormat(start, end time.Time, timezone string) string {
	start = model.ToTime(start, timezone)
	end = model.ToTime(end, timezone)
	return fmt.Sprintf("%s %s - %s", start.Format("01/02"), start.Format("15:04"), end.Format("15:04"))
}
