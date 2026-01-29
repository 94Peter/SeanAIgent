package appointment

import (
	"seanAIgent/internal/booking/domain/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestGetPipelineApptWithTrainDateWithTrainFilter(t *testing.T) {
	// Mock query
	q := bson.M{"user_id": "u1"}

	t.Run("WithFilterTrainDateByAfterTime", func(t *testing.T) {
		start := time.Now()
		filter := repository.NewFilterTrainDateByAfterTime(start)

		pipeline := getPipelineApptWithTrainDateWithTrainFilter(q, filter)

		// Expected stages:
		// 0: $match (q)
		// 1: $lookup (training_date)
		// 2: $unwind (training_date_info)
		// 3: $match (training_date_info.start_date >= start)
		// 4: $sort
		// 5: $project
		assert.Len(t, pipeline, 6)

		// Verify stage 0 ($match q)
		stage0 := pipeline[0]
		assert.Equal(t, q, stage0[0].Value)

		// Verify stage 3 (Train filter)
		stage3 := pipeline[3]
		assert.Equal(t, "$match", stage3[0].Key)

		// Verify content of stage 3 match
		matchContent := stage3[0].Value.(bson.D)
		// "training_date_info.start_date" -> {"$gte": start}
		assert.Equal(t, "training_date_info.start_date", matchContent[0].Key)
		cond := matchContent[0].Value.(bson.D)
		assert.Equal(t, "$gte", cond[0].Key)
		assert.Equal(t, start, cond[0].Value)
	})

	t.Run("WithNoOpFilter", func(t *testing.T) {
		// Using a filter that defaults to empty bson.D{}
		// Currently only FilterTrainDateByAfterTime is handled in switch
		// We can mock another filter or use one that isn't handled if available
		// repository.NewFilterTrainDateByIds is not handled in the switch in aggr_user_appts_detail.go

		filter := repository.NewFilterTrainDateByIds("id1")

		pipeline := getPipelineApptWithTrainDateWithTrainFilter(q, filter)

		// Expected stages:
		// 0-2: Lookup/Unwind
		// 3: Empty bson.D{} (Default case)
		// 4: Sort
		// 5: Project
		assert.Len(t, pipeline, 6)

		// Verify stage 3 is empty
		stage3 := pipeline[3]
		assert.Empty(t, stage3)
	})
}
