package handler

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/94peter/vulpes/ezapi"
	"github.com/94peter/vulpes/log"
	"github.com/gin-gonic/gin"

	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase"
	uccore "seanAIgent/internal/booking/usecase/core"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	writeTrain "seanAIgent/internal/booking/usecase/traindate/write"
	"seanAIgent/internal/db/model"
	"seanAIgent/internal/service/lineliff"
	"seanAIgent/templates"
	"seanAIgent/templates/forms/manageTrainingDate"
)

func NewTrainingUseCaseSet(registry *usecase.Registry) TrainingUseCaseSet {
	return TrainingUseCaseSet{}
}

type TrainingUseCaseSet struct {
	batchCreateTrainDateUC uccore.WriteUseCase[[]writeTrain.ReqCreateTrainDate, []*entity.TrainDate]
	deleteTrainDateUC      uccore.WriteUseCase[writeTrain.ReqDeleteTrainDate, *entity.TrainDate]
	queryFutureTrainUC     uccore.ReadUseCase[readTrain.ReqQueryFutureTrain, []*entity.TrainDateHasApptState]
}

func NewTrainingApi(enableCSRF bool, schedule TrainingUseCaseSet) WebAPI {
	return &trainingAPI{
		TrainingUseCaseSet: schedule,
		enableCSRF:         enableCSRF,
	}
}

type trainingAPI struct {
	TrainingUseCaseSet
	enableCSRF bool
	once       sync.Once
}

func (api *trainingAPI) InitRouter(r ezapi.Router) {
	api.once.Do(func() {
		r.GET("/training", api.getForm)
		r.GET("/:lang/training", api.getForm)

		r.POST("/training-date/add", api.addTrainingDate)
		r.POST("/training-date/delete", api.deleteTrainingDate)
	})
}

func (api *trainingAPI) postErrorHandler(c *gin.Context, showErr error) {
	log.Warnf("post error: %v", showErr)
	dbTrainingDate, err := api.queryFutureTrainUC.Execute(c.Request.Context(), readTrain.ReqQueryFutureTrain{
		TimeAfter: time.Now(),
	})
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
	dbTrainingDates, err := inputTrainingDateToReqCreateTrainDateSlice(&input, userId)
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
	_, err = api.batchCreateTrainDateUC.Execute(c.Request.Context(), dbTrainingDates)
	if err != nil {
		api.postErrorHandler(c, err)
		return
	}
	trainingDates, err := api.queryFutureTrainUC.Execute(
		c.Request.Context(),
		readTrain.ReqQueryFutureTrain{
			TimeAfter: time.Now(),
		},
	)
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
	_, err = api.deleteTrainDateUC.Execute(c.Request.Context(), writeTrain.ReqDeleteTrainDate{
		TrainDateID: input.ID,
		UserID:      getUserID(c),
	})
	if err != nil {
		api.postErrorHandler(c, err)
		return
	}
	c.Writer.WriteHeader(http.StatusOK)
}

func (api *trainingAPI) getForm(c *gin.Context) {
	lineliffid := lineliff.GetTrainingDataLiffId()
	isAdmin := isAdmin(c)

	dbTrainingDate, err := api.queryFutureTrainUC.Execute(
		c.Request.Context(),
		readTrain.ReqQueryFutureTrain{
			TimeAfter: time.Now(),
		},
	)
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
			Image:       "https://storage.94peter.dev/cdn-cgi/image/width=1200,height=630,quality=80,format=auto/https://storage.94peter.dev/images/UAC.png",
		},
	)
	r := newTemplRenderer(c.Request.Context(), http.StatusOK, com)
	c.Render(http.StatusOK, r)
}

func sliceDbTrainingDateToTrainingDate(dbTrainingDate []*entity.TrainDateHasApptState) []*manageTrainingDate.TrainingDate {
	trainingDate := make([]*manageTrainingDate.TrainingDate, len(dbTrainingDate))
	for i, v := range dbTrainingDate {
		trainingDate[i] = dbTrainingDateToTrainingDate(v)
	}
	return trainingDate
}

func dbTrainingDateToTrainingDate(dbTrainingDate *entity.TrainDateHasApptState) *manageTrainingDate.TrainingDate {
	startDate := model.ToTime(dbTrainingDate.StartDate, dbTrainingDate.Timezone)
	endDate := model.ToTime(dbTrainingDate.EndDate, dbTrainingDate.Timezone)
	return &manageTrainingDate.TrainingDate{
		ID:            dbTrainingDate.ID,
		Date:          startDate.Format("2006/01/02"),
		Start:         startDate.Format("15:04"),
		End:           endDate.Format("15:04"),
		Location:      dbTrainingDate.Location,
		Capacity:      dbTrainingDate.Capacity,
		BookedCount:   len(dbTrainingDate.UserAppointments),
		FormattedDate: formattedDate(startDate),
	}
}

func inputTrainingDateToReqCreateTrainDateSlice(
	trainingDate *manageTrainingDate.InputTrainingDate, userId string,
) ([]writeTrain.ReqCreateTrainDate, error) {
	dateTimeRangeSlice, err := trainingDate.Dates()
	if err != nil {
		return nil, err
	}
	dbTrainingDates := make([]writeTrain.ReqCreateTrainDate, len(dateTimeRangeSlice))
	for i, v := range dateTimeRangeSlice {
		dbTrainingDate := writeTrain.ReqCreateTrainDate{
			CoachID:   userId,
			Location:  trainingDate.Location,
			Capacity:  trainingDate.Capacity,
			StartTime: v.Start,
			EndTime:   v.End,
		}
		dbTrainingDates[i] = dbTrainingDate
	}
	return dbTrainingDates, nil
}

// func formattedDate(date time.Time) string {
// 	return date.Format("2006/01/02 (Mon)")
// }

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
