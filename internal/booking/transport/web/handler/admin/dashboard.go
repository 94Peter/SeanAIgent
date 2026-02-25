package admin

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/transport/util/lineutil"
	"seanAIgent/internal/booking/transport/web/handler"
	"seanAIgent/internal/booking/usecase"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	uccore "seanAIgent/internal/booking/usecase/core"
	readStats "seanAIgent/internal/booking/usecase/stats/read"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	"seanAIgent/templates"
	"seanAIgent/templates/admin"

	"github.com/94peter/botreplyer/provider/line/mid"
	"github.com/94peter/vulpes/ezapi"
	"github.com/gin-gonic/gin"
)

func NewAdminApi(registry *usecase.Registry) handler.WebAPI {
	return &adminAPI{
		adminQueryTrainRangeUC:  registry.AdminQueryTrainRange,
		findTrainHasApptsByIdUC: registry.FindTrainHasApptsById,
		adminCheckInUC:          registry.AdminCheckIn,
		adminToggleCheckInUC:    registry.AdminToggleCheckIn,
		adminCreateLeaveUC:      registry.AdminCreateLeave,
		adminRestoreFromLeaveUC: registry.AdminRestoreFromLeave,
		adminCreateWalkInUC:     registry.AdminCreateWalkIn,
		adminQueryStudentsUC:    registry.AdminQueryStudents,
		queryAllUserApptStatsUC: registry.QueryAllUserApptStats,
		getUserMonthlyStatsUC:   registry.GetUserMonthlyStats,
	}
}

type adminAPI struct {
	adminQueryTrainRangeUC  uccore.ReadUseCase[readTrain.ReqAdminQueryTrainRange, []*entity.TrainDateHasApptState]
	findTrainHasApptsByIdUC uccore.ReadUseCase[readTrain.ReqFindTrainHasApptsById, *entity.TrainDateHasApptState]
	adminCheckInUC          writeAppt.AdminCheckInUseCase
	adminToggleCheckInUC    writeAppt.AdminToggleCheckInUseCase
	adminCreateLeaveUC      writeAppt.AdminCreateLeaveUseCase
	adminRestoreFromLeaveUC writeAppt.AdminRestoreFromLeaveUseCase
	adminCreateWalkInUC     writeAppt.AdminCreateWalkInUseCase
	adminQueryStudentsUC    readStats.AdminQueryStudentsUseCase
	queryAllUserApptStatsUC uccore.ReadUseCase[readStats.ReqQueryAllUserApptStats, []*entity.UserApptStats]
	getUserMonthlyStatsUC   readStats.GetUserMonthlyStatsUseCase
	once                    sync.Once
}

func (api *adminAPI) InitRouter(r ezapi.Router) {
	api.once.Do(func() {
		r.GET("/v2/admin/dashboard", api.getDashboard)
		r.GET("/:lang/v2/admin/dashboard", api.getDashboard)
		api.adminGroup(r)
	})
}

func (api *adminAPI) adminGroup(r ezapi.Router) {
	r.GET("/v2/admin/analytics", api.getAnalytics)
	r.GET("/:lang/v2/admin/analytics", api.getAnalytics)
	r.GET("/v2/admin/checkin/:sessionId", api.getCheckinPage)
	r.GET("/:lang/v2/admin/checkin/:sessionId", api.getCheckinPage)
	r.POST("/v2/admin/checkin/submit", api.submitCheckin)
	r.POST("/v2/admin/checkin/toggle", api.toggleCheckin)
	r.POST("/v2/admin/checkin/leave", api.createLeave)
	r.POST("/v2/admin/checkin/restore", api.restoreFromLeave)
	r.POST("/v2/admin/checkin/walkin", api.createWalkIn)
	r.GET("/v2/admin/students/search", api.searchStudents)

	r.GET("/v2/admin/users/report", api.getUserReport)
	r.GET("/:lang/v2/admin/users/report", api.getUserReport)
	r.GET("/v2/admin/users/:userId", api.getUserDetail)
	r.GET("/:lang/v2/admin/users/:userId", api.getUserDetail)
}

