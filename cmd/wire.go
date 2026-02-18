//go:build wireinject
// +build wireinject

//go:generate wire
package cmd

import (
	"seanAIgent/internal/booking/domain/service"
	"seanAIgent/internal/booking/infra/db"
	"seanAIgent/internal/booking/transport/web"
	"seanAIgent/internal/booking/transport/web/handler"
	"seanAIgent/internal/booking/usecase"
	"seanAIgent/internal/mcp"
	"seanAIgent/internal/mcp/tool"
	v1service "seanAIgent/internal/service"

	"github.com/google/wire"
	"github.com/mark3labs/mcp-go/server"
)

func InitializeWeb() web.WebService {
	wire.Build(
		// 1. 提供資料庫與 Repo (內含 wire.Bind 介面綁定)
		db.InfraSet,

		// 2. 提供 Domain Service
		service.NewTrainDateService,

		// 3. 提供包裝過的 UseCase 與 Registry
		usecase.UseCaseSet,

		// 4. 提供 API 需要的 UseCaseSet
		handler.NewBookingUseCaseSet,
		handler.NewTrainingUseCaseSet,
		handler.NewV2BookingUseCaseSet,
		// 5. 提供 WebConfig
		ProvideWebConfig,
		// 5. 提供 WebService
		web.InitWeb,
	)
	return nil
}

func toolSet() []server.ServerTool {
	return []server.ServerTool{
		tool.ProvideCreateTrainingCoursesTool(),
		tool.ProvideDeleteLeaveByIdTool(),
		tool.ProvideQueryLeaveByDateTool(),
		tool.ProvideQueryTrainingByRangeTool(),
	}
}

func InitializeMCP(svc v1service.TrainingDateService) mcp.Server {
	wire.Build(
		// 1. 提供資料庫與 Repo (內含 wire.Bind 介面綁定)
		db.InfraSet,

		// 2. 提供 Domain Service
		service.NewTrainDateService,

		// 3. 提供包裝過的 UseCase 與 Registry
		usecase.UseCaseSet,
		toolSet,
		// 4. 提供 MCP 需要的 UseCaseSet
		mcp.InitMcpServer,
	)
	return nil
}

func GetMigrationUseCaseSet() usecase.MigrationRegistry {
	wire.Build(
		// 1. 提供資料庫與 Repo (內含 wire.Bind 介面綁定)
		db.InfraSet,

		// 2. 提供包裝過的 UseCase 與 Registry
		usecase.MigrationUseCaseSet,
	)
	return usecase.MigrationRegistry{}
}
