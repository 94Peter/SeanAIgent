package timeutil

import (
	"fmt"
	"sync"
	"time"
)

var locationCache sync.Map

// ToLocation 轉換 time.Time 到指定時區，並快取 Location 物件
func ToLocation(t time.Time, timezone string) time.Time {
	if timezone == "" {
		return t
	}
	loc, err := GetLocation(timezone)
	if err != nil {
		return t
	}
	return t.In(loc)
}

// GetLocation 獲取時區物件 (帶快取)
func GetLocation(timezone string) (*time.Location, error) {
	if val, ok := locationCache.Load(timezone); ok {
		return val.(*time.Location), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("load location %s fail: %w", timezone, err)
	}

	locationCache.Store(timezone, loc)
	return loc, nil
}

// ParseDateTime 解析日期與時間字串為指定時區的 time.Time
// date: "2006-01-02", timeStr: "15:04"
func ParseDateTime(date, timeStr, timezone string) (time.Time, error) {
	loc, err := GetLocation(timezone)
	if err != nil {
		loc = time.Local
	}
	return time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", date, timeStr), loc)
}
