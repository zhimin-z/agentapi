package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/coder/openagent/lib/httpapi"
	"github.com/coder/openagent/lib/logctx"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the server",
	Long:  `Run the server`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		ctx := logctx.WithLogger(context.Background(), logger)
		process, cleanup, err := httpapi.SetupProcess(ctx, "claude")
		if err != nil {
			logger.Error("Failed to setup process", "error", err)
			os.Exit(1)
		}
		defer cleanup()
		srv := httpapi.NewServer(ctx, process, 8080)
		logger.Info("Starting server on port 8080")
		go func() {
			if err := process.Wait(); err != nil {
				logger.Error("Process exited with error", "error", err)
			}
			if err := srv.Stop(ctx); err != nil {
				logger.Error("Failed to stop server", "error", err)
			}
		}()
		if err := srv.Start(); err != nil && err != context.Canceled && err != http.ErrServerClosed {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	},
}
