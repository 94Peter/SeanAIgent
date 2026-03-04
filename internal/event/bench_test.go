package event

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"testing"
)

// 一般 JSON Payload (觸發反射)
type JSONPayload struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// 自定義 Payload (不使用反射，手動處理相同資料)
type FastPayload struct {
	ID    int64
	Name  string
	Email string
}

func (p *FastPayload) Marshal() ([]byte, error) {
	// 手動序列化：[8 bytes ID][1 byte NameLen][Name][1 byte EmailLen][Email]
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, p.ID)
	
	buf.WriteByte(byte(len(p.Name)))
	buf.WriteString(p.Name)
	
	buf.WriteByte(byte(len(p.Email)))
	buf.WriteString(p.Email)
	
	return buf.Bytes(), nil
}

func (p *FastPayload) Unmarshal(data []byte) error {
	if len(data) < 10 {
		return fmt.Errorf("invalid data")
	}
	buf := bytes.NewReader(data)
	_ = binary.Read(buf, binary.LittleEndian, &p.ID)
	
	nameLen, _ := buf.ReadByte()
	nameBuf := make([]byte, nameLen)
	_, _ = buf.Read(nameBuf)
	p.Name = string(nameBuf)
	
	emailLen, _ := buf.ReadByte()
	emailBuf := make([]byte, emailLen)
	_, _ = buf.Read(emailBuf)
	p.Email = string(emailBuf)
	
	return nil
}

func BenchmarkEventSystem(b *testing.B) {
	topic := "bench.topic"

	b.Run("JSON_Reflection", func(b *testing.B) {
		payload := JSONPayload{ID: 12345678, Name: "Peter Parker", Email: "peter.parker@starkindustries.com"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			evt := NewTypedEvent("evt_id", topic, payload)
			_ = evt.Data() 

			var target JSONPayload
			_ = json.Unmarshal(evt.Data(), &target)
		}
	})

	b.Run("Custom_Manual_NoReflect", func(b *testing.B) {
		payload := FastPayload{ID: 12345678, Name: "Peter Parker", Email: "peter.parker@starkindustries.com"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			evt := NewTypedEvent("evt_id", topic, payload)
			_ = evt.Data() 

			var target FastPayload
			u := any(&target).(Unmarshaler)
			_ = u.Unmarshal(evt.Data()) 
		}
	})

	b.Run("Data_Cache_Hit", func(b *testing.B) {
		payload := JSONPayload{ID: 12345678, Name: "Peter Parker", Email: "peter.parker@starkindustries.com"}
		evt := NewTypedEvent("evt_id", topic, payload)
		_ = evt.Data()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = evt.Data()
		}
	})
}

func BenchmarkTypedSubscriberHandle(b *testing.B) {
	ctx := context.Background()
	
	b.Run("Handle_JSON", func(b *testing.B) {
		handler := func(ctx context.Context, e Event, p JSONPayload) error { return nil }
		sub := NewTypedSubscriber("sub_id", handler)
		payload := JSONPayload{ID: 12345678, Name: "Peter Parker", Email: "peter.parker@starkindustries.com"}
		evt := NewTypedEvent("evt_id", "topic", payload)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = sub.Handle(ctx, evt)
		}
	})

	b.Run("Handle_Custom_Manual", func(b *testing.B) {
		handler := func(ctx context.Context, e Event, p FastPayload) error { return nil }
		sub := NewTypedSubscriber("sub_id", handler)
		payload := FastPayload{ID: 12345678, Name: "Peter Parker", Email: "peter.parker@starkindustries.com"}
		evt := NewTypedEvent("evt_id", "topic", payload)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = sub.Handle(ctx, evt)
		}
	})
}
