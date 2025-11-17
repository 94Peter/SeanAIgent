/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/94peter/botreplyer"
	"github.com/94peter/botreplyer/llm"
	"github.com/94peter/botreplyer/provider/line"
	"github.com/94peter/botreplyer/provider/line/flexmsg"
	lineMid "github.com/94peter/botreplyer/provider/line/mid"
	"github.com/94peter/botreplyer/provider/line/reply/textreply"
	"github.com/94peter/vulpes/db/mgo"
	"github.com/94peter/vulpes/ezapi"
	_ "github.com/94peter/vulpes/ezapi/session"
	"github.com/94peter/vulpes/log"
	"github.com/94peter/vulpes/storage"
	"github.com/invopop/ctxi18n"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms/googleai"

	"seanAIgent/internal/db/factory"
	"seanAIgent/internal/handler"

	"seanAIgent/internal/service"
	"seanAIgent/internal/service/linemsg"
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
		fmt.Println("teacher called")

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

		// db connection
		dbCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		err := factory.InitializeDb(
			dbCtx,
			factory.WithMongoDB(viper.GetString("database.uri"),
				viper.GetString("database.db")))
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
		storageCtx, storageCancel := context.WithCancel(context.Background())
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
		factory.InjectStore(func(stores *factory.Stores) {
			svc := service.InitService(
				service.WithTrainingStore(stores.TrainingDateStore),
				service.WithAppointmentStore(stores.AppointmentStore),
			)
			handler.InitTrainingApi(svc, viper.GetBool("http.csrf.enabled"))
			handler.InitBookingApi(svc, viper.GetBool("http.csrf.enabled"))
			handler.InitCheckinApi(svc, viper.GetBool("http.csrf.enabled"))

			checkinReplyer = linemsg.NewStartCheckinReply(stores.TrainingDateStore)
			appointmentState = linemsg.NewAppointmentStateReply(stores.AppointmentStore, r2storage)
		})
		handler.InitComponentApi()

		conversationMgr, llmCancel, err := newConversationMgr(context.Background())
		if err != nil {
			log.Fatal(err.Error())
		}
		cancelSlice = append(cancelSlice, llmCancel)

		// init botreplyer
		botctx, cancel := context.WithTimeout(context.Background(), time.Minute)

		err = botreplyer.InitBotReplyer(
			botctx,
			botreplyer.WithLineConfig(
				line.WithChannelSecret(viper.GetString("linebot.channel_secret")),
				line.WithChannelToken(viper.GetString("linebot.channel_access_token")),
				line.WithReplies(
					appointmentState,
					linemsg.NewStartBookingReply(),
					checkinReplyer,
					linemsg.NewLLMReply(conversationMgr),
				),
				line.WithAdminUserId(viper.GetString("linebot.admin_user_id")),
			),
		)
		cancel()
		if err != nil {
			log.Err(err)
			return
		}

		sigChan := make(chan os.Signal, 1)
		// 註冊要接收的訊號：SIGINT(Ctrl+C), SIGTERM(kill)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		apiCtx, apicancel := context.WithCancel(context.Background())
		go func(ctx context.Context) {
			log.Infof("Starting server on port %d", viper.GetInt("http.port"))
			if err := ezapi.RunGin(
				ctx,
				ezapi.WithPort(viper.GetInt("http.port")),
				ezapi.WithMiddleware(
					ezapi.RequestLogger(),
					lineMid.LineLiff(),
					ezapi.I18n("zh-tw", locales.LocaleExist),
					lineMid.CheckAdmin(botreplyer.GetFollowStore()),
				),
				ezapi.WithSession(
					viper.GetBool("http.session.enabled"),
					viper.GetString("http.session.store"),
					viper.GetString("http.session.cookie_name"),
					viper.GetInt("http.session.max_age"),
					viper.GetStringSlice("http.session.key_pairs")...,
				),
				ezapi.WithCsrf(
					viper.GetBool("http.csrf.enabled"),
					viper.GetString("http.csrf.secret"),
					viper.GetString("http.csrf.field_name"),
					viper.GetStringSlice("http.csrf.ignore_paths")...,
				),
				ezapi.WithStaticFS("/assets", http.Dir("assets")),
			); err != nil {
				log.Info(err.Error())
			}
		}(apiCtx)
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
	llmmodel, err := googleai.New(llmCtx, googleai.WithAPIKey(viper.GetString("llm.googleai.api_key")), googleai.WithDefaultModel("gemini-2.5-flash"))
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
