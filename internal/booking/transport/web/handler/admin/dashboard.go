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
		adminQueryTrainRangeUC:       registry.AdminQueryTrainRange,
		findTrainHasApptsByIdUC:      registry.FindTrainHasApptsById,
		adminCheckInUC:               registry.AdminCheckIn,
		adminToggleCheckInUC:         registry.AdminToggleCheckIn,
		adminCreateLeaveUC:           registry.AdminCreateLeave,
		adminRestoreFromLeaveUC:      registry.AdminRestoreFromLeave,
		adminCreateWalkInUC:          registry.AdminCreateWalkIn,
		adminQueryStudentsUC:         registry.AdminQueryStudents,
		adminBatchUpdateAttendanceUC: registry.AdminBatchUpdateAttendance,
		queryAllUserApptStatsUC:      registry.QueryAllUserApptStats,
		getUserMonthlyStatsUC:        registry.GetUserMonthlyStats,
		queryMonthlyUserReportsUC:    registry.QueryMonthlyUserReports,
		getBusinessAnalyticsUC:       registry.GetBusinessAnalytics,
		getUserDetailUC:              registry.GetUserDetail,
	}
}

type adminAPI struct {
	adminQueryTrainRangeUC       uccore.ReadUseCase[readTrain.ReqAdminQueryTrainRange, []*entity.TrainDateHasApptState]
	findTrainHasApptsByIdUC      uccore.ReadUseCase[readTrain.ReqFindTrainHasApptsById, *entity.TrainDateHasApptState]
	adminCheckInUC               writeAppt.AdminCheckInUseCase
	adminToggleCheckInUC         writeAppt.AdminToggleCheckInUseCase
	adminCreateLeaveUC           writeAppt.AdminCreateLeaveUseCase
	adminRestoreFromLeaveUC      writeAppt.AdminRestoreFromLeaveUseCase
	adminCreateWalkInUC          writeAppt.AdminCreateWalkInUseCase
	adminQueryStudentsUC         readStats.AdminQueryStudentsUseCase
	adminBatchUpdateAttendanceUC writeAppt.AdminBatchUpdateAttendanceUseCase
	queryAllUserApptStatsUC      uccore.ReadUseCase[readStats.ReqQueryAllUserApptStats, []*entity.UserApptStats]
	getUserMonthlyStatsUC        readStats.GetUserMonthlyStatsUseCase
	queryMonthlyUserReportsUC    readStats.QueryMonthlyUserReportsUseCase
	getBusinessAnalyticsUC       readStats.GetBusinessAnalyticsUseCase
	getUserDetailUC              readStats.GetUserDetailUseCase
	once                         sync.Once
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
	r.POST("/v2/admin/checkin/batch-update", api.batchUpdateAttendance)
	r.GET("/v2/admin/students/search", api.searchStudents)

	r.GET("/v2/admin/users/report", api.getUserReport)
	r.GET("/:lang/v2/admin/users/report", api.getUserReport)
	r.GET("/v2/admin/users/report/export", api.exportUserReport)
	r.GET("/v2/admin/users/:userId", api.getUserDetail)
	r.GET("/:lang/v2/admin/users/:userId", api.getUserDetail)
}

func (api *adminAPI) exportUserReport(c *gin.Context) {
	now := time.Now()
	yearStr := c.DefaultQuery("year", fmt.Sprintf("%d", now.Year()))
	monthStr := c.DefaultQuery("month", fmt.Sprintf("%d", int(now.Month())))

	var year, month int
	fmt.Sscanf(yearStr, "%d", &year)
	fmt.Sscanf(monthStr, "%d", &month)

	// 1. 獲取該月所有資料 (不分頁)
	resp, err := api.queryMonthlyUserReportsUC.Execute(c.Request.Context(), readStats.ReqQueryMonthlyUserReports{
		Year:  year,
		Month: month,
		Page:  1,
		Limit: 1000, // 假設單月家長不超過 1000 位
	})

	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	// 2. 生成 CSV 內容
	fileName := fmt.Sprintf("SeanAIgent_Report_%d_%02d.csv", year, month)
	c.Writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	
	// 寫入 UTF-8 BOM 以免 Excel 亂碼
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	fmt.Fprintln(c.Writer, "家長姓名,UserID,孩子姓名,總預約,出席次數,請假次數,缺席次數,出席率")

	for _, u := range resp.UserStats {
		// 寫入家長匯總列
		parentRate := 0.0
		if u.TotalBookings > 0 {
			parentRate = float64(u.AttendedCount) / float64(u.TotalBookings)
		}
		fmt.Fprintf(c.Writer, "%s,%s,---(家長匯總)---,%d,%d,%d,%d,%.2f%%\n",
			u.UserName, u.UserID, u.TotalBookings, u.AttendedCount, u.LeaveCount, u.AbsentCount, parentRate*100)

		// 寫入孩子明細列
		for _, child := range u.Children {
			childRate := 0.0
			if child.TotalBookings > 0 {
				childRate = float64(child.AttendedCount) / float64(child.TotalBookings)
			}
			fmt.Fprintf(c.Writer, ",,%s,%d,%d,%d,%d,%.2f%%\n",
				child.ChildName, child.TotalBookings, child.AttendedCount, child.LeaveCount, child.AbsentCount, childLimitRate(childRate)*100)
		}
	}
}

