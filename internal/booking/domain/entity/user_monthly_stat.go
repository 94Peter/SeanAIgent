package entity

import (
	"errors"
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

func (s *UserMonthlyStat) Validate() error {
	if s.UserID == "" {
		return errors.New("user_id is required")
	}
	if s.Year < 2024 || s.Month < 1 || s.Month > 12 {
		return errors.New("invalid year or month")
	}
	if s.TotalBookings < 0 || s.AttendedCount < 0 || s.AbsentCount < 0 || s.LeaveCount < 0 {
		return errors.New("counters cannot be negative")
	}
	// 邏輯檢查：總數不應小於各項分類總和
	if s.TotalBookings < (s.AttendedCount + s.AbsentCount + s.LeaveCount) {
		return errors.New("total bookings mismatch with status counters")
	}
	return nil
}
