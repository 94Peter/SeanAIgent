package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase"
	readappt "seanAIgent/internal/booking/usecase/appointment/read"
	writeappt "seanAIgent/internal/booking/usecase/appointment/write"
	readstats "seanAIgent/internal/booking/usecase/stats/read"
	readtrain "seanAIgent/internal/booking/usecase/traindate/read"
	"seanAIgent/internal/service/lineliff"
	"seanAIgent/templates"
	"seanAIgent/templates/forms/booking_v2"

	"github.com/94peter/vulpes/ezapi"
	"github.com/gin-gonic/gin"
)

func NewV2BookingUseCaseSet(registry *usecase.Registry) V2BookingUseCaseSet {
	return V2BookingUseCaseSet{
		createApptUC:            registry.CreateAppt,
		cancelApptUC:            registry.CancelAppt,
		getUserMonthlyStatsUC:   registry.GetUserMonthlyStats,
		queryTwoWeeksScheduleUC: registry.QueryTwoWeeksSchedule,
		queryUserBookingsUC:     registry.QueryUserBookings,
	}
}

type V2BookingUseCaseSet struct {
	createApptUC            writeappt.CreateApptUseCase
	cancelApptUC            writeappt.CancelApptUseCase
	getUserMonthlyStatsUC   readstats.GetUserMonthlyStatsUseCase
	queryTwoWeeksScheduleUC readtrain.QueryTwoWeeksScheduleUseCase
	queryUserBookingsUC     readappt.QueryUserBookingsUseCase
}

func NewV2BookingApi(enableCSRF bool, bookingUseCaseSet V2BookingUseCaseSet) WebAPI {
	return &v2BookingAPI{
		V2BookingUseCaseSet: bookingUseCaseSet,
		enableCSRF:          enableCSRF,
	}
}

type v2BookingAPI struct {
	V2BookingUseCaseSet
	enableCSRF bool
	once       sync.Once
}

func (api *v2BookingAPI) InitRouter(r ezapi.Router) {
	api.once.Do(func() {
		// v2 booking
		r.GET("/v2/training/booking", api.getBookingV2Form)
		r.GET("/:lang/v2/training/booking", api.getBookingV2Form)
		// v2 api
		r.POST("/api/v2/bookings", api.createBookingV2)
		r.DELETE("/api/v2/bookings/:bookingId", api.cancelBookingV2)
		r.POST("/api/v2/bookings/:bookingId/leave", api.submitLeaveV2)
		r.DELETE("/api/v2/bookings/:bookingId/leave", api.cancelLeaveV2)
		r.GET("/api/v2/my-bookings", api.getMyBookingsV2)
		r.GET("/api/v2/calendar/weeks", api.getCalendarWeeksV2)
		r.GET("/api/v2/calendar/stats", api.getCalendarStatsV2)
	})
}

