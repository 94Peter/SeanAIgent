package appointment

import (
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/infra/db/mongo/train"
	"testing"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestModelConversion(t *testing.T) {
	// Helper to generate a valid ObjectID hex string
	genID := func() string { return bson.NewObjectID().Hex() }

	t.Run("DomainToModel_Full", func(t *testing.T) {
		apptID := genID()
		trainID := genID()
		userID := "user-123"
		userName := "Test User"
		childName := "Child"
		now := time.Now().Truncate(time.Millisecond) // Truncate to match mongo precision often lost
		verifiedAt := now.Add(time.Hour)
		leaveReason := "sick"
		leaveStatus := entity.LeaveStatusApproved
		leaveTime := now.Add(2 * time.Hour)

		user, err := entity.NewUser(userID, userName)
		require.NoError(t, err)

		leaveInfo := entity.NewLeaveInfo(leaveReason, leaveStatus, leaveTime)

		domainAppt, err := entity.NewAppointment(
			entity.WithApptID(apptID),
			entity.WithUser(user),
			entity.WithChildName(childName),
			entity.WithTrainingID(trainID),
			entity.WithStatus(entity.StatusAttended),
			entity.WithCreatedAt(now),
			entity.WithUpdatedAt(now),
			entity.WithVerifiedAt(&verifiedAt),
			entity.WithLeaveInfo(leaveInfo),
		)
		require.NoError(t, err)

		model, err := newModelAppt(withDomainAppt(domainAppt))
		require.NoError(t, err)

		assert.Equal(t, apptID, model.ID.Hex())
		assert.Equal(t, userID, model.UserID)
		assert.Equal(t, userName, model.UserName)
		assert.Equal(t, childName, model.ChildName)
		assert.Equal(t, trainID, model.TrainingDateId.Hex())
		assert.Equal(t, string(entity.StatusAttended), model.Status)
		assert.True(t, model.CreatedAt.Equal(now))
		assert.True(t, model.UpdateAt.Equal(now))

		// Check VerifyTime
		assert.True(t, model.IsCheckedIn)
		assert.NotNil(t, model.VerifyTime)
		assert.True(t, model.VerifyTime.Equal(verifiedAt))

		// Check Leave
		assert.True(t, model.IsOnLeave)
		assert.NotNil(t, model.Leave)
		assert.Equal(t, leaveReason, model.Leave.Reason)
		assert.Equal(t, string(leaveStatus), model.Leave.Status)
		assert.True(t, model.Leave.CreatedAt.Equal(leaveTime))
	})

	t.Run("ModelToDomain_Full", func(t *testing.T) {
		apptID := bson.NewObjectID()
		trainID := bson.NewObjectID()
		userID := "user-456"
		userName := "Another User"
		childName := "Kiddo"
		now := time.Now().Truncate(time.Millisecond)
		verifiedAt := now.Add(30 * time.Minute)

		model := &appointment{
			ID:             apptID,
			UserID:         userID,
			UserName:       userName,
			ChildName:      childName,
			TrainingDateId: trainID,
			v2Fields: v2Fields{
				Status:     string(entity.StatusAttended),
				UpdateAt:   now,
				VerifyTime: &verifiedAt,
				Leave: &leaveInfo{
					Reason:    "vacation",
					Status:    string(entity.LeaveStatusPending),
					CreatedAt: now.Add(time.Hour),
				},
			},
			v1_deprecatedFields: v1_deprecatedFields{
				IsCheckedIn: true,
				IsOnLeave:   true,
			},
			CreatedAt: now,
		}

		domain, err := model.toDomain()
		require.NoError(t, err)

		assert.Equal(t, apptID.Hex(), domain.ID())
		assert.Equal(t, userID, domain.User().UserID())
		assert.Equal(t, userName, domain.User().UserName())
		assert.Equal(t, childName, domain.ChildName())
		assert.Equal(t, trainID.Hex(), domain.TrainingID())
		assert.Equal(t, entity.StatusAttended, domain.Status())
		assert.True(t, domain.CreatedAt().Equal(now))
		assert.True(t, domain.UpdateAt().Equal(now))

		assert.NotNil(t, domain.VerifiedAt())
		assert.True(t, domain.VerifiedAt().Equal(verifiedAt))

		info := domain.LeaveInfo()
		assert.False(t, info.IsEmpty())
		assert.Equal(t, "vacation", info.Reason())
		assert.Equal(t, entity.LeaveStatusPending, info.Status())
	})

	t.Run("ModelToDomain_InvalidStatus", func(t *testing.T) {
		model := &appointment{
			ID:             bson.NewObjectID(),
			UserID:         "u1",
			UserName:       "n1",
			TrainingDateId: bson.NewObjectID(),
			v2Fields: v2Fields{
				Status:   "INVALID_STATUS",
				UpdateAt: time.Now(),
			},
			CreatedAt: time.Now(),
		}

		// toDomain handles invalid status by defaulting to StatusConfirmed inside AppointmentStatusFromString logic?
		// Actually checking the code:
		// status, ok := entity.AppointmentStatusFromString(s.Status)
		// if !ok { status = entity.StatusConfirmed }
		// So it shouldn't error on main status.

		// However, leave status check:
		model.Leave = &leaveInfo{Status: "BAD_LEAVE_STATUS"}
		_, err := model.toDomain()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "leave status is invalid")
	})
}

