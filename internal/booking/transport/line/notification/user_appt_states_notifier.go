package notification

import (
	"context"
	"fmt"
	"seanAIgent/internal/booking/domain/repository"
	"time"

	"github.com/94peter/botreplyer/provider/line/notify"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type UserApptStatsNotifier interface {
	notify.LineNotify
}

func NewUserApptStatsNotifier(repo repository.StatsRepository) UserApptStatsNotifier {
	return &userApptStats{repo: repo}
}

// 使用者預約狀況推播通知
type userApptStats struct {
	repo repository.StatsRepository
}

func (n *userApptStats) GetNotification(ctx context.Context) []*notify.NotificationContent {
	now := time.Now()
	var start, end time.Time
	var isReview bool // 是否為回顧（1號）
	switch now.Day() {
	case 1:
		// 上個月 1 號 ~ 本月 1 號
		start = time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
		end = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Add(-time.Millisecond)
		isReview = true
	case 15:
		// 本月 1 號 ~ 下個月 1 號
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Millisecond)
		isReview = false
	default:
		return nil
	}

	users, err := n.repo.GetAllUserApptStats(ctx, repository.NewFilterUserApptStatsByTrainTimeRange(
		start, end,
	))
	if err != nil {
		return nil
	}

	var notifications []*notify.NotificationContent
	for _, user := range users {
		var msgText string
		month := int(start.Month())

		if isReview {
			msgText = fmt.Sprintf("嗨 %s 👋，幫你整理了 %d 月的學習報表唷！\n\n上個月總共有 %d 堂預約，完成了 %d 堂，請假 %d 堂。看到你的進度真的太棒了！✨",
				user.UserName, month, user.TotalAppointment, user.CheckedInCount, user.OnLeaveCount)
		} else {
			msgText = fmt.Sprintf("嗨 %s 👋，跟你分享一下 %d 月目前的預約進度唷！\n\n這個月目前預約了 %d 堂，已完成 %d 堂，請假 %d 堂。繼續保持，加油加油！🔥",
				user.UserName, month, user.TotalAppointment, user.CheckedInCount, user.OnLeaveCount)
		}

		if len(user.ChildState) > 0 {
			msgText += "\n\n學員的小紀錄："
			for _, child := range user.ChildState {
				msgText += fmt.Sprintf("\n📍 %s：預約了 %d 堂（報到 %d 次、請假 %d 次）",
					child.ChildName, len(child.Appointments), child.CheckedInCount, child.OnLeaveCount)
			}
		}

		notifications = append(notifications, &notify.NotificationContent{
			UserIDs: user.UserID,
			Message: []linebot.SendingMessage{linebot.NewTextMessage(msgText)},
		})
	}
	return notifications
}
