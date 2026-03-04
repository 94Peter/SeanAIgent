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
) []event.Subscriber {
	// 1. 處理單一預約狀態變更
	statusChangeHandler := func(ctx context.Context, e event.Event, p domain.AppointmentStatusChanged) error {
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

		return aggregateAndUpsert(bgCtx, statsRepo, p.UserID, year, month)
	}

	// 2. 處理批次更新後的刷新請求
	refreshHandler := func(ctx context.Context, e event.Event, p domain.UserStatsRefreshRequested) error {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return aggregateAndUpsert(bgCtx, statsRepo, p.UserID, p.Year, p.Month)
	}

	return []event.Subscriber{
		event.NewTypedSubscriber("user_monthly_stats_status_change", domain.TopicAppointmentStatusChanged, statusChangeHandler),
		event.NewTypedSubscriber("user_monthly_stats_refresh", domain.TopicUserStatsRefreshRequested, refreshHandler),
	}
}

func aggregateAndUpsert(ctx context.Context, repo repository.StatsRepository, userID string, year, month int) error {
	stat, err := repo.AggregateUserMonthlyStats(ctx, userID, year, month)
	if err != nil {
		return fmt.Errorf("aggregate fail: %w", err)
	}

	if err := repo.UpsertUserMonthlyStats(ctx, stat); err != nil {
		return fmt.Errorf("upsert fail: %w", err)
	}

	log.Infof("StatsSubscriber: updated stats for user %s (%d/%d)", userID, year, month)
	return nil
}
