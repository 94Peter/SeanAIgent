package event

import (
	"context"
	"log"
	"sync"
)

type internalBus struct {
	store       EventStore
	subscribers map[string][]Subscriber
	mu          sync.RWMutex
}

func NewBus(store EventStore) Bus {
	return &internalBus{
		store:       store,
		subscribers: make(map[string][]Subscriber),
	}
}

func (b *internalBus) Subscribe(topic string, s Subscriber) {
	b.mu.Lock()
	b.subscribers[topic] = append(b.subscribers[topic], s)
	b.mu.Unlock()

	// 啟動追趕機制：找出該訂閱者漏掉的歷史事件
	go b.catchUp(context.Background(), topic, s)
}

func (b *internalBus) catchUp(ctx context.Context, topic string, s Subscriber) {
	unprocessed, err := b.store.FindUnprocessedEvents(ctx, s.ID(), topic)
	if err != nil {
		log.Printf("EventBus: catch-up error for %s: %v", s.ID(), err)
		return
	}

	for _, e := range unprocessed {
		b.handleEvent(ctx, s, e)
	}
}

func (b *internalBus) Publish(ctx context.Context, e Event) {
	// 1. 持久化事件 (稽核與防丟)
	if err := b.store.Save(ctx, e); err != nil {
		log.Printf("EventBus: fail to save event %s: %v", e.ID(), err)
		// 即使存檔失敗也繼續嘗試發送即時訊息，但會紀錄錯誤
	}

	b.mu.RLock()
	subs, ok := b.subscribers[e.Topic()]
	b.mu.RUnlock()

	if !ok {
		return
	}

	for _, s := range subs {
		go b.handleEvent(ctx, s, e)
	}
}

func (b *internalBus) handleEvent(ctx context.Context, s Subscriber, e Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("EventBus: subscriber %s panicked: %v", s.ID(), r)
		}
	}()

	if err := s.Handle(ctx, e); err != nil {
		log.Printf("EventBus: subscriber %s handle error: %v", s.ID(), err)
		return
	}

	// 處理成功後，更新進度
	if err := b.store.UpdateProgress(ctx, s.ID(), e.ID()); err != nil {
		log.Printf("EventBus: fail to update progress for %s: %v", s.ID(), err)
	}
}
