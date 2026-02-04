package appointment

import (
	"context"
	"fmt"
	"os"
	"seanAIgent/internal/booking/domain/repository"
	"testing"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestGetQueryByFilterAppt(t *testing.T) {
	t.Run("FilterApptByTrainID", func(t *testing.T) {
		oid := bson.NewObjectID()
		filter := repository.NewFilterApptByTrainID(oid.Hex())
		query, err := getQueryByFilterAppt(filter)
		assert.NoError(t, err)
		assert.Equal(t, bson.M{"training_date_id": oid}, query)
	})

	t.Run("FilterApptByIDs", func(t *testing.T) {
		oid1 := bson.NewObjectID()
		oid2 := bson.NewObjectID()
		filter := repository.NewFilterApptByIDs(oid1.Hex(), oid2.Hex())
		query, err := getQueryByFilterAppt(filter)
		assert.NoError(t, err)
		assert.Equal(t, bson.M{"_id": bson.M{"$in": []bson.ObjectID{oid1, oid2}}}, query)
	})

	t.Run("FilterAppointmentByUserID", func(t *testing.T) {
		filter := repository.NewFilterApptByUserID("user-456")
		query, err := getQueryByFilterAppt(filter)
		assert.NoError(t, err)
		assert.Equal(t, bson.M{"user_id": "user-456"}, query)
	})
}

var cleanDb func()

func TestMain(m *testing.M) {
	// 檢查是否有傳入 bench 參數
	// 只有在需要跑測試時才啟動容器 (可選)
	ctx := context.Background()
	clean, closeFunc, err := mgo.InitTestContainer(ctx)
	if err != nil {
		fmt.Println("Setup failed")
		os.Exit(1)
	}
	cleanDb = clean

	code := m.Run() // 這裡會根據你的 -bench=xxx 自動篩選

	closeFunc()
	os.Exit(code)
}
