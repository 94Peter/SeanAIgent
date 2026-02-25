package handler

import (
	"bytes"
	"errors"
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

	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/transport/util/lineutil"
	"seanAIgent/internal/booking/usecase"
	readAppt "seanAIgent/internal/booking/usecase/appointment/read"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"
	uccore "seanAIgent/internal/booking/usecase/core"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	"seanAIgent/internal/util/timeutil"
	"seanAIgent/templates"
	"seanAIgent/templates/forms/bookTraining"
	"seanAIgent/templates/forms/checkin"
	"seanAIgent/templates/forms/myBookings"
)

func NewBookingUseCaseSet(registry *usecase.Registry) BookingUseCaseSet {
	return BookingUseCaseSet{
		userQueryFutureTrainUC:   registry.UserQueryFutureTrain,
		userQueryTrainById:       registry.UserQueryTrainByID,
		createApptUC:             registry.CreateAppt,
		cancelApptUC:             registry.CancelAppt,
		createLeaveUC:            registry.CreateLeave,
		cancelLeaveUC:            registry.CancelLeave,
		queryUserBookingsUC:      registry.QueryUserBookings,
		findNearestTrainByTimeUC: registry.FindNearestTrainByTime,
		findTranHasApptByIdUC:    registry.FindTrainHasApptsById,
		checkinUC:                registry.CheckIn,
	}
}

type BookingUseCaseSet struct {
	userQueryFutureTrainUC uccore.ReadUseCase[
		readTrain.ReqUserQueryFutureTrain, []*entity.TrainDateHasUserApptState]
	userQueryTrainById uccore.ReadUseCase[
		readTrain.ReqUserQueryTrainByID, *entity.TrainDateHasUserApptState]
	createApptUC        uccore.WriteUseCase[writeAppt.ReqCreateAppt, []*entity.Appointment]
	cancelApptUC        uccore.WriteUseCase[writeAppt.ReqCancelAppt, *entity.Appointment]
	createLeaveUC       uccore.WriteUseCase[writeAppt.ReqCreateLeave, *entity.Appointment]
	cancelLeaveUC       uccore.WriteUseCase[writeAppt.ReqCancelLeave, *entity.Appointment]
	queryUserBookingsUC uccore.ReadUseCase[
		readAppt.ReqQueryUserBookings, *readAppt.RespQueryUserBookings,
	]
	checkinUC                uccore.WriteUseCase[writeAppt.ReqCheckIn, []*entity.Appointment]
	findNearestTrainByTimeUC uccore.ReadUseCase[
		readTrain.ReqFindNearestTrainByTime, *entity.TrainDateHasApptState,
	]
	findTranHasApptByIdUC uccore.ReadUseCase[
		readTrain.ReqFindTrainHasApptsById, *entity.TrainDateHasApptState,
	]
}

func NewBookingApi(enableCSRF bool, bookingUseCaseSet BookingUseCaseSet) WebAPI {
	return &bookingAPI{
		BookingUseCaseSet: bookingUseCaseSet,
		enableCSRF:        enableCSRF,
	}
}

type bookingAPI struct {
	BookingUseCaseSet
	enableCSRF bool
	once       sync.Once
}

func (api *bookingAPI) InitRouter(r ezapi.Router) {
	api.once.Do(func() {
		// 建立活動表單
		r.GET("/training/booking", api.getBookingForm)
		r.GET("/:lang/training/booking", api.getBookingForm)
		r.GET("/my-bookings", api.getMyBookingsPage)
		r.GET("/:lang/my-bookings", api.getMyBookingsPage)
		r.GET("/my-bookings/items", api.getMyBookingsNextPage)
		r.GET("/:lang/my-bookings/items", api.getMyBookingsNextPage)
		r.POST("/booking/create", api.bookTraining)
		// 取消預約
		r.POST("/booking/delete", api.bookingCancel)
		r.GET("/booking/summary/:type", api.bookingSummary)
		r.GET("/booking/leave-request-form/:bookingId", api.getLeaveRequestForm) // New route for leave request modal
		// 請假
		r.POST("/booking/:bookingId/leave", api.submitLeaveRequest) // New route for submitting leave request
		r.DELETE("/booking/:bookingId/leave", api.cancelLeave)
		// 簽到
		r.GET("/:lang/admin/checkin", api.getCheckinPage)
		r.GET("/admin/checkin", api.getCheckinPage)
		r.POST("/admin/checkin/submit", api.submitCheckin)
	})
}

