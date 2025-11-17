package mcp

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var tools []server.ServerTool

func AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	tools = append(tools, server.ServerTool{Tool: tool, Handler: handler})
}

func Start() {
	s := server.NewMCPServer(
		"Calculator Demo",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	s.AddTools(tools...)
	fmt.Println("starting server")
	server.NewStreamableHTTPServer(s).Start(":9080")
}
