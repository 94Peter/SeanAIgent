package db

import (
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/infra/db/core"
	"seanAIgent/internal/booking/usecase"

	"github.com/google/wire"
)

var InfraSet = wire.NewSet(
	// 2. 提供具體的實作函數
	// 如果 NewDbRepoAndIdGenerate 需要 *factory.Stores 作為參數，Wire 會自動傳入
	NewDbRepoAndIdGenerate,

	// 3. 重要：綁定介面
	// 假設 UseCase 接收的是 usecase.Repository 介面
	// 我們告訴 Wire：core.DbRepository 滿足了那個介面
	wire.Bind(new(usecase.Repository), new(core.DbRepository)),
	wire.Bind(new(repository.TrainRepository), new(core.DbRepository)),
)
