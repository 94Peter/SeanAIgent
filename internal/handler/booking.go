package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/94peter/vulpes/ezapi"
	"github.com/94peter/vulpes/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap/buffer"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"seanAIgent/internal/db/model"
	"seanAIgent/internal/service"
	"seanAIgent/internal/service/lineliff"
	"seanAIgent/templates"
	"seanAIgent/templates/forms/bookTraining"
	"seanAIgent/templates/forms/myBookings"
)

type bookingAPI struct {
	svc        service.Service
	enableCSRF bool
}

var initBookingApiOnce sync.Once

func InitBookingApi(service service.Service, enableCSRF bool) {
	initBookingApiOnce.Do(func() {
		api := &bookingAPI{
			svc:        service,
			enableCSRF: enableCSRF,
		}

		ezapi.RegisterGinApi(func(r ezapi.Router) {
			// 建立活動表單

			r.GET("/training/booking", api.getBookingForm)
			r.GET("/:lang/training/booking", api.getBookingForm)
			r.GET("/my-bookings", api.getMyBookingsPage)
			r.GET("/:lang/my-bookings", api.getMyBookingsPage)
			r.POST("/booking/create", api.bookTraining)
			// 取消預約
			r.POST("/booking/delete", api.bookingCancel)
			r.GET("/booking/summary/:type", api.bookingSummary)
			r.GET("/booking/leave-request-form/:bookingId", api.getLeaveRequestForm) // New route for leave request modal
			// 請假
			r.POST("/booking/:bookingId/leave", api.submitLeaveRequest) // New route for submitting leave request
		})

	})
}

