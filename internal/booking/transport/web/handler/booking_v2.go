package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase"
	writeappt "seanAIgent/internal/booking/usecase/appointment/write"
	"seanAIgent/internal/service/lineliff"
	"seanAIgent/templates"
	"seanAIgent/templates/forms/booking_v2"

	"github.com/94peter/vulpes/ezapi"
	"github.com/gin-gonic/gin"
)

func NewV2BookingUseCaseSet(registry *usecase.Registry) V2BookingUseCaseSet {
	return V2BookingUseCaseSet{
		createApptUC: registry.CreateAppt,
		cancelApptUC: registry.CancelAppt,
	}
}

type V2BookingUseCaseSet struct {
	createApptUC writeappt.CreateApptUseCase
	cancelApptUC writeappt.CancelApptUseCase
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
	year := c.Query("year")
	month := c.Query("month")

	// Mock data based on the provided year/month
	// In a real application, you would calculate stats for the specific month.
	c.JSON(http.StatusOK, gin.H{
		"total_sessions": 10, // Dummy monthly value
		"total_leave":    1,
		"children": []gin.H{
			{"name": "小明", "completed": 6, "leave": 0, "absent": 0, "avg_week": 1.5},
			{"name": "小華", "completed": 4, "leave": 1, "absent": 0, "avg_week": 1.0},
			{"name": "小東", "completed": 2, "leave": 0, "absent": 0, "avg_week": 0.5},
		},
		"year":  year,
		"month": month,
	})
}

