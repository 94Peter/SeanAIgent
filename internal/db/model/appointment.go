package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const AppointmentCollectionName = "appointment"

var appointmentCollection = mgo.NewCollectDef(AppointmentCollectionName, func() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "training_date_id", Value: 1}, {Key: "child_name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "training_date_id", Value: 1}},
		},
	}
})

func init() {
	mgo.RegisterIndex(appointmentCollection)
}

func NewAppointment() *Appointment {
	return &Appointment{
		Index: appointmentCollection,
		ID:    bson.NewObjectID(),
	}
}

type Appointment struct {
	CreatedAt      time.Time `bson:"created_at"`
	mgo.Index      `bson:"-"`
	UserID         string        `bson:"user_id"`
	UserName       string        `bson:"user_name"`
	ChildName      string        `bson:"child_name,omitempty"`
	ID             bson.ObjectID `bson:"_id"`
	TrainingDateId bson.ObjectID `bson:"training_date_id"`
	IsCheckedIn    bool          `bson:"is_checked_in"`
	IsOnLeave      bool          `bson:"is_on_leave"`
}

func (s *Appointment) GetId() any {
	if s.ID.IsZero() {
		return nil
	}
	return s.ID
}

func (s *Appointment) SetId(id any) {
	oid, ok := id.(bson.ObjectID)
	if !ok {
		return
	}
	s.ID = oid
}

func (p *Appointment) Validate() error {
	return nil
}
