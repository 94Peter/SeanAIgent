package entity

import "time"

// 完整的預約狀態
type TrainDateHasApptState struct {
	StartDate         time.Time         `json:"start_date"`
	EndDate           time.Time         `json:"end_date"`
	ID                string            `json:"id"`
	Date              string            `json:"date"`
	Location          string            `json:"location"`
	Timezone          string            `json:"timezone"`
	UserAppointments  []UserAppointment `json:"user_appointments"`
	Capacity          int               `json:"capacity"`
	AvailableCapacity int               `json:"available_capacity"`
}

type UserAppointment struct {
	CreatedAt   time.Time `json:"created_at"`
	ID          string    `json:"id"`
	UserName    string    `json:"user_name"`
	UserID      string    `json:"user_id"`
	ChildName   string    `json:"child_name,omitempty"`
	IsCheckedIn bool      `json:"is_checked_in"`
	IsOnLeave   bool      `json:"is_on_leave"`
}

// 使用者角度的預約狀態
type TrainDateHasUserApptState struct {
	StartDate         time.Time         `json:"start_date"`
	EndDate           time.Time         `json:"end_date"`
	ID                string            `json:"id"`
	Date              string            `json:"date"`
	Location          string            `json:"location"`
	Timezone          string            `json:"timezone"`
	UserAppointments  []UserAppointment `json:"user_appointments"`
	AllUsers          []string          `json:"other_users"`
	Capacity          int               `json:"capacity"`
	AvailableCapacity int               `json:"available_capacity"`
}
