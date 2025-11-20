package replyfunc

import (
	"context"

	"github.com/94peter/botreplyer/provider/line/reply"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

var MyJoinGroupReply reply.JoinGroupReplyFunc = func(ctx context.Context) ([]linebot.SendingMessage, error) {
	return nil, nil
}
