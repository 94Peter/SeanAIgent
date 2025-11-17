package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"seanAIgent/internal/db/model"
	mymcp "seanAIgent/internal/mcp"
)

func init() {
	mymcp.AddTool(
		mcp.NewTool(
			"create_course_sessions",
			mcp.WithDescription("Create course sessions for a user with multiple schedules"),
			mcp.WithString("line_user_id",
				mcp.Required(),
				mcp.Description("Line user ID of the person creating the sessions"),
			),
			mcp.WithString("location",
				mcp.Description("Location or room of the course sessions"),
			),
			mcp.WithString("time_zone",
				mcp.DefaultString("Asia/Taipei"),
				mcp.Description("Time zone of user's location"),
			),
			mcp.WithNumber("capacity",
				mcp.Required(),
				mcp.Description("Maximum number of participants per session"),
			),
			mcp.WithArray("schedule",
				mcp.Required(),
				mcp.Description("List of session objects, each with date, start_time, end_time"),
				mcp.Items(map[string]any{
					"type": "object",
					"properties": map[string]any{
						"date": map[string]any{
							"type":        "string",
							"description": "Date in YYYY-MM-DD format",
						},
						"start_time": map[string]any{
							"type":        "string",
							"description": "Start time in HH:MM format",
						},
						"end_time": map[string]any{
							"type":        "string",
							"description": "End time in HH:MM format",
						},
					},
					"required": []string{"date", "start_time", "end_time"},
				}),
			),
		),
		mcp.NewTypedToolHandler(createTrainingCoursesHandler),
	)
}

type createTrainingCoursesArgs struct {
	UserId   string      `json:"line_user_id"`
	Location string      `json:"location"`
	TimeZone string      `json:"time_zone"`
	Capacity int         `json:"capacity"`
	Schedule []*schedule `json:"schedule"`
}

type schedule struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func (s *schedule) GetStartTime(timeZone string) (time.Time, error) {
	return toTime(s.Date, s.StartTime, timeZone)
}

func (s *schedule) GetEndTime(timeZone string) (time.Time, error) {
	return toTime(s.Date, s.EndTime, timeZone)
}

func toTime(date, timeStr, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to UTC or local if timezone is invalid
		loc = time.UTC
	}
	return time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("%s %s", date, timeStr), loc)
}

func createTrainingCoursesHandler(ctx context.Context, request mcp.CallToolRequest, args createTrainingCoursesArgs) (*mcp.CallToolResult, error) {
	if trainingDateService == nil {
		return nil, fmt.Errorf("trainingDateService is not initialized")
	}

	trainingDates := make([]*model.TrainingDate, len(args.Schedule))
	var startTime, endTime time.Time
	var err error
	for i, schedule := range args.Schedule {

		startTime, err = schedule.GetStartTime(args.TimeZone)
		if err != nil {
			return nil, err
		}
		endTime, err = schedule.GetEndTime(args.TimeZone)
		if err != nil {
			return nil, err
		}
		trainingDates[i] = model.NewTrainingDate()
		trainingDates[i].Date = schedule.Date
		trainingDates[i].Capacity = args.Capacity
		trainingDates[i].Location = args.Location
		trainingDates[i].Timezone = args.TimeZone
		trainingDates[i].UserID = args.UserId
		trainingDates[i].StartDate = startTime
		trainingDates[i].EndDate = endTime
	}
	_, err = trainingDateService.AddTrainingDates(ctx, trainingDates)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText("ok"), nil
}