func childLimitRate(rate float64) float64 {
	if rate > 1.0 { return 1.0 }
	return rate
}

func apptToRecord(appt *entity.Appointment) *admin.CheckinRecord {
	status := "Pending"
	if appt.Status() == entity.StatusAttended {
		status = "CheckedIn"
	} else if appt.Status() == entity.StatusCancelledLeave {
		status = "Leave"
	} else if appt.Status() == entity.StatusAbsent {
		status = "Absent"
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
		} else if appt.IsAbsent {
			status = "Absent"
		}

		bookings = append(bookings, &admin.CheckinRecord{
			BookingID:   appt.ID,
			ChildName:   appt.ChildName,
			ParentName:  appt.UserName,
			Status:      status,
			IsWalkIn:    appt.IsWalkIn,
			IsGuest:     appt.IsGuest,
			ContactInfo: appt.ContactInfo,
			LeaveReason: appt.LeaveReason,
		})
	}

	model := &admin.CheckinPageModel{
		SessionID:   sessionID,
		DateDisplay: trainData.StartDate.Format("2006/01/02"),
		TimeDisplay: fmt.Sprintf("%s - %s", trainData.StartDate.Format("15:04"), trainData.EndDate.Format("15:04")),
		Location:    trainData.Location,
		Capacity:    trainData.Capacity,
		Bookings:    bookings,
		IsStarted:   time.Now().After(trainData.StartDate),
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
	_, err := api.adminToggleCheckInUC.Execute(c.Request.Context(), writeAppt.ReqAdminToggleCheckIn{
		BookingID: bookingID,
	})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	// In batch mode, we don't return row fragments anymore as state is managed by Alpine.js
	c.Status(http.StatusOK)
}

