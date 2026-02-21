package train

import (
	"context"
	"fmt"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

type cachedTrainRepo struct {
	delegate repository.TrainRepository
	cache    *cache.Cache
	sfGroup  *singleflight.Group
}

func NewCachedTrainRepository(delegate repository.TrainRepository) repository.TrainRepository {
	return &cachedTrainRepo{
		delegate: delegate,
		cache:    cache.New(5*time.Minute, 10*time.Minute),
		sfGroup:  &singleflight.Group{},
	}
}

func (r *cachedTrainRepo) UserQueryTrainDateHasApptState(
	ctx context.Context, userID string, filter repository.FilterTrainDate,
) ([]*entity.TrainDateHasUserApptState, repository.RepoError) {
	// 僅對按時間範圍的查詢做快取
	if f, ok := filter.(repository.FilterTrainingDateByTimeRange); ok {
		// 建立 Cache Key: schedule:userID:startTime:endTime
		cacheKey := fmt.Sprintf("schedule:%s:%s:%s", userID, f.StartTime.Format("2006-01-02"), f.EndTime.Format("2006-01-02"))
		if val, found := r.cache.Get(cacheKey); found {
			return val.([]*entity.TrainDateHasUserApptState), nil
		}

		// 使用 SingleFlight 解決緩存擊穿
		res, err, _ := r.sfGroup.Do(cacheKey, func() (interface{}, error) {
			data, repoErr := r.delegate.UserQueryTrainDateHasApptState(ctx, userID, filter)
			if repoErr != nil {
				return nil, repoErr
			}
			r.cache.Set(cacheKey, data, cache.DefaultExpiration)
			return data, nil
		})

		if err != nil {
			return nil, err.(repository.RepoError)
		}
		return res.([]*entity.TrainDateHasUserApptState), nil
	}

	return r.delegate.UserQueryTrainDateHasApptState(ctx, userID, filter)
}

func (r *cachedTrainRepo) CleanTrainCache(ctx context.Context, userID string) repository.RepoError {
	// 因為我們不知道具體是哪個時間範圍被快取了，最簡單的方法是清除該用戶的所有相關快取
	// 或者如果 go-cache 支持按前綴刪除（它不支持），我們可能需要換種方式。
	// 對於 go-cache，我們可以簡單地 Flush 或不處理，但更好的做法是使用 userID 作為 Key 的一部分
	// 這裡我們先用 Flush 簡化，但在生產環境建議用 Redis 或更細粒度的控制。
	// 由於這是 memory cache 且 TTL 很短 (5m)，不 Clean 也許能接受，但為了正確性：
	r.cache.Flush() 
	return r.delegate.CleanTrainCache(ctx, userID)
}

// Delegate other methods
func (r *cachedTrainRepo) SaveTrainDate(ctx context.Context, training *entity.TrainDate) repository.RepoError {
	return r.delegate.SaveTrainDate(ctx, training)
}
func (r *cachedTrainRepo) SaveManyTrainDates(ctx context.Context, trainings []*entity.TrainDate) repository.RepoError {
	return r.delegate.SaveManyTrainDates(ctx, trainings)
}
func (r *cachedTrainRepo) UpdateManyTrainDates(ctx context.Context, trainings []*entity.TrainDate) repository.RepoError {
	return r.delegate.UpdateManyTrainDates(ctx, trainings)
}
func (r *cachedTrainRepo) DeleteTrainingDate(ctx context.Context, training *entity.TrainDate) repository.RepoError {
	return r.delegate.DeleteTrainingDate(ctx, training)
}
func (r *cachedTrainRepo) FindTrainDateByID(ctx context.Context, id string) (*entity.TrainDate, repository.RepoError) {
	return r.delegate.FindTrainDateByID(ctx, id)
}
func (r *cachedTrainRepo) FindTrainDates(ctx context.Context, filter repository.FilterTrainDate) ([]*entity.TrainDate, repository.RepoError) {
	return r.delegate.FindTrainDates(ctx, filter)
}
func (r *cachedTrainRepo) QueryTrainDateHasAppointmentState(ctx context.Context, filter repository.FilterTrainDate) ([]*entity.TrainDateHasApptState, repository.RepoError) {
	return r.delegate.QueryTrainDateHasAppointmentState(ctx, filter)
}
func (r *cachedTrainRepo) CheckOverlap(ctx context.Context, coachID string, tr entity.TimeRange) (bool, repository.RepoError) {
	return r.delegate.CheckOverlap(ctx, coachID, tr)
}
func (r *cachedTrainRepo) HasAnyOverlap(ctx context.Context, coachID string, tr []entity.TimeRange) (bool, repository.RepoError) {
	return r.delegate.HasAnyOverlap(ctx, coachID, tr)
}
func (r *cachedTrainRepo) DeductCapacity(ctx context.Context, trainingID string, count int) repository.RepoError {
	return r.delegate.DeductCapacity(ctx, trainingID, count)
}
func (r *cachedTrainRepo) IncreaseCapacity(ctx context.Context, trainingID string, count int) repository.RepoError {
	return r.delegate.IncreaseCapacity(ctx, trainingID, count)
}
