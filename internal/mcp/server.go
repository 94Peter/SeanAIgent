package mcp

import (
	"seanAIgent/internal/booking/usecase"
	"seanAIgent/internal/service"

	"github.com/94peter/vulpes/log"
	"github.com/mark3labs/mcp-go/server"
)

type Server interface {
	Start()
}

type mcpServer struct {
	s *server.MCPServer
}

func (s *mcpServer) Start() {
	log.Info("Starting MCP server on port 9080")
	server.NewStreamableHTTPServer(s.s).Start(":9080")
}

func InitMcpServer(svc service.TrainingDateService, registry *usecase.Registry, tools []server.ServerTool) Server {
	s := server.NewMCPServer(
		"Calculator Demo",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	s.AddTools(tools...)
	return &mcpServer{s: s}
}
