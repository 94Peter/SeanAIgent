package domain

import (
	"time"
)

const (
	TopicAppointmentStatusChanged = "booking.appointment.status_changed"
	TopicUserStatsRefreshRequested = "booking.stats.refresh_requested"
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

// UserStatsRefreshRequested 用於大批次更新後，請求重新計算特定用戶在特定月份的統計
type UserStatsRefreshRequested struct {
	UserID     string    `json:"user_id"`
	Year       int       `json:"year"`
	Month      int       `json:"month"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (e *UserStatsRefreshRequested) Marshal() ([]byte, error) {
	return nil, nil
}

func (e *UserStatsRefreshRequested) Unmarshal(data []byte) error {
	return nil
}
