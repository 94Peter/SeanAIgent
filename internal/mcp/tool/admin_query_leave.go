package tool

import (
	"context"
	"encoding/json"
	"fmt"
	mymcp "seanAIgent/internal/mcp"
	"time"

	"github.com/94peter/vulpes/log"
	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	mymcp.AddTool(
		mcp.NewTool("query_all_courses_and_absence_records_on_date",
			// Description: Crucial for the LLM Agent's reasoning process
			mcp.WithDescription("Retrieves a list of all scheduled training courses on a specific date, including the total enrollment and the corresponding list of participants who have submitted absence requests (leave records) for each course. Use this when the user asks for a summary of courses and who is absent on a particular date."),

			mcp.WithString("date",
				mcp.Description("The specific date to query all courses and their absence records for. Use the standardized YYYY-MM-DD format (e.g., 2025-12-04). This parameter is required."),
			),
			mcp.WithString("time_zone",
				mcp.Description("The time zone for the query, using an IANA Time Zone Database name (e.g., 'Asia/Taipei', 'America/New_York', or 'Europe/London'). If omitted, the server's default time zone will be assumed."),
			),
		),
		queryLeaveByDateHandler,
	)
}

func queryLeaveByDateHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Info("queryLeaveByDateHandler")
	if trainingDateService == nil {
		log.Error("trainingDateService is not initialized")
		return nil, fmt.Errorf("trainingDateService is not initialized")
	}

	date, err := request.RequireString("date")
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

	startDateTime, err := time.ParseInLocation("2006-01-02", date, timeZoneLocation)
	if err != nil {
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, err := trainingDateService.QueryLeaveByTrainingDate(ctx, startDateTime)

	if err != nil {
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	result, _ := json.Marshal(data)
	return mcp.NewToolResultText(string(result)), nil
}