func (api *bookingAPI) getMyBookingsPage(c *gin.Context) {
	userId := getUserID(c)

	dbBookings, err := api.svc.QueryUserBookings(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	viewModel := modelToMyBookingsViewModel(dbBookings)
	viewModel.EnableCSRF = api.enableCSRF

	com := templates.Layout(
		myBookings.MyBookingsPage(viewModel),
		lineliff.GetBookingLiffId(),
		&templates.OgMeta{
			Title:       "我的預約",
			Description: "管理您的預約記錄",
			Image:       "https://images.pexels.com/photos/2558605/pexels-photo-2558605.jpeg",
		},
	)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func (api *bookingAPI) getBookingForm(c *gin.Context) {
	lineliffid := lineliff.GetBookingLiffId()
	userId := getUserID(c)
	displayName := getUserDisplayName(c)

	dbTrainingDate, err := api.svc.QueryFutureTrainingDateAppointmentState(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bookableDays := modelTrainingDateToBookTrainingDate(dbTrainingDate)

	com := templates.Layout(
		bookTraining.BookTrainingPage(lineliffid, displayName, bookableDays, api.enableCSRF),
		lineliffid,
		&templates.OgMeta{
			Title:       "Sean AIgent",
			Description: "Sean 的預課服務",
			Image:       "https://images.pexels.com/photos/2558605/pexels-photo-2558605.jpeg",
		},
	)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func (api *bookingAPI) bookingSummary(c *gin.Context) {
	userId := getUserID(c)
	if userId == "" {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "user not logged in",
		})
		return
	}
	dbTrainingDate, err := api.svc.QueryFutureTrainingDateAppointmentState(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	viewTrainingDate := modelTrainingDateToBookTrainingDate(dbTrainingDate)
	slotUsersName := make(map[string][]string)
	for _, td := range dbTrainingDate {
		if td.AppointmentUserNames == nil {
			slotUsersName[td.ID.Hex()] = []string{}
		} else {
			slotUsersName[td.ID.Hex()] = td.AppointmentUserNames
		}
	}

	var message buffer.Buffer
	for _, td := range viewTrainingDate {
		message.WriteString(fmt.Sprintf("%s\n", td.DateDisplay))
		for _, s := range td.Slots {
			message.WriteString(fmt.Sprintf("%s-%s @%s (%d/%d)\n", s.StartTime, s.EndTime, s.Location, s.BookedCount, s.Capacity))
			userNames := slotUsersName[s.ID]
			for i, n := range userNames {
				message.WriteString(fmt.Sprintf("%d. %s\n", i+1, n))
			}
			message.WriteString("\n")
		}
		message.WriteString("-------------------\n")
	}
	c.JSON(http.StatusOK, map[string]interface{}{
		"message": message.String(),
	})
}

func (api *bookingAPI) getLeaveRequestForm(c *gin.Context) {
	bookingID := c.Param("bookingId")
	if bookingID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	com := myBookings.LeaveRequestModal(bookingID, api.enableCSRF, lineliff.GetBookingLiffId())
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func (api *bookingAPI) submitLeaveRequest(c *gin.Context) {
	bookingID := c.Param("bookingId")
	if bookingID == "" {
		addToastTrigger(c, "提交失敗", "預約 ID 缺失。", "error")
		c.Status(http.StatusBadRequest)
		return
	}

	var input struct {
		Reason string `form:"reason"`
	}
	if err := c.ShouldBind(&input); err != nil {
		addToastTrigger(c, "提交失敗", fmt.Sprintf("參數錯誤: %v", err), "error")
		c.Status(http.StatusBadRequest)
		return
	}

	userID := getUserID(c)
	if userID == "" {
		addToastTrigger(c, "提交失敗", "使用者未登入。", "error")
		c.Status(http.StatusUnauthorized)
		return
	}

	// TODO: Implement the actual service call to update booking status to "Leave Requested"
	// For now, just simulate success
	log.Infof("Leave request submitted for BookingID: %s by UserID: %s with Reason: %s", bookingID, userID, input.Reason)
	leave, err := api.svc.CreateLeave(c, bookingID, userID, input.Reason)
	if err != nil {
		log.Error(err.Error())
		addToastTrigger(c, "提交失敗", fmt.Sprintf("提交請假申請失敗: %v", err), "error")
		c.Status(http.StatusInternalServerError)
		return
	}
	if leave == nil {
		addToastTrigger(c, "提交失敗", "提交請假申請失敗。", "error")
		c.Status(http.StatusInternalServerError)
		return
	}
	msg := "教練我要請假"
	detail, err := api.svc.GetLeave(c, leave.ID.Hex())
	if err != nil {
		log.Error(err.Error())
		addToastTrigger(c, "請假申請成功", "您的請假申請已送出，請等待教練審核。", "success")
		c.JSON(http.StatusOK, gin.H{"liffMessage": msg})
		return
	}

	trainDate, err := api.svc.GetTrainingDateDetail(c, detail.TrainingInfo.ID.Hex())
	if err != nil {
		log.Error(err.Error())
		msg, err = service.RenderTemplate("leave_msg", map[string]string{
			"ChildName": detail.ChildName,
			"Date": api.svc.TrainingDateRangeFormat(
				detail.TrainingInfo.StartDate,
				detail.TrainingInfo.EndDate,
				detail.TrainingInfo.Timezone,
			),
			"Reason": detail.Reason,
		})
		if err != nil {
			log.Err(err)
		}
	} else {
		fmt.Println("no error", lineliff.GetBookingLiffUrl(), "ddddd")
		msg, err = service.RenderTemplate("leave_msg", map[string]any{
			"ChildName": detail.ChildName,
			"Date": api.svc.TrainingDateRangeFormat(
				detail.TrainingInfo.StartDate,
				detail.TrainingInfo.EndDate,
				detail.TrainingInfo.Timezone,
			),
			"Reason":      detail.Reason,
			"RemainQuota": fmt.Sprintf("%d", trainDate.Capacity-trainDate.TotalAppointments),
			"BookedList":  trainDate.AppointmentUserNames,
			"BookingURL":  lineliff.GetBookingLiffUrl(),
		})
		if err != nil {
			log.Err(err)
		}
	}

	addToastTrigger(c, "請假申請成功", "您的請假申請已送出，請等待教練審核。", "success")
	c.JSON(http.StatusOK, gin.H{"liffMessage": msg})
}

func (api *bookingAPI) bookingCancel(c *gin.Context) {
	var input bookTraining.InputCancelBookingForm
	err := c.ShouldBind(&input)
	if err != nil {
		api.postErrorHandler(c, err, "", "")
		return
	}
	userId := getUserID(c)
	if userId == "" {
		api.postErrorHandler(c, fmt.Errorf("user not logged in"), "", "")
		return
	}
	cancelled, err := api.svc.AppointmentCancel(c, input.BookingId, userId)
	if err != nil {
		api.postErrorHandler(c, err, input.BookingId, userId)
		return
	}
	if cancelled == nil {
		api.postErrorHandler(c, fmt.Errorf("booking not cancelled"), input.BookingId, userId)
		return
	}
	addToastTrigger(c, "取消成功", "預約已成功取消。", "success")
	c.Status(http.StatusOK)
}

func (api *bookingAPI) returnSlotContent(c *gin.Context, slotId, userId string) {
	dbTrainingDate, err := api.svc.QueryFutureTrainingDateAppointmentState(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	bookableDays := modelTrainingDateToBookTrainingDate(dbTrainingDate)
	var selectedSlot *bookTraining.BookableSlot
	for _, v := range bookableDays {
		for _, slot := range v.Slots {
			if slot.ID == slotId {
				selectedSlot = slot
				break
			}
		}
		if selectedSlot != nil {
			break
		}
	}

	if selectedSlot == nil {
		c.Status(http.StatusNotFound)
		return
	}

	// Render components to buffer to combine them for OOB swap
	var buf bytes.Buffer
	// The main target content is the list of user bookings
	bookTraining.UserBookingsList(selectedSlot.CurrentUserBookings, selectedSlot.ID, api.enableCSRF).Render(c.Request.Context(), &buf)
	// The OOB content is the updated count
	bookTraining.BookedCountOOB(selectedSlot).Render(c.Request.Context(), &buf)

	c.Data(http.StatusOK, "text/html; charset=utf-8", buf.Bytes())
}

func (api *bookingAPI) bookTraining(c *gin.Context) {

	var input bookTraining.InputBookingTrainingForm

	err := c.ShouldBind(&input)

	if err != nil {
		api.postErrorHandler(c, err, "", "")
		return
	}

	userId := getUserID(c)
	if userId == "" {
		api.postErrorHandler(c, fmt.Errorf("user not logged in"), "", "")
		return
	}

	// Parse child names

	rawNames := strings.Split(input.ChildName, " ")

	var childNames []string
	titleCaser := cases.Title(language.English)
	nameMap := make(map[string]bool)
	for _, name := range rawNames {
		trimmedName := titleCaser.String(strings.TrimSpace(name))
		if nameMap[trimmedName] {
			continue
		}
		if trimmedName != "" {
			childNames = append(childNames, trimmedName)
			nameMap[trimmedName] = true
		}
	}

	if len(childNames) == 0 {
		// Trigger a toast for empty name
		addToastTrigger(c, "無效操作", "請至少輸入一個孩子的姓名。", "warning")
		c.Status(http.StatusNoContent) // No content to swap
		return
	}

	// Capacity Check
	dbTrainingDate, err := api.svc.QueryFutureTrainingDateAppointmentState(c, "") // Check global state for capacity
	if err != nil {
		api.postErrorHandler(c, err, input.SlotId, userId)
		return
	}

	var targetSlot *model.AggrTrainingDateAppointState
	for _, slot := range dbTrainingDate {
		if slot.ID.Hex() == input.SlotId {
			targetSlot = slot
			break
		}
	}

	if targetSlot == nil {
		api.postErrorHandler(c, fmt.Errorf("slot not found"), input.SlotId, userId)
		return
	}

	if (targetSlot.Capacity - targetSlot.TotalAppointments) < len(childNames) {
		remaining := targetSlot.Capacity - targetSlot.TotalAppointments
		description := fmt.Sprintf("名額不足，僅剩 %d 位。", remaining)
		addToastTrigger(c, "預約失敗", description, "error")
		c.Status(http.StatusNoContent) // No content to swap, just show toast
		return
	}

	userName := getUserDisplayName(c)
	if err = api.svc.AppointmentTrainingDates(c, input.SlotId, userId, userName, childNames); err != nil {
		api.postErrorHandler(c, err, input.SlotId, userId)
		return
	}

	// Success: trigger toast and return updated content

	bookedNamesStr := strings.Join(childNames, ", ")
	description := fmt.Sprintf("%s 已成功加入預約。", bookedNamesStr)
	addToastTrigger(c, "預約成功", description, "success")

	api.returnSlotContent(c, input.SlotId, userId)

}

func (api *bookingAPI) postErrorHandler(c *gin.Context, err error, slotId, userId string) {
	log.Warnf("post error: %v", err)
	addToastTrigger(c, "操作失敗", err.Error(), "error")
	api.returnSlotContent(c, slotId, userId)
}

const cancellableDuration = 1 * time.Hour

func modelToMyBookingsViewModel(dbBookings []*model.AggrUserAppointment) *myBookings.MyBookingsPageModel {
	bookingsByDate := make(map[string][]*myBookings.BookingItem)
	var startDate, endDate time.Time
	var reason string
	for _, dbBooking := range dbBookings {
		if dbBooking.TrainingDate == nil {
			continue // Skip if training date info is missing
		}
		td := dbBooking.TrainingDate
		startDate = model.ToTime(td.StartDate, td.Timezone)
		endDate = model.ToTime(td.EndDate, td.Timezone)
		dateDisplay := formattedDate(startDate)

		childName := dbBooking.ChildName
		if childName == "" {
			childName = dbBooking.UserName // Fallback to parent's name
		}

		if dbBooking.IsOnLeave {
			reason = dbBooking.LeaveInfo.Reason
		} else {
			reason = ""
		}
		item := &myBookings.BookingItem{
			BookingID:     dbBooking.ID.Hex(),
			ChildName:     childName,
			StartTime:     startDate.Format("15:04"),
			EndTime:       endDate.Format("15:04"),
			Location:      td.Location,
			IsCancellable: time.Since(dbBooking.CreatedAt) < cancellableDuration,
			IsOnLeave:     dbBooking.IsOnLeave,
			OnLeaveReason: reason,
		}
		bookingsByDate[dateDisplay] = append(bookingsByDate[dateDisplay], item)
	}

	// Sort dates
	uniqueDates := make([]string, 0, len(bookingsByDate))
	for dateStr := range bookingsByDate {
		uniqueDates = append(uniqueDates, dateStr)
	}
	sort.Strings(uniqueDates)

	// Create final model
	bookingGroups := make([]*myBookings.BookingGroup, 0, len(uniqueDates))
	for _, dateStr := range uniqueDates {
		group := &myBookings.BookingGroup{
			DateDisplay: dateStr,
			Bookings:    bookingsByDate[dateStr],
		}
		bookingGroups = append(bookingGroups, group)
	}

	return &myBookings.MyBookingsPageModel{
		Bookings: bookingGroups,
	}
}

func modelTrainingDateToBookTrainingDate(dbTrainingDate []*model.AggrTrainingDateAppointState) []*bookTraining.BookableDay {
	// Group slots by date string
	slotsByDate := make(map[string][]*model.AggrTrainingDateAppointState)
	for _, v := range dbTrainingDate {
		slotsByDate[v.Date] = append(slotsByDate[v.Date], v)
	}

	// Get unique dates and sort them
	uniqueDates := make([]string, 0, len(slotsByDate))
	for dateStr := range slotsByDate {
		uniqueDates = append(uniqueDates, dateStr)
	}
	sort.Strings(uniqueDates)

	// Build the final BookableDay slice
	bookableDays := make([]*bookTraining.BookableDay, 0, len(uniqueDates))
	var startTime, endTime time.Time
	for _, dateStr := range uniqueDates {
		slotsForDay := slotsByDate[dateStr]

		// Sort slots within the day by start time
		sort.Slice(slotsForDay, func(i, j int) bool {
			return slotsForDay[i].StartDate.Before(slotsForDay[j].StartDate)
		})

		// Create the view model for the day
		dayVM := &bookTraining.BookableDay{
			Slots: make([]*bookTraining.BookableSlot, 0, len(slotsForDay)),
		}

		// Set the display date from the first slot of the day
		if len(slotsForDay) > 0 {
			dayVM.DateDisplay = formattedDate(slotsForDay[0].StartDate)
		}

		// Transform each slot model into a slot view model
		for _, slotModel := range slotsForDay {
			userBookings := make([]*bookTraining.UserBooking, 0, len(slotModel.UserAppointments))
			for _, userBookingModel := range slotModel.UserAppointments {
				userBookings = append(userBookings, &bookTraining.UserBooking{
					BookingID: userBookingModel.ID.Hex(),
					ChildName: userBookingModel.ChildName,
					IsOnLeave: userBookingModel.IsOnLeave,
				})
			}
			startTime = model.ToTime(slotModel.StartDate, slotModel.Timezone)
			endTime = model.ToTime(slotModel.EndDate, slotModel.Timezone)
			slotVM := &bookTraining.BookableSlot{
				ID:                    slotModel.ID.Hex(),
				StartTime:             startTime.Format("15:04"),
				EndTime:               endTime.Format("15:04"),
				Location:              slotModel.Location,
				Capacity:              slotModel.Capacity,
				BookedCount:           slotModel.TotalAppointments,
				IsBookedByCurrentUser: len(userBookings) > 0,
				CurrentUserBookings:   userBookings,
			}
			dayVM.Slots = append(dayVM.Slots, slotVM)
		}
		bookableDays = append(bookableDays, dayVM)
	}

	return bookableDays
}
