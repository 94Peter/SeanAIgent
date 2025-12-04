package db

import (
	"context"
	"time"

	"seanAIgent/internal/db/model"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CloseDbFunc func(ctx context.Context) error

type TrainingDateStore interface {
	Find(ctx context.Context, q bson.M) ([]*model.TrainingDate, error)
	Add(ctx context.Context, trainingDate *model.TrainingDate) error
	AddMany(ctx context.Context, trainingDates []*model.TrainingDate) error
	Delete(ctx context.Context, id string) error
	QueryTrainingDataHasAppointment(ctx context.Context, q bson.M) ([]*model.AggrTrainingDateHasAppoint, error)
	QueryTrainingDateAppointmentState(ctx context.Context, id string, q bson.M) ([]*model.AggrTrainingDateAppointState, error)
	QueryTrainingDateHasCheckinList(ctx context.Context, now time.Time) (*model.AggrTrainingdateHasCheckinItems, error)
}

type AppointmentStore interface {
	Add(ctx context.Context, appointment *model.Appointment) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.Appointment, error)
	QueryUserAppointments(ctx context.Context, userID string) ([]*model.AggrUserAppointment, error)
	AddMany(ctx context.Context, appointments []*model.Appointment) error
	UpdateCheckinStatus(ctx context.Context, slotID string, checkedInBookingIDs []string) error
	UpdateIsOnLeave(ctx context.Context, bookingID bson.ObjectID, isOnLeave bool) error // Updated method
	CreateLeave(ctx context.Context, leave *model.Leave) (*model.Leave, error)
	GetLeave(ctx context.Context, leaveID string) (*model.AggrLeaveHasAppointmentHasTraining, error)
	QueryLeaveByDate(ctx context.Context, q bson.M) ([]*model.AggrTrainingHasAppointOnLeave, error)
	CancelLeave(ctx context.Context, leaveID string) (*model.Leave, error)
	AppointmentState(ctx context.Context, q bson.M) ([]*model.AggrAppointmentState, error)
}