func (api *bookingAPI) getMyBookingsPage(c *gin.Context) {
	userId := getUserID(c)
	t := time.Now()
	// userId = "Ufa91de91be0274e3cc9851918a8e9660"
	// t = time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	dbBookings, err := api.queryUserBookingsUC.Execute(c.Request.Context(), readAppt.ReqQueryUserBookings{
		UserID:         userId,
		TrainDateAfter: t,
	})
	if err != nil {
		c.Error(err)
		return
	}

	viewModel := modelToMyBookingsViewModel(dbBookings.Appts, dbBookings.Cursor)
	viewModel.EnableCSRF = api.enableCSRF

	com := templates.Layout(
		myBookings.MyBookingsPage(viewModel),
		lineutil.GetBookingLiffId(),
		&templates.OgMeta{
			Title:       "我的預約",
			Description: "管理您的預約記錄",
			Image:       "https://storage.94peter.dev/cdn-cgi/image/width=1200,height=630,quality=80,format=auto/https://storage.94peter.dev/images/UAC.png",
		},
	)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func (api *bookingAPI) getMyBookingsNextPage(c *gin.Context) {
	userId := getUserID(c)
	userId = "Ufa91de91be0274e3cc9851918a8e9660"
	cursor := c.Query("cursor")
	if cursor == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	dbBookings, err := api.queryUserBookingsUC.Execute(c.Request.Context(), readAppt.ReqQueryUserBookings{
		UserID: userId,
		Cursor: cursor,
	})
	if err != nil {
		c.Error(err)
		return
	}

	viewModel := modelToMyBookingsViewModel(dbBookings.Appts, dbBookings.Cursor)
	viewModel.EnableCSRF = api.enableCSRF

	// Return only the list content, not the full page layout
	com := myBookings.BookingList(viewModel)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func (api *bookingAPI) getBookingForm(c *gin.Context) {
	lineliffid := lineutil.GetBookingLiffId()
	userId := getUserID(c)
	displayName := getUserDisplayName(c)

	dbTrainingDate, err := api.userQueryFutureTrainUC.Execute(
		c.Request.Context(),
		readTrain.ReqUserQueryFutureTrain{
			UserID:    userId,
			TimeAfter: time.Now(),
		},
	)
	if err != nil {

	}

	bookableDays := modelTrainingDateToBookTrainingDate(dbTrainingDate)

	com := templates.Layout(
		bookTraining.BookTrainingPage(lineliffid, displayName, bookableDays, api.enableCSRF),
		lineliffid,
		&templates.OgMeta{
			Title:       "Sean AIgent",
			Description: "Sean 的預課服務",
			Image:       "https://storage.94peter.dev/cdn-cgi/image/width=1200,height=630,quality=80,format=auto/https://storage.94peter.dev/images/UAC.png",
		},
	)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func (api *bookingAPI) bookingSummary(c *gin.Context) {
	userId := getUserID(c)
	if userId == "" {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"message": "user not logged in",
		})
		return
	}
	dbTrainingDate, err := api.userQueryFutureTrainUC.Execute(
		c.Request.Context(),
		readTrain.ReqUserQueryFutureTrain{
			UserID:    userId,
			TimeAfter: time.Now(),
		},
	)
	if err != nil {
		ErrorHandler(c, err)
		return
	}

	viewTrainingDate := modelTrainingDateToBookTrainingDate(dbTrainingDate)
	slotUsersName := make(map[string][]string)
	for _, td := range dbTrainingDate {
		if len(td.AllUsers) == 0 {
			slotUsersName[td.ID] = []string{}
		} else {
			slotUsersName[td.ID] = td.AllUsers
		}
	}

	var message buffer.Buffer
	message.WriteString("✨ 點擊下方連結開始預約，加入我們一起進步！\n\n")
	message.WriteString(fmt.Sprintf("👉 %s\n\n", lineutil.GetBookingV2LiffUrl()))
	message.WriteString("📋 課程預約與出席名單\n")
	message.WriteString("-------------------\n")
	for _, td := range viewTrainingDate {
		message.WriteString(fmt.Sprintf("📅 %s\n", td.DateDisplay))
		for _, s := range td.Slots {
			message.WriteString(fmt.Sprintf("%s-%s 📍%s (%d/%d)\n", s.StartTime, s.EndTime, s.Location, s.BookedCount, s.Capacity))
			userNames := slotUsersName[s.ID]
			for i, n := range userNames {
				message.WriteString(fmt.Sprintf("%d. %s\n", i+1, n))
			}
			message.WriteString("\n")
		}
		message.WriteString("-------------------\n")
	}

	message.WriteString("")
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

	com := myBookings.LeaveRequestModal(bookingID, api.enableCSRF, lineutil.GetBookingLiffId())
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
	var err error
	var input struct {
		Reason string `form:"reason"`
	}
	if err := c.ShouldBind(&input); err != nil {
		addToastTrigger(c, "提交失敗", fmt.Sprintf("參數錯誤: %v", err), "error")
		c.Status(http.StatusBadRequest)
		return
	}

	if input.Reason == "" {
		addToastTrigger(c, "提交失敗", "請假原因不能為空。", "error")
		c.Status(http.StatusBadRequest)
		return
	}

	userID := getUserID(c)
	if userID == "" {
		addToastTrigger(c, "提交失敗", "使用者未登入。", "error")
		c.Status(http.StatusUnauthorized)
		return
	}
	userName := getUserDisplayName(c)
	domainUser, err := entity.NewUser(userID, userName)
	if err != nil {
		log.Error(err.Error())
		addToastTrigger(c, "提交失敗", fmt.Sprintf("提交請假申請失敗: %v", err), "error")
		c.Status(http.StatusInternalServerError)
		return
	}

	// TODO: Implement the actual service call to update booking status to "Leave Requested"
	// For now, just simulate success
	log.Infof("Leave request submitted for BookingID: %s by UserID: %s with Reason: %s", bookingID, userID, input.Reason)
	appt, err := api.createLeaveUC.Execute(c, writeAppt.ReqCreateLeave{
		AppointmentID: bookingID,
		User:          domainUser,
		Reason:        input.Reason,
	})

	if err != nil {
		log.Error(err.Error())
		addToastTrigger(c, "提交失敗", fmt.Sprintf("提交請假申請失敗: %v", err), "error")
		c.Status(http.StatusInternalServerError)
		return
	}

	var msg string
	train, err := api.userQueryTrainById.Execute(c.Request.Context(), readTrain.ReqUserQueryTrainByID{
		UserID:      userID,
		TrainDateID: appt.TrainingID(),
	})
	if err != nil {
		log.Error(err.Error())
		msg, err = lineutil.RenderTemplate("leave_msg", map[string]string{
			"ChildName": appt.ChildName(),
			"Date": TrainDateRangeFormat(
				train.StartDate,
				train.EndDate,
				train.Timezone,
			),
			"Reason": appt.LeaveInfo().Reason(),
		})
		if err != nil {
			log.Err(err)
		}
	} else {
		msg, err = lineutil.RenderTemplate("leave_msg", map[string]any{
			"ChildName": appt.ChildName(),
			"Date": TrainDateRangeFormat(
				train.StartDate,
				train.EndDate,
				train.Timezone,
			),
			"Reason":      appt.LeaveInfo().Reason(),
			"RemainQuota": train.AvailableCapacity,
			"BookedList":  train.AllUsers,
			"BookingURL":  lineutil.GetBookingLiffUrl(),
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
	_, err = api.cancelApptUC.Execute(c.Request.Context(), writeAppt.ReqCancelAppt{
		ApptID: input.BookingId,
		UserID: userId,
	})
	if err != nil {
		api.postErrorHandler(c, err, input.BookingId, userId)
		return
	}
	addToastTrigger(c, "取消成功", "預約已成功取消。", "success")
	c.Status(http.StatusOK)
}

func (api *bookingAPI) returnSlotContent(c *gin.Context, slotId, userId string) {
	dbTrainingDate, err := api.userQueryFutureTrainUC.Execute(
		c.Request.Context(),
		readTrain.ReqUserQueryFutureTrain{
			UserID:    userId,
			TimeAfter: time.Now(),
		},
	)
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
	input.ChildName = strings.ReplaceAll(input.ChildName, "，", ",")
	rawNames := strings.Split(input.ChildName, ",")

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

	userName := getUserDisplayName(c)
	domainUser, err := entity.NewUser(userId, userName)
	if err != nil {
		api.postErrorHandler(c, err, input.SlotId, userId)
	}

	_, err = api.createApptUC.Execute(c.Request.Context(), writeAppt.ReqCreateAppt{
		TrainDateID: input.SlotId,
		User:        domainUser,
		ChildNames:  childNames,
	})
	if err != nil {
		switch {
		case errors.Is(err, writeAppt.ErrCreateApptDeductCapacityFail):
			addToastTrigger(c, "預約失敗", "名額不足或時段已結束。", "error")
			c.Status(http.StatusNoContent) // No content to swap, just show toast
			return
		case errors.Is(err, writeAppt.ErrCreateApptSaveApptFail):
			addToastTrigger(c, "預約失敗", "預約失敗，請稍後再試。", "error")
			c.Status(http.StatusNoContent) // No content to swap
			return
		default:
			api.postErrorHandler(c, err, input.SlotId, userId)
		}
	}

	// Success: trigger toast and return updated content
	bookedNamesStr := strings.Join(childNames, ", ")
	description := fmt.Sprintf("%s 已成功加入預約。", bookedNamesStr)
	addToastTrigger(c, "預約成功", description, "success")

	api.returnSlotContent(c, input.SlotId, userId)
}

func (api *bookingAPI) postErrorHandler(c *gin.Context, err error, slotId, userId string) {
	log.Warnf("post error: %v: %v", err, errors.Unwrap(err))
	addToastTrigger(c, "操作失敗", err.Error(), "error")
	api.returnSlotContent(c, slotId, userId)
}

const cancellableDuration = 24 * time.Hour

func modelToMyBookingsViewModel(dbBookings []*entity.AppointmentWithTrainDate, cursor string) *myBookings.MyBookingsPageModel {
	bookingsByDate := make(map[string][]*myBookings.BookingItem)
	var startDate, endDate time.Time
	var reason string
	for _, dbBooking := range dbBookings {
		if dbBooking.TrainDate.ID == "" {
			continue // Skip if training date info is missing
		}
		td := dbBooking.TrainDate
		startDate = timeutil.ToLocation(td.StartDate, td.Timezone)
		endDate = timeutil.ToLocation(td.EndDate, td.Timezone)
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
			BookingID:     dbBooking.ID,
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
		Bookings:   bookingGroups,
		NextCursor: cursor,
	}
}

func modelTrainingDateToBookTrainingDate(
	dbTrainingDate []*entity.TrainDateHasUserApptState,
) []*bookTraining.BookableDay {
	// Group slots by date string
	slotsByDate := make(map[string][]*entity.TrainDateHasUserApptState, 100)
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
					BookingID: userBookingModel.ID,
					ChildName: userBookingModel.ChildName,
					IsOnLeave: userBookingModel.IsOnLeave,
				})
			}
			startTime = timeutil.ToLocation(slotModel.StartDate, slotModel.Timezone)
			endTime = timeutil.ToLocation(slotModel.EndDate, slotModel.Timezone)
			slotVM := &bookTraining.BookableSlot{
				ID:                    slotModel.ID,
				StartTime:             startTime.Format("15:04"),
				EndTime:               endTime.Format("15:04"),
				Location:              slotModel.Location,
				Capacity:              slotModel.Capacity,
				BookedCount:           len(slotModel.AllUsers),
				IsBookedByCurrentUser: len(userBookings) > 0,
				CurrentUserBookings:   userBookings,
			}
			dayVM.Slots = append(dayVM.Slots, slotVM)
		}
		bookableDays = append(bookableDays, dayVM)
	}

	return bookableDays
}

