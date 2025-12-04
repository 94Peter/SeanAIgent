package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/94peter/vulpes/log"
	"github.com/mark3labs/mcp-go/mcp"

	mymcp "seanAIgent/internal/mcp"
)

func init() {
	mymcp.AddTool(
		mcp.NewTool("query_training_courses_by_range",
			// Description: Crucial for the LLM Agent's reasoning process
			mcp.WithDescription("Queries the training course schedule for a specific user over a date range. **The query results will be returned based on the specified time zone.** Use this when the user asks about courses, classes, or their schedule."),

			// Parameter 1: Linebot UserId (Required)
			mcp.WithString("line_user_id",
				mcp.Required(),
				mcp.Description("The unique identifier (ID) of the Linebot user."),
			),

			// Parameter 2: Start Date (Optional)
			mcp.WithString("start_date",
				mcp.Description("The start date of the period to query. Use the standardized YYYY-MM-DD format (e.g., 2025-11-01). If omitted, the query starts from today."),
			),

			// Parameter 3: End Date (Optional)
			mcp.WithString("end_date",
				mcp.Description("The end date of the period to query. Use the standardized YYYY-MM-DD format (e.g., 2025-11-08). If omitted, the query defaults to the same day as the start_date."),
			),

			// Parameter 4: Time Zone (Optional) - NEW
			mcp.WithString("time_zone",
				mcp.Description("The time zone for the query, using an IANA Time Zone Database name (e.g., 'Asia/Taipei', 'America/New_York', or 'Europe/London'). If omitted, the server's default time zone will be assumed."),
			),
		),
		queryTrainingCoursesByRangeHandler,
	)
}

func queryTrainingCoursesByRangeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Info("queryTrainingCoursesByRangeHandler")
	if trainingDateService == nil {
		log.Error("trainingDateService is not initialized")
		return nil, fmt.Errorf("trainingDateService is not initialized")
	}

	startDate, err := request.RequireString("start_date")
	if err != nil {
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	endDate, err := request.RequireString("end_date")
	if err != nil {
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	timeZone, err := request.RequireString("time_zone")
	if err != nil {
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	timeZoneLocation, err := time.LoadLocation(timeZone)
	if err != nil {
		log.Err(err)
		return nil, err
	}

	startDateTime, err := time.ParseInLocation("2006-01-02", startDate, timeZoneLocation)
	if err != nil {
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	endDateTime, err := time.ParseInLocation("2006-01-02", endDate, timeZoneLocation)
	if err != nil {
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	if endDateTime.Equal(startDateTime) {
		endDateTime = startDateTime.Add(24*time.Hour - time.Second)
	}
	data, err := trainingDateService.QueryDateTimeRangeTrainingDate(ctx, startDateTime, endDateTime)

	if err != nil {
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, _ := json.Marshal(data)
	return mcp.NewToolResultText(string(result)), nil
}
