package event

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CustomPayload struct {
	Value string
}

func (p *CustomPayload) Marshal() ([]byte, error) {
	return []byte("CUSTOM:" + p.Value), nil
}

func (p *CustomPayload) Unmarshal(data []byte) error {
	s := string(data)
	if len(s) < 7 || s[:7] != "CUSTOM:" {
		return fmt.Errorf("invalid custom format: %s", s)
	}
	p.Value = s[7:]
	return nil
}

func TestCustomMarshalerUnmarshaler(t *testing.T) {
	memStore := &mockStore{
		events:   make(map[string][]Event),
		progress: make(map[string]string),
	}
	bus := NewBus(memStore)
	ctx := context.Background()
	topic := "custom.topic"

	// 1. Publish Custom Event
	payload := CustomPayload{Value: "Hello"}
	evt := NewTypedEvent("evt_custom", topic, payload)
	
	// Verify data is custom encoded
	data := evt.Data()
	assert.Equal(t, "CUSTOM:Hello", string(data))

	bus.Publish(ctx, evt)

	// 2. Subscribe with TypedSubscriber
	var received CustomPayload
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(ctx context.Context, e Event, p CustomPayload) error {
		defer wg.Done()
		received = p
		return nil
	}

	sub := NewTypedSubscriber("custom_sub", handler)
	bus.Subscribe(topic, sub)

	wg.Wait()

	// 3. Verify
	assert.Equal(t, "Hello", received.Value)
}

func TestCustomMarshalerUnmarshalerPointer(t *testing.T) {
	// This test checks if it works when T is a pointer type
	// Note: We might need to adjust implementation if it panics
	memStore := &mockStore{
		events:   make(map[string][]Event),
		progress: make(map[string]string),
	}
	bus := NewBus(memStore)
	ctx := context.Background()
	topic := "custom.pointer.topic"

	// 1. Publish Custom Event
	payload := &CustomPayload{Value: "Pointer"}
	evt := NewTypedEvent("evt_pointer", topic, payload)
	
	data := evt.Data()
	assert.Equal(t, "CUSTOM:Pointer", string(data))

	bus.Publish(ctx, evt)

	// 2. Subscribe with TypedSubscriber[*CustomPayload]
	var received *CustomPayload
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(ctx context.Context, e Event, p *CustomPayload) error {
		defer wg.Done()
		received = p
		return nil
	}

	sub := NewTypedSubscriber[*CustomPayload]("pointer_sub", handler)
	bus.Subscribe(topic, sub)

	wg.Wait()
	assert.NotNil(t, received)
	assert.Equal(t, "Pointer", received.Value)
}
