package cmd

import (
	"fmt"
	"os"

	"github.com/hugodutka/openagent/cmd/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "openagent",
	Short: "OpenAgent CLI tool",
	Long:  `OpenAgent is a CLI tool for running various commands.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(server.ServerCmd)
}
