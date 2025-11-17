/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// checkHealthCmd represents the checkHealth command
var checkHealthCmd = &cobra.Command{
	Use:   "checkHealth",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		port := viper.Get("http.port")
		url := fmt.Sprintf("http://localhost:%d/health", port)
		resp, err := http.Get(url)
		if err != nil {
			os.Exit(1)
		}
		if resp.StatusCode != 200 {
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(checkHealthCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkHealthCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkHealthCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
