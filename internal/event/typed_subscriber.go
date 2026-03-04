package event

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
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

	// 處理指標型別的初始化
	rv := reflect.ValueOf(&payload).Elem()
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	// 取得實際用於 Unmarshal 的對象
	target := any(&payload)
	if u, ok := target.(Unmarshaler); ok {
		if err := u.Unmarshal(e.Data()); err != nil {
			return fmt.Errorf("TypedSubscriber[%s]: custom unmarshal fail: %w", s.id, err)
		}
	} else if u, ok := any(payload).(Unmarshaler); ok {
		if err := u.Unmarshal(e.Data()); err != nil {
			return fmt.Errorf("TypedSubscriber[%s]: custom unmarshal fail: %w", s.id, err)
		}
	} else {
		if err := json.Unmarshal(e.Data(), &payload); err != nil {
			return fmt.Errorf("TypedSubscriber[%s]: unmarshal fail: %w", s.id, err)
		}
	}

	return s.handler(ctx, e, payload)
}

// TypedEvent 是泛型事件包裝器，方便 Producer 發送
type TypedEvent[T any] struct {
	id         string
	topic      string
	occurredAt time.Time
	payload    T
	data       []byte
	marshalOnce sync.Once
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
	e.marshalOnce.Do(func() {
		// 優先檢查是否實作了 Marshaler 介面 (包含檢查指標)
		if m, ok := any(e.payload).(Marshaler); ok {
			if b, err := m.Marshal(); err == nil {
				e.data = b
				return
			}
		}
		if m, ok := any(&e.payload).(Marshaler); ok {
			if b, err := m.Marshal(); err == nil {
				e.data = b
				return
			}
		}
		// 回退到 JSON
		e.data, _ = json.Marshal(e.payload)
	})
	return e.data
}
