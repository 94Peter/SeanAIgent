package model

import (
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const leaveCollectionName = "leave"

var leaveCollection = mgo.NewCollectDef(leaveCollectionName, func() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "booking_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
})

func init() {
	mgo.RegisterIndex(leaveCollection)
}

func NewLeave() *Leave {
	return &Leave{
		Index: leaveCollection,
		ID:    bson.NewObjectID(),
	}
}

// Leave represents a leave request for a booking.
type Leave struct {
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	mgo.Index `bson:"-"`
	UserID    string        `bson:"user_id"`
	ChildName string        `bson:"childName"`
	Reason    string        `bson:"reason"`
	Status    LeaveStatus   `bson:"status"`
	ID        bson.ObjectID `bson:"_id,omitempty"`
	BookingID bson.ObjectID `bson:"booking_id"`
}

func (s *Leave) GetId() any {
	if s.ID.IsZero() {
		return nil
	}
	return s.ID
}

func (s *Leave) SetId(id any) {
	oid, ok := id.(bson.ObjectID)
	if !ok {
		return
	}
	s.ID = oid
}

func (p *Leave) Validate() error {
	return nil
}

// LeaveStatus defines the possible states of a leave request.
type LeaveStatus string

const (
	LeaveStatusNone     LeaveStatus = "none" // No leave request or not applicable
	LeaveStatusPending  LeaveStatus = "pending"
	LeaveStatusApproved LeaveStatus = "approved"
	LeaveStatusRejected LeaveStatus = "rejected"
)
