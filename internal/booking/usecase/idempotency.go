package usecase

import (
	"sync"
	"time"
)

// IdempotencyManager 負責追蹤並防止重複處理相同的請求 Key
type IdempotencyManager interface {
	// CheckAndSet 檢查 Key 是否已存在，若不存在則設定並回傳 true，否則回傳 false
	CheckAndSet(key string) bool
	// Delete 移除 Key (通常用於發生錯誤需要重試時)
	Delete(key string)
}

type idempotencyManager struct {
	processedKeys sync.Map // key: string, value: time.Time (處理時間)
	ttl           time.Duration
}

func NewIdempotencyManager(ttl time.Duration) IdempotencyManager {
	m := &idempotencyManager{
		ttl: ttl,
	}
	// 定期清理過期的 Key
	go m.gc()
	return m
}

func (m *idempotencyManager) CheckAndSet(key string) bool {
	if key == "" {
		return true // 若無 Key，視為非冪等請求，直接通過 (由其他機制保護)
	}
	
	_, loaded := m.processedKeys.LoadOrStore(key, time.Now())
	return !loaded
}

func (m *idempotencyManager) Delete(key string) {
	if key == "" {
		return
	}
	m.processedKeys.Delete(key)
}

func (m *idempotencyManager) gc() {
	ticker := time.NewTicker(m.ttl / 2)
	for range ticker.C {
		now := time.Now()
		m.processedKeys.Range(func(key, value interface{}) bool {
			if t, ok := value.(time.Time); ok {
				if now.Sub(t) > m.ttl {
					m.processedKeys.Delete(key)
				}
			}
			return true
		})
	}
}
