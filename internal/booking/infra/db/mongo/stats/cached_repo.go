package stats

import (
	"context"
	"strconv"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

type cachedStatsRepo struct {
	delegate repository.StatsRepository
	cache    *cache.Cache
	sfGroup  *singleflight.Group
}

func NewCachedStatsRepository(delegate repository.StatsRepository) repository.StatsRepository {
	return &cachedStatsRepo{
		delegate: delegate,
		cache:    cache.New(5*time.Minute, 10*time.Minute),
		sfGroup:  &singleflight.Group{},
	}
}

func (r *cachedStatsRepo) GetAllUserApptStats(
	ctx context.Context, filter repository.FilterUserApptStats,
) ([]*entity.UserApptStats, repository.RepoError) {
	// 暫不對 GetAll 做快取，因為過濾條件可能很多變
	return r.delegate.GetAllUserApptStats(ctx, filter)
}

func (r *cachedStatsRepo) GetUserApptStats(
	ctx context.Context, userID string, filter repository.FilterUserApptStats,
) (*entity.UserApptStats, repository.RepoError) {
	// 僅針對時間範圍過濾做快取 (通常是用於月度統計)
	if f, ok := filter.(repository.FilterUserApptStatsByTrainTimeRange); ok {
		// 優化：改用字串拼接與 strconv 避免 fmt.Sprintf 的反射開銷
		cacheKey := "stats:" + userID + ":" + strconv.Itoa(f.TrainStart.Year()) + ":" + strconv.Itoa(int(f.TrainStart.Month()))
		if val, found := r.cache.Get(cacheKey); found {
			return val.(*entity.UserApptStats), nil
		}

		// 使用 SingleFlight 解決緩存擊穿
		res, err, _ := r.sfGroup.Do(cacheKey, func() (interface{}, error) {
			data, repoErr := r.delegate.GetUserApptStats(ctx, userID, filter)
			if repoErr != nil {
				return nil, repoErr
			}
			r.cache.Set(cacheKey, data, cache.DefaultExpiration)
			return data, nil
		})

		if err != nil {
			return nil, err.(repository.RepoError)
		}
		return res.(*entity.UserApptStats), nil
	}

	return r.delegate.GetUserApptStats(ctx, userID, filter)
}

func (r *cachedStatsRepo) CleanStatsCache(ctx context.Context, userID string, year, month int) repository.RepoError {
	cacheKey := "stats:" + userID + ":" + strconv.Itoa(year) + ":" + strconv.Itoa(month)
	r.cache.Delete(cacheKey)
	return r.delegate.CleanStatsCache(ctx, userID, year, month)
}
