package entity

import "time"

type TimeRange struct {
	start time.Time
	end   time.Time
}

func NewTimeRange(start, end time.Time) (TimeRange, error) {
	if end.Before(start) {
		return TimeRange{}, ErrTrainingInvalidTime
	}
	return TimeRange{start, end}, nil
}

func (t TimeRange) Start() time.Time {
	return t.start
}

func (t TimeRange) End() time.Time {
	return t.end
}
