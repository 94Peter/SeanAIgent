package write

import (
	"context"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqBatchSyncMonthlyStats struct {
	Year  int
	Month int
}

type RespBatchSyncMonthlyStats struct {
	TotalProcessed int
}

type BatchSyncMonthlyStatsUseCase core.WriteUseCase[ReqBatchSyncMonthlyStats, *RespBatchSyncMonthlyStats]

type batchSyncMonthlyStatsUseCaseRepo interface {
	repository.StatsRepository
}

func NewBatchSyncMonthlyStatsUseCase(repo batchSyncMonthlyStatsUseCaseRepo) BatchSyncMonthlyStatsUseCase {
	return &batchSyncMonthlyStatsUseCase{repo: repo}
}

type batchSyncMonthlyStatsUseCase struct {
	repo batchSyncMonthlyStatsUseCaseRepo
}

func (uc *batchSyncMonthlyStatsUseCase) Name() string {
	return "BatchSyncMonthlyStats"
}

func (uc *batchSyncMonthlyStatsUseCase) Execute(
	ctx context.Context, req ReqBatchSyncMonthlyStats,
) (*RespBatchSyncMonthlyStats, core.UseCaseError) {
	// 1. 一次性聚合該月所有用戶的數據 (比 N 次單獨聚合快得多)
	stats, err := uc.repo.AggregateAllUsersMonthlyStats(ctx, req.Year, req.Month)
	if err != nil {
		return nil, core.NewDBError("BATCH_SYNC_STATS", "AGGREGATE_FAIL", "aggregate all users fail", core.ErrInternal).Wrap(err)
	}

	if len(stats) == 0 {
		return &RespBatchSyncMonthlyStats{TotalProcessed: 0}, nil
	}

	// 2. 使用 BulkWrite 一次性存入 (比 N 次 UpdateOne 快得多)
	if err := uc.repo.UpsertManyUserMonthlyStats(ctx, stats); err != nil {
		return nil, core.NewDBError("BATCH_SYNC_STATS", "BULK_UPSERT_FAIL", "bulk upsert fail", core.ErrInternal).Wrap(err)
	}

	return &RespBatchSyncMonthlyStats{
		TotalProcessed: len(stats),
	}, nil
}
