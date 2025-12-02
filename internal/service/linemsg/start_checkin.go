package linemsg

import (
	"context"
	"errors"
	"fmt"
	"seanAIgent/internal/db"
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

func NewStartCheckinReply(training db.TrainingDateStore) textreply.LineKeywordReply {
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
		cfg:             &cfg,
		training:        training,
		msgStartCheckin: fmt.Sprintf("大家好!!教練要簽到啦!!\n教練請點擊下方的按鈕開始簽到吧!!\n%s", lineliff.GetCheckinLiffUrl()),
	}
}

type startCheckinReply struct {
	cfg             *startBookingCfg
	training        db.TrainingDateStore
	msgStartCheckin string
}

func (r *startCheckinReply) MessageTextReply(ctx context.Context, typ linebot.EventSourceType, groupID, userID, msg string, mysession sessions.Session) ([]linebot.SendingMessage, textreply.DelayedMessage, error) {
	if slices.Contains(r.cfg.Keywords, strings.ToLower(msg)) {
		if !session.IsAdmin(mysession) {
			return nil, nil, nil
		}
		var err error
		var sendingMsgs []linebot.SendingMessage

		now := time.Now().Add(10 * time.Minute)
		data, err := r.training.QueryTrainingDateHasCheckinList(ctx, now)
		if err != nil {
			switch {
			case errors.Is(err, db.ErrNotFound):
				sendingMsgs = []linebot.SendingMessage{
					linebot.NewTextMessage("目前沒有課程可以簽到ㄛ!!"),
				}
				return sendingMsgs, nil, nil
			default:
				return nil, nil, err
			}
		}
		if len(data.CheckinItems) == 0 {
			sendingMsgs = []linebot.SendingMessage{
				linebot.NewTextMessage("這時段沒有人預約喔!!"),
			}
			return sendingMsgs, nil, nil
		}

		sendingMsgs = []linebot.SendingMessage{
			linebot.NewTextMessage(r.msgStartCheckin),
		}
		return sendingMsgs, nil, err
	}
	return nil, nil, nil
}
