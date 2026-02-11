package handler

import (
	"time"

	"github.com/94peter/botreplyer/provider/line/mid"
	"github.com/gin-gonic/gin"
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
	start = toTime(start, timezone)
	end = toTime(end, timezone)
	return start.Format("01/02 15:04") + " - " + end.Format("15:04")
}

func toTime(t time.Time, timezone string) time.Time {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}
	return t.In(loc)
}
