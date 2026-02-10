/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/94peter/botreplyer"
	"github.com/94peter/botreplyer/llm"
	"github.com/94peter/botreplyer/provider/line"
	"github.com/94peter/botreplyer/provider/line/flexmsg"
	"github.com/94peter/botreplyer/provider/line/notify"
	"github.com/94peter/botreplyer/provider/line/reply/textreply"
	"github.com/94peter/vulpes/db/mgo"
	_ "github.com/94peter/vulpes/ezapi/session"
	"github.com/94peter/vulpes/log"
	"github.com/94peter/vulpes/storage"
	"github.com/invopop/ctxi18n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms/googleai"
	"go.opentelemetry.io/otel"

	"seanAIgent/internal/booking/infra/db"
	"seanAIgent/internal/booking/transport/line/notification"
	"seanAIgent/internal/booking/transport/web"
	"seanAIgent/internal/db/factory"
	"seanAIgent/internal/service"
	"seanAIgent/internal/service/lineliff"
	"seanAIgent/internal/service/linemsg"
	"seanAIgent/internal/service/linemsg/replyfunc"
	"seanAIgent/locales"
)

// teacherCmd represents the teacher command
var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		mainCtx, mainCancel := context.WithCancel(context.Background())
		defer mainCancel()
		shutdown, err := initTracer(mainCtx)
		if err != nil {
			log.Fatalf("initTracer fail: %v", err)
		}
		defer func() {
			// 建立一個跟 mainCtx 脫鉤的 Context
			// 給予 5 秒鐘的緩衝時間，確保即使 mainCtx 結束了，監控數據也能發完
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			log.Info("正在關閉監控並推送最後數據...")
			shutdown(ctx)
		}()

		lineliff.InitLineLiff(viper.GetStringMapString("liffids"))
		// load locales
		if err := ctxi18n.Load(locales.Content); err != nil {
			log.Fatalf("error loading locales: %v", err)
		}

		// Load flex message templates
		flexCfgFile := viper.GetString("linebot.flex_message.config_file")
		if err := flexmsg.Load(flexCfgFile); err != nil {
			log.Fatalf("Failed to load flex message templates: %v", err)
		}

		// Load text message templates
		msgCfgFile := viper.GetString("linebot.message.config_file")
		if err := textreply.Load(msgCfgFile); err != nil {
			log.Fatalf("Failed to load message templates: %v", err)
		}

		// Load service response templates
		responseCfgFile := viper.GetString("service.response_templates")
		if err := service.LoadTemplateMsg(responseCfgFile); err != nil {
			log.Fatalf("Failed to load response templates: %v", err)
		}

		dbTracer := otel.Tracer("Mongodb")
		// db connection
		dbCtx, cancel := context.WithTimeout(mainCtx, time.Minute)
		err = factory.InitializeDb(
			dbCtx,
			factory.WithMongoDB(
				viper.GetString("database.uri"),
				viper.GetString("database.db"),
			),
			factory.WithMongoDBPoolSize(
				viper.GetUint64("database.max_pool_size"),
				viper.GetUint64("database.min_pool_size"),
			),
			factory.WithTracer(dbTracer),
		)
		if err != nil {
			log.Fatal(err.Error())
		}
		defer func() {
			closeCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
			err = mgo.Close(closeCtx)
			if err != nil {
				log.Fatal(err.Error())
			}
			cancel()
		}()
		cancel()
		var cancelSlice []context.CancelFunc
		storageCtx, storageCancel := context.WithCancel(mainCtx)
		cancelSlice = append(cancelSlice, storageCancel)
		r2storage, err := storage.New(storageCtx,
			storage.WithAccessKey(viper.GetString("storage.r2.access_key_id")),
			storage.WithSecretKey(viper.GetString("storage.r2.secret_access_key")),
			storage.WithEndpoint(viper.GetString("storage.r2.endpoint")),
			storage.WithBucket(viper.GetString("storage.r2.bucket")),
		)
		if err != nil {
			log.Fatal(err.Error())
		}
		var checkinReplyer textreply.LineKeywordReply
		var appointmentState textreply.LineKeywordReply
		var catchUpCheckIn textreply.LineKeywordReply
		var userApptStatsNotify notification.UserApptStatsNotifier
		var webService web.WebService
		factory.InjectStore(func(stores *factory.Stores) {
			_ = service.InitService(
				service.WithTrainingStore(stores.TrainingDateStore),
				service.WithAppointmentStore(stores.AppointmentStore),
			)
			webService = InitializeWeb()
			// v1 handler

			// v2 handler
			dbRepo := db.NewDbRepoAndIdGenerate()
			userApptStatsNotify = notification.NewUserApptStatsNotifier(dbRepo)
			// web.InitWeb(
			// 	routerGroup,
			// 	// app.NewScheduleService(dbRepo, idGen),
			// 	// app.NewAppointmentService(dbRepo, idGen),
			// 	viper.GetBool("http.csrf.enabled"),
			// )

			checkinReplyer = linemsg.NewStartCheckinReply(stores.TrainingDateStore)
			appointmentState = linemsg.NewAppointmentStateReply(stores.AppointmentStore, r2storage)
			catchUpCheckIn = linemsg.NewCatchUpCheckInReply(stores.TrainingDateStore)
		})

		conversationMgr, llmCancel, err := newConversationMgr(mainCtx)
		if err != nil {
			log.Fatal(err.Error())
		}
		cancelSlice = append(cancelSlice, llmCancel)

		// init botreplyer
		botctx, cancel := context.WithTimeout(mainCtx, time.Minute)

		notifyService := notify.NewLineNotificationService()
		notifyService.RegisterNotification(
			"user-appt-stats", userApptStatsNotify,
		)
		err = botreplyer.InitBotReplyer(
			botctx,
			botreplyer.WithLineConfig(
				line.WithChannelSecret(viper.GetString("linebot.channel_secret")),
				line.WithChannelToken(viper.GetString("linebot.channel_access_token")),
				line.WithReplies(
					appointmentState,
					linemsg.NewStartBookingReply(),
					checkinReplyer,
					catchUpCheckIn,
					linemsg.NewLLMReply(conversationMgr),
				),
				line.WithAdminUserId(viper.GetString("linebot.admin_user_id")),
				line.WithNotificationService(notifyService),
			),
			botreplyer.WithJoinGroupReplyFunc(replyfunc.MyJoinGroupReply),
		)
		cancel()
		if err != nil {
			log.Err(err)
			return
		}

		sigChan := make(chan os.Signal, 1)
		// 註冊要接收的訊號：SIGINT(Ctrl+C), SIGTERM(kill)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		apiCtx, apicancel := context.WithCancel(mainCtx)

		go webService.Run(apiCtx)
		cancelSlice = append(cancelSlice, apicancel)

		// 等待訊號
		sig := <-sigChan
		log.Infof("Received signal: %s", sig)
		for _, c := range cancelSlice {
			c()
		}
		log.Info("Server is shutting down...")
		time.Sleep(time.Second)
	},
}

func init() {
	serveCmd.AddCommand(consoleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// teacherCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// teacherCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func newConversationMgr(ctx context.Context) (llm.ConversationMgr, context.CancelFunc, error) {
	llmCtx, llmCancel := context.WithCancel(ctx)
	llmmodel, err := googleai.New(llmCtx,
		googleai.WithAPIKey(viper.GetString("llm.googleai.api_key")),
		googleai.WithDefaultModel(viper.GetString("llm.model")))
	if err != nil {
		llmCancel()
		return nil, nil, err
	}

	conversationMgr, err := llm.NewConversationMgr(
		llmmodel,
		viper.GetString("llm.config_file"),
		viper.GetStringSlice("llm.mcp_server"),
		llm.WithConversationMemoryMongo(
			viper.GetString("database.uri"),
			viper.GetString("database.db"),
			viper.GetString("llm.memory_collection"),
		),
	)
	if err != nil {
		llmCancel()
		return nil, nil, err
	}
	return conversationMgr, llmCancel, nil
}
