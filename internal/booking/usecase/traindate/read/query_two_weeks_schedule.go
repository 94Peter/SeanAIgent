package read

import (
	"context"
	"fmt"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
	"sync"
	"time"
)

type ReqQueryTwoWeeksSchedule struct {
	UserID        string
	ReferenceDate time.Time
	Direction     string // "next" or "prev"
}

type WeekVO struct {
	ID        string   `json:"id"`
	Days      []*DayVO `json:"days"`
	IsCurrent bool     `json:"is_current"`
}

type DayVO struct {
	FullDate    string    `json:"full_date"`
	DateDisplay string    `json:"date_display"`
	DayOfWeek   string    `json:"day_of_week"`
	IsToday     bool      `json:"is_today"`
	Slots       []*SlotVO `json:"slots"`
}

type SlotVO struct {
	ID             string        `json:"id"`
	TimeDisplay    string        `json:"time_display"`
	EndTimeDisplay string        `json:"end_time_display"`
	CourseName     string        `json:"course_name"`
	Location       string        `json:"location"`
	Capacity       int           `json:"capacity"`
	BookedCount    int           `json:"booked_count"`
	Attendees      []*AttendeeVO `json:"attendees"`
	EndDate        time.Time     `json:"end_date"`
	IsFull         bool          `json:"is_full"`
	IsEmpty        bool          `json:"is_empty"`
}

type AttendeeVO struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // "Booked", "Leave"
	BookingID   string    `json:"booking_id"`
	BookingTime time.Time `json:"booking_time"`
}

type QueryTwoWeeksScheduleUseCase core.ReadUseCase[ReqQueryTwoWeeksSchedule, []*WeekVO]

type queryTwoWeeksScheduleUseCase struct {
	repo repository.TrainRepository
}

func NewQueryTwoWeeksScheduleUseCase(repo repository.TrainRepository) QueryTwoWeeksScheduleUseCase {
	return &queryTwoWeeksScheduleUseCase{repo: repo}
}

func (uc *queryTwoWeeksScheduleUseCase) Name() string {
	return "QueryTwoWeeksSchedule"
}

func (uc *queryTwoWeeksScheduleUseCase) Execute(
	ctx context.Context, req ReqQueryTwoWeeksSchedule,
) ([]*WeekVO, core.UseCaseError) {
	var start, end time.Time
	if req.Direction == "prev" {
		end = req.ReferenceDate.AddDate(0, 0, -1)
		start = end.AddDate(0, 0, -13)
	} else {
		start = req.ReferenceDate.AddDate(0, 0, 1)
		end = start.AddDate(0, 0, 13)
	}

	// 取得這段時間的所有課程
	trainDates, err := uc.repo.UserQueryTrainDateHasApptState(
		ctx, req.UserID, repository.NewFilterTrainDataByTimeRange(start, end),
	)
	if err != nil {
		return nil, core.NewDBError("QUERY_TWO_WEEKS", "QUERY_FAIL", err.Error(), core.ErrInternal)
	}

	return groupToWeeks(trainDates, start, end), nil
}

func groupToWeeks(data []*entity.TrainDateHasUserApptState, start, end time.Time) []*WeekVO {
	weeksMap := make(map[string]*WeekVO)
	dateToDayMap := make(map[string]*DayVO)
	var weekIDs []string

	// 初始化所有日期，確保沒課的天數也會顯示
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		year, week := d.ISOWeek()
		weekID := fmt.Sprintf("%d-W%02d", year, week)
		if _, ok := weeksMap[weekID]; !ok {
			weeksMap[weekID] = &WeekVO{ID: weekID, Days: make([]*DayVO, 0)}
			weekIDs = append(weekIDs, weekID)
		}
		day := &DayVO{
			FullDate:    d.Format("2006-01-02"),
			DateDisplay: fmt.Sprintf("%d", d.Day()),
			DayOfWeek:   d.Format("Mon"),
			IsToday:     d.Format("2006-01-02") == time.Now().Format("2006-01-02"),
			Slots:       []*SlotVO{},
		}
		weeksMap[weekID].Days = append(weeksMap[weekID].Days, day)
		dateToDayMap[day.FullDate] = day
	}

	// 填入實際課程
	for _, td := range data {
		fullDate := td.StartDate.Format("2006-01-02")
		if day, ok := dateToDayMap[fullDate]; ok {
			loc := getLocation(td.Timezone)
			startDate := td.StartDate.In(loc)
			endDate := td.EndDate.In(loc)
			day.Slots = append(day.Slots, &SlotVO{
				ID:             td.ID,
				TimeDisplay:    startDate.Format("15:04"),
				EndTimeDisplay: endDate.Format("15:04"),
				CourseName:     "@" + td.Location,
				Location:       td.Location,
				Capacity:       td.Capacity,
				BookedCount:    td.Capacity - td.AvailableCapacity,
				IsFull:         td.AvailableCapacity <= 0,
				Attendees:      transformAttendees(td.UserAppointments, td.EndDate),
				EndDate:        td.EndDate,
			})
		}
	}

	// 填充空白 Slot
	for _, w := range weeksMap {
		for _, d := range w.Days {
			if len(d.Slots) == 0 {
				d.Slots = append(d.Slots, &SlotVO{IsEmpty: true})
			}
		}
	}

	res := make([]*WeekVO, 0)
	for _, id := range weekIDs {
		res = append(res, weeksMap[id])
	}
	return res
}

func transformAttendees(appts []entity.UserAppointment, courseEndDate time.Time) []*AttendeeVO {
	res := make([]*AttendeeVO, 0)
	now := time.Now()
	for _, a := range appts {
		status := "Booked"
		if a.IsOnLeave {
			status = "Leave"
		} else if a.IsCheckedIn {
			status = "CheckedIn"
		} else if now.After(courseEndDate) {
			status = "Absent"
		}
		res = append(res, &AttendeeVO{
			Name: a.ChildName, Status: status, BookingID: a.ID, BookingTime: a.CreatedAt,
		})
	}
	return res
}

var (
	locationCache   = make(map[string]*time.Location)
	locationCacheMu sync.RWMutex
)

func getLocation(location string) *time.Location {
	locationCacheMu.RLock()
	loc, ok := locationCache[location]
	locationCacheMu.RUnlock()
	if ok {
		return loc
	}

	loc, err := time.LoadLocation(location)
	if err != nil {
		return time.UTC
	}

	locationCacheMu.Lock()
	locationCache[location] = loc
	locationCacheMu.Unlock()
	return loc
}
