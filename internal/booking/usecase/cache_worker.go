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
		// Channel full, drop or handle accordingly. For cache cleaning, dropping is usually acceptable
		// but we might want to log it.
	}
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
		trainingStart := trainDate.Period().Start()
		_ = w.repo.CleanStatsCache(bgCtx, job.UserID, trainingStart.Year(), int(trainingStart.Month()))
		_ = w.repo.CleanTrainCache(bgCtx, job.UserID)
	}
}
