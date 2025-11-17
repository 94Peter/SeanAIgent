package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/94peter/vulpes/ezapi"
	"github.com/94peter/vulpes/log"
	"github.com/gin-gonic/gin"

	"seanAIgent/internal/db/model"
	"seanAIgent/internal/service"
	"seanAIgent/internal/service/lineliff"
	"seanAIgent/templates"
	"seanAIgent/templates/forms/checkin"
)

type checkinAPI struct {
	svc        service.Service
	enableCSRF bool
}

var initCheckinApiOnce sync.Once

func InitCheckinApi(service service.Service, enableCSRF bool) {
	initCheckinApiOnce.Do(func() {
		api := &checkinAPI{
			svc:        service,
			enableCSRF: enableCSRF,
		}

		ezapi.RegisterGinApi(func(r ezapi.Router) {
			r.GET("/:lang/admin/checkin", api.getCheckinPage)
			r.GET("/admin/checkin", api.getCheckinPage)
			r.POST("/admin/checkin/submit", api.submitCheckin)
		})
	})
}

const tenMin = 10 * time.Minute

func (api *checkinAPI) getCheckinPage(c *gin.Context) {
	lineliffid := lineliff.GetCheckinLiffId() // Assuming same liffId for now

	var viewModel *checkin.CheckinPageModel

	if !isAdmin(c) {
		viewModel = &checkin.CheckinPageModel{
			ErrorMessage: "您沒有權限開啟此頁面。",
		}
	} else {
		// Determine current time slot
		now := time.Now().Add(tenMin)

		// Query bookings for this slot
		checkinList, err := api.svc.QueryCheckinList(c, now)
		// checkinList, err := api.svc.QueryCheckinList(c, time.Date(2025, 11, 4, 14, 0, 0, 0, now.Location()))
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

	viewModel.EnableCSRF = api.enableCSRF

	com := templates.Layout(
		checkin.CheckinPage(viewModel, lineliffid),
		lineliffid,
		&templates.OgMeta{
			Title:       "簽到管理",
			Description: "管理訓練時段簽到",
			Image:       "https://images.pexels.com/photos/2558605/pexels-photo-2558605.jpeg",
		},
	)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func (api *checkinAPI) submitCheckin(c *gin.Context) {
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

	// Update checkin status in DB
	if err := api.svc.UpdateCheckinStatus(c, input.SlotID, input.CheckedInBookingIDs); err != nil {
		log.Errorf("Failed to update checkin status: %v", err)
		addToastTrigger(c, "提交失敗", "更新簽到狀態失敗，請稍後再試。", "error")
		c.Status(http.StatusInternalServerError)
		return
	}

	// Re-query checkin list to get updated status for message
	now := time.Now().Add(tenMin)

	updatedCheckinList, err := api.svc.QueryCheckinList(c, now)
	//updatedCheckinList, err := api.svc.QueryCheckinList(c, time.Date(2025, 11, 4, 14, 0, 0, 0, now.Location()))
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

func modelToCheckinPageModel(data *model.AggrTrainingdateHasCheckinItems) *checkin.CheckinPageModel {
	checkinItems := make([]*checkin.CheckinItem, 0, len(data.CheckinItems))
	onleaveItems := make([]*checkin.CheckinItem, 0, len(data.OnLeaveItems))
	for _, dbItem := range data.CheckinItems {
		childName := dbItem.ChildName
		if childName == "" {
			childName = dbItem.UserName // Fallback to parent's name
		}
		checkinItems = append(checkinItems, &checkin.CheckinItem{
			BookingID:   dbItem.ID.Hex(),
			ChildName:   childName,
			IsCheckedIn: dbItem.IsCheckedIn,
		})
	}
	for _, dbItem := range data.OnLeaveItems {
		childName := dbItem.ChildName
		if childName == "" {
			childName = dbItem.UserName // Fallback to parent's name
		}
		onleaveItems = append(onleaveItems, &checkin.CheckinItem{
			BookingID:   dbItem.ID.Hex(),
			ChildName:   childName,
			IsCheckedIn: dbItem.IsCheckedIn, // On-leave items are considered checked in
		})
	}
	var startTime, endTime time.Time
	startTime = model.ToTime(data.StartDate, data.TimeZone)
	endTime = model.ToTime(data.EndDate, data.TimeZone)

	return &checkin.CheckinPageModel{
		SlotID: data.ID.Hex(), // Assuming all items are for the same slot
		SlotInfo: fmt.Sprintf("%s %s - %s @ %s",
			formattedDate(startTime),
			startTime.Format("15:04"),
			endTime.Format("15:04"),
			data.Location),
		CheckinItems: checkinItems,
		OnLeaveItems: onleaveItems,
	}
}

func messageCheckinResult(checkinResults *model.AggrTrainingdateHasCheckinItems) string {
	// Construct message
	var startTime, endTime time.Time
	startTime = model.ToTime(checkinResults.StartDate, checkinResults.TimeZone)
	endTime = model.ToTime(checkinResults.EndDate, checkinResults.TimeZone)
	var msgBuilder strings.Builder
	msgBuilder.WriteString("簽到結果通知\n\n")
	msgBuilder.WriteString(fmt.Sprintf("課程：%s %s - %s @ %s\n",
		startTime.Format("01/02 (Mon)"),
		startTime.Format("15:04"),
		endTime.Format("15:04"),
		checkinResults.Location))

	totalBooked := len(checkinResults.CheckinItems)
	checkedInCount := 0
	var checkedInNames []string
	var absentNames []string

	for _, item := range checkinResults.CheckinItems {
		if item.IsCheckedIn {
			checkedInCount++
			checkedInNames = append(checkedInNames, item.ChildName)
		} else {
			absentNames = append(absentNames, item.ChildName)
		}
	}

	msgBuilder.WriteString(fmt.Sprintf("總預約人數：%d 人\n", totalBooked))
	msgBuilder.WriteString(fmt.Sprintf("已簽到人數：%d 人\n", checkedInCount))
	msgBuilder.WriteString(fmt.Sprintf("未簽到人數：%d 人\n", totalBooked-checkedInCount))
	msgBuilder.WriteString(fmt.Sprintf("請假人數：%d 人\n\n", len(checkinResults.OnLeaveItems)))

	msgBuilder.WriteString("已簽到學員：\n")
	if len(checkedInNames) == 0 {
		msgBuilder.WriteString("- 無\n")
	} else {
		for _, name := range checkedInNames {
			msgBuilder.WriteString(fmt.Sprintf("- %s\n", name))
		}
	}

	msgBuilder.WriteString("\n未簽到學員：\n")
	if len(absentNames) == 0 {
		msgBuilder.WriteString("- 無\n")
	} else {
		for _, name := range absentNames {
			msgBuilder.WriteString(fmt.Sprintf("- %s\n", name))
		}
	}

	msgBuilder.WriteString("\n請假學員：\n")
	if len(checkinResults.OnLeaveItems) == 0 {
		msgBuilder.WriteString("- 無\n")
	} else {
		for _, item := range checkinResults.OnLeaveItems {
			msgBuilder.WriteString(fmt.Sprintf("- %s\n", item.ChildName))
		}
	}

	return msgBuilder.String()
}
