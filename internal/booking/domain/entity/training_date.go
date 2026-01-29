package entity

import (
	"errors"
	"math"
	"time"
)

type TrainDateStatus string

const (
	TrainDateStatusActive   TrainDateStatus = "ACTIVE"
	TrainDateStatusInactive TrainDateStatus = "INACTIVE"
)

type trainDateOpt func(*TrainDate) error

func WithTrainDateID(id string) trainDateOpt {
	return func(td *TrainDate) error {
		td.id = id
		return nil
	}
}

func WithTrainDateUserID(userID string) trainDateOpt {
	return func(td *TrainDate) error {
		td.userID = userID
		return nil
	}
}

func WithTrainDateLocation(location string) trainDateOpt {
	return func(td *TrainDate) error {
		td.location = location
		return nil
	}
}

func WithTrainDateMaxCapacity(maxCapacity int) trainDateOpt {
	return func(td *TrainDate) error {
		td.maxCapacity = maxCapacity
		return nil
	}
}

func WithTrainDatePeriod(period TimeRange) trainDateOpt {
	return func(td *TrainDate) error {
		td.period = period
		return nil
	}
}

func WithTrainDateTimezone(timezone string) trainDateOpt {
	return func(td *TrainDate) error {
		td.timezone = timezone
		return nil
	}
}

func WithTrainDateAvailableCapacity(availableCapacity int) trainDateOpt {
	return func(td *TrainDate) error {
		td.availableCapacity = availableCapacity
		return nil
	}
}

func WithTrainDateCreatedAt(createdAt time.Time) trainDateOpt {
	return func(td *TrainDate) error {
		td.createdAt = createdAt
		return nil
	}
}

func WithTrainDateStatus(status TrainDateStatus) trainDateOpt {
	return func(td *TrainDate) error {
		td.status = status
		return nil
	}
}

func WithTrainDateUpdatedAt(updatedAt time.Time) trainDateOpt {
	return func(td *TrainDate) error {
		td.updatedAt = updatedAt
		return nil
	}
}

func WithBasicTrainDate(id, userID, location string, maxCapacity int, period TimeRange) trainDateOpt {
	return func(td *TrainDate) error {
		td.id = id
		td.userID = userID
		td.location = location
		td.maxCapacity = maxCapacity
		td.period = period
		td.availableCapacity = maxCapacity
		td.createdAt = time.Now()
		td.updatedAt = time.Now()
		td.timezone = period.start.Location().String()
		return nil
	}
}

func NewTrainDate(opts ...trainDateOpt) (*TrainDate, error) {
	td := &TrainDate{
		availableCapacity: math.MinInt8,
		status:            TrainDateStatusActive,
	}
	for _, opt := range opts {
		err := opt(td)
		if err != nil {
			return nil, err
		}
	}
	if td.availableCapacity == math.MinInt8 {
		return nil, ErrTrainingInvalidAvailableCapacity
	}
	return td, nil
}

type TrainDate struct {
	period            TimeRange
	createdAt         time.Time
	updatedAt         time.Time
	id                string
	userID            string
	location          string
	timezone          string
	status            TrainDateStatus
	maxCapacity       int
	availableCapacity int
}

func (s *TrainDate) Delete() error {
	if s.availableCapacity != s.maxCapacity {
		return ErrTrainingHasAppointments
	}
	s.status = TrainDateStatusInactive
	return nil
}

// ReserveSpot 預約名額 (關鍵行為)
func (s *TrainDate) ReserveSpot(count int) error {
	if count <= 0 {
		return ErrTrainingReserveCountInvalid
	}
	if s.availableCapacity < count {
		return ErrTrainingCapacityNotEnough
	}

	if time.Now().After(s.period.start) {
		return ErrTrainingOver
	}

	s.availableCapacity -= count
	return nil
}

// ReleaseSpot 歸還名額 (當取消預約或請假成功時)
func (s *TrainDate) ReleaseSpot(count int) error {
	if count <= 0 {
		return ErrTrainingReleaseCountInvalid
	}
	// 確保歸還後不會超過最大名額 (資料完整性檢查)
	if s.availableCapacity+count > s.maxCapacity {
		s.availableCapacity = s.maxCapacity
		return nil
	}

	s.availableCapacity += count
	return nil
}

// CanVerifyAttendance 檢查目前是否處於教練點名時間 (前10分鐘)
func (s *TrainDate) CanVerifyAttendance() bool {
	now := time.Now()
	// 前 10 分鐘
	openTime := s.period.start.Add(-10 * time.Minute)
	return now.After(openTime)
}

// IsFull 檢查是否滿額
func (s *TrainDate) IsFull() bool {
	return s.availableCapacity <= 0
}

func (p *TrainDate) Validate() error {
	return nil
}

func (p *TrainDate) StartDateWithTimeZone() (time.Time, error) {
	return toTime(p.period.start, p.location)
}

func (p *TrainDate) EndDateWithTimeZone() (time.Time, error) {
	return toTime(p.period.end, p.location)
}

func toTime(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, nil
	}
	return t.In(loc), nil
}

// Getter
func (p *TrainDate) ID() string {
	return p.id
}

func (p *TrainDate) UserID() string {
	return p.userID
}

func (p *TrainDate) Location() string {
	return p.location
}

func (p *TrainDate) MaxCapacity() int {
	return p.maxCapacity
}

func (p *TrainDate) AvailableCapacity() int {
	return p.availableCapacity
}

func (p *TrainDate) Period() TimeRange {
	return p.period
}

func (p *TrainDate) Timezone() string {
	return p.timezone
}

func (p *TrainDate) Status() TrainDateStatus {
	return p.status
}

func (p *TrainDate) CreatedAt() time.Time {
	return p.createdAt
}

func (p *TrainDate) UpdatedAt() time.Time {
	return p.updatedAt
}

// Error Definition
var (
	ErrTrainingInvalidAvailableCapacity = errors.New("TRAINING_INVALID_AVAILABLE_CAPACITY")
	ErrTrainingCoachBusy                = errors.New("TRAINING_COACH_BUSY")
	ErrTrainingInvalidTime              = errors.New("TRAINING_INVALID_TIME")
	ErrTrainingReserveCountInvalid      = errors.New("TRAINING_RESERVE_COUNT_INVALID")
	ErrTrainingCapacityNotEnough        = errors.New("TRAINING_CAPACITY_NOT_ENOUGH")
	ErrTrainingOver                     = errors.New("TRAINING_OVER")
	ErrTrainingReleaseCountInvalid      = errors.New("TRAINING_RELEASE_COUNT_INVALID")
	ErrTrainingNotFound                 = errors.New("TRAINING_NOT_FOUND")
	ErrTrainingHasAppointments          = errors.New("TRAINING_HAS_APPOINTMENTS")
)
