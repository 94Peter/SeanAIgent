package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/94peter/vulpes/ezapi"
	"github.com/94peter/vulpes/log"
	"github.com/gin-gonic/gin"

	"seanAIgent/internal/db/model"
	"seanAIgent/internal/service"
	"seanAIgent/internal/service/lineliff"
	"seanAIgent/templates"
	"seanAIgent/templates/forms/manageTrainingDate"
)

type trainingAPI struct {
	svc        service.Service
	enableCSRF bool
}

var initTrainingApiOnce sync.Once

func InitTrainingApi(service service.Service, enableCSRF bool) {
	initTrainingApiOnce.Do(func() {
		api := &trainingAPI{
			svc:        service,
			enableCSRF: enableCSRF,
		}

		ezapi.RegisterGinApi(func(r ezapi.Router) {
			// 建立活動表單
			r.GET("/training", api.getForm)
			r.GET("/:lang/training", api.getForm)

			r.POST("/training-date/add", api.addTrainingDate)
			r.POST("/training-date/delete", api.deleteTrainingDate)
		})
	})
}

func (api *trainingAPI) postErrorHandler(c *gin.Context, showErr error) {
	log.Warnf("post error: %v", showErr)
	dbTrainingDate, err := api.svc.FutureTrainingDate(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	r := newTemplRenderer(
		c.Request.Context(), http.StatusOK,
		manageTrainingDate.AllTrainingDateRow(sliceDbTrainingDateToTrainingDate(dbTrainingDate), true, api.enableCSRF),
	)
	c.Render(http.StatusOK, r)
	addToastTrigger(c, "操作失敗", showErr.Error(), "error")
}

func (api *trainingAPI) addTrainingDate(c *gin.Context) {
	var input manageTrainingDate.InputTrainingDate
	err := c.ShouldBind(&input)
	if err != nil {
		api.postErrorHandler(c, err)
		return
	}
	userId := getUserID(c)
	if userId == "" {
		api.postErrorHandler(c, fmt.Errorf("user not logged in"))
		return
	}
	dbTrainingDates, err := inputTrainingDateToDbTrainingDate(&input, userId)
	if err != nil {
		api.postErrorHandler(c, err)
		return
	}
	for _, v := range dbTrainingDates {
		if err = v.Validate(); err != nil {
			api.postErrorHandler(c, err)
			return
		}
	}
	trainingDates, err := api.svc.AddTrainingDates(c, dbTrainingDates)
	if err != nil {
		api.postErrorHandler(c, err)
		return
	}

	r := newTemplRenderer(
		c.Request.Context(), http.StatusOK,
		manageTrainingDate.AllTrainingDateRow(sliceDbTrainingDateToTrainingDate(trainingDates), true, api.enableCSRF),
	)
	c.Render(http.StatusOK, r)
}

func (api *trainingAPI) deleteTrainingDate(c *gin.Context) {
	var input manageTrainingDate.InputDeleteTrainingDate
	err := c.ShouldBind(&input)
	if err != nil {
		api.postErrorHandler(c, err)
		return
	}
	if err = api.svc.DeleteTrainingDate(c, input.ID); err != nil {
		api.postErrorHandler(c, err)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}

func (api *trainingAPI) getForm(c *gin.Context) {
	lineliffid := lineliff.GetTrainingDataLiffId()
	isAdmin := isAdmin(c)

	dbTrainingDate, err := api.svc.FutureTrainingDate(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defaultDate := roundUpToNearest5Min(time.Now())
	com := templates.Layout(
		manageTrainingDate.MultiTrainingDateForm(
			&manageTrainingDate.MultiTrainingDateFormModel{
				Dates: sliceDbTrainingDateToTrainingDate(dbTrainingDate),
				DefaultDate: &manageTrainingDate.InputTrainingDate{
					Date:  defaultDate.Format("2006-01-02"),
					Start: defaultDate.Format("15:04"),
					End:   defaultDate.Add(1 * time.Hour).Format("15:04"),
				},
				IsAdmin:    isAdmin,
				LiffId:     lineliffid,
				EnableCSRF: api.enableCSRF,
			},
		), lineliffid,
		&templates.OgMeta{
			Title:       "Sean AIgent",
			Description: "Sean 訓練時程管理服務",
			Image:       "https://images.pexels.com/photos/2558605/pexels-photo-2558605.jpeg",
		},
	)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func sliceDbTrainingDateToTrainingDate(dbTrainingDate []*model.AggrTrainingDateHasAppoint) []*manageTrainingDate.TrainingDate {
	trainingDate := make([]*manageTrainingDate.TrainingDate, len(dbTrainingDate))
	for i, v := range dbTrainingDate {
		trainingDate[i] = dbTrainingDateToTrainingDate(v)
	}
	return trainingDate
}

func dbTrainingDateToTrainingDate(dbTrainingDate *model.AggrTrainingDateHasAppoint) *manageTrainingDate.TrainingDate {
	startDate := model.ToTime(dbTrainingDate.StartDate, dbTrainingDate.Timezone)
	endDate := model.ToTime(dbTrainingDate.EndDate, dbTrainingDate.Timezone)
	return &manageTrainingDate.TrainingDate{
		ID:            dbTrainingDate.ID.Hex(),
		Date:          startDate.Format("2006/01/02"),
		Start:         startDate.Format("15:04"),
		End:           endDate.Format("15:04"),
		Location:      dbTrainingDate.Location,
		Capacity:      dbTrainingDate.Capacity,
		BookedCount:   dbTrainingDate.TotalAppointments,
		FormattedDate: formattedDate(startDate),
	}
}

func inputTrainingDateToDbTrainingDate(trainingDate *manageTrainingDate.InputTrainingDate, userId string) ([]*model.TrainingDate, error) {
	dateTimeRangeSlice, err := trainingDate.Dates()
	if err != nil {
		return nil, err
	}
	dbTrainingDates := make([]*model.TrainingDate, len(dateTimeRangeSlice))
	for i, v := range dateTimeRangeSlice {
		dbTrainingDate := model.NewTrainingDate()
		dbTrainingDate.UserID = userId
		dbTrainingDate.Date = v.Start.Format("2006-01-02")
		dbTrainingDate.Location = trainingDate.Location
		dbTrainingDate.Capacity = trainingDate.Capacity
		dbTrainingDate.Timezone = trainingDate.Timezone
		dbTrainingDate.StartDate = v.Start
		dbTrainingDate.EndDate = v.End
		dbTrainingDates[i] = dbTrainingDate
	}
	return dbTrainingDates, nil
}

func formattedDate(date time.Time) string {
	return date.Format("2006/01/02 (Mon)")
}

func roundUpToNearest5Min(t time.Time) time.Time {
	min := t.Minute()
	// 計算需要加的分鐘數，使分鐘能被 5 整除
	add := 5 - (min % 5)
	if add == 5 {
		add = 0 // 已經是整除
	}
	// 先清除秒和納秒
	t = t.Truncate(time.Minute)
	return t.Add(time.Duration(add) * time.Minute)
}
