/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"seanAIgent/internal/booking/usecase/core"
	"seanAIgent/internal/db/factory"
	"sync"
	"time"

	"github.com/94peter/vulpes/db/mgo"
	"github.com/94peter/vulpes/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/trace/noop"
)

// v1tov2Cmd represents the v1tov2 command
var v1tov2Cmd = &cobra.Command{
	Use:   "v1tov2",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mainCtx, mainCancel := context.WithCancel(context.Background())
		defer mainCancel()

		dbCtx, cancel := context.WithTimeout(mainCtx, time.Minute)
		err := factory.InitializeDb(
			dbCtx,
			factory.WithMongoDB(
				viper.GetString("database.uri"),
				viper.GetString("database.db"),
			),
			factory.WithMongoDBPoolSize(
				viper.GetUint64("database.max_pool_size"),
				viper.GetUint64("database.min_pool_size"),
			),
			factory.WithTracer(noop.NewTracerProvider().Tracer("migration")),
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

		fmt.Println("v1tov2 called")
		uc := GetMigrationUseCaseSet()
		ctx := cmd.Context()
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := uc.TrainDataMigrationV1ToV2.Execute(ctx, core.Empty{})
			if err != nil {
				fmt.Println("TrainDataMigrationV1ToV2 fail:", err, errors.Unwrap(err))
			}
		}()
		go func() {
			defer wg.Done()
			_, err := uc.ApptMigrationV1ToV2.Execute(ctx, core.Empty{})
			if err != nil {
				fmt.Println("ApptMigrationV1ToV2 fail:", err, errors.Unwrap(err))
			}
		}()
		wg.Wait()
		return nil
	},
}

func init() {
	migrateCmd.AddCommand(v1tov2Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// v1tov2Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// v1tov2Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
