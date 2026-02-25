package train

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

type cachedTrainRepo struct {
	delegate repository.TrainRepository
	cache    *cache.Cache
	sfGroup  *singleflight.Group

	// 用於紀錄每個用戶關聯的快取 Key，以便精準刪除
	// key: string (userID), value: *sync.Map (key: cacheKey, value: struct{}{})
	userKeys sync.Map
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
	if f, ok := filter.(repository.FilterTrainingDateByTimeRange); ok {
		// 優化：改用字串拼接避免 fmt.Sprintf 的反射開銷
		cacheKey := "schedule:" + userID + ":" + f.StartTime.Format("2006-01-02") + ":" + f.EndTime.Format("2006-01-02")

		// 紀錄 Key 關聯 (使用 sync.Map 降低鎖競爭)
		actual, _ := r.userKeys.LoadOrStore(userID, &sync.Map{})
		userMap := actual.(*sync.Map)
		userMap.Store(cacheKey, struct{}{})

		if val, found := r.cache.Get(cacheKey); found {
			return val.([]*entity.TrainDateHasUserApptState), nil
		}

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
	if userID != "" {
		if actual, ok := r.userKeys.Load(userID); ok {
			userMap := actual.(*sync.Map)
			userMap.Range(func(key, value interface{}) bool {
				r.cache.Delete(key.(string))
				return true
			})
			r.userKeys.Delete(userID)
		}
	} else {
		// 如果沒有指定 userID，則維持原有的 Flush 行為（全域清理）
		r.cache.Flush()
		// 同時清空所有用戶的 Key 紀錄
		r.userKeys.Range(func(key, value interface{}) bool {
			r.userKeys.Delete(key)
			return true
		})
	}

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
func (r *cachedTrainRepo) AdminDeductCapacity(ctx context.Context, trainingID string, count int) repository.RepoError {
	return r.delegate.AdminDeductCapacity(ctx, trainingID, count)
}
func (r *cachedTrainRepo) IncreaseCapacity(ctx context.Context, trainingID string, count int) repository.RepoError {
	return r.delegate.IncreaseCapacity(ctx, trainingID, count)
}
func (r *cachedTrainRepo) FindPastTrainDateIDs(ctx context.Context, cutoff time.Time, limit uint16) ([]string, repository.RepoError) {
	return r.delegate.FindPastTrainDateIDs(ctx, cutoff, limit)
}
