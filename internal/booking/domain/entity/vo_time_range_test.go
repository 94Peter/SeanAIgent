package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTimeRange(t *testing.T) {
	now := time.Now()

	t.Run("Success", func(t *testing.T) {
		tr, err := NewTimeRange(now, now.Add(time.Hour))
		require.NoError(t, err)
		assert.Equal(t, now, tr.Start())
		assert.Equal(t, now.Add(time.Hour), tr.End())
	})

	t.Run("Fail_EndBeforeStart", func(t *testing.T) {
		_, err := NewTimeRange(now, now.Add(-time.Hour))
		assert.ErrorIs(t, err, ErrTrainingInvalidTime)
	})
}
