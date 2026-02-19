package entity

import "time"

type AppointmentWithTrainDate struct {
	CreatedAt      time.Time   `json:"created_at"`
	LeaveInfo      LeaveInfoUI `json:"leave_info"`
	ID             string      `json:"id"`
	UserID         string      `json:"user_id"`
	UserName       string      `json:"user_name"`
	ChildName      string      `json:"child_name,omitempty"`
	TrainingDateId string      `json:"training_date_id"`
	TrainDate      TrainDateUI `json:"training_date_info"`
	Status         string      `json:"status"`
	IsOnLeave      bool        `json:"is_on_leave"`
	IsCheckedIn    bool        `json:"is_checked_in"`
}

type TrainDateUI struct {
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
	ID                string    `json:"id"`
	Date              string    `json:"date"`
	Location          string    `json:"location"`
	Timezone          string    `json:"timezone"`
	Capacity          int       `json:"capacity"`
	AvailableCapacity int       `json:"available_capacity"`
}

type LeaveInfoUI struct {
	CreatedAt time.Time `json:"created_at"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
}
