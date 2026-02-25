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

type CronTask struct {
	Name string `mapstructure:"name"`
	Spec string `mapstructure:"spec"`
	Path string `mapstructure:"path"` // 內部 API 路徑
	URL  string `mapstructure:"url"`  // 外部 Webhook URL
}

// cronCmd represents the cron command
var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "執行定時背景任務 (由 YAML 設定驅動)",
	Long:  `讀取設定檔中的 cron.tasks 區塊，定時呼叫內部 API 或外部 Webhook 執行任務。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. 讀取設定
		baseURL := viper.GetString("cron.base_url")
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}

		var tasks []CronTask
		err := viper.UnmarshalKey("cron.tasks", &tasks)
		if err != nil {
			fmt.Printf("讀取 Cron 任務設定失敗: %v\n", err)
			return
		}

		if len(tasks) == 0 {
			fmt.Println("未偵測到任何排程任務，請檢查設定檔中的 cron.tasks 區塊。")
			return
		}

		// 2. 初始化 Cron (使用秒級精確度或標準分級精確度，這裡使用標準)
		c := cron.New()

		for _, task := range tasks {
			task := task // 閉包捕獲
			if task.Spec == "" {
				fmt.Printf("跳過任務 [%s]: 缺少 spec 排程設定。\n", task.Name)
				continue
			}

			var targetURL string
			if task.Path != "" {
				targetURL = baseURL + task.Path
			} else if task.URL != "" {
				targetURL = task.URL
			}

			if targetURL == "" {
				fmt.Printf("跳過任務 [%s]: 缺少 path 或 url 設定。\n", task.Name)
				continue
			}

			_, err := c.AddFunc(task.Spec, func() {
				fmt.Printf("[%s] 開始執行任務: %s -> %s\n", time.Now().Format("2006-01-02 15:04:05"), task.Name, targetURL)
				triggerAPI(targetURL)
			})

			if err != nil {
				fmt.Printf("註冊任務 [%s] 失敗: %v\n", task.Name, err)
			} else {
				fmt.Printf("已註冊任務 [%s], 排程: %s\n", task.Name, task.Spec)
			}
		}

		// 3. 啟動 Cron
		c.Start()
		fmt.Println("Cron 服務已啟動，按 Ctrl+C 結束...")

		// 4. 設定訊號監聽
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// 5. 阻塞主線程
		sig := <-sigChan
		fmt.Printf("\n收到訊號: %v，正在關閉排程服務...\n", sig)

		// 6. 優雅關閉
		stopCtx := c.Stop()
		select {
		case <-stopCtx.Done():
			fmt.Println("所有執行中任務已完成，程式正式退出。")
		case <-time.After(30 * time.Second):
			fmt.Println("關閉超時，強制退出。")
		}
	},
}

func init() {
	rootCmd.AddCommand(cronCmd)
}

// 統一發送 POST 請求
func triggerAPI(url string) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("建立請求失敗: %v\n", err)
		return
	}

	// 這裡可以預留統一的 Secret 驗證
	// if secret := viper.GetString("cron.secret"); secret != "" {
	//     req.Header.Set("X-Cron-Secret", secret)
	// }

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("API 執行失敗: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("任務執行成功 (HTTP %d)\n", resp.StatusCode)
	} else {
		fmt.Printf("任務執行失敗 (HTTP %d)\n", resp.StatusCode)
	}
}
