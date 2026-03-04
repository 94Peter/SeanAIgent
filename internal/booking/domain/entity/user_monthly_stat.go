package entity

import (
	"time"
)

type UserMonthlyStat struct {
	UserID           string    `json:"user_id" bson:"user_id"`
	UserName         string    `json:"user_name" bson:"user_name"`
	Year             int       `json:"year" bson:"year"`
	Month            int       `json:"month" bson:"month"`
	TotalBookings    int       `json:"total_bookings" bson:"total_bookings"`
	AttendedCount    int       `json:"attended_count" bson:"attended_count"`
	AbsentCount      int       `json:"absent_count" bson:"absent_count"`
	LeaveCount       int       `json:"leave_count" bson:"leave_count"`
	LastUpdatedAt    time.Time `json:"last_updated_at" bson:"last_updated_at"`
}

func NewUserMonthlyStat(userID, userName string, year, month int) *UserMonthlyStat {
	return &UserMonthlyStat{
		UserID:        userID,
		UserName:      userName,
		Year:          year,
		Month:         month,
		LastUpdatedAt: time.Now(),
	}
}
