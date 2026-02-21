package usecase

import (
	"context"
	"time"

	"github.com/94peter/vulpes/log"
)

type cleaningJob struct {
	UserID      string
	TrainingID string
}

type CacheWorker interface {
	Clean(uid string, tid string)
	// CleanSync 進行同步清理，確保關鍵資料即時失效
	// 直接傳入 startTime 避免再次查詢資料庫
	CleanSync(ctx context.Context, uid string, tid string, startTime time.Time)
	Start(ctx context.Context)
}

type cacheWorker struct {
	repo    Repository
	jobChan chan cleaningJob
}

func NewCacheWorker(repo Repository) CacheWorker {
	return &cacheWorker{
		repo:    repo,
		jobChan: make(chan cleaningJob, 1000),
	}
}

func (w *cacheWorker) Clean(uid string, tid string) {
	select {
	case w.jobChan <- cleaningJob{UserID: uid, TrainingID: tid}:
	default:
		// Channel full, drop or handle accordingly.
	}
}

func (w *cacheWorker) CleanSync(ctx context.Context, uid string, tid string, startTime time.Time) {
	// 1. 清理該用戶的排程快取 (精準清理)
	_ = w.repo.CleanTrainCache(ctx, uid)
	
	// 2. 清理統計快取 (同步呼叫確保一致性)
	_ = w.repo.CleanStatsCache(ctx, uid, startTime.Year(), int(startTime.Month()))
}

func (w *cacheWorker) Start(ctx context.Context) {
	// Start 5 workers
	for i := 0; i < 5; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("CacheWorker Panic Recovered: %v", r)
				}
			}()
			for {
				select {
				case job := <-w.jobChan:
					w.process(job)
				case <-ctx.Done():
					// 收到退出訊號，退出
					return
				}
			}
		}()
	}
}

func (w *cacheWorker) process(job cleaningJob) {
	bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	trainDate, err := w.repo.FindTrainDateByID(bgCtx, job.TrainingID)
	if err == nil {
		w.CleanSync(bgCtx, job.UserID, job.TrainingID, trainDate.Period().Start())
	}
}
