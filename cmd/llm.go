/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"time"

	"github.com/94peter/botreplyer/llm"
	"github.com/94peter/vulpes/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms/googleai"
)

// llmCmd represents the llm command
var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("llm called")
		llmCtx, llmCancel := context.WithCancel(context.Background())
		llmmodel, err := googleai.New(llmCtx, googleai.WithAPIKey(viper.GetString("llm.googleai.api_key")), googleai.WithDefaultModel("gemini-2.5-flash"))
		if err != nil {
			panic(fmt.Errorf("Failed to create llm: %v", err))
		}
		defer llmCancel()

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
			panic(fmt.Errorf("Failed to create conversation: %v", err))
		}
		ctx := context.Background()
		conversation, err := conversationMgr.NewConversation(ctx, "1234567890")
		if err != nil {
			panic(fmt.Errorf("Failed to create conversation: %v", err))
		}
		result, err := conversation.Chat(ctx,
			"幫我建立課程,接下來三個星期的每個星期日的下午2點到4點，最多10人，地點在TKU", map[string]any{
				"line_user_id": "1234567890",
				"today":        time.Now().Format(time.RFC3339),
				"weekday":      time.Now().Weekday().String(),
			})

		if err != nil {
			log.Fatalf("Failed to format prompt: %v", err)
			return
		}

		fmt.Println("Agent result:", result)

		time.Sleep(time.Second * 3)
		result, err = conversation.Chat(ctx, "ok", map[string]any{
			"line_user_id": "1234567890",
			"today":        time.Now().Format(time.RFC3339),
			"weekday":      time.Now().Weekday().String(),
		})
		fmt.Println("Agent result:", result)
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// llmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// llmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
