// 補簽到
package linemsg

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"seanAIgent/internal/booking/domain/entity"
	"seanAIgent/internal/booking/usecase/core"
	readTrain "seanAIgent/internal/booking/usecase/traindate/read"
	"seanAIgent/internal/booking/transport/util/lineutil"
	"slices"
	"strings"
	"time"

	"github.com/94peter/botreplyer/provider/line/reply/textreply"
	"github.com/gin-contrib/sessions"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type keyWordsCfg struct {
	Keywords []string `yaml:"keywords"`
}

func NewCatchUpCheckInReply(
	adminQueryRecentTrainUC core.ReadUseCase[readTrain.ReqAdminQueryRecentTrain, []*entity.TrainDate],
) textreply.LineKeywordReply {
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
		cfg:                     &cfg,
		adminQueryRecentTrainUC: adminQueryRecentTrainUC,
	}
}

type catchUpCheckInReply struct {
	cfg                     *keyWordsCfg
	adminQueryRecentTrainUC core.ReadUseCase[readTrain.ReqAdminQueryRecentTrain, []*entity.TrainDate]
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

		// 查詢過去72小時的課程
		trainings, err := r.adminQueryRecentTrainUC.Execute(ctx, readTrain.ReqAdminQueryRecentTrain{
			StartTime: pastTime,
			EndTime:   now,
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
			startTime := training.Period().Start()
			// 如果有時區資訊，可以考慮在這裡處理，但簽到頁面主要依賴 RFC3339 傳遞時間
			quickReplyButtons = append(quickReplyButtons, linebot.NewQuickReplyButton("",
				linebot.NewMessageAction(startTime.In(time.Local).Format("2006-01-02 15:04"), startTime.Format(time.RFC3339)),
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
	strBuf.WriteString(lineutil.GetCheckinLiffUrl())
	strBuf.WriteString("?time=")
	strBuf.WriteString(url.QueryEscape(msg))

	var sendingMsgs []linebot.SendingMessage
	sendingMsgs = []linebot.SendingMessage{
		linebot.NewTextMessage("請點選下列連結進行補簽：\n" + strBuf.String()),
	}
	mysession.Delete("topic")
	return sendingMsgs, nil, err

}
