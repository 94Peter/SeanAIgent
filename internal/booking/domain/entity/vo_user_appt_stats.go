package entity

import (
	"time"
)

type UserApptStats struct {
	UserID           string        `json:"user_id"`
	UserName         string        `json:"user_name"`
	ChildState       []*childState `json:"child_state"`
	CheckedInCount   int           `json:"checked_in_count"`
	OnLeaveCount     int           `json:"on_leave_count"`
	TotalAppointment int           `json:"total_appointment"`
}

type childState struct {
	ChildName      string             `json:"child_name"`
	Appointments   []*appointmentInfo `json:"appointments"`
	CheckedInCount int                `json:"checked_in_count"`
	OnLeaveCount   int                `json:"on_leave_count"`
}

type appointmentInfo struct {
	AppointmentDate time.Time `json:"appointment_date"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Location        string    `json:"location"`
	Timezone        string    `json:"timezone"`
	Capacity        int       `json:"capacity"`
	IsCheckedIn     bool      `json:"is_checked_in"`
	IsOnLeave       bool      `json:"is_on_leave"`
}
