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

func NewUserMonthlyStatsSubscriber(
	trainRepo repository.TrainRepository,
	statsRepo repository.StatsRepository,
) event.Subscriber {
	handler := func(ctx context.Context, e event.Event, p domain.AppointmentStatusChanged) error {
		// 1. 取得課程資訊以確定年月
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if p.TrainingID == "" {
			log.Warnf("UserMonthlyStatsSubscriber: TrainingID is empty for booking %s, might be legacy event. Skipping.", p.BookingID)
			return nil
		}

		trainDate, err := trainRepo.FindTrainDateByID(bgCtx, p.TrainingID)
		if err != nil {
			return fmt.Errorf("UserMonthlyStatsSubscriber: find traindate fail (ID: %s): %w", p.TrainingID, err)
		}

		year := trainDate.Period().Start().Year()
		month := int(trainDate.Period().Start().Month())

		// 2. 執行核心運算：重新聚合該用戶該月的數據
		stat, err := statsRepo.AggregateUserMonthlyStats(bgCtx, p.UserID, year, month)
		if err != nil {
			return fmt.Errorf("UserMonthlyStatsSubscriber: aggregate fail: %w", err)
		}

		// 3. 資料持久化：寫入預聚合表
		if err := statsRepo.UpsertUserMonthlyStats(bgCtx, stat); err != nil {
			return fmt.Errorf("UserMonthlyStatsSubscriber: upsert fail: %w", err)
		}

		log.Infof("UserMonthlyStatsSubscriber: updated stats for user %s (%d/%d)", p.UserID, year, month)
		return nil
	}

	return event.NewTypedSubscriber("user_monthly_stats_processor", handler)
}
