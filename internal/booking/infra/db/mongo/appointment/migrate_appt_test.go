package appointment

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"testing"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Test_MigrateV1ToV2apptRepoImpl(t *testing.T) {
	findLeaveByBookingId := func(ctx context.Context, bookID string) (*leave, error) {
		oid, err := bson.ObjectIDFromHex(bookID)
		if err != nil {
			return nil, err
		}
		leav := newLeave()
		err = mgo.FindOne(ctx, leav, bson.M{"booking_id": oid})
		if err != nil {
			return nil, err
		}
		if leav.BookingID.IsZero() {
			return nil, nil
		}
		return leav, err
	}
	t.Run("CreateLeave", func(t *testing.T) {
		defer cleanDb()
		migrateRepo := NewApptRepository()
		ctx := t.Context()
		apptID := bson.NewObjectID().Hex()
		trainID := bson.NewObjectID().Hex()
		userID := "user-migrate"
		user, err := entity.NewUser(userID, "User Migrate")
		if err != nil {
			t.Fatal(err)
		}
		now := time.Now().Truncate(time.Millisecond)
		appt, err := entity.NewAppointment(
			entity.WithApptID(apptID),
			entity.WithUser(user),
			entity.WithChildName("ChildMigrate"),
			entity.WithTrainingID(trainID),
			entity.WithStatus(entity.StatusConfirmed),
			entity.WithCreatedAt(now),
			entity.WithUpdatedAt(now),
		)
		if err != nil {
			t.Fatal(err)
		}
		err = migrateRepo.SaveAppointment(ctx, appt)
		if err != nil {
			t.Fatal(err)
		}

		// check leave not exist
		leav, err := findLeaveByBookingId(ctx, apptID)
		require.ErrorIs(t, err, mongo.ErrNoDocuments)
		require.Nil(t, leav)

		err = appt.AppendLeaveRecord("reason", now.Add(3*time.Hour))
		require.NoError(t, err)

		require.False(t, appt.LeaveInfo().IsEmpty())

		err = migrateRepo.UpdateAppt(ctx, appt)
		if err != nil {
			t.Fatal(err)
		}
		leav, err = findLeaveByBookingId(ctx, apptID)
		require.NoError(t, err)
		require.NotNil(t, leav)
		require.Equal(t, apptID, leav.BookingID.Hex())
		require.Equal(t, userID, leav.UserID)
		require.Equal(t, "ChildMigrate", leav.ChildName)
		require.Equal(t, "reason", leav.Reason)
		require.Equal(t, string(entity.LeaveStatusApproved), string(leav.Status))

		// delete leave
		err = appt.CancelLeave(userID)
		require.NoError(t, err)
		require.True(t, appt.LeaveInfo().IsEmpty())
		err = migrateRepo.UpdateAppt(ctx, appt)
		if err != nil {
			t.Fatal(err)
		}
		leav, err = findLeaveByBookingId(ctx, apptID)
		require.ErrorIs(t, err, mongo.ErrNoDocuments)
		require.Nil(t, leav)

		err = appt.MarkAsAttended(now)
		require.NoError(t, err)
		err = migrateRepo.UpdateAppt(ctx, appt)
		if err != nil {
			t.Fatal(err)
		}
		leav, err = findLeaveByBookingId(ctx, apptID)
		require.ErrorIs(t, err, mongo.ErrNoDocuments)
		require.Nil(t, leav)

	})
}