func (api *adminAPI) createLeave(c *gin.Context) {
	if !mid.IsAdmin(c) {
		c.Status(http.StatusUnauthorized)
		return
	}

	bookingID := c.PostForm("bookingId")
	_, err := api.adminCreateLeaveUC.Execute(c.Request.Context(), writeAppt.ReqAdminCreateLeave{
		BookingID: bookingID,
		Reason:    "教練現場標記請假",
	})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (api *adminAPI) restoreFromLeave(c *gin.Context) {
	if !mid.IsAdmin(c) {
		c.Status(http.StatusUnauthorized)
		return
	}

	bookingID := c.PostForm("bookingId")
	_, err := api.adminRestoreFromLeaveUC.Execute(c.Request.Context(), writeAppt.ReqAdminRestoreFromLeave{
		BookingID: bookingID,
	})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	c.Status(http.StatusOK)
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

	if _, err := api.adminCreateWalkInUC.Execute(c.Request.Context(), req); err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	// Prepend to list using the NEW batch row component
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Note: We wrap it in a special div to tell Alpine.js to pick it up or just refresh
	// For simplicity in batch mode, we tell the page to refresh after walk-in to re-sync Alpine state
	c.Writer.Header().Set("HX-Refresh", "true")
	c.Status(http.StatusOK)
}

func (api *adminAPI) renderStatsOOB(c *gin.Context, sessionID string) {
	// Deprecated in batch mode: stats are calculated locally by Alpine.js
}

func (api *adminAPI) batchUpdateAttendance(c *gin.Context) {
	if !mid.IsAdmin(c) {
		c.Status(http.StatusUnauthorized)
		return
	}

	var req writeAppt.ReqAdminBatchUpdateAttendance
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	count, err := api.adminBatchUpdateAttendanceUC.Execute(c.Request.Context(), req)
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"updated_count": count,
	})
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
	resp, err := api.getBusinessAnalyticsUC.Execute(c.Request.Context(), readStats.ReqGetBusinessAnalytics{
		MonthsLimit: 12,
	})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	historicalStats := make([]*admin.MonthlyStat, 0, len(resp.HistoricalStats))
	for i := len(resp.HistoricalStats) - 1; i >= 0; i-- {
		s := resp.HistoricalStats[i]
		historicalStats = append(historicalStats, &admin.MonthlyStat{
			Month:         fmt.Sprintf("%d-%02d", s.Year, s.Month),
			BookedCount:   s.TotalBookings,
			AttendedCount: s.AttendedCount,
		})
	}

	metrics := &admin.BusinessMetrics{
		AvgAttendanceRate: resp.Metrics.AvgAttendanceRate,
		RetentionRate:     resp.Metrics.RetentionRate,
		ActiveStudents:    resp.Metrics.ActiveStudents,
		RevenueGrowth:     resp.Metrics.RevenueGrowth,
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

	resp, err := api.getUserDetailUC.Execute(c.Request.Context(), readStats.ReqGetUserDetail{
		UserID: userID,
	})
	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	var filteredRecords []*admin.UserMonthlyRecord
	var filterStats *admin.UserOverallStats
	availableMonths := make([]string, 0, len(resp.MonthlyRecords))

	for _, r := range resp.MonthlyRecords {
		availableMonths = append(availableMonths, r.MonthValue)
	}

	if monthQuery == "all" {
		// 轉換所有記錄
		for _, r := range resp.MonthlyRecords {
			bookings := make([]*admin.ChildBookingRecord, 0, len(r.Bookings))
			for _, b := range r.Bookings {
				bookings = append(bookings, &admin.ChildBookingRecord{
					ChildName: b.ChildName,
					Date:      b.Date,
					Time:      b.Time,
					Location:  b.Location,
					Status:    b.Status,
				})
			}
			filteredRecords = append(filteredRecords, &admin.UserMonthlyRecord{
				MonthDisplay: r.MonthDisplay,
				MonthValue:   r.MonthValue,
				Bookings:     bookings,
			})
		}
		filterStats = &admin.UserOverallStats{
			TotalBookings:  resp.OverallStats.TotalBookings,
			TotalAttended:  resp.OverallStats.TotalAttended,
			TotalLeave:     resp.OverallStats.TotalLeave,
			TotalAbsent:    resp.OverallStats.TotalAbsent,
			AttendanceRate: resp.OverallStats.AttendanceRate,
		}
	} else {
		// 找出特定月份
		var selectedMonth *readStats.UserMonthlyRecordVO
		for _, r := range resp.MonthlyRecords {
			if r.MonthValue == monthQuery {
				selectedMonth = r
				break
			}
		}

		if selectedMonth != nil {
			bookings := make([]*admin.ChildBookingRecord, 0, len(selectedMonth.Bookings))
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
				bookings = append(bookings, &admin.ChildBookingRecord{
					ChildName: rec.ChildName,
					Date:      rec.Date,
					Time:      rec.Time,
					Location:  rec.Location,
					Status:    rec.Status,
				})
			}
			filteredRecords = append(filteredRecords, &admin.UserMonthlyRecord{
				MonthDisplay: selectedMonth.MonthDisplay,
				MonthValue:   selectedMonth.MonthValue,
				Bookings:     bookings,
			})
			rate := 0.0
			if b > 0 {
				rate = float64(a) / float64(b)
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
		UserID:          resp.UserID,
		LineDisplayName: resp.LineDisplayName,
		CurrentMonth:    monthQuery,
		AvailableMonths: availableMonths,
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
	now := time.Now()
	yearStr := c.DefaultQuery("year", fmt.Sprintf("%d", now.Year()))
	monthStr := c.DefaultQuery("month", fmt.Sprintf("%d", int(now.Month())))
	pageStr := c.DefaultQuery("page", "1")
	search := c.Query("search")

	var year, month int
	var page int64
	fmt.Sscanf(yearStr, "%d", &year)
	fmt.Sscanf(monthStr, "%d", &month)
	fmt.Sscanf(pageStr, "%d", &page)

	resp, err := api.queryMonthlyUserReportsUC.Execute(c.Request.Context(), readStats.ReqQueryMonthlyUserReports{
		Year:   year,
		Month:  month,
		Page:   page,
		Limit:  50, // 預設每頁 50 筆
		Search: search,
	})

	if err != nil {
		handler.ErrorHandler(c, err)
		return
	}

	userStats := make([]*admin.UserAccountStat, 0, len(resp.UserStats))
	for _, s := range resp.UserStats {
		children := make([]*admin.ChildMonthlyStat, 0, len(s.Children))
		for _, cs := range s.Children {
			rate := 0.0
			if cs.TotalBookings > 0 {
				rate = float64(cs.AttendedCount) / float64(cs.TotalBookings)
			}
			children = append(children, &admin.ChildMonthlyStat{
				ChildName:      cs.ChildName,
				Bookings:       cs.TotalBookings,
				Attended:       cs.AttendedCount,
				Leave:          cs.LeaveCount,
				Absent:         cs.AbsentCount,
				AttendanceRate: rate,
			})
		}

		parentRate := 0.0
		if s.TotalBookings > 0 {
			parentRate = float64(s.AttendedCount) / float64(s.TotalBookings)
		}

		userStats = append(userStats, &admin.UserAccountStat{
			UserID:          s.UserID,
			LineDisplayName: s.UserName,
			TotalBookings:   s.TotalBookings,
			TotalAttended:   s.AttendedCount,
			TotalLeave:      s.LeaveCount,
			TotalAbsent:     s.AbsentCount,
			AttendanceRate:  parentRate,
			Children:        children,
		})
	}

	model := &admin.UserReportModel{
		Year:      year,
		Month:     month,
		UserStats: userStats,
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
