package infra

import (
	"context"
	"fmt"
	"seanAIgent/internal/booking/domain"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/event"
	"time"

	"github.com/94peter/vulpes/log"
)

type cacheSubscriber struct {
	repo repository.TrainRepository
	// 這裡可能還需要 StatsRepository 或在某個組合後的介面
	statsRepo repository.StatsRepository
}

func NewCacheSubscriber(repo repository.TrainRepository, statsRepo repository.StatsRepository) event.Subscriber {
	handler := func(ctx context.Context, e event.Event, p domain.AppointmentStatusChanged) error {
		// 背景清理任務
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if p.TrainingID == "" {
			log.Warnf("CacheSubscriber: TrainingID is empty for booking %s, might be legacy event. Skipping.", p.BookingID)
			return nil
		}

		trainDate, err := repo.FindTrainDateByID(bgCtx, p.TrainingID)
		if err != nil {
			return fmt.Errorf("CacheSubscriber: find traindate fail (ID: %s): %w", p.TrainingID, err)
		}

		startTime := trainDate.Period().Start()
		
		// 1. 清理該用戶的排程快取
		_ = repo.CleanTrainCache(bgCtx, p.UserID)
		
		// 2. 清理統計快取
		_ = statsRepo.CleanStatsCache(bgCtx, p.UserID, startTime.Year(), int(startTime.Month()))
		
		log.Infof("CacheSubscriber: cleaned cache for user %s, training %s", p.UserID, p.TrainingID)
		return nil
	}

	return event.NewTypedSubscriber("cache_worker_v2", handler)
}