func (api *bookingAPI) cancelLeave(c *gin.Context) {
	bookingID := c.Param("bookingId")
	if bookingID == "" {
		addToastTrigger(c, "取消請假失敗", "預約 ID 缺失。", "error")
		c.Status(http.StatusBadRequest)
		return
	}
	userID := getUserID(c)
	if userID == "" {
		addToastTrigger(c, "取消請假失敗", "使用者未登入。", "error")
		c.Status(http.StatusUnauthorized)
		return
	}
	_, err := api.cancelLeaveUC.Execute(
		c.Request.Context(),
		writeAppt.ReqCancelLeave{ApptID: bookingID, UserID: userID},
	)
	if err != nil {
		addToastTrigger(c, "取消請假失敗", err.Error(), "error")
		c.Status(http.StatusInternalServerError)
		return
	}
	addToastTrigger(c, "取消請假成功", "請假已成功取消。", "success")
	c.Status(http.StatusOK)
}

func (api *bookingAPI) getCheckinPage(c *gin.Context) {
	lineliffid := lineutil.GetCheckinLiffId() // Assuming same liffId for now

	var viewModel *checkin.CheckinPageModel
	var queryTime time.Time
	if !isAdmin(c) {
		viewModel = &checkin.CheckinPageModel{
			ErrorMessage: "您沒有權限開啟此頁面。",
		}
	} else {
		// Determine current time slot
		if trainingTimeStr := c.Query("time"); trainingTimeStr != "" {
			parseTime, err := time.Parse(time.RFC3339, trainingTimeStr)
			if err != nil {
				viewModel = &checkin.CheckinPageModel{
					ErrorMessage: "時間格式錯誤",
				}
			}
			queryTime = parseTime
		} else {
			const defaultBeforeTrainStartDuration = 10 * time.Minute
			queryTime = time.Now().Add(defaultBeforeTrainStartDuration)
		}

		if !queryTime.IsZero() {
			// Query bookings for this slot
			checkinList, err := api.findNearestTrainByTimeUC.Execute(
				c.Request.Context(),
				readTrain.ReqFindNearestTrainByTime{TimeAfter: queryTime},
			)
			if err != nil {
				log.Errorf("Failed to query checkin list: %v", err)
				viewModel = &checkin.CheckinPageModel{
					ErrorMessage: "無法載入簽到列表，請稍後再試。",
				}
			} else {
				// Transform to view model
				viewModel = modelToCheckinPageModel(checkinList)
			}
		}
	}

	viewModel.EnableCSRF = api.enableCSRF
	viewModel.StartTime = queryTime.Format(time.RFC3339)

	com := templates.Layout(
		checkin.CheckinPage(viewModel, lineliffid),
		lineliffid,
		&templates.OgMeta{
			Title:       "簽到管理",
			Description: "管理訓練時段簽到",
			Image:       "https://storage.94peter.dev/cdn-cgi/image/width=1200,height=630,quality=80,format=auto/https://storage.94peter.dev/images/UAC.png",
		},
	)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func (api *bookingAPI) submitCheckin(c *gin.Context) {
	if !isAdmin(c) {
		addToastTrigger(c, "權限不足", "您沒有權限進行此操作。", "error")
		c.Status(http.StatusUnauthorized)
		return
	}

	var input checkin.InputCheckinForm
	// c.ShouldBind(&input) will bind form data, including arrays of checkboxes
	if err := c.ShouldBind(&input); err != nil {
		addToastTrigger(c, "提交失敗", fmt.Sprintf("參數錯誤: %v", err), "error")
		c.Status(http.StatusBadRequest)
		return
	}
	var err error
	// Update checkin status in DB
	_, err = api.checkinUC.Execute(c.Request.Context(), writeAppt.ReqCheckIn{
		TrainDateID:         input.SlotID,
		CheckedInBookingIDs: input.CheckedInBookingIDs,
	})
	if err != nil {
		log.Errorf("Failed to update checkin status: " + getAllErrorMessage(err))
		addToastTrigger(c, "提交失敗", "更新簽到狀態失敗，請稍後再試。", "error")
		c.Status(http.StatusInternalServerError)
		return
	}

	updatedCheckinList, err := api.findTranHasApptByIdUC.Execute(
		c.Request.Context(), readTrain.ReqFindTrainHasApptsById{TrainID: input.SlotID},
	)
	if err != nil {
		log.Errorf("Failed to re-query checkin list for notification after submit: %v", err)
		addToastTrigger(c, "提交失敗", "無法取得簽到結果，請稍後再試。", "error")
		c.Status(http.StatusInternalServerError)
		return
	}

	// Construct LIFF message content
	liffMessage := messageCheckinResult(updatedCheckinList)
	log.Debugf("Sending LIFF message JSON: %s", liffMessage)

	c.JSON(http.StatusOK, gin.H{
		"liffMessage": liffMessage,
	})
}

func modelToCheckinPageModel(data *entity.TrainDateHasApptState) *checkin.CheckinPageModel {
	if data == nil {
		return &checkin.CheckinPageModel{
			ErrorMessage: "查無資料",
		}
	}
	leaveCount := 0
	for _, appt := range data.UserAppointments {
		if appt.IsOnLeave {
			leaveCount++
		}
	}
	checkinItems := make([]*checkin.CheckinItem, 0, len(data.UserAppointments)-leaveCount)
	onleaveItems := make([]*checkin.CheckinItem, 0, leaveCount)
	for _, dbItem := range data.UserAppointments {
		childName := dbItem.ChildName
		if childName == "" {
			childName = dbItem.UserName // Fallback to parent's name
		}
		if dbItem.IsOnLeave {
			onleaveItems = append(onleaveItems, &checkin.CheckinItem{
				BookingID:   dbItem.ID,
				ChildName:   childName,
				IsCheckedIn: dbItem.IsCheckedIn,
			})
		} else {
			checkinItems = append(checkinItems, &checkin.CheckinItem{
				BookingID:   dbItem.ID,
				ChildName:   childName,
				IsCheckedIn: dbItem.IsCheckedIn,
			})
		}
	}
	var startTime, endTime time.Time
	startTime = timeutil.ToLocation(data.StartDate, data.Timezone)
	endTime = timeutil.ToLocation(data.EndDate, data.Timezone)

	return &checkin.CheckinPageModel{
		SlotID: data.ID, // Assuming all items are for the same slot
		SlotInfo: fmt.Sprintf("%s %s - %s @ %s",
			formattedDate(startTime),
			startTime.Format("15:04"),
			endTime.Format("15:04"),
			data.Location),
		CheckinItems: checkinItems,
		OnLeaveItems: onleaveItems,
	}
}

func messageCheckinResult(checkinResults *entity.TrainDateHasApptState) string {
	// Construct message
	var startTime, endTime time.Time
	startTime = timeutil.ToLocation(checkinResults.StartDate, checkinResults.Timezone)
	endTime = timeutil.ToLocation(checkinResults.EndDate, checkinResults.Timezone)
	var msgBuilder strings.Builder
	msgBuilder.WriteString("簽到結果通知\n\n")
	msgBuilder.WriteString(fmt.Sprintf("課程：%s %s - %s @ %s\n",
		startTime.Format("01/02 (Mon)"),
		startTime.Format("15:04"),
		endTime.Format("15:04"),
		checkinResults.Location))

	// Count checked in and absent users
	leaveCount := 0
	absentCount := 0
	for _, item := range checkinResults.UserAppointments {
		if item.IsOnLeave {
			leaveCount++
		} else if !item.IsCheckedIn {
			absentCount++
		}
	}
	checkedInCount := len(checkinResults.UserAppointments) - leaveCount - absentCount
	totalBooked := len(checkinResults.UserAppointments) - leaveCount

	checkedInNames := make([]string, 0, checkedInCount)
	absentNames := make([]string, 0, absentCount)
	leaveNames := make([]string, 0, leaveCount)

	for _, item := range checkinResults.UserAppointments {
		if item.IsCheckedIn {
			checkedInNames = append(checkedInNames, item.ChildName)
		} else if item.IsOnLeave {
			leaveNames = append(leaveNames, item.ChildName)
		} else {
			absentNames = append(absentNames, item.ChildName)
		}
	}

	msgBuilder.WriteString(fmt.Sprintf("總預約人數：%d 人\n", totalBooked))
	msgBuilder.WriteString(fmt.Sprintf("已簽到人數：%d 人\n", checkedInCount))
	msgBuilder.WriteString(fmt.Sprintf("未簽到人數：%d 人\n", totalBooked-checkedInCount))
	msgBuilder.WriteString(fmt.Sprintf("請假人數：%d 人\n\n", leaveCount))

	msgBuilder.WriteString("已簽到學員：\n")
	if len(checkedInNames) == 0 {
		msgBuilder.WriteString("- 無\n")
	} else {
		for _, name := range checkedInNames {
			msgBuilder.WriteString("- ")
			msgBuilder.WriteString(name)
			msgBuilder.WriteString("\n")
		}
	}

	msgBuilder.WriteString("\n未簽到學員：\n")
	if len(absentNames) == 0 {
		msgBuilder.WriteString("- 無\n")
	} else {
		for _, name := range absentNames {
			msgBuilder.WriteString("- ")
			msgBuilder.WriteString(name)
			msgBuilder.WriteString("\n")
		}
	}

	msgBuilder.WriteString("\n請假學員：\n")
	if leaveCount == 0 {
		msgBuilder.WriteString("- 無\n")
	} else {
		for _, name := range leaveNames {
			msgBuilder.WriteString("- ")
			msgBuilder.WriteString(name)
			msgBuilder.WriteString("\n")
		}
	}

	return msgBuilder.String()
}
