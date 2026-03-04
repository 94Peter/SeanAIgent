package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TypedHandler 是具有強型別 Payload 的處理函數
type TypedHandler[T any] func(ctx context.Context, e Event, payload T) error

// NewTypedSubscriber 建立一個強型別的訂閱者
func NewTypedSubscriber[T any](id string, handler TypedHandler[T]) Subscriber {
	return &typedSubscriber[T]{
		id:      id,
		handler: handler,
	}
}

type typedSubscriber[T any] struct {
	id      string
	handler TypedHandler[T]
}

func (s *typedSubscriber[T]) ID() string { return s.id }

func (s *typedSubscriber[T]) Handle(ctx context.Context, e Event) error {
	var payload T
	// 只有在最終的 Subscriber 這裡才會進行一次 Unmarshal
	if err := json.Unmarshal(e.Data(), &payload); err != nil {
		return fmt.Errorf("TypedSubscriber[%s]: unmarshal fail: %w", s.id, err)
	}
	return s.handler(ctx, e, payload)
}

// TypedEvent 是泛型事件包裝器，方便 Producer 發送
type TypedEvent[T any] struct {
	id         string
	topic      string
	occurredAt time.Time
	payload    T
}

func NewTypedEvent[T any](id, topic string, payload T) Event {
	return &TypedEvent[T]{
		id:         id,
		topic:      topic,
		occurredAt: time.Now(),
		payload:    payload,
	}
}

func (e *TypedEvent[T]) ID() string          { return e.id }
func (e *TypedEvent[T]) Topic() string       { return e.topic }
func (e *TypedEvent[T]) OccurredAt() time.Time { return e.occurredAt }
func (e *TypedEvent[T]) Data() []byte {
	b, _ := json.Marshal(e.payload)
	return b
}
