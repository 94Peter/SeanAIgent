package train

import (
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"testing"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Helper struct to insert appointments for aggregation tests
type testAppointment struct {
	CreatedAt      time.Time `bson:"created_at"`
	mgo.Index      `bson:"-"`
	UserID         string        `bson:"user_id"`
	UserName       string        `bson:"user_name"`
	ChildName      string        `bson:"child_name,omitempty"`
	Status         string        `bson:"status"`
	ID             bson.ObjectID `bson:"_id"`
	TrainingDateId bson.ObjectID `bson:"training_date_id"`
}

func (*testAppointment) Validate() error {
	return nil
}

func (a *testAppointment) GetId() any {
	if a.ID.IsZero() {
		return nil
	}
	return a.ID
}

func (a *testAppointment) SetId(id any) {
	oid, ok := id.(bson.ObjectID)
	if !ok {
		return
	}
	a.ID = oid
}

var testAppointmentCollection = mgo.NewCollectDef("appointment", func() []mongo.IndexModel {
	return []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "training_date_id", Value: 1}},
			Options: options.Index().SetUnique(false), // Relax unique for test simplicity if needed
		},
	}
})

func TestScheduleRepo_Integration(t *testing.T) {
	drapAllDb, closeFunc, err := mgo.InitTestContainer(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer closeFunc()

	repo := NewTrainRepository()

	t.Run("CRUD", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()
		timeRange, _ := entity.NewTimeRange(
			time.Date(2025, 2, 3, 12, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 3, 14, 0, 0, 0, time.UTC),
		)

		trainID := bson.NewObjectID().Hex()
		trainDate, err := entity.NewTrainDate(
			entity.WithBasicTrainDate(
				trainID, "coach1", "Gym A", 10, timeRange),
		)
		require.NoError(t, err)

		// Save
		err = repo.SaveTrainDate(ctx, trainDate)
		require.NoError(t, err)

		// Find by ID
		found, err := repo.FindTrainDateByID(ctx, trainID)
		require.NoError(t, err)
		assert.Equal(t, trainID, found.ID())
		assert.Equal(t, "coach1", found.UserID())

		// Update via Save (or specialized update if needed, but Save usually overwrites or UpdateMany)
		// Testing UpdateManyTrainDates
		err = repo.UpdateManyTrainDates(ctx, []*entity.TrainDate{found})
		require.NoError(t, err)

		// Delete
		err = repo.DeleteTrainingDate(ctx, found)
		require.NoError(t, err)
	})

	t.Run("QueryTrainDateHasAppointmentState", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()

		// 1. Prepare Training Date
		trainID := bson.NewObjectID()
		tr, _ := entity.NewTimeRange(time.Now(), time.Now().Add(time.Hour))
		trainDate, _ := entity.NewTrainDate(
			entity.WithBasicTrainDate(
				trainID.Hex(), "coach_aggr", "Gym D", 10, tr),
		)
		err = repo.SaveTrainDate(ctx, trainDate)
		require.NoError(t, err)

		// 2. Insert Appointments manually
		user1ID := "user1"
		user2ID := "user2"

		appt1 := &testAppointment{
			Index:          testAppointmentCollection,
			ID:             bson.NewObjectID(),
			UserID:         user1ID,
			UserName:       "User One",
			ChildName:      "Child One",
			TrainingDateId: trainID,
			Status:         "CONFIRMED",
			CreatedAt:      time.Now(),
		}

		appt2 := &testAppointment{
			Index:          testAppointmentCollection,
			ID:             bson.NewObjectID(),
			UserID:         user2ID,
			UserName:       "User Two",
			ChildName:      "Child Two",
			TrainingDateId: trainID,
			Status:         "ATTENDED",
			CreatedAt:      time.Now(),
		}

		// Insert appts using mgo.Save (as it accepts interface{})
		_, err = mgo.Save(ctx, appt1)
		require.NoError(t, err)
		_, err = mgo.Save(ctx, appt2)
		require.NoError(t, err)

		testApp := &testAppointment{
			Index: testAppointmentCollection,
			ID:    appt2.ID,
		}
		err = mgo.FindById(ctx, testApp)
		require.NoError(t, err)
		assert.Equal(t, appt2.UserName, testApp.UserName)

		// 3. Test QueryTrainDateHasAppointmentState (Admin view)
		filter := repository.NewFilterTrainDateByIds(trainID.Hex())
		results, err := repo.QueryTrainDateHasAppointmentState(ctx, filter)
		require.NoError(t, err)
		require.Len(t, results, 1)

		state := results[0]
		assert.Equal(t, trainID.Hex(), state.ID)
		assert.Len(t, state.UserAppointments, 2) // Should have 2 appointments
		assert.Equal(t, appt1.ID.Hex(), state.UserAppointments[0].ID)
		assert.True(t, state.UserAppointments[1].IsCheckedIn)

		// 4. Test UserQueryTrainDateHasApptState (User view)
		// Query for user1
		userResults, err := repo.UserQueryTrainDateHasApptState(ctx, user1ID, filter)
		require.NoError(t, err)
		require.Len(t, userResults, 1)

		userState := userResults[0]
		assert.Equal(t, trainID.Hex(), userState.ID)
		// UserAppointments should only contain user1's appointment
		require.Len(t, userState.UserAppointments, 1)
		assert.Equal(t, user1ID, userState.UserAppointments[0].UserID)
	})
}