func apptToRecord(appt *entity.Appointment) *admin.CheckinRecord {
	status := "Pending"
	if appt.Status() == entity.StatusAttended {
		status = "CheckedIn"
	} else if appt.Status() == entity.StatusCancelledLeave {
		status = "Leave"
	}
	return &admin.CheckinRecord{
		BookingID:   appt.ID(),
		ChildName:   appt.ChildName(),
		ParentName:  appt.User().UserName(),
		Status:      status,
		IsWalkIn:    appt.IsWalkIn(),
		IsGuest:     appt.IsGuest(),
		ContactInfo: appt.ContactInfo(),
	}
}

func (api *adminAPI) getCheckinPage(c *gin.Context) {
	sessionID := c.Param("sessionId")

	trainData, err := api.findTrainHasApptsByIdUC.Execute(c.Request.Context(), readTrain.ReqFindTrainHasApptsById{
		TrainID: sessionID,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load checkin data")
		return
	}

	var bookings []*admin.CheckinRecord
	for _, appt := range trainData.UserAppointments {
		status := "Pending"
		if appt.IsCheckedIn {
			status = "CheckedIn"
		} else if appt.IsOnLeave {
			status = "Leave"
		}

		bookings = append(bookings, &admin.CheckinRecord{
			BookingID:   appt.ID,
			ChildName:   appt.ChildName,
			ParentName:  appt.UserName,
			Status:      status,
			IsWalkIn:    appt.IsWalkIn,
			IsGuest:     appt.IsGuest,
			ContactInfo: appt.ContactInfo,
		})
	}

	model := &admin.CheckinPageModel{
		SessionID:   sessionID,
		DateDisplay: trainData.StartDate.Format("2006/01/02"),
		TimeDisplay: fmt.Sprintf("%s - %s", trainData.StartDate.Format("15:04"), trainData.EndDate.Format("15:04")),
		Location:    trainData.Location,
		Capacity:    trainData.Capacity,
		Bookings:    bookings,
	}

	com := templates.Layout(
		admin.AdminCheckin(model),
		"",
		&templates.OgMeta{
			Title:       "場次點名管理 | Sean AIgent",
			Description: "即時進行學員簽到、請假與臨時加人管理",
			Image:       "",
		},
	)

	c.Render(http.StatusOK, handler.Renderer{
		Ctx:       c.Request.Context(),
		Status:    http.StatusOK,
		Component: com,
	})
}

func (api *adminAPI) submitCheckin(c *gin.Context) {
	// Traditional bulk submit if needed
	c.Status(http.StatusOK)
}

func (api *adminAPI) toggleCheckin(c *gin.Context) {
	if !mid.IsAdmin(c) {
		c.Status(http.StatusUnauthorized)
		return
	}

	bookingID := c.PostForm("bookingId")
	appt, err := api.adminToggleCheckInUC.Execute(c.Request.Context(), writeAppt.ReqAdminToggleCheckIn{
		BookingID: bookingID,
	})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	// Return updated row AND stats OOB
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	admin.CheckinRow(apptToRecord(appt)).Render(c.Request.Context(), c.Writer)
	api.renderStatsOOB(c, appt.TrainingID())
}

func (api *adminAPI) createLeave(c *gin.Context) {
	if !mid.IsAdmin(c) {
		c.Status(http.StatusUnauthorized)
		return
	}

	bookingID := c.PostForm("bookingId")
	appt, err := api.adminCreateLeaveUC.Execute(c.Request.Context(), writeAppt.ReqAdminCreateLeave{
		BookingID: bookingID,
		Reason:    "教練現場標記請假",
	})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	// Use hx-swap-oob to add to leave list
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(c.Writer, "<div hx-swap-oob=\"afterbegin:#leave-list\">")
	admin.LeaveRow(apptToRecord(appt)).Render(c.Request.Context(), c.Writer)
	fmt.Fprintf(c.Writer, "</div>")

	api.renderStatsOOB(c, appt.TrainingID())
}

func (api *adminAPI) restoreFromLeave(c *gin.Context) {
	if !mid.IsAdmin(c) {
		c.Status(http.StatusUnauthorized)
		return
	}

	bookingID := c.PostForm("bookingId")
	appt, err := api.adminRestoreFromLeaveUC.Execute(c.Request.Context(), writeAppt.ReqAdminRestoreFromLeave{
		BookingID: bookingID,
	})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(c.Writer, "<div hx-swap-oob=\"afterbegin:#active-list\">")
	admin.CheckinRow(apptToRecord(appt)).Render(c.Request.Context(), c.Writer)
	fmt.Fprintf(c.Writer, "</div>")

	api.renderStatsOOB(c, appt.TrainingID())
}

func (api *adminAPI) createWalkIn(c *gin.Context) {
	if !mid.IsAdmin(c) {
		c.Status(http.StatusUnauthorized)
		return
	}

	var req writeAppt.ReqAdminCreateWalkIn
	if err := c.ShouldBind(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	appt, err := api.adminCreateWalkInUC.Execute(c.Request.Context(), req)
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	// Prepend to active list
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(c.Writer, "<div hx-swap-oob=\"afterbegin:#active-list\">")
	admin.CheckinRow(apptToRecord(appt)).Render(c.Request.Context(), c.Writer)
	fmt.Fprintf(c.Writer, "</div>")

	api.renderStatsOOB(c, appt.TrainingID())
}

func (api *adminAPI) renderStatsOOB(c *gin.Context, sessionID string) {
	trainData, err := api.findTrainHasApptsByIdUC.Execute(c.Request.Context(), readTrain.ReqFindTrainHasApptsById{
		TrainID: sessionID,
	})
	if err != nil {
		return
	}
	var attended, leave int
	for _, appt := range trainData.UserAppointments {
		if appt.IsCheckedIn {
			attended++
		}
		if appt.IsOnLeave {
			leave++
		}
	}
	// Wrap in container with hx-swap-oob to replace #checkin-stats
	fmt.Fprintf(c.Writer, "<div id=\"checkin-stats\" hx-swap-oob=\"true\" class=\"flex gap-4\">")
	admin.CheckinStats(len(trainData.UserAppointments), attended, leave).Render(c.Request.Context(), c.Writer)
	fmt.Fprintf(c.Writer, "</div>")
}

func (api *adminAPI) searchStudents(c *gin.Context) {
	keyword := c.Query("q")
	sessionID := c.Query("sessionId") // Pass this from frontend
	if keyword == "" {
		return
	}

	stats, err := api.adminQueryStudentsUC.Execute(c.Request.Context(), readStats.ReqAdminQueryStudents{
		SearchKeyword: keyword,
	})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, s := range stats {
		for _, child := range s.ChildState {
			admin.SearchResultRow(s.UserID, child.ChildName, s.UserName, s.UserID, sessionID).Render(c.Request.Context(), c.Writer)
		}
	}
}

func (api *adminAPI) getAnalytics(c *gin.Context) {
	historicalStats := []*admin.MonthlyStat{
		{Month: "2025-03", BookedCount: 120, AttendedCount: 95},
		{Month: "2025-04", BookedCount: 135, AttendedCount: 110},
		{Month: "2025-05", BookedCount: 150, AttendedCount: 125},
		{Month: "2025-06", BookedCount: 180, AttendedCount: 140},
		{Month: "2025-07", BookedCount: 210, AttendedCount: 185},
		{Month: "2025-08", BookedCount: 190, AttendedCount: 170},
		{Month: "2025-09", BookedCount: 220, AttendedCount: 200},
		{Month: "2025-10", BookedCount: 240, AttendedCount: 210},
		{Month: "2025-11", BookedCount: 230, AttendedCount: 205},
		{Month: "2025-12", BookedCount: 260, AttendedCount: 230},
		{Month: "2026-01", BookedCount: 200, AttendedCount: 180},
		{Month: "2026-02", BookedCount: 215, AttendedCount: 195},
	}

	metrics := &admin.BusinessMetrics{
		AvgAttendanceRate: 0.88,
		RetentionRate:     0.92,
		ActiveStudents:    45,
		RevenueGrowth:     7.5,
	}

	model := &admin.AnalyticsModel{
		HistoricalStats: historicalStats,
		Metrics:         metrics,
	}

	com := templates.Layout(
		admin.AdminAnalytics(model),
		"",
		&templates.OgMeta{
			Title:       "經營分析看板 | Sean AIgent",
			Description: "深度分析訓練營經營數據與成長趨勢",
			Image:       "",
		},
	)

	c.Render(http.StatusOK, handler.Renderer{
		Ctx:       c.Request.Context(),
		Status:    http.StatusOK,
		Component: com,
	})
}

func checkUser(c *gin.Context, lineliffid string) bool {
	userID := getUserID(c)
	if userID == "" {
		com := templates.Layout(
			nil,
			lineliffid,
			&templates.OgMeta{
				Title:       "訓練場次看板 | Sean AIgent",
				Description: "即時監控訓練場次預約與簽到狀態",
				Image:       "",
			},
		)

		c.Render(http.StatusOK, handler.Renderer{
			Ctx:       c.Request.Context(),
			Status:    http.StatusOK,
			Component: com,
		})
		return false
	}
	if !isAdmin(c) {
		return false
	}
	return true
}

func (api *adminAPI) getDashboard(c *gin.Context) {
	lineliffid := lineutil.GetAdminDashboardLiffId()
	if !checkUser(c, lineliffid) {
		return
	}
	now := time.Now()
	startTime := now.AddDate(0, 0, -7)
	endTime := now.AddDate(0, 0, 7)

	trainDates, err := api.adminQueryTrainRangeUC.Execute(c.Request.Context(), readTrain.ReqAdminQueryTrainRange{
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load dashboard data")
		return
	}

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)

	var todaySessions, upcomingSessions, pastSessions []*admin.SessionSummary

	for _, td := range trainDates {
		s := &admin.SessionSummary{
			ID:          td.ID,
			DateDisplay: td.StartDate.Format("2006/01/02 (Mon)"),
			TimeDisplay: fmt.Sprintf("%s - %s", td.StartDate.Format("15:04"), td.EndDate.Format("15:04")),
			Location:    td.Location,
			Capacity:    td.Capacity,
			BookedCount: len(td.UserAppointments),
		}

		var attended, leave int
		for _, appt := range td.UserAppointments {
			if appt.IsCheckedIn {
				attended++
			}
			if appt.IsOnLeave {
				leave++
			}
		}
		s.AttendedCount = attended
		s.LeaveCount = leave
		s.PendingCount = s.BookedCount - attended - leave

		if !td.StartDate.Before(todayStart) && td.StartDate.Before(todayEnd) {
			s.DateDisplay = td.StartDate.Format("01/02 (今日)")
			todaySessions = append(todaySessions, s)
		} else if !td.StartDate.Before(todayEnd) {
			upcomingSessions = append(upcomingSessions, s)
		} else {
			pastSessions = append(pastSessions, s)
		}
	}

	model := &admin.DashboardModel{
		TodaySessions:    todaySessions,
		UpcomingSessions: upcomingSessions,
		PastSessions:     pastSessions,
	}

	com := templates.Layout(
		admin.AdminDashboard(model),
		lineliffid,
		&templates.OgMeta{
			Title:       "訓練場次看板 | Sean AIgent",
			Description: "即時監控訓練場次預約與簽到狀態",
			Image:       "",
		},
	)

	c.Render(http.StatusOK, handler.Renderer{
		Ctx:       c.Request.Context(),
		Status:    http.StatusOK,
		Component: com,
	})
}

func (api *adminAPI) getUserDetail(c *gin.Context) {
	userID := c.Param("userId")
	monthQuery := c.DefaultQuery("month", "all")

	allMonthlyRecords := []*admin.UserMonthlyRecord{
		{
			MonthDisplay: "2026年 2月",
			MonthValue:   "2026-02",
			Bookings: []*admin.ChildBookingRecord{
				{ChildName: "小明", Date: "02/24 (週二)", Time: "14:00", Location: "TKU 操場", Status: "CheckedIn"},
				{ChildName: "小紅", Date: "02/24 (週二)", Time: "14:00", Location: "TKU 操場", Status: "CheckedIn"},
				{ChildName: "小明", Date: "02/16 (週一)", Time: "18:30", Location: "中和運動中心", Status: "Absent"},
				{ChildName: "小紅", Date: "02/24 (週二)", Time: "14:00", Location: "TKU 操場", Status: "Leave"},
			},
		},
		{
			MonthDisplay: "2026年 1月",
			MonthValue:   "2026-01",
			Bookings: []*admin.ChildBookingRecord{
				{ChildName: "小明", Date: "01/20 (週二)", Time: "14:00", Location: "TKU 操場", Status: "CheckedIn"},
				{ChildName: "小紅", Date: "01/13 (週二)", Time: "14:00", Location: "TKU 操場", Status: "CheckedIn"},
			},
		},
	}

	overallStats := &admin.UserOverallStats{
		TotalBookings:  45,
		TotalAttended:  38,
		TotalLeave:     4,
		TotalAbsent:    3,
		AttendanceRate: 0.92,
	}

	var filteredRecords []*admin.UserMonthlyRecord
	var filterStats *admin.UserOverallStats

	if monthQuery == "all" {
		filteredRecords = allMonthlyRecords
		filterStats = overallStats
	} else {
		var selectedMonth *admin.UserMonthlyRecord
		for _, r := range allMonthlyRecords {
			if r.MonthValue == monthQuery {
				selectedMonth = r
				filteredRecords = append(filteredRecords, r)
				break
			}
		}

		if selectedMonth != nil {
			var b, a, l, abs int
			for _, rec := range selectedMonth.Bookings {
				b++
				switch rec.Status {
				case "CheckedIn":
					a++
				case "Leave":
					l++
				case "Absent":
					abs++
				}
			}
			rate := 0.0
			if b-l > 0 {
				rate = float64(a) / float64(b-l)
			}
			filterStats = &admin.UserOverallStats{
				TotalBookings:  b,
				TotalAttended:  a,
				TotalLeave:     l,
				TotalAbsent:    abs,
				AttendanceRate: rate,
			}
		} else {
			filterStats = &admin.UserOverallStats{}
		}
	}

	model := &admin.UserDetailModel{
		UserID:          userID,
		LineDisplayName: "Peter_Admin",
		CurrentMonth:    monthQuery,
		AvailableMonths: []string{"2026-02", "2026-01"},
		FilterStats:     filterStats,
		MonthlyRecords:  filteredRecords,
	}

	com := templates.Layout(
		admin.UserDetail(model),
		"",
		&templates.OgMeta{
			Title:       "學員預約明細 | Sean AIgent",
			Description: "檢視學員每月預約與出席狀況",
			Image:       "",
		},
	)

	c.Render(http.StatusOK, handler.Renderer{
		Ctx:       c.Request.Context(),
		Status:    http.StatusOK,
		Component: com,
	})
}

func (api *adminAPI) getUserReport(c *gin.Context) {
	model := &admin.UserReportModel{
		Year:  2026,
		Month: 2,
		UserStats: []*admin.UserAccountStat{
			{
				UserID:          "Ufa91de91be0274e3cc9851918a8e9660",
				LineDisplayName: "Peter_Admin",
				TotalBookings:   20,
				TotalAttended:   18,
				TotalLeave:      1,
				TotalAbsent:     1,
				AttendanceRate:  0.9,
				Children: []*admin.ChildMonthlyStat{
					{
						ChildName:      "小明",
						Bookings:       12,
						Attended:       10,
						Leave:          1,
						Absent:         1,
						AttendanceRate: 0.83,
					},
					{
						ChildName:      "小紅",
						Bookings:       8,
						Attended:       8,
						Leave:          0,
						Absent:         0,
						AttendanceRate: 1.0,
					},
				},
			},
			{
				UserID:          "U1234567890abcdef",
				LineDisplayName: "Strong_Parent",
				TotalBookings:   15,
				TotalAttended:   4,
				TotalLeave:      8,
				TotalAbsent:     3,
				AttendanceRate:  0.26,
				Children: []*admin.ChildMonthlyStat{
					{
						ChildName:      "小強",
						Bookings:       15,
						Attended:       4,
						Leave:          8,
						Absent:         3,
						AttendanceRate: 0.26,
					},
				},
			},
		},
	}

	com := templates.Layout(
		admin.UserReport(model),
		"",
		&templates.OgMeta{
			Title:       "學員數據月報表 | Sean AIgent",
			Description: "檢視家長帳號與各個孩子的出席趨勢",
			Image:       "",
		},
	)

	c.Render(http.StatusOK, handler.Renderer{
		Ctx:       c.Request.Context(),
		Status:    http.StatusOK,
		Component: com,
	})
}

func isAdmin(c *gin.Context) bool {
	return mid.IsAdmin(c)
}

func getUserID(c *gin.Context) string {
	return mid.GetLineLiffUserId(c)
}
