package domain

import (
	"time"
)

const (
	TopicAppointmentStatusChanged = "booking.appointment.status_changed"
)

// AppointmentStatusChanged 預約狀態變更事件 Payload
type AppointmentStatusChanged struct {
	BookingID  string    `json:"booking_id"`
	UserID     string    `json:"user_id"`
	TrainingID string    `json:"training_id"`
	OldStatus  string    `json:"old_status"`
	NewStatus  string    `json:"new_status"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (e *AppointmentStatusChanged) Marshal() ([]byte, error) {
	return nil, nil // 使用預設 JSON
}

func (e *AppointmentStatusChanged) Unmarshal(data []byte) error {
	return nil // 使用預設 JSON
}
