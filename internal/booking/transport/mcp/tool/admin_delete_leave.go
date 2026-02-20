package tool

import (
	"context"
	"fmt"
	writeAppt "seanAIgent/internal/booking/usecase/appointment/write"

	"github.com/94peter/vulpes/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func ProvideDeleteLeaveByIdTool() server.ServerTool {
	return server.ServerTool{
		Tool: mcp.NewTool("delete_absence_records_by_id",
			// Description: Crucial for the LLM Agent's reasoning process
			mcp.WithDescription("Deletes (cancels) a specific submitted absence request (leave record) using its unique Absence ID. This is used to cancel a leave application. Requires the Absence ID, which can often be retrieved using 'query_all_courses_and_absence_records_on_date' first."),

			mcp.WithString("absence_id",
				mcp.Description("The unique identifier (ID) of the absence record to be deleted/cancelled. This is a mandatory parameter."),
				mcp.Required(),
			),
			mcp.WithString("line_user_id",
				mcp.Description("The unique identifier (ID) of the Linebot user."),
				mcp.Required(),
			),
		),
		Handler: deleteLeaveByIdHandler,
	}
}

func deleteLeaveByIdHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Info("deleteLeaveByIdHandler")

	id, err := request.RequireString("absence_id")
	if err != nil {
		err = fmt.Errorf("require id error: %w", err)
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	lineUserID, err := request.RequireString("line_user_id")
	if err != nil {
		err = fmt.Errorf("require line_user_id error: %w", err)
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	_, err = cancelLeaveUC.Execute(ctx, writeAppt.ReqCancelLeave{
		ApptID: id,
		UserID: lineUserID,
	})
	if err != nil {
		err = fmt.Errorf("cancel leave error: %w", err)
		log.Err(err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText("ok"), nil
}
