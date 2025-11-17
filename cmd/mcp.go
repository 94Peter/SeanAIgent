/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/94peter/vulpes/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"seanAIgent/internal/db/factory"
	"seanAIgent/internal/mcp"
	"seanAIgent/internal/mcp/tool"
	_ "seanAIgent/internal/mcp/tool"
	"seanAIgent/internal/service"
)

// parentCmd represents the parent command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mcp called")
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

		factory.InjectStore(func(stores *factory.Stores) {
			svc := service.InitService(
				service.WithTrainingStore(stores.TrainingDateStore),
				service.WithAppointmentStore(stores.AppointmentStore),
			)
			tool.InitTool(svc)
		})

		mcp.Start()
	},
}

func init() {
	serveCmd.AddCommand(mcpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// parentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// parentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
