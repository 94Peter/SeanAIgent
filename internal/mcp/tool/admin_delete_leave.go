package tool

import (
	"context"
	"fmt"
	mymcp "seanAIgent/internal/mcp"

	"github.com/94peter/vulpes/log"
	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	mymcp.AddTool(
		mcp.NewTool("delete_absence_records_by_id",
			// Description: Crucial for the LLM Agent's reasoning process
			mcp.WithDescription("Deletes (cancels) a specific submitted absence request (leave record) using its unique Absence ID. This is used to cancel a leave application. Requires the Absence ID, which can often be retrieved using 'query_all_courses_and_absence_records_on_date' first."),

			mcp.WithString("absence_id",
				mcp.Description("The unique identifier (ID) of the absence record to be deleted/cancelled. This is a mandatory parameter."),
			),
		),
		deleteLeaveByIdHandler,
	)
}

func deleteLeaveByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Info("queryTrainingCoursesByRangeHandler")
	if trainingDateService == nil {
		log.Error("trainingDateService is not initialized")
		return nil, fmt.Errorf("trainingDateService is not initialized")
	}

	id, err := request.RequireString("id")
	if err != nil {
		err = fmt.Errorf("require id error: %w", err)
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	err = trainingDateService.CancelLeave(ctx, id)
	if err != nil {
		err = fmt.Errorf("cancel leave error: %w", err)
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText("ok"), nil
}
