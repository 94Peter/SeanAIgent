package read

import (
	"seanAIgent/internal/booking/domain/entity"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGroupToWeeks(t *testing.T) {
	now := time.Now()
	// Start from a Monday to make it predictable
	offset := (int(now.Weekday()) + 6) % 7
	thisMonday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -offset)
	start := thisMonday
	end := start.AddDate(0, 0, 13)

	data := []*entity.TrainDateHasUserApptState{
		{
			ID:                "td1",
			StartDate:         thisMonday.Add(10 * time.Hour),
			EndDate:           thisMonday.Add(11 * time.Hour),
			Location:          "Taipei",
			Timezone:          "Asia/Taipei",
			Capacity:          10,
			AvailableCapacity: 5,
		},
		{
			ID:                "td2",
			StartDate:         thisMonday.AddDate(0, 0, 1).Add(14 * time.Hour),
			EndDate:           thisMonday.AddDate(0, 0, 1).Add(15 * time.Hour),
			Location:          "Hsinchu",
			Timezone:          "Asia/Taipei",
			Capacity:          10,
			AvailableCapacity: 10,
		},
	}

	weeks := groupToWeeks(data, start, end)

	assert.Len(t, weeks, 2)

	// Week 1
	assert.Len(t, weeks[0].Days, 7)
	// Day 1 (Monday) should have 1 slot
	assert.Len(t, weeks[0].Days[0].Slots, 1)
	assert.Equal(t, "td1", weeks[0].Days[0].Slots[0].ID)
	assert.Equal(t, "Taipei", weeks[0].Days[0].Slots[0].Location)

	// Day 2 (Tuesday) should have 1 slot
	assert.Len(t, weeks[0].Days[1].Slots, 1)
	assert.Equal(t, "td2", weeks[0].Days[1].Slots[0].ID)

	// Day 3 (Wednesday) should have an empty slot (IsEmpty: true)
	assert.Len(t, weeks[0].Days[2].Slots, 1)
	assert.True(t, weeks[0].Days[2].Slots[0].IsEmpty)
}

