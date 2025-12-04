package mongodb

import (
	"context"

	"seanAIgent/internal/db"
	"seanAIgent/internal/db/model"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func NewAppointmentStore() db.AppointmentStore {
	return &appointmentStore{}
}

type appointmentStore struct{}

func (s *appointmentStore) Add(ctx context.Context, appointment *model.Appointment) error {
	_, err := mgo.Save(ctx, appointment)
	return err
}

func (s *appointmentStore) Delete(ctx context.Context, id string) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	appointment := model.NewAppointment()
	appointment.ID = oid
	_, err = mgo.DeleteById(ctx, appointment)
	return err
}

func (s *appointmentStore) Get(ctx context.Context, id string) (*model.Appointment, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	appointment := model.NewAppointment()
	appointment.ID = oid
	err = mgo.FindById(ctx, appointment)
	return appointment, err
}

func (s *appointmentStore) AddMany(ctx context.Context, appointments []*model.Appointment) error {
	opers, err := mgo.NewBulkOperation(model.AppointmentCollectionName)
	if err != nil {
		return err
	}
	for _, a := range appointments {
		opers.InsertOne(a)
	}
	_, err = opers.Execute(ctx)
	return err
}

func (s *appointmentStore) QueryUserAppointments(ctx context.Context, userID string) ([]*model.AggrUserAppointment, error) {
	appointment := model.NewAggUserAppointment()
	results, err := mgo.PipeFind(ctx, appointment, bson.M{"user_id": userID})

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (s *appointmentStore) UpdateCheckinStatus(ctx context.Context, slotID string, checkedInBookingIDs []string) error {
	oid, err := bson.ObjectIDFromHex(slotID)
	if err != nil {
		return err
	}

	// Convert checkedInBookingIDs to ObjectIDs
	checkedInOIDs := make([]bson.ObjectID, len(checkedInBookingIDs))
	for i, id := range checkedInBookingIDs {
		checkedInOIDs[i], err = bson.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
	}

	// 1. Set IsCheckedIn to true for specified bookings in this slot
	filterCheckedIn := bson.D{
		{Key: "training_date_id", Value: oid},
		{Key: "_id", Value: bson.M{"$in": checkedInOIDs}},
	}

	updateCheckedIn := bson.D{{Key: "$set", Value: bson.M{"is_checked_in": true}}}
	_, err = mgo.UpdateMany(ctx, model.NewAppointment(), filterCheckedIn, updateCheckedIn)
	if err != nil {
		return err
	}

	// 2. Set IsCheckedIn to false for other bookings in this slot
	filterNotCheckedIn := bson.D{
		{Key: "training_date_id", Value: oid},
		{Key: "_id", Value: bson.M{"$nin": checkedInOIDs}},
	}

	updateNotCheckedIn := bson.D{{Key: "$set", Value: bson.M{"is_checked_in": false}}}
	_, err = mgo.UpdateMany(ctx, model.NewAppointment(), filterNotCheckedIn, updateNotCheckedIn)
	return nil
}

func (s *appointmentStore) UpdateIsOnLeave(ctx context.Context, id bson.ObjectID, isOnLeave bool) error {
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.M{"is_on_leave": isOnLeave}}}
	_, err := mgo.UpdateOne(ctx, model.NewAppointment(), filter, update)
	return err
}

func (s *appointmentStore) CreateLeave(ctx context.Context, leave *model.Leave) (*model.Leave, error) {
	return mgo.Save(ctx, leave)
}

func (s *appointmentStore) GetLeave(ctx context.Context, id string) (*model.AggrLeaveHasAppointmentHasTraining, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	leave := model.NewAggrLeaveHasAppointmentHasTraining()
	err = mgo.PipeFindOne(ctx, leave, bson.M{"_id": oid})
	return leave, err
}

func (s *appointmentStore) QueryLeaveByDate(ctx context.Context, q bson.M) ([]*model.AggrTrainingHasAppointOnLeave, error) {
	aggr := model.NewAggrTrainingHasAppointOnLeave()
	results, err := mgo.PipeFind(ctx, aggr, q)
	return results, err
}

func (s *appointmentStore) CancelLeave(ctx context.Context, leaveID string) (*model.Leave, error) {
	oid, err := bson.ObjectIDFromHex(leaveID)
	if err != nil {
		return nil, err
	}
	leave := model.NewLeave()
	leave.ID = oid
	err = mgo.FindById(ctx, leave)
	if err != nil {
		return nil, err
	}
	_, err = mgo.DeleteById(ctx, leave)
	return leave, err
}

func (s *appointmentStore) AppointmentState(ctx context.Context, q bson.M) ([]*model.AggrAppointmentState, error) {
	appointment := model.NewAggrAppointmentState()
	results, err := mgo.PipeFind(ctx, appointment, q)
	return results, err
}
