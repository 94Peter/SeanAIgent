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
	"go.opentelemetry.io/otel"
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
		mainCtx, mainCancel := context.WithCancel(context.Background())
		defer mainCancel()
		// init tracer
		shutdown, err := initTracer(mainCtx)
		if err != nil {
			log.Fatalf("initTracer fail: %v", err)
		}
		defer shutdown(mainCtx)

		dbTracer := otel.Tracer("Mongodb")
		// db connection
		dbCtx, cancel := context.WithTimeout(mainCtx, time.Minute)
		err = mgo.InitConnection(
			dbCtx,
			viper.GetString("database.db"),
			dbTracer,
			mgo.WithURI(viper.GetString("database.uri")),
			mgo.WithMinPoolSize(viper.GetUint64("database.min_pool_size")),
			mgo.WithMaxPoolSize(viper.GetUint64("database.max_pool_size")),
		)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = mgo.SyncIndexes(dbCtx)
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

		mcpServer := InitializeMCP()
		mcpServer.Start()
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
	// teacherCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
