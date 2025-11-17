package linemsg

import (
	"context"
	"fmt"
	"seanAIgent/internal/db"
	"seanAIgent/internal/db/model"
	"slices"
	"strconv"
	"time"

	"github.com/94peter/botreplyer/provider/line/reply/textreply"
	"github.com/94peter/botreplyer/session"
	"github.com/94peter/vulpes/export/csv"
	"github.com/94peter/vulpes/storage"
	"github.com/gin-contrib/sessions"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type appointmentStateCfg struct {
	Keywords []string `yaml:"keywords"`
}

func NewAppointmentStateReply(appointmentStore db.AppointmentStore, storage storage.Storage) textreply.LineKeywordReply {
	yamlCfg, ok := textreply.GetNode("appointment_statistic")
	if !ok {
		panic("create_class is not defined")
	}
	var cfg appointmentStateCfg
	err := yamlCfg.Decode(&cfg)
	if err != nil {
		panic(fmt.Errorf("decode create_class config: %w", err))
	}

	return &appointmentStateReply{
		cfg:              &cfg,
		appointmentStore: appointmentStore,
		storage:          storage,
	}
}

type appointmentStateReply struct {
	cfg              *appointmentStateCfg
	appointmentStore db.AppointmentStore
	storage          storage.Storage
}

func (r *appointmentStateReply) MessageTextReply(ctx context.Context, typ linebot.EventSourceType, groupID, userID, msg string, mysession sessions.Session) ([]linebot.SendingMessage, textreply.DelayedMessage, error) {
	if !session.IsAdmin(mysession) {
		return []linebot.SendingMessage{
			linebot.NewTextMessage("您沒有權限查看"),
		}, nil, nil
	}
	if typ != linebot.EventSourceTypeUser {
		return nil, nil, nil
	}
	if slices.Contains(r.cfg.Keywords, msg) {
		mysession.Set("topic", "appointment_state")
		quickReplyButtons := make([]*linebot.QuickReplyButton, 0, 3)
		now := time.Now()
		var quickDate time.Time
		for i := -1; i <= 1; i++ {
			quickDate = now.AddDate(0, i, 0)
			selectStr := quickDate.Format("2006/01")
			quickReplyButtons = append(quickReplyButtons, linebot.NewQuickReplyButton("",
				linebot.NewMessageAction(selectStr, selectStr),
			))
		}

		sendingMsgs := []linebot.SendingMessage{
			linebot.NewTextMessage("請選擇月份").WithQuickReplies(
				linebot.NewQuickReplyItems(quickReplyButtons...),
			),
		}
		return sendingMsgs, nil, nil
	}
	topic := mysession.Get("topic")
	if topic == nil || (topic != nil && topic != "appointment_state") {
		return nil, nil, nil
	}
	selectTime, err := time.Parse("2006/01", msg)
	if err != nil {
		return nil, nil, err
	}
	fileURL, err := r.updateCsvAndGetUrl(ctx, selectTime.Year(), selectTime.Month())
	var sendingMsgs []linebot.SendingMessage

	sendingMsgs = []linebot.SendingMessage{
		linebot.NewTemplateMessage(
			"下載報表",
			linebot.NewButtonsTemplate(
				"", // thumbnail image URL (optional)
				"CSV 檔案下載",
				"點擊下方按鈕下載報表。",
				linebot.NewURIAction("下載 CSV", fileURL),
			),
		),
	}
	mysession.Delete("topic")
	return sendingMsgs, nil, err
}

const keyTpl = "appointment_state_%d_%d.csv"

func (r *appointmentStateReply) updateCsvAndGetUrl(ctx context.Context, year int, month time.Month) (string, error) {
	filter := bson.M{
		"$expr": bson.M{
			"$and": bson.A{
				bson.M{"$eq": bson.A{bson.M{"$year": "$start_date"}, year}},
				bson.M{"$eq": bson.A{bson.M{"$month": "$start_date"}, month}},
			},
		},
	}
	results, err := r.appointmentStore.AppointmentState(ctx, filter)
	if err != nil {
		return "", err
	}
	key := fmt.Sprintf(keyTpl, year, month)
	return csv.Upload(ctx, r.storage, key, newAppointmentStateCsvMarshaler(results))
}

func newAppointmentStateCsvMarshaler(data []*model.AggrAppointmentState) csv.CSVMarshaler {
	return &csvMarshaler{data: data}
}

type csvMarshaler struct {
	data []*model.AggrAppointmentState
}

func (m *csvMarshaler) MarshalCSV() (headers []string, rows [][]string, err error) {
	headers = []string{"Index", "Parent", "Child", "Total Appointment"}
	var index int
	for _, v := range m.data {
		for _, vv := range v.ChildState {
			index++
			row := make([]string, 0, 4+len(vv.Appointments))
			row = append(row, strconv.Itoa(index),
				v.UserName, vv.ChildName,
				strconv.Itoa(len(vv.Appointments)))
			for _, vvv := range vv.Appointments {
				row = append(row, vvv.StartDate.Format("01/02"))
			}
			rows = append(rows, row)
		}
	}
	return headers, rows, nil
}
