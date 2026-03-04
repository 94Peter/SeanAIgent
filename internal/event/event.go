package event

import (
	"context"
	"time"
)

// Event 是所有系統事件的基礎介面
type Event interface {
	ID() string        // 唯一的事件 ID
	Topic() string     // 事件主題
	OccurredAt() time.Time
	Data() []byte      // 優化: 使用原始位元組，避免 Store 與 Bus 進行反射處理
}

// Subscriber 定義了訂閱者介面
type Subscriber interface {
	ID() string
	Handle(ctx context.Context, e Event) error
}

// EventStore 負責事件的持久化與稽核
type EventStore interface {
	Save(ctx context.Context, e Event) error
	FindUnprocessedEvents(ctx context.Context, subscriberID string, topic string) ([]Event, error)
	UpdateProgress(ctx context.Context, subscriberID string, eventID string) error
}

// Bus 負責分發事件
type Bus interface {
	Publish(ctx context.Context, e Event)
	Subscribe(topic string, s Subscriber)
}