// getBookingV2Form renders the V2 Booking Page with mock data
func (api *v2BookingAPI) getBookingV2Form(c *gin.Context) {
	lineliffid := lineliff.GetBookingLiffId()
	displayName := getUserDisplayName(c)
	if displayName == "" {
		displayName = "測試家長"
	}

	// --- Generate Dummy Data for V2 ---
	modelV2 := &booking_v2.BookingV2Model{
		LiffID: lineliffid,
		CurrentUser: &booking_v2.UserContext{
			DisplayName:      displayName,
			UserID:           getUserID(c),
			FrequentChildren: []string{"小明", "小華", "小東"},
		},
		Stats: &booking_v2.StatsSummary{
			TotalSessions: 48,
			TotalLeave:    5,
			Children: []*booking_v2.ChildStat{
				{Name: "小明", Completed: 24, Leave: 2, Absent: 1, AvgWeek: 1.9},
				{Name: "小華", Completed: 18, Leave: 4, Absent: 0, AvgWeek: 1.4},
				{Name: "小東", Completed: 6, Leave: 0, Absent: 1, AvgWeek: 0.5},
			},
		},
		MyBookings: []*booking_v2.MyBookingItem{
			{
				ID:          "booking-1",
				DateDisplay: "02/14 (六) 14:30",
				Title:       "師大班 @ 師大附中",
				Attendees: []*booking_v2.Attendee{
					{Name: "小明", Status: "Booked", BookingTime: time.Now(), BookingID: "101"},
					{Name: "小華", Status: "Leave", BookingTime: time.Now().Add(-48 * time.Hour), BookingID: "102"},
				},
			},
		},
		Weeks: []*booking_v2.WeekData{
			{
				ID:        "2026-W07",
				IsCurrent: true,
				Days: []*booking_v2.DayData{
					{DateDisplay: "8", DayOfWeek: "Sun", IsToday: false, FullDate: "2026-02-08", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "9", DayOfWeek: "Mon", IsToday: false, FullDate: "2026-02-09", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "10", DayOfWeek: "Tue", IsToday: false, FullDate: "2026-02-10", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "11", DayOfWeek: "Wed", IsToday: false, FullDate: "2026-02-11", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "12", DayOfWeek: "Thu", IsToday: false, FullDate: "2026-02-12", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{
						DateDisplay: "13",
						DayOfWeek:   "Fri",
						FullDate:    "2026-02-13",
						IsToday:     true,
						Slots: []*booking_v2.SlotData{
							{
								ID:          "slot-1",
								TimeDisplay: "16:30",
								CourseName:  "足球進階班",
								Location:    "北安球場",
								BookedCount: 12,
								Capacity:    20,
								Attendees: []*booking_v2.Attendee{
									{Name: "小明", Status: "Booked",
										BookingTime: time.Now(), BookingID: "1"},
									{Name: "小華", Status: "Booked",
										BookingTime: time.Now().Add(-time.Hour * 48), BookingID: "2"},
								},
							},
							{
								ID:          "slot-2",
								TimeDisplay: "18:30",
								CourseName:  "體能開發",
								Location:    "北安球場",
								IsEmpty:     false,
								BookedCount: 5,
								Capacity:    10,
								Attendees: []*booking_v2.Attendee{
									{Name: "小明", Status: "Leave"},
								},
							},
						},
					},
					{DateDisplay: "14", DayOfWeek: "Sat", IsToday: false, FullDate: "2026-02-14", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
				},
			},
			{
				ID:        "2026-W08",
				IsCurrent: false,
				Days: []*booking_v2.DayData{
					{DateDisplay: "15", DayOfWeek: "Sun", IsToday: false, FullDate: "2026-02-15", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "16", DayOfWeek: "Mon", IsToday: false, FullDate: "2026-02-16", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "17", DayOfWeek: "Tue", IsToday: false, FullDate: "2026-02-17", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "18", DayOfWeek: "Wed", IsToday: false, FullDate: "2026-02-18", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "19", DayOfWeek: "Thu", IsToday: false, FullDate: "2026-02-19", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "20", DayOfWeek: "Fri", IsToday: false, FullDate: "2026-02-20", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
					{DateDisplay: "21", DayOfWeek: "Sat", IsToday: false, FullDate: "2026-02-21", Slots: []*booking_v2.SlotData{{IsEmpty: true}}},
				},
			},
		},
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
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

// API V2 Mock Handlers

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

	var items []gin.H
	if listType == "history" {
		items = []gin.H{
			{
				"booking_id":   "b_hist_1",
				"date_display": "02/07 (六) 14:30",
				"title":        "師大班 @ 師大附中",
				"attendees": []gin.H{
					{"name": "小明", "status": "CheckedIn", "booking_time": "2026-02-07T14:30:00Z", "booking_id": "h1"},
				},
			},
			{
				"booking_id":   "b_hist_2",
				"date_display": "01/31 (六) 14:30",
				"title":        "師大班 @ 師大附中",
				"attendees": []gin.H{
					{"name": "小明", "status": "Absent", "booking_time": "2026-01-31T14:30:00Z", "booking_id": "h2"},
				},
			},
		}
	} else {
		// Upcoming (Mock same as initial load for consistency)
		items = []gin.H{
			{
				"booking_id":   "b_up_1",
				"date_display": "02/14 (六) 14:30",
				"title":        "師大班 @ 師大附中",
				"attendees": []gin.H{
					{"name": "小明", "status": "Booked", "booking_time": time.Now().Format(time.RFC3339), "booking_id": "101"},
					{"name": "小華", "status": "Leave", "booking_time": time.Now().Format(time.RFC3339), "booking_id": "102"},
				},
			},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items":       items,
		"next_cursor": "",
		"has_more":    false,
	})
}

func (api *v2BookingAPI) getCalendarWeeksV2(c *gin.Context) {
	startDateStr := c.Query("start_date")
	direction := c.Query("direction") // "next" or "prev"

	refDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		refDate = time.Now()
	}

	var startGenerationDate time.Time
	// We want to load 2 weeks
	weeksToLoad := 2

	if direction == "prev" {
		// If refDate is the first day of current view, we want the 2 weeks BEFORE it.
		// e.g. ref is Feb 15. We want Jan 31 - Feb 14 (approx).
		// Actually, we usually align to weeks (Sunday/Monday).
		// For simplicity in mock: just subtract 14 days from refDate.
		startGenerationDate = refDate.AddDate(0, 0, -7*weeksToLoad)
	} else {
		// If refDate is the last day of current view, we want the 2 weeks AFTER it.
		// e.g. ref is Feb 28. We want Mar 1 - Mar 14.
		startGenerationDate = refDate.AddDate(0, 0, 1)
	}

	weeks := make([]gin.H, 0, weeksToLoad)

	// Helper to generate a mock day
	generateMockDay := func(date time.Time) gin.H {
		dayName := date.Format("Mon")
		return gin.H{
			"date_display": fmt.Sprintf("%d", date.Day()),
			"day_of_week":  dayName,
			"is_today":     date.Format("2006-01-02") == time.Now().Format("2006-01-02"),
			"full_date":    date.Format("2006-01-02"),
			"slots": []gin.H{
				{
					"id":           fmt.Sprintf("slot-mock-%d", date.Unix()),
					"time_display": "16:30",
					"course_name":  "足球練習",
					"location":     "球場",
					"capacity":     20,
					"booked_count": 0,
					"is_empty":     false,
					"attendees":    []gin.H{},
				},
			},
		}
	}

	currentDate := startGenerationDate
	for w := 0; w < weeksToLoad; w++ {
		weekDays := make([]gin.H, 0, 7)
		// Assuming we want to align slightly or just give 7 days chunks
		// Let's just generate 7 consecutive days for each week
		for d := 0; d < 7; d++ {
			weekDays = append(weekDays, generateMockDay(currentDate))
			currentDate = currentDate.AddDate(0, 0, 1)
		}

		_, weekIso := currentDate.AddDate(0, 0, -7).ISOWeek()

		weeks = append(weeks, gin.H{
			"id":         fmt.Sprintf("%d-W%d", currentDate.Year(), weekIso),
			"is_current": false,
			"days":       weekDays,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"weeks": weeks,
	})
}
