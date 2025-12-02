package linemsg

import (
	"context"
	"fmt"
	"seanAIgent/internal/service/lineliff"
	"slices"
	"strings"

	"github.com/94peter/botreplyer/provider/line/reply/textreply"
	"github.com/gin-contrib/sessions"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type startBookingCfg struct {
	Keywords []string `yaml:"keywords"`
}

func NewStartBookingReply() textreply.LineKeywordReply {
	yamlCfg, ok := textreply.GetNode("start_booking")
	if !ok {
		panic("create_class is not defined")
	}
	var cfg startBookingCfg
	err := yamlCfg.Decode(&cfg)
	if err != nil {
		panic(fmt.Errorf("decode create_class config: %w", err))
	}

	return &startBookingReply{
		cfg:             &cfg,
		msgStartBooking: fmt.Sprintf("大家好!!我是SeanAIgent!!由我來為大家提供約課的服務喔!!\n請點擊下方的按鈕開始約課吧!!\n%s", lineliff.GetBookingLiffUrl()),
	}
}

type startBookingReply struct {
	cfg             *startBookingCfg
	msgStartBooking string
}

func (r *startBookingReply) MessageTextReply(ctx context.Context, typ linebot.EventSourceType, groupID, userID, msg string, session sessions.Session) ([]linebot.SendingMessage, textreply.DelayedMessage, error) {
	if slices.Contains(r.cfg.Keywords, strings.ToLower(msg)) {
		var err error
		var sendingMsgs []linebot.SendingMessage

		sendingMsgs = []linebot.SendingMessage{
			linebot.NewTextMessage(r.msgStartBooking),
		}
		return sendingMsgs, nil, err
	}
	return nil, nil, nil
}
