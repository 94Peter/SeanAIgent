package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFilterApptsWithTrainDateByCursor(t *testing.T) {
	t.Run("EncodeDecode", func(t *testing.T) {
		cursor := &FilterApptsWithTrainDateByCursor{
			LastStartDate: time.Now(),
			LastID:        "123",
			PageSize:      10,
		}
		encoded := cursor.Encode()
		decoded := &FilterApptsWithTrainDateByCursor{}
		err := decoded.Decode(encoded)
		assert.NoError(t, err)
		assert.Equal(t, cursor.LastStartDate.Truncate(0), decoded.LastStartDate)
		assert.Equal(t, cursor.LastID, decoded.LastID)
		assert.Equal(t, cursor.PageSize, decoded.PageSize)
	})
}
