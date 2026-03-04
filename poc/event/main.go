package main

import (
	"context"
	"fmt"
	"seanAIgent/internal/event"
	"sync"
	"time"
)

// 1. 定義業務 Payload (必須有 JSON tag 方便自動序列化)
type AppointmentStatusChanged struct {
	BookingID string `json:"booking_id"`
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	UserID    string `json:"user_id"`
}

func main() {
	fmt.Println("🚀 Sean AIgent Event System POC")
	fmt.Println("================================")

	// 2. 初始化環境 (這裡使用簡單的內存 Store 做展示)
	store := NewMemStore()
	bus := event.NewBus(store)
	ctx := context.Background()

	// 3. 定義訂閱者 (例如：經營分析模組)
	statsSubscriber := event.NewTypedSubscriber(
		"stats_processor_v1",
		func(ctx context.Context, e event.Event, p AppointmentStatusChanged) error {
			fmt.Printf("📊 [統計模組] 收到事件 %s: 學員 %s 狀態變更為 %s\n", 
				e.ID(), p.UserID, p.NewStatus)
			
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("✅ [統計模組] 已完成 BookingID %s 的預聚合更新\n", p.BookingID)
			return nil
		},
	)

	// 4. 定義另一個訂閱者 (例如：通知模組)
	notifySubscriber := event.NewTypedSubscriber(
		"notify_service_v1",
		func(ctx context.Context, e event.Event, p AppointmentStatusChanged) error {
			fmt.Printf("🔔 [通知模組] 發送 LINE 推播給 %s: 您的預約 %s 已更新！\n", p.UserID, p.BookingID)
			return nil
		},
	)

	// 5. 註冊訂閱
	topic := "booking.appt_changed"
	bus.Subscribe(topic, statsSubscriber)
	bus.Subscribe(topic, notifySubscriber)

	// 6. 模擬業務邏輯發布事件
	fmt.Println("\n[業務流程] 正在處理批次簽到...")
	
	payload := AppointmentStatusChanged{
		BookingID: "BOOKING_999",
		OldStatus: "CONFIRMED",
		NewStatus: "ATTENDED",
		UserID:    "USER_PETER",
	}
	
	evt := event.NewTypedEvent("evt_abc_123", topic, payload)
	bus.Publish(ctx, evt)

	fmt.Println("[業務流程] 簽到完成，API 已回傳 (異步任務持續執行中...)\n")

	// 為了在控制台看到異步結果，我們稍微等一下
	time.Sleep(2 * time.Second)
	fmt.Println("\n================================")
	fmt.Println("🎉 POC 執行完畢")
}

// --- 以下為簡單的內存 Store 實作，僅供 POC 執行 ---

type memStore struct {
	events   map[string][]event.Event
	progress map[string]string
	mu       sync.RWMutex
}

func NewMemStore() event.EventStore {
	return &memStore{
		events:   make(map[string][]event.Event),
		progress: make(map[string]string),
	}
}

func (m *memStore) Save(ctx context.Context, e event.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[e.Topic()] = append(m.events[e.Topic()], e)
	return nil
}

func (m *memStore) UpdateProgress(ctx context.Context, subID, evtID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.progress[subID] = evtID
	return nil
}

func (m *memStore) FindUnprocessedEvents(ctx context.Context, subID, topic string) ([]event.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := m.events[topic]
	last := m.progress[subID]
	if last == "" {
		return all, nil
	}
	var found bool
	var result []event.Event
	for _, e := range all {
		if found {
			result = append(result, e)
		}
		if e.ID() == last {
			found = true
		}
	}
	return result, nil
}
