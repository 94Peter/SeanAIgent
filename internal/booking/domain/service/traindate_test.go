package service

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTrainRepository
type MockTrainRepository struct {
	mock.Mock
}

func (m *MockTrainRepository) SaveTrainDate(ctx context.Context, training *entity.TrainDate) repository.RepoError {
	args := m.Called(ctx, training)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.RepoError)
}
func (m *MockTrainRepository) SaveManyTrainDates(ctx context.Context, trainings []*entity.TrainDate) repository.RepoError {
	args := m.Called(ctx, trainings)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.RepoError)
}
func (m *MockTrainRepository) UpdateManyTrainDates(ctx context.Context, trainings []*entity.TrainDate) repository.RepoError {
	args := m.Called(ctx, trainings)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.RepoError)
}
func (m *MockTrainRepository) DeleteTrainingDate(ctx context.Context, training *entity.TrainDate) repository.RepoError {
	args := m.Called(ctx, training)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.RepoError)
}
func (m *MockTrainRepository) FindTrainDateByID(ctx context.Context, id string) (*entity.TrainDate, repository.RepoError) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Get(1).(repository.RepoError)
	}
	return args.Get(0).(*entity.TrainDate), nil
}
func (m *MockTrainRepository) FindTrainDates(ctx context.Context, filter repository.FilterTrainDate) ([]*entity.TrainDate, repository.RepoError) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(repository.RepoError)
	}
	return args.Get(0).([]*entity.TrainDate), nil
}
func (m *MockTrainRepository) QueryTrainDateHasAppointmentState(ctx context.Context, filter repository.FilterTrainDate) ([]*entity.TrainDateHasApptState, repository.RepoError) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(repository.RepoError)
	}
	return args.Get(0).([]*entity.TrainDateHasApptState), nil
}
func (m *MockTrainRepository) UserQueryTrainDateHasApptState(ctx context.Context, userID string, filter repository.FilterTrainDate) ([]*entity.TrainDateHasUserApptState, repository.RepoError) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(repository.RepoError)
	}
	return args.Get(0).([]*entity.TrainDateHasUserApptState), nil
}
func (m *MockTrainRepository) CheckOverlap(ctx context.Context, coachID string, tr entity.TimeRange) (bool, repository.RepoError) {
	args := m.Called(ctx, coachID, tr)
	if args.Get(1) == nil {
		return args.Bool(0), nil
	}
	return args.Bool(0), args.Get(1).(repository.RepoError)
}
func (m *MockTrainRepository) HasAnyOverlap(ctx context.Context, coachID string, tr []entity.TimeRange) (bool, repository.RepoError) {
	args := m.Called(ctx, coachID, tr)
	if args.Get(1) == nil {
		return args.Bool(0), nil
	}
	return args.Bool(0), args.Get(1).(repository.RepoError)
}
func (m *MockTrainRepository) DeductCapacity(ctx context.Context, trainingID string, count int) repository.RepoError {
	args := m.Called(ctx, trainingID, count)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.RepoError)
}
func (m *MockTrainRepository) IncreaseCapacity(ctx context.Context, trainingID string, count int) repository.RepoError {
	args := m.Called(ctx, trainingID, count)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.RepoError)
}

func TestCheckTrainerAvailability(t *testing.T) {
	mockRepo := new(MockTrainRepository)
	svc := NewTrainDateService(mockRepo)
	ctx := context.Background()
	now := time.Now()
	tr, _ := entity.NewTimeRange(now, now.Add(time.Hour))
	td, _ := entity.NewTrainDate(
		entity.WithBasicTrainDate("id1", "coach1", "Gym", 10, tr),
	)

	t.Run("Success_NoOverlap", func(t *testing.T) {
		mockRepo.On("CheckOverlap", ctx, "coach1", tr).Return(false, nil).Once()

		err := svc.CheckTrainerAvailability(ctx, td)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Fail_Overlap", func(t *testing.T) {
		mockRepo.On("CheckOverlap", ctx, "coach1", tr).Return(true, nil).Once()

		err := svc.CheckTrainerAvailability(ctx, td)
		assert.ErrorIs(t, err, ErrTrainerTimeOverlap)
		mockRepo.AssertExpectations(t)
	})
}

func TestCheckAnyOverlap(t *testing.T) {
	mockRepo := new(MockTrainRepository)
	svc := NewTrainDateService(mockRepo)
	ctx := context.Background()
	now := time.Now()
	tr1, _ := entity.NewTimeRange(now, now.Add(time.Hour))
	ranges := []entity.TimeRange{tr1}

	t.Run("Success_NoOverlap", func(t *testing.T) {
		mockRepo.On("HasAnyOverlap", ctx, "coach1", ranges).Return(false, nil).Once()

		err := svc.CheckAnyOverlap(ctx, "coach1", ranges)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Fail_Overlap", func(t *testing.T) {
		mockRepo.On("HasAnyOverlap", ctx, "coach1", ranges).Return(true, nil).Once()

		err := svc.CheckAnyOverlap(ctx, "coach1", ranges)
		assert.ErrorIs(t, err, ErrTrainerTimeOverlap)
		mockRepo.AssertExpectations(t)
	})
}
