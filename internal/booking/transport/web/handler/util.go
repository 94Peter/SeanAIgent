package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/94peter/botreplyer/provider/line/mid"
	"github.com/gin-gonic/gin"

	"seanAIgent/internal/util/timeutil"
)

func getUserID(c *gin.Context) string {
	return mid.GetLineLiffUserId(c)
}

func getUserDisplayName(c *gin.Context) string {
	return mid.GetLineLiffUserName(c)
}

func isAdmin(c *gin.Context) bool {
	return mid.IsAdmin(c)
}

func formattedDate(date time.Time) string {
	return date.Format("2006/01/02 (Mon)")
}

func TrainDateRangeFormat(start, end time.Time, timezone string) string {
	start = timeutil.ToLocation(start, timezone)
	end = timeutil.ToLocation(end, timezone)
	return start.Format("01/02 15:04") + " - " + end.Format("15:04")
}

func getAllErrorMessage(err error) string {
	var msgs []string
	for err != nil {
		msgs = append(msgs, err.Error())
		err = errors.Unwrap(err) // 取得下一層錯誤
	}
	return strings.Join(msgs, ": ")
}
