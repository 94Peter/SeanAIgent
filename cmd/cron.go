/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// cronCmd represents the cron command
var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. 初始化 Cron
		c := cron.New()
		spec := "0 0 1,15 * *"

		_, err := c.AddFunc(spec, func() {
			fmt.Println("執行定時任務...")
			triggerWebhook(viper.GetString("cron.user_stats_notify_url"))
		})

		if err != nil {
			fmt.Println("Cron 任務添加失敗:", err)
			return
		}

		// 2. 啟動 Cron
		c.Start()
		fmt.Printf("服務已啟動，排程: %s。按 Ctrl+C 結束...\n", spec)

		// 3. 設定訊號監聽
		// 建立一個頻道來接收訊號
		sigChan := make(chan os.Signal, 1)
		// 監聽 SIGINT (Ctrl+C) 和 SIGTERM (Docker stop)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// 4. 阻塞主線程，直到收到訊號
		sig := <-sigChan
		fmt.Printf("\n收到訊號: %v，正在關閉服務...\n", sig)

		// 5. 優雅關閉 Cron
		// Stop() 會回傳一個 Context，可以用來等待所有執行中的任務結束
		ctx := c.Stop()

		// 等待任務完成（可選：設定超時強制結束）
		select {
		case <-ctx.Done():
			fmt.Println("所有任務已完成，程式正式退出。")
		case <-time.After(30 * time.Second):
			fmt.Println("關閉超時，強制退出。")
		}
	},
}

func init() {
	serveCmd.AddCommand(cronCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cronCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cronCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// 發送空 Body 的 POST 請求
func triggerWebhook(url string) {
	// 建議建立自定義 Client 並設定超時，避免永久等待
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// 第三個參數傳入 nil 代表空 Body
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		fmt.Printf("[%s] 請求發送失敗: %v\n", time.Now().Format("15:04:05"), err)
		return
	}
	// 務必關閉 Body 以釋放資源
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("[%s] API 觸發成功，狀態碼: %d\n", time.Now().Format("15:04:05"), resp.StatusCode)
	} else {
		fmt.Printf("[%s] API 回傳失敗，狀態碼: %d\n", time.Now().Format("15:04:05"), resp.StatusCode)
	}
}
