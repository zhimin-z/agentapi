package server

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/hugodutka/openagent/lib/httpapi"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the server",
	Long:  `Run the server`,
	Run: func(cmd *cobra.Command, args []string) {
		srv := httpapi.NewServer(8080)
		fmt.Println("Starting server on port 8080")
		if err := srv.Start(); err != nil && err != context.Canceled {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}