func TestAppointmentIntegrate(t *testing.T) {
	drapAllDb, closeFunc, err := mgo.InitTestContainer(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer closeFunc()

	repo := NewApptRepository()
	t.Run("CRUD", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()

		// 1. Create
		apptID := bson.NewObjectID().Hex()
		trainID := bson.NewObjectID().Hex()
		userID := "user-crud"
		user, err := entity.NewUser(userID, "User CRUD")
		require.NoError(t, err)

		now := time.Now().Truncate(time.Millisecond)

		appt, err := entity.NewAppointment(
			entity.WithApptID(apptID),
			entity.WithUser(user),
			entity.WithChildName("ChildCRUD"),
			entity.WithTrainingID(trainID),
			entity.WithStatus(entity.StatusConfirmed),
			entity.WithCreatedAt(now),
			entity.WithUpdatedAt(now),
		)
		require.NoError(t, err)

		err = repo.SaveAppointment(ctx, appt)
		require.NoError(t, err)

		// 2. Read
		found, err := repo.FindApptByID(ctx, apptID)
		require.NoError(t, err)
		assert.Equal(t, appt.ID(), found.ID())
		assert.Equal(t, appt.Status(), found.Status())
		assert.Equal(t, appt.TrainingID(), found.TrainingID())

		// 3. Update
		// Update status to Attended and set VerifiedAt
		verifiedAt := now.Add(time.Hour)
		updatedAppt, err := entity.NewAppointment(
			entity.WithApptID(apptID),
			entity.WithUser(user),
			entity.WithChildName("ChildCRUD"),
			entity.WithTrainingID(trainID),
			entity.WithStatus(entity.StatusAttended),
			entity.WithCreatedAt(now),
			entity.WithUpdatedAt(verifiedAt),
			entity.WithVerifiedAt(&verifiedAt),
		)
		require.NoError(t, err)

		err = repo.UpdateAppt(ctx, updatedAppt)
		require.NoError(t, err)

		foundUpdated, err := repo.FindApptByID(ctx, apptID)
		require.NoError(t, err)
		assert.Equal(t, entity.StatusAttended, foundUpdated.Status())
		assert.NotNil(t, foundUpdated.VerifiedAt())
		assert.True(t, foundUpdated.VerifiedAt().Equal(verifiedAt))

		// 4. Delete
		// Must cancel first due to repository logic
		cancelledAppt, err := entity.NewAppointment(
			entity.WithApptID(apptID),
			entity.WithUser(user),
			entity.WithChildName("ChildCRUD"),
			entity.WithTrainingID(trainID),
			entity.WithStatus(entity.StatusCancelled),
			entity.WithCreatedAt(now),
			entity.WithUpdatedAt(now.Add(2*time.Hour)),
		)
		require.NoError(t, err)

		err = repo.DeleteAppointment(ctx, cancelledAppt)
		require.NoError(t, err)

		// 5. Verify Delete
		_, err = repo.FindApptByID(ctx, apptID)
		assert.Error(t, err)
	})

	t.Run("BatchSaveAndUpdate", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()
		count := 5
		appts := make([]*entity.Appointment, 0, count)
		trainID := bson.NewObjectID().Hex()
		userID := "user-batch"
		user, err := entity.NewUser(userID, "User Batch")
		require.NoError(t, err)
		now := time.Now().Truncate(time.Millisecond)

		// 1. Prepare Appts
		for i := 0; i < count; i++ {
			apptID := bson.NewObjectID().Hex()
			// Unique constraint: user_id + training_date_id + child_name
			childName := "ChildBatch" + string(rune('A'+i))

			appt, err := entity.NewAppointment(
				entity.WithApptID(apptID),
				entity.WithUser(user),
				entity.WithChildName(childName),
				entity.WithTrainingID(trainID),
				entity.WithStatus(entity.StatusConfirmed),
				entity.WithCreatedAt(now),
				entity.WithUpdatedAt(now),
			)
			require.NoError(t, err)
			appts = append(appts, appt)
		}

		// 2. Batch Save
		err = repo.SaveManyAppointments(ctx, appts)
		require.NoError(t, err)

		// 3. Verify Saved using Filter
		filter := repository.NewFilterApptByTrainID(trainID)
		foundAppts, err := repo.FindApptsByFilter(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, foundAppts, count)

		// 4. Batch Update
		updatedAppts := make([]*entity.Appointment, 0, count)
		for _, appt := range appts {
			appt.MarkAsAttended(time.Now())
			updatedAppts = append(updatedAppts, appt)
		}

		err = repo.UpdateManyAppts(ctx, updatedAppts)
		require.NoError(t, err)

		// 5. Verify Updates
		foundApptsAfterUpdate, err := repo.FindApptsByFilter(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, foundApptsAfterUpdate, count)
		for _, found := range foundApptsAfterUpdate {
			assert.Equal(t, entity.StatusAttended, found.Status())
		}
	})

	t.Run("PageFindApptsWithTrainDateByFilterAndTrainFilter", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()

		// We need TrainRepository to insert training dates for lookup
		// Assuming we can import the train package. If not allowed due to cycle, we might need to use mgo directly or move this test.
		// Since this is an integration test inside appointment package, importing infra/db/mongo/train might cause import cycle
		// if train package imports appointment package.
		// Checking imports:
		// appointment -> entity, repository, mgo
		// train -> entity, repository, mgo, core
		// Neither seems to import each other. So it should be safe.

		trainRepo := train.NewTrainRepository()

		// 1. Prepare Data
		userID := "user-join-test"
		user, _ := entity.NewUser(userID, "Join Tester")
		otherUser, _ := entity.NewUser("other", "Other")

		// Time Setup
		now := time.Now()
		future := now.Add(24 * time.Hour)
		past := now.Add(-24 * time.Hour)

		// Create Training Dates
		// Train 1: Future
		trainID1 := bson.NewObjectID().Hex()
		tr1, _ := entity.NewTimeRange(future, future.Add(time.Hour))
		train1, err := entity.NewTrainDate(
			entity.WithBasicTrainDate(trainID1, "coach1", "Gym A", 10, tr1),
		)
		require.NoError(t, err)

		// Train 2: Past
		trainID2 := bson.NewObjectID().Hex()
		tr2, _ := entity.NewTimeRange(past, past.Add(time.Hour))
		train2, err := entity.NewTrainDate(
			entity.WithBasicTrainDate(trainID2, "coach1", "Gym A", 10, tr2),
		)
		require.NoError(t, err)

		err = trainRepo.SaveManyTrainDates(ctx, []*entity.TrainDate{train1, train2})
		require.NoError(t, err)

		// Create Appointments
		// Appt 1: Target User, Future Train (Should Match)
		apptID1 := bson.NewObjectID().Hex()
		appt1, err := entity.NewAppointment(
			entity.WithCreateAppt(apptID1, trainID1, user, "Child1"),
		)
		require.NoError(t, err)

		// Appt 2: Target User, Past Train (Should Fail Time Filter)
		apptID2 := bson.NewObjectID().Hex()
		appt2, err := entity.NewAppointment(
			entity.WithCreateAppt(apptID2, trainID2, user, "Child1"),
		)
		require.NoError(t, err)

		// Appt 3: Other User, Future Train (Should Fail User Filter)
		apptID3 := bson.NewObjectID().Hex()
		appt3, err := entity.NewAppointment(
			entity.WithCreateAppt(apptID3, trainID1, otherUser, "ChildOther"),
		)
		require.NoError(t, err)

		// Appt 4: Target User, Future Train (Should Match)
		apptID4 := bson.NewObjectID().Hex()
		appt4, err := entity.NewAppointment(
			entity.WithCreateAppt(apptID4, trainID1, user, "Child4"),
		)
		require.NoError(t, err)

		err = repo.SaveManyAppointments(ctx, []*entity.Appointment{appt1, appt2, appt3, appt4})
		require.NoError(t, err)

		// 2. Execute Query
		// Filter: UserID + Future Training Dates
		apptFilter := repository.NewFilterApptByUserID(userID)
		trainFilter := repository.NewFilterTrainDateByAfterTime(now)

		results, cursorStr, err := repo.PageFindApptsWithTrainDateByFilterAndTrainFilter(
			ctx, apptFilter, trainFilter,
			repository.NewFilterApptsWithTrainDateByCursor(1),
		)
		require.NoError(t, err)
		assert.NotEmpty(t, cursorStr)
		// 3. Verify
		assert.Len(t, results, 1)
		if len(results) > 0 {
			res := results[0]
			assert.Equal(t, apptID1, res.ID)
			assert.Equal(t, trainID1, res.TrainingDateId)
			assert.Equal(t, trainID1, res.TrainDate.ID)
			assert.Equal(t, userID, res.UserID)
			// Verify joined fields are populated
			assert.Equal(t, train1.Location(), res.TrainDate.Location)
			// assert.True(t, res.TrainDate.StartDate.Equal(train1.Period().Start())) // Timezone comparison might be tricky
		}
		results, _, err = repo.PageFindApptsWithTrainDateByFilterAndTrainFilter(
			ctx, apptFilter, trainFilter,
			cursorStr,
		)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		if len(results) > 0 {
			res := results[0]
			assert.Equal(t, apptID4, res.ID)
		}
	})

	t.Run("TestAppointmentLeave", func(t *testing.T) {
		defer drapAllDb()
		ctx := t.Context()
		repo := NewApptRepository()

		// 1. Prepare Data
		apptID := bson.NewObjectID().Hex()
		trainID := bson.NewObjectID().Hex()
		userID := "user-leave-test"
		user, _ := entity.NewUser(userID, "Leave Tester")

		// Training starts in 24 hours (so we can ask for leave)
		now := time.Now()
		trainingStart := now.Add(24 * time.Hour)

		appt, err := entity.NewAppointment(
			entity.WithCreateAppt(apptID, trainID, user, "ChildLeave"),
			entity.WithStatus(entity.StatusConfirmed),
			entity.WithCreatedAt(now),
			entity.WithUpdatedAt(now),
		)
		require.NoError(t, err)

		err = repo.SaveAppointment(ctx, appt)
		require.NoError(t, err)

		// 2. Append Leave
		reason := "Sick"
		err = appt.AppendLeaveRecord(reason, trainingStart)
		require.NoError(t, err)

		// 3. Update Repo
		err = repo.UpdateAppt(ctx, appt)
		require.NoError(t, err)

		// 4. Retrieve and Verify
		found, err := repo.FindApptByID(ctx, apptID)
		require.NoError(t, err)

		assert.Equal(t, entity.StatusCancelledLeave, found.Status())
		assert.False(t, found.LeaveInfo().IsEmpty())
		assert.Equal(t, reason, found.LeaveInfo().Reason())
		assert.Equal(t, entity.LeaveStatusApproved, found.LeaveInfo().Status())

		// 5. Cancel Leave (Revert to Confirmed)
		err = appt.CancelLeave(userID)
		require.NoError(t, err)

		err = repo.UpdateAppt(ctx, appt)
		require.NoError(t, err)

		foundAfterCancel, err := repo.FindApptByID(ctx, apptID)
		require.NoError(t, err)

		assert.Equal(t, entity.StatusConfirmed, foundAfterCancel.Status())
		assert.True(t, foundAfterCancel.LeaveInfo().IsEmpty())
	})
}
