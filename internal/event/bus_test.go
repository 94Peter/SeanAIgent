package event

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 業務 Payload 定義
type AppointmentPayload struct {
	BookingID string `json:"bookingId"`
	Status    string `json:"status"`
}

func TestEventSystemOptimization(t *testing.T) {
	ctx := t.Context()
	uri := os.Getenv("TEST_MONGO_URI")
	if uri == "" {
		_, closeFunc, err := mgo.InitTestContainer(ctx)
		require.NoError(t, err)
		defer closeFunc()
	}

	t.Run("TypedEventAndSubscriberFlow", func(t *testing.T) {
		memStore := &mockStore{
			events:   make(map[string][]Event),
			progress: make(map[string]string),
		}
		bus := NewBus(memStore)
		topic := "booking.status_changed"

		// 1. 發布強型別事件
		payload := AppointmentPayload{BookingID: "B001", Status: "ATTENDED"}
		evt := NewTypedEvent("evt_1", topic, payload)
		bus.Publish(ctx, evt)

		// 2. 建立強型別訂閱者
		var receivedPayload AppointmentPayload
		var wg sync.WaitGroup
		wg.Add(1)

		handler := func(ctx context.Context, e Event, p AppointmentPayload) error {
			receivedPayload = p
			wg.Done()
			return nil
		}

		// 使用泛型包裝器，自動處理 Unmarshal
		sub := NewTypedSubscriber("stats_processor", topic, handler)
		bus.Subscribe(topic, sub)

		wg.Wait()

		// 驗證
		assert.Equal(t, "B001", receivedPayload.BookingID)
		assert.Equal(t, "ATTENDED", receivedPayload.Status)
	})
}

// MockStore 與 genericEvent 定義 (省略，同前次)
type mockStore struct {
	events   map[string][]Event
	progress map[string]string
}

func (m *mockStore) Save(ctx context.Context, e Event) error {
	m.events[e.Topic()] = append(m.events[e.Topic()], e)
	return nil
}

func (m *mockStore) UpdateProgress(ctx context.Context, subID, evtID string) error {
	m.progress[subID] = evtID
	return nil
}

func (m *mockStore) FindUnprocessedEvents(ctx context.Context, subID, topic string) ([]Event, error) {
	all := m.events[topic]
	last := m.progress[subID]
	if last == "" { return all, nil }
	var found bool
	var result []Event
	for _, e := range all {
		if found { result = append(result, e) }
		if e.ID() == last { found = true }
	}
	return result, nil
}
