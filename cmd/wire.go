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
	"seanAIgent/internal/booking/transport/mcp"
	"seanAIgent/internal/booking/transport/mcp/tool"

	"github.com/google/wire"
	"github.com/mark3labs/mcp-go/server"
	"github.com/94peter/vulpes/db/mgo"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func ProvideDatabase() *mongo.Database {
	return mgo.GetDatabase()
}

func InitializeWeb() (web.WebService, error) {
	wire.Build(
		ProvideDatabase,
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
	return nil, nil
}

func toolSet() []server.ServerTool {
	return []server.ServerTool{
		tool.ProvideCreateTrainingCoursesTool(),
		tool.ProvideDeleteLeaveByIdTool(),
		tool.ProvideQueryLeaveByDateTool(),
		tool.ProvideQueryTrainingByRangeTool(),
	}
}

func InitializeMCP() (mcp.Server, error) {
	wire.Build(
		ProvideDatabase,
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
	return nil, nil
}

func GetUseCaseRegistry() (*usecase.Registry, error) {
	wire.Build(
		ProvideDatabase,
		db.InfraSet,
		service.NewTrainDateService,
		usecase.UseCaseSet,
	)
	return nil, nil
}
