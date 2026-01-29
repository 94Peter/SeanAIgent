package db

import (
	"seanAIgent/internal/booking/infra/db/core"
	"seanAIgent/internal/booking/infra/db/mongo"
)

func NewDbRepoAndIdGenerate() core.DbRepository {
	return mongo.NewRepoAndIdGenerate()
}
