// 補簽到
package linemsg

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"seanAIgent/internal/db"
	"seanAIgent/internal/db/model"
	"seanAIgent/internal/service/lineliff"
	"slices"
	"strings"
	"time"

	"github.com/94peter/botreplyer/provider/line/reply/textreply"
	"github.com/gin-contrib/sessions"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type keyWordsCfg struct {
	Keywords []string `yaml:"keywords"`
}

func NewCatchUpCheckInReply(training db.TrainingDateStore) textreply.LineKeywordReply {
	yamlCfg, ok := textreply.GetNode("catch_up_check_in")
	if !ok {
		panic("yaml catch_up_check_in is not setting")
	}
	var cfg keyWordsCfg
	err := yamlCfg.Decode(&cfg)
	if err != nil {
		panic(fmt.Errorf("decode create_class config: %w", err))
	}

	return &catchUpCheckInReply{
		cfg:      &cfg,
		training: training,
	}
}

type catchUpCheckInReply struct {
	cfg      *keyWordsCfg
	training db.TrainingDateStore
}

func (r *catchUpCheckInReply) MessageTextReply(
	ctx context.Context, typ linebot.EventSourceType,
	groupID, userID, msg string, mysession sessions.Session,
) ([]linebot.SendingMessage, textreply.DelayedMessage, error) {
	if slices.Contains(r.cfg.Keywords, strings.ToLower(msg)) {
		mysession.Set("topic", "catch_up_check_in")
		var sendingMsgs []linebot.SendingMessage
		now := time.Now()
		pastTime := now.Add(time.Hour * -72)
		timeQ := bson.M{
			"$gte": pastTime, "$lte": now,
		}
		// 查詢過去72小時的課程
		trainings, err := r.training.Find(ctx, bson.M{
			"$or": bson.A{
				bson.M{
					"start_date": timeQ,
				},
				bson.M{
					"end_date": timeQ,
				},
			},
		})
		if err != nil {
			return nil, nil, err
		}
		if len(trainings) == 0 {
			sendingMsgs = []linebot.SendingMessage{
				linebot.NewTextMessage("過去72小時沒有相關的課程可以補簽到"),
			}
			mysession.Delete("topic")
			return sendingMsgs, nil, nil
		}

		quickReplyButtons := make([]*linebot.QuickReplyButton, 0, len(trainings))
		for _, training := range trainings {
			selectTime := model.ToTime(training.StartDate, training.Timezone)
			quickReplyButtons = append(quickReplyButtons, linebot.NewQuickReplyButton("",
				linebot.NewMessageAction(selectTime.Format("2006-01-02 15:04"), selectTime.Format(time.RFC3339)),
			))
		}
		sendingMsgs = []linebot.SendingMessage{
			linebot.NewTextMessage("請選擇補簽時段").WithQuickReplies(
				linebot.NewQuickReplyItems(quickReplyButtons...),
			),
		}
		return sendingMsgs, nil, nil
	}

	topic := mysession.Get("topic")
	if topic == nil || (topic != nil && topic != "catch_up_check_in") {
		return nil, nil, nil
	}
	_, err := time.Parse(time.RFC3339, msg)
	if err != nil {
		return nil, nil, err
	}
	var strBuf bytes.Buffer
	strBuf.WriteString(lineliff.GetCheckinLiffUrl())
	strBuf.WriteString("?time=")
	strBuf.WriteString(url.QueryEscape(msg))

	sendingMsgs := []linebot.SendingMessage{
		linebot.NewTextMessage("請點選下列連結進行補簽：\n" + strBuf.String()),
	}
	mysession.Delete("topic")
	return sendingMsgs, nil, err

}
