package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppointment(t *testing.T) {
	user, _ := NewUser("u1", "User")
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		appt, err := NewAppointment(
			WithCreateAppt("a1", "t1", user, "Child"),
		)
		require.NoError(t, err)
		assert.Equal(t, "a1", appt.ID())
		assert.Equal(t, StatusConfirmed, appt.Status())
	})

	t.Run("Fail_Validate", func(t *testing.T) {
		// Missing ID
		appt, err := NewAppointment(
			WithTrainingID("t1"),
			WithUser(user),
			WithChildName("Child"),
			WithStatus(StatusConfirmed),
			WithCreatedAt(now),
			WithUpdatedAt(now),
		)
		assert.ErrorIs(t, err, ErrAppointmentInvalid)
		assert.Nil(t, appt)
	})
}

func TestAppointment_CancelAsMistake(t *testing.T) {
	user, _ := NewUser("u1", "User")
	
	t.Run("Success", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		err := appt.CancelAsMistake("u1")
		require.NoError(t, err)
		assert.Equal(t, StatusCancelled, appt.Status())
	})

	t.Run("Fail_NotUser", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		err := appt.CancelAsMistake("u2")
		assert.ErrorIs(t, err, ErrAppointmentNotBelongToUser)
	})

	t.Run("Fail_Timeout", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		// Hack creation time
		oldTime := time.Now().Add(-25 * time.Hour)
		WithCreatedAt(oldTime)(appt)

		err := appt.CancelAsMistake("u1")
		assert.ErrorIs(t, err, ErrAppointmentCancelTimeout)
	})

	t.Run("Fail_InvalidStatus", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		_ = appt.CancelAsMistake("u1") // Cancelled

		err := appt.CancelAsMistake("u1")
		assert.ErrorIs(t, err, ErrAppointmentInvalidStatus)
	})
}

func TestAppointment_MarkAsAttended(t *testing.T) {
	user, _ := NewUser("u1", "User")
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		// Training starts in 5 mins (within 10min window)
		trainStart := now.Add(5 * time.Minute)
		
		err := appt.MarkAsAttended(trainStart)
		require.NoError(t, err)
		assert.Equal(t, StatusAttended, appt.Status())
		assert.NotNil(t, appt.VerifiedAt())
	})

	t.Run("Fail_TooEarly", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		// Training starts in 20 mins (too early)
		trainStart := now.Add(20 * time.Minute)

		err := appt.MarkAsAttended(trainStart)
		assert.ErrorIs(t, err, ErrAppointmentCheckInNotOpen)
	})

	t.Run("Fail_TooLate", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		// Training started 4 days ago (max 3 days)
		trainStart := now.Add(-4 * 24 * time.Hour)

		err := appt.MarkAsAttended(trainStart)
		assert.ErrorIs(t, err, ErrAppointmentCheckInTooLate)
	})

	t.Run("Fail_OnLeave", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		// Set to Leave
		WithStatus(StatusCancelledLeave)(appt)

		trainStart := now.Add(5 * time.Minute)
		err := appt.MarkAsAttended(trainStart)
		assert.ErrorIs(t, err, ErrAppointmentOnLeave)
	})
}

func TestAppointment_AppendLeaveRecord(t *testing.T) {
	user, _ := NewUser("u1", "User")
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		// Training starts in 3 hours (> 2 hours)
		trainStart := now.Add(3 * time.Hour)

		err := appt.AppendLeaveRecord("Sick", trainStart)
		require.NoError(t, err)
		assert.Equal(t, StatusCancelledLeave, appt.Status())
		assert.Equal(t, "Sick", appt.LeaveInfo().Reason())
	})

	t.Run("Fail_TooLate", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		// Training starts in 1 hour (< 2 hours)
		trainStart := now.Add(1 * time.Hour)

		err := appt.AppendLeaveRecord("Sick", trainStart)
		assert.ErrorIs(t, err, ErrAppointmentLeaveTooLate)
	})

	t.Run("Fail_EmptyReason", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		trainStart := now.Add(3 * time.Hour)

		err := appt.AppendLeaveRecord("", trainStart)
		assert.ErrorIs(t, err, ErrAppointmentLeaveReasonEmpty)
	})

	t.Run("Fail_AlreadyCancelled", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		WithStatus(StatusCancelled)(appt)
		
		trainStart := now.Add(3 * time.Hour)
		err := appt.AppendLeaveRecord("Sick", trainStart)
		assert.ErrorIs(t, err, ErrAppointmentCannotLeave)
	})
}

func TestAppointment_CancelLeave(t *testing.T) {
	user, _ := NewUser("u1", "User")
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		// Setup leave
		_ = appt.AppendLeaveRecord("Sick", now.Add(3*time.Hour))

		err := appt.CancelLeave("u1")
		require.NoError(t, err)
		assert.Equal(t, StatusConfirmed, appt.Status())
		assert.True(t, appt.LeaveInfo().IsEmpty())
	})

	t.Run("Fail_NotUser", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		_ = appt.AppendLeaveRecord("Sick", now.Add(3*time.Hour))

		err := appt.CancelLeave("u2")
		assert.ErrorIs(t, err, ErrAppointmentNotBelongToUser)
	})

	t.Run("Fail_NoLeave", func(t *testing.T) {
		appt, _ := NewAppointment(WithCreateAppt("a1", "t1", user, "Child"))
		err := appt.CancelLeave("u1")
		assert.ErrorIs(t, err, ErrAppointmentLeaveNotApproved)
	})
}
