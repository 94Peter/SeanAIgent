package tool

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	writeTrain "seanAIgent/internal/booking/usecase/traindate/write"
	"seanAIgent/internal/util/timeutil"

	"github.com/mark3labs/mcp-go/server"
)

func ProvideCreateTrainingCoursesTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool(
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
		Handler: mcp.NewTypedToolHandler(createTrainingCoursesHandler),
	}
}

type createTrainingCoursesArgs struct {
	UserId   string      `json:"line_user_id"`
	Location string      `json:"location"`
	TimeZone string      `json:"time_zone"`
	Schedule []*schedule `json:"schedule"`
	Capacity int         `json:"capacity"`
}

type schedule struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func (s *schedule) GetStartTime(timeZone string) (time.Time, error) {
	return timeutil.ParseDateTime(s.Date, s.StartTime, timeZone)
}

func (s *schedule) GetEndTime(timeZone string) (time.Time, error) {
	return timeutil.ParseDateTime(s.Date, s.EndTime, timeZone)
}

func createTrainingCoursesHandler(ctx context.Context, request mcp.CallToolRequest, args createTrainingCoursesArgs) (*mcp.CallToolResult, error) {
	var startTime, endTime time.Time
	var err error
	reqs := make([]writeTrain.ReqCreateTrainDate, len(args.Schedule))
	for i, schedule := range args.Schedule {
		startTime, err = schedule.GetStartTime(args.TimeZone)
		if err != nil {
			return nil, err
		}
		endTime, err = schedule.GetEndTime(args.TimeZone)
		if err != nil {
			return nil, err
		}

		reqs[i] = writeTrain.ReqCreateTrainDate{
			StartTime: startTime,
			EndTime:   endTime,
			CoachID:   args.UserId,
			Location:  args.Location,
			Capacity:  args.Capacity,
		}
	}
	_, err = batchCreateTrainDateUC.Execute(ctx, reqs)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText("ok"), nil
}