func (api *v2BookingAPI) getCalendarStatsV2(c *gin.Context) {
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	var year, month int
	fmt.Sscanf(yearStr, "%d", &year)
	fmt.Sscanf(monthStr, "%d", &month)

	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not logged in"})
		return
	}

	stats, err := api.getUserMonthlyStatsUC.Execute(c.Request.Context(), readstats.ReqGetUserMonthlyStats{
		UserID: userID,
		Year:   year,
		Month:  month,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// getBookingV2Form renders the V2 Booking Page with real data
func (api *v2BookingAPI) getBookingV2Form(c *gin.Context) {
	lineliffid := lineliff.GetBookingV2LiffId()
	userID := getUserID(c)

	ctx := c.Request.Context()
	now := time.Now()
	var err error
	var stats *readstats.UserMonthlyStatsVO
	// 1. Fetch current month stats
	if userID != "" {
		stats, err = api.getUserMonthlyStatsUC.Execute(ctx, readstats.ReqGetUserMonthlyStats{
			UserID: userID,
			Year:   now.Year(),
			Month:  int(now.Month()),
		})
		if err != nil {
			c.Error(err)
			return
		}
	} else {
		stats = &readstats.UserMonthlyStatsVO{Children: []*readstats.ChildStatVO{}}
	}

	// 2. Fetch initial 2 weeks schedule (starting from this Monday)
	// Calculate this week's Monday, then subtract 1 day to get last Sunday as ReferenceDate
	offset := (int(now.Weekday()) + 6) % 7
	thisMonday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -offset)
	lastSunday := thisMonday.AddDate(0, 0, -1)

	weeks, err := api.queryTwoWeeksScheduleUC.Execute(ctx, readtrain.ReqQueryTwoWeeksSchedule{
		UserID:        userID,
		ReferenceDate: lastSunday,
		Direction:     "next",
	})
	if err != nil {
		c.Error(err)
		return
	}

	// 3. Fetch current user bookings
	userBookings, err := api.queryUserBookingsUC.Execute(ctx, readappt.ReqQueryUserBookings{
		UserID:         userID,
		TrainDateAfter: now,
	})
	if err != nil {
		c.Error(err)
		return
	}

	// Transform data to ViewModel
	modelV2 := &booking_v2.BookingV2Model{
		LiffID: lineliffid,
		CurrentUser: &booking_v2.UserContext{
			DisplayName:      getUserDisplayName(c),
			UserID:           userID,
			FrequentChildren: statsToFrequentChildren(stats),
		},
		Stats:      statsToStatsSummary(stats),
		MyBookings: apptsToMyBookingsV2(userBookings.Appts),
		Weeks:      transformToWeeksVO(weeks),
	}

	com := templates.Layout(
		booking_v2.Page(modelV2),
		lineliffid,
		&templates.OgMeta{
			Title:       "Sean AIgent V2",
			Description: "Sean 的新版預課服務",
			Image:       "https://storage.94peter.dev/cdn-cgi/image/width=1200,height=630,quality=80,format=auto/https://storage.94peter.dev/images/UAC.png",
		},
	)
	r := newTemplRenderer(ctx, http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func statsToFrequentChildren(stats *readstats.UserMonthlyStatsVO) []string {
	res := make([]string, 0)
	for _, child := range stats.Children {
		res = append(res, child.Name)
	}
	return res
}

func statsToStatsSummary(stats *readstats.UserMonthlyStatsVO) *booking_v2.StatsSummary {
	children := make([]*booking_v2.ChildStat, 0)
	for _, child := range stats.Children {
		children = append(children, &booking_v2.ChildStat{
			Name:      child.Name,
			Completed: child.Completed,
			Leave:     child.Leave,
			Absent:    child.Absent,
			AvgWeek:   child.AvgWeek,
		})
	}
	return &booking_v2.StatsSummary{
		TotalSessions: stats.TotalSessions,
		TotalLeave:    stats.TotalLeave,
		Children:      children,
	}
}

func apptsToMyBookingsV2(appts []*entity.AppointmentWithTrainDate) []*booking_v2.MyBookingItem {
	groups := make(map[string]*booking_v2.MyBookingItem)
	order := make([]string, 0)

	for _, appt := range appts {
		key := appt.TrainingDateId
		if _, ok := groups[key]; !ok {
			startDate := appt.TrainDate.StartDate.In(taipeiLoc)
			groups[key] = &booking_v2.MyBookingItem{
				ID: key,
				DateDisplay: fmt.Sprintf("%s (%s) %s",
					startDate.Format("01/02"),
					startDate.Format("Mon"),
					startDate.Format("15:04")),
				Title:     appt.TrainDate.Location,
				Attendees: make([]*booking_v2.Attendee, 0),
			}
			order = append(order, key)
		}
		groups[key].Attendees = append(groups[key].Attendees, &booking_v2.Attendee{
			Name:        appt.ChildName,
			Status:      uiApptStatusTransform(appt),
			BookingTime: appt.CreatedAt,
			BookingID:   appt.ID,
		})
	}

	res := make([]*booking_v2.MyBookingItem, 0)
	for _, key := range order {
		res = append(res, groups[key])
	}
	return res
}

func uiApptStatusTransform(appt *entity.AppointmentWithTrainDate) string {
	if appt.IsOnLeave {
		return "Leave"
	}
	if appt.IsCheckedIn {
		return "CheckedIn"
	}
	if appt.Status == string(entity.StatusAttended) {
		return "CheckedIn"
	}
	// If the course has ended and user didn't check in or leave, it's Absent
	if time.Now().After(appt.TrainDate.EndDate) {
		return "Absent"
	}
	return "Booked"
}

func transformToWeeksVO(weeks []*readtrain.WeekVO) []*booking_v2.WeekData {
	res := make([]*booking_v2.WeekData, 0, len(weeks))
	for _, w := range weeks {
		days := make([]*booking_v2.DayData, 0, len(w.Days))
		for _, d := range w.Days {
			slots := make([]*booking_v2.SlotData, 0, len(d.Slots))
			for _, s := range d.Slots {
				attendees := make([]*booking_v2.Attendee, 0, len(s.Attendees))
				for _, a := range s.Attendees {
					attendees = append(attendees, &booking_v2.Attendee{
						Name:        a.Name,
						Status:      a.Status,
						BookingID:   a.BookingID,
						BookingTime: a.BookingTime,
					})
				}
				slots = append(slots, &booking_v2.SlotData{
					ID:             s.ID,
					TimeDisplay:    s.TimeDisplay,
					EndTimeDisplay: s.EndTimeDisplay,
					CourseName:     s.CourseName,
					Location:       s.Location,
					Capacity:       s.Capacity,
					BookedCount:    s.BookedCount,
					Attendees:      attendees,
					IsFull:         s.IsFull,
					IsEmpty:        s.IsEmpty,
					IsPast:         time.Now().After(s.EndDate),
				})
			}
			days = append(days, &booking_v2.DayData{
				DateDisplay: d.DateDisplay,
				DayOfWeek:   d.DayOfWeek,
				FullDate:    d.FullDate,
				IsToday:     d.IsToday,
				Slots:       slots,
			})
		}
		res = append(res, &booking_v2.WeekData{
			ID:        w.ID,
			Days:      days,
			IsCurrent: w.IsCurrent,
		})
	}
	return res
}

func (api *v2BookingAPI) createBookingV2(c *gin.Context) {
	var req struct {
		SlotID       string   `json:"slot_id"`
		StudentNames []string `json:"student_names"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	userId := getUserID(c)
	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not logged in"})
		return
	}

	userName := getUserDisplayName(c)
	domainUser, err := entity.NewUser(userId, userName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create user"})
		return
	}
	appts, errUC := api.createApptUC.Execute(c.Request.Context(), writeappt.ReqCreateAppt{
		TrainDateID: req.SlotID,
		User:        domainUser,
		ChildNames:  req.StudentNames,
	})
	if errUC != nil {
		c.JSON(getStatus(errUC.Type()), gin.H{"success": false, "message": errUC.Message()})
		return
	}

	newBookings := make([]gin.H, 0)
	for _, appt := range appts {
		newBookings = append(newBookings, gin.H{
			"booking_id":   appt.ID(),
			"name":         appt.ChildName(),
			"status":       uiBookingStatusTransform(appt),
			"booking_time": time.Now().Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "預約成功",
		"new_bookings": newBookings,
	})
}

func uiBookingStatusTransform(appt *entity.Appointment) string {
	switch appt.Status() {
	case entity.StatusConfirmed:
		return "Booked"
	case entity.StatusCancelledLeave:
		return "Leave"
	default:
		return ""

	}
}

func (api *v2BookingAPI) cancelBookingV2(c *gin.Context) {
	bookingID := c.Param("bookingId")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Booking ID required"})
		return
	}
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not logged in"})
		return
	}

	_, errUC := api.cancelApptUC.Execute(c.Request.Context(), writeappt.ReqCancelAppt{
		ApptID: bookingID,
		UserID: userID,
	})
	if errUC != nil {
		c.JSON(getStatus(errUC.Type()), gin.H{"success": false, "message": errUC.Message()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "預約已取消",
	})
}

func (api *v2BookingAPI) submitLeaveV2(c *gin.Context) {
	bookingID := c.Param("bookingId")
	var req struct {
		Reason string `json:"reason"`
	}
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Booking ID required"})
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "請假申請已送出",
	})
}

func (api *v2BookingAPI) cancelLeaveV2(c *gin.Context) {
	bookingID := c.Param("bookingId")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Booking ID required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "已取消請假並恢復預約",
		"current_status": "Booked",
	})
}

func (api *v2BookingAPI) getMyBookingsV2(c *gin.Context) {
	listType := c.Query("type") // "upcoming" or "history"
	userID := getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not logged in"})
		return
	}

	var trainDateAfter time.Time
	if listType == "history" {
		trainDateAfter = time.Time{} // All time, but we'll need to filter for past dates in a real repo query if needed
	} else {
		trainDateAfter = time.Now()
	}

	resp, errUC := api.queryUserBookingsUC.Execute(c.Request.Context(), readappt.ReqQueryUserBookings{
		UserID:         userID,
		TrainDateAfter: trainDateAfter,
	})
	if errUC != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": errUC.Error()})
		return
	}

	// For history, we only want past courses. For upcoming, we only want future courses.
	// The UC currently only supports TrainDateAfter.
	// We might need to filter manually here for "history".
	now := time.Now()
	filteredAppts := make([]*entity.AppointmentWithTrainDate, 0)
	for _, appt := range resp.Appts {
		if listType == "history" {
			if appt.TrainDate.EndDate.Before(now) {
				filteredAppts = append(filteredAppts, appt)
			}
		} else {
			if appt.TrainDate.EndDate.After(now) {
				filteredAppts = append(filteredAppts, appt)
			}
		}
	}

	items := apptsToMyBookingsV2(filteredAppts)

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"next_cursor": resp.Cursor,
		"has_more":    resp.Cursor != "",
	})
}

var taipeiLoc = time.FixedZone("Asia/Taipei", 8*60*60)

func (api *v2BookingAPI) getCalendarWeeksV2(c *gin.Context) {
	startDateStr := c.Query("start_date")
	direction := c.Query("direction") // "next" or "prev"

	refDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		refDate = time.Now()
	}
	refDate = time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, taipeiLoc)

	// Add date range restrictions
	now := time.Now().In(taipeiLoc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, taipeiLoc)
	limitPrev := today.AddDate(0, -3, 0) // 3 months ago
	limitNext := today.AddDate(0, 6, 0)  // 6 months from now

	if direction == "prev" && !refDate.After(limitPrev) {
		c.JSON(http.StatusOK, gin.H{"weeks": []*booking_v2.WeekData{}})
		return
	}
	if direction == "next" && !refDate.Before(limitNext) {
		c.JSON(http.StatusOK, gin.H{"weeks": []*booking_v2.WeekData{}})
		return
	}

	ctx := c.Request.Context()
	userID := getUserID(c)
	fmt.Println("getCalendarWeeksV2", startDateStr, direction, refDate)
	weeks, errUC := api.queryTwoWeeksScheduleUC.Execute(ctx, readtrain.ReqQueryTwoWeeksSchedule{
		UserID:        userID,
		ReferenceDate: refDate,
		Direction:     direction,
	})
	fmt.Println("getCalendarWeeksV2", weeks)
	if errUC != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": errUC.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"weeks": transformToWeeksVO(weeks),
	})
}
