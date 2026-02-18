package read

import (
	"context"
	"fmt"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
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
	ID          string        `json:"id"`
	TimeDisplay string        `json:"time_display"`
	CourseName  string        `json:"course_name"`
	Location    string        `json:"location"`
	Capacity    int           `json:"capacity"`
	BookedCount int           `json:"booked_count"`
	Attendees   []*AttendeeVO `json:"attendees"`
	IsFull      bool          `json:"is_full"`
	IsEmpty     bool          `json:"is_empty"`
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

func (uc *queryTwoWeeksScheduleUseCase) Execute(ctx context.Context, req ReqQueryTwoWeeksSchedule) ([]*WeekVO, core.UseCaseError) {
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
	var weekIDs []string

	// 初始化所有日期，確保沒課的天數也會顯示
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		year, week := d.ISOWeek()
		weekID := fmt.Sprintf("%d-W%02d", year, week)
		if _, ok := weeksMap[weekID]; !ok {
			weeksMap[weekID] = &WeekVO{ID: weekID, Days: make([]*DayVO, 0)}
			weekIDs = append(weekIDs, weekID)
		}
		weeksMap[weekID].Days = append(weeksMap[weekID].Days, &DayVO{
			FullDate:    d.Format("2006-01-02"),
			DateDisplay: fmt.Sprintf("%d", d.Day()),
			DayOfWeek:   d.Format("Mon"),
			IsToday:     d.Format("2006-01-02") == time.Now().Format("2006-01-02"),
			Slots:       []*SlotVO{},
		})
	}

	// 填入實際課程
	for _, td := range data {
		year, week := td.StartDate.ISOWeek()
		weekID := fmt.Sprintf("%d-W%02d", year, week)
		if w, ok := weeksMap[weekID]; ok {
			for _, day := range w.Days {
				if day.FullDate == td.StartDate.Format("2006-01-02") {
					day.Slots = append(day.Slots, &SlotVO{
						ID:          td.ID,
						TimeDisplay: td.StartDate.Format("15:04"),
						CourseName:  td.Date, // 暫用 Date 欄位
						Location:    td.Location,
						Capacity:    td.Capacity,
						BookedCount: td.Capacity - td.AvailableCapacity,
						IsFull:      td.AvailableCapacity <= 0,
						Attendees:   transformAttendees(td.UserAppointments),
					})
				}
			}
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

func transformAttendees(appts []entity.UserAppointment) []*AttendeeVO {
	res := make([]*AttendeeVO, 0)
	for _, a := range appts {
		status := "Booked"
		if a.IsOnLeave {
			status = "Leave"
		}
		res = append(res, &AttendeeVO{
			Name: a.ChildName, Status: status, BookingID: a.ID, BookingTime: a.CreatedAt,
		})
	}
	return res
}
