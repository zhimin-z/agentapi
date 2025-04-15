package cmd

import (
	"fmt"
	"os"

	"github.com/coder/agentapi/cmd/attach"
	"github.com/coder/agentapi/cmd/server"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agentapi",
	Short: "AgentAPI CLI tool",
	Long:  `AgentAPI is a CLI tool for running various commands.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(server.ServerCmd)
	rootCmd.AddCommand(attach.AttachCmd)
}
