// 產生 LINE 通知訊息提醒使用者每個月的預約上課狀況的 UseCase
package read

import (
	"context"
	"seanAIgent/internal/booking/usecase/core"
	"time"
)

type resNotifyUserApptStats struct {
}

func NewNotifyUserApptStatsUseCase() core.ReadUseCase[time.Time, *resNotifyUserApptStats] {
	return &notifyUserApptStatsUseCase{}
}

type notifyUserApptStatsUseCase struct {
}

func (uc *notifyUserApptStatsUseCase) Name() string {
	return "NotifyUserApptStats"
}

func (uc *notifyUserApptStatsUseCase) Execute(
	ctx context.Context, req time.Time,
) (*resNotifyUserApptStats, core.UseCaseError) {
	return nil, nil
}
