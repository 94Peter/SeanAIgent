package read

import (
	"context"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/domain/repository"
	"seanAIgent/internal/booking/usecase/core"
)

type ReqAdminQueryStudents struct {
	SearchKeyword string
}

type AdminQueryStudentsUseCase core.ReadUseCase[ReqAdminQueryStudents, []*entity.UserApptStats]

type adminQueryStudentsUseCase struct {
	repo repository.StatsRepository
}

func NewAdminQueryStudentsUseCase(repo repository.StatsRepository) AdminQueryStudentsUseCase {
	return &adminQueryStudentsUseCase{repo: repo}
}

func (uc *adminQueryStudentsUseCase) Name() string {
	return "AdminQueryStudents"
}

func (uc *adminQueryStudentsUseCase) Execute(ctx context.Context, req ReqAdminQueryStudents) ([]*entity.UserApptStats, core.UseCaseError) {
	// For simplicity, we fetch all and let the caller filter, 
	// or we can implement filtering in the repo if it grows too large.
	stats, err := uc.repo.GetAllUserApptStats(ctx, nil)
	if err != nil {
		return nil, ErrAdminQueryStudentsFail.Wrap(err)
	}
	
	// Optional: Keyword filtering logic here if repo doesn't support it
	
	return stats, nil
}

var (
	ErrAdminQueryStudentsFail = core.NewDBError(
		"ADMIN_QUERY_STUDENTS", "QUERY_FAIL", "admin query students fail", core.ErrInternal)
)
