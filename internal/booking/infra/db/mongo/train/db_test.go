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
	IsCheckedIn    bool          `bson:"is_checked_in"`
	IsOnLeave      bool          `bson:"is_on_leave"`
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

		// Verify Delete
		_, err = repo.FindTrainDateByID(ctx, trainID)
		assert.Error(t, err) // Should be not found
	})

	t.Run("Capacity", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()
		timeRange, _ := entity.NewTimeRange(
			time.Date(2025, 2, 4, 10, 0, 0, 0, time.UTC),
			time.Date(2025, 2, 4, 12, 0, 0, 0, time.UTC),
		)
		trainID := bson.NewObjectID().Hex()
		trainDate, _ := entity.NewTrainDate(
			entity.WithBasicTrainDate(
				trainID, "coach1", "Gym B", 10, timeRange),
		)
		err := repo.SaveTrainDate(ctx, trainDate)
		require.NoError(t, err)

		// Deduct
		err = repo.DeductCapacity(ctx, trainID, 2)
		require.NoError(t, err)

		found, _ := repo.FindTrainDateByID(ctx, trainID)
		assert.Equal(t, 8, found.AvailableCapacity())

		// Increase
		err = repo.IncreaseCapacity(ctx, trainID, 1)
		require.NoError(t, err)

		found, _ = repo.FindTrainDateByID(ctx, trainID)
		assert.Equal(t, 9, found.AvailableCapacity())
	})

	t.Run("Overlap", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()
		baseStart := time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC)
		baseEnd := baseStart.Add(2 * time.Hour)
		timeRange, _ := entity.NewTimeRange(baseStart, baseEnd)

		trainID := bson.NewObjectID().Hex()
		trainDate, _ := entity.NewTrainDate(
			entity.WithBasicTrainDate(
				trainID, "coach_overlap", "Gym C", 5, timeRange),
		)
		err := repo.SaveTrainDate(ctx, trainDate)
		require.NoError(t, err)

		// Check Overlap
		overlapRange, _ := entity.NewTimeRange(baseStart.Add(1*time.Hour), baseStart.Add(3*time.Hour))
		isOverlap, err := repo.CheckOverlap(ctx, "coach_overlap", overlapRange)
		require.NoError(t, err)
		assert.True(t, isOverlap)

		noOverlapRange, _ := entity.NewTimeRange(baseStart.Add(3*time.Hour), baseStart.Add(5*time.Hour))
		isOverlap, err = repo.CheckOverlap(ctx, "coach_overlap", noOverlapRange)
		require.NoError(t, err)
		assert.False(t, isOverlap)

		// Check HasAnyOverlap
		isAnyOverlap, err := repo.HasAnyOverlap(ctx, "coach_overlap", []entity.TimeRange{noOverlapRange, overlapRange})
		require.NoError(t, err)
		assert.True(t, isAnyOverlap)
	})

	t.Run("Aggregations", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()
		var err error
		// 1. Create TrainDate
		start := time.Date(2025, 4, 1, 10, 0, 0, 0, time.UTC)
		end := start.Add(2 * time.Hour)
		tr, _ := entity.NewTimeRange(start, end)
		trainID := bson.NewObjectID()

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
			IsCheckedIn:    false,
			IsOnLeave:      false,
		}

		appt2 := &testAppointment{
			Index:          testAppointmentCollection,
			ID:             bson.NewObjectID(),
			UserID:         user2ID,
			UserName:       "User Two",
			ChildName:      "Child Two",
			TrainingDateId: trainID,
			Status:         "CONFIRMED",
			CreatedAt:      time.Now(),
			IsCheckedIn:    true,
			IsOnLeave:      false,
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

		// AllUsers should contain names of all users (User One, User Two)
		// Note: user_appointments in aggregation might be filtered, but "allUsers" logic depends on implementation
		// Looking at aggr_user_train_has_appts.go:
		// "allUsers" maps from all appointments where is_on_leave is false.
		assert.Contains(t, userState.AllUsers, "Child One")
		assert.Contains(t, userState.AllUsers, "Child Two")
	})

	t.Run("SaveManyAndFind", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()
		// Create 2 train dates
		tr1, _ := entity.NewTimeRange(
			time.Date(2025, 5, 1, 10, 0, 0, 0, time.UTC),
			time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC),
		)
		td1, _ := entity.NewTrainDate(entity.WithBasicTrainDate(bson.NewObjectID().Hex(), "c1", "L1", 5, tr1))

		tr2, _ := entity.NewTimeRange(
			time.Date(2025, 5, 2, 10, 0, 0, 0, time.UTC),
			time.Date(2025, 5, 2, 12, 0, 0, 0, time.UTC),
		)
		td2, _ := entity.NewTrainDate(entity.WithBasicTrainDate(bson.NewObjectID().Hex(), "c1", "L1", 5, tr2))

		err = repo.SaveManyTrainDates(ctx, []*entity.TrainDate{td1, td2})
		require.NoError(t, err)

		// Find by time range covering both
		filter := repository.NewFilterTrainDataByTimeRange(
			time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 5, 3, 0, 0, 0, 0, time.UTC),
		)
		found, err := repo.FindTrainDates(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(found), 2)
	})

	t.Run("NearestTime", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()
		// Create 2 train dates
		tr1, _ := entity.NewTimeRange(
			time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC),
			time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC),
		)
		td1, _ := entity.NewTrainDate(entity.WithBasicTrainDate(bson.NewObjectID().Hex(), "c1", "L1", 5, tr1))
		tr2, _ := entity.NewTimeRange(
			time.Date(2025, 7, 2, 10, 0, 0, 0, time.UTC),
			time.Date(2025, 7, 2, 12, 0, 0, 0, time.UTC),
		)
		td2, _ := entity.NewTrainDate(entity.WithBasicTrainDate(bson.NewObjectID().Hex(), "c1", "L1", 5, tr2))

		err := repo.SaveManyTrainDates(ctx, []*entity.TrainDate{td1, td2})
		require.NoError(t, err)

		// Find by time range covering both
		filter := repository.NewFilterTrainDateByEndTime(time.Date(2025, 7, 1, 11, 0, 0, 0, time.UTC))
		found, err := repo.FindTrainDates(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 2, len(found))
		assert.Equal(t, td1.ID(), found[0].ID())
	})
}
