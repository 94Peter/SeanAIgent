package linemsg

import (
	"context"
	"time"

	"github.com/94peter/botreplyer/llm"
	"github.com/94peter/botreplyer/provider/line/reply/textreply"
	"github.com/94peter/botreplyer/session"

	"github.com/gin-contrib/sessions"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func NewLLMReply(mgr llm.ConversationMgr) textreply.LineKeywordReply {
	return &llmReply{
		mgr: mgr,
	}
}

type llmReply struct {
	mgr llm.ConversationMgr
}

func (r *llmReply) MessageTextReply(ctx context.Context, typ linebot.EventSourceType, groupID, userID, msg string, mysession sessions.Session) ([]linebot.SendingMessage, textreply.DelayedMessage, error) {
	if !session.IsAdmin(mysession) {
		return nil, nil, nil
	}
	if typ != linebot.EventSourceTypeUser {
		return nil, nil, nil
	}
	conversation, err := r.mgr.NewConversation(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	today := time.Now()
	response, err := conversation.Chat(ctx, msg, map[string]any{
		"line_user_id": userID,
		"today":        today.Format(time.RFC3339),
		"weekday":      today.Weekday().String(),
	})
	if err != nil {
		return nil, nil, err
	}

	return []linebot.SendingMessage{
		linebot.NewTextMessage(response),
	}, nil, nil
}
