package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrainDate(t *testing.T) {
	now := time.Now()
	period, _ := NewTimeRange(now, now.Add(time.Hour))

	t.Run("Success", func(t *testing.T) {
		td, err := NewTrainDate(
			WithBasicTrainDate("id1", "coach1", "Gym", 10, period),
		)
		require.NoError(t, err)
		assert.Equal(t, "id1", td.ID())
		assert.Equal(t, 10, td.AvailableCapacity())
		assert.Equal(t, TrainDateStatusActive, td.Status())
	})

	t.Run("Fail_InvalidAvailableCapacity", func(t *testing.T) {
		// NewTrainDate initializes availableCapacity to MinInt8.
		// If WithBasicTrainDate or similar opt doesn't set it, it should fail.
		// But WithBasicTrainDate sets it.
		// Let's try creating without setting basic info.
		td, err := NewTrainDate(
			WithTrainDateID("id1"),
		)
		assert.ErrorIs(t, err, ErrTrainingInvalidAvailableCapacity)
		assert.Nil(t, td)
	})
}

func TestTrainDate_ReserveSpot(t *testing.T) {
	now := time.Now()
	// Future training
	futureStart := now.Add(24 * time.Hour)
	period, _ := NewTimeRange(futureStart, futureStart.Add(time.Hour))

	td, _ := NewTrainDate(WithBasicTrainDate("id1", "coach1", "Gym", 10, period))

	t.Run("Success", func(t *testing.T) {
		err := td.ReserveSpot(1)
		require.NoError(t, err)
		assert.Equal(t, 9, td.AvailableCapacity())
	})

	t.Run("Fail_InvalidCount", func(t *testing.T) {
		err := td.ReserveSpot(0)
		assert.ErrorIs(t, err, ErrTrainingReserveCountInvalid)
		err = td.ReserveSpot(-1)
		assert.ErrorIs(t, err, ErrTrainingReserveCountInvalid)
	})

	t.Run("Fail_CapacityNotEnough", func(t *testing.T) {
		// Current capacity 9
		err := td.ReserveSpot(10)
		assert.ErrorIs(t, err, ErrTrainingCapacityNotEnough)
	})

	t.Run("Fail_TrainingOver", func(t *testing.T) {
		pastStart := now.Add(-2 * time.Hour)
		pastPeriod, _ := NewTimeRange(pastStart, pastStart.Add(time.Hour))
		pastTd, _ := NewTrainDate(WithBasicTrainDate("id2", "coach1", "Gym", 10, pastPeriod))

		err := pastTd.ReserveSpot(1)
		assert.ErrorIs(t, err, ErrTrainingOver)
	})
}

func TestTrainDate_ReleaseSpot(t *testing.T) {
	now := time.Now()
	period, _ := NewTimeRange(now, now.Add(time.Hour))
	// Max 10, currently 10
	td, _ := NewTrainDate(WithBasicTrainDate("id1", "coach1", "Gym", 10, period))

	t.Run("Success_RestoreToMax", func(t *testing.T) {
		// Reduce first
		_ = td.ReserveSpot(1) // 9
		err := td.ReleaseSpot(1)
		require.NoError(t, err)
		assert.Equal(t, 10, td.AvailableCapacity())
	})

	t.Run("Success_CapAtMax", func(t *testing.T) {
		// Currently 10
		err := td.ReleaseSpot(1)
		require.NoError(t, err)
		assert.Equal(t, 10, td.AvailableCapacity())
	})

	t.Run("Fail_InvalidCount", func(t *testing.T) {
		err := td.ReleaseSpot(0)
		assert.ErrorIs(t, err, ErrTrainingReleaseCountInvalid)
	})
}

func TestTrainDate_CanVerifyAttendance(t *testing.T) {
	now := time.Now()

	t.Run("True_InWindow", func(t *testing.T) {
		// Starts in 5 mins (Window is Start - 10 mins)
		start := now.Add(5 * time.Minute)
		period, _ := NewTimeRange(start, start.Add(time.Hour))
		td, _ := NewTrainDate(WithBasicTrainDate("id1", "c", "l", 10, period))

		assert.True(t, td.CanVerifyAttendance())
	})

	t.Run("True_AfterStart", func(t *testing.T) {
		start := now.Add(-5 * time.Minute)
		period, _ := NewTimeRange(start, start.Add(time.Hour))
		td, _ := NewTrainDate(WithBasicTrainDate("id1", "c", "l", 10, period))

		assert.True(t, td.CanVerifyAttendance())
	})

	t.Run("False_TooEarly", func(t *testing.T) {
		// Starts in 20 mins
		start := now.Add(20 * time.Minute)
		period, _ := NewTimeRange(start, start.Add(time.Hour))
		td, _ := NewTrainDate(WithBasicTrainDate("id1", "c", "l", 10, period))

		assert.False(t, td.CanVerifyAttendance())
	})
}

func TestTrainDate_Delete(t *testing.T) {
	now := time.Now()
	// Use future time to ensure ReserveSpot doesn't fail with ErrTrainingOver
	start := now.Add(time.Hour)
	period, _ := NewTimeRange(start, start.Add(time.Hour))

	t.Run("Success", func(t *testing.T) {
		td, _ := NewTrainDate(WithBasicTrainDate("id1", "c", "l", 10, period))
		err := td.Delete()
		require.NoError(t, err)
		assert.Equal(t, TrainDateStatusInactive, td.Status())
	})

	t.Run("Fail_HasAppointments", func(t *testing.T) {
		td, _ := NewTrainDate(WithBasicTrainDate("id1", "c", "l", 10, period))
		err := td.ReserveSpot(1) // Capacity changed
		require.NoError(t, err)

		err = td.Delete()
		assert.ErrorIs(t, err, ErrTrainingHasAppointments)
	})
}
