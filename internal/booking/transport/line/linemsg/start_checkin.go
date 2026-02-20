package linemsg

import (
	"context"
	"fmt"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase/core"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	"seanAIgent/internal/service/lineliff"
	"slices"
	"strings"
	"time"

	"github.com/94peter/botreplyer/provider/line/reply/textreply"
	"github.com/94peter/botreplyer/session"

	"github.com/gin-contrib/sessions"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type startCheckinCfg struct {
	Keywords []string `yaml:"keywords"`
}

func NewStartCheckinReply(
	findNearestTrainByTimeUC core.ReadUseCase[readTrain.ReqFindNearestTrainByTime, *entity.TrainDateHasApptState],
) textreply.LineKeywordReply {
	yamlCfg, ok := textreply.GetNode("start_checkin")
	if !ok {
		panic("create_class is not defined")
	}
	var cfg startBookingCfg
	err := yamlCfg.Decode(&cfg)
	if err != nil {
		panic(fmt.Errorf("decode create_class config: %w", err))
	}

	return &startCheckinReply{
		cfg:                      &cfg,
		findNearestTrainByTimeUC: findNearestTrainByTimeUC,
		msgStartCheckin:          fmt.Sprintf("大家好!!教練要簽到啦!!\n教練請點擊下方的按鈕開始簽到吧!!\n%s", lineliff.GetCheckinLiffUrl()),
	}
}

type startCheckinReply struct {
	cfg                      *startBookingCfg
	findNearestTrainByTimeUC core.ReadUseCase[readTrain.ReqFindNearestTrainByTime, *entity.TrainDateHasApptState]
	msgStartCheckin          string
}

func (r *startCheckinReply) MessageTextReply(ctx context.Context, typ linebot.EventSourceType, groupID, userID, msg string, mysession sessions.Session) ([]linebot.SendingMessage, textreply.DelayedMessage, error) {
	if slices.Contains(r.cfg.Keywords, strings.ToLower(msg)) {
		if !session.IsAdmin(mysession) {
			return nil, nil, nil
		}
		var sendingMsgs []linebot.SendingMessage

		now := time.Now().Add(10 * time.Minute)
		data, ucErr := r.findNearestTrainByTimeUC.Execute(ctx, readTrain.ReqFindNearestTrainByTime{
			TimeAfter: now,
		})
		if ucErr != nil {
			if ucErr.Type() == core.ErrNotFound {
				sendingMsgs = []linebot.SendingMessage{
					linebot.NewTextMessage("目前沒有課程可以簽到ㄛ!!"),
				}
				return sendingMsgs, nil, nil
			}
			return nil, nil, ucErr
		}
		if len(data.UserAppointments) == 0 {
			sendingMsgs = []linebot.SendingMessage{
				linebot.NewTextMessage("這時段沒有人預約喔!!"),
			}
			return sendingMsgs, nil, nil
		}

		sendingMsgs = []linebot.SendingMessage{
			linebot.NewTextMessage(r.msgStartCheckin),
		}
		return sendingMsgs, nil, nil
	}
	return nil, nil, nil
}
