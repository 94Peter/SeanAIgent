package stats

import (
	"os"
	"testing"
	"time"

	"seanAIgent/internal/booking/domain/repository"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleRepo_Integration(t *testing.T) {
	drapAllDb, closeFunc, err := mgo.InitTestContainer(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer closeFunc()

	file, err := os.Open("../test_data/training.json")
	require.NoError(t, err)
	err = mgo.Import(t.Context(), "training_date", file)
	require.NoError(t, err)

	file, err = os.Open("../test_data/appointment.json")
	require.NoError(t, err)
	err = mgo.Import(t.Context(), "appointment", file)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		defer drapAllDb()
		repo := NewStatsRepository()
		ctx := t.Context()
		filter := repository.NewFilterUserApptStatsByTrainTimeRange(
			time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 12, 31, 23, 59, 999, 999, time.UTC),
		)
		stats, err := repo.GetAllUserApptStats(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 2, len(stats))
		assert.Equal(t, 2, stats[0].TotalAppointment)
		assert.Equal(t, 59, stats[1].TotalAppointment)
	})
}
