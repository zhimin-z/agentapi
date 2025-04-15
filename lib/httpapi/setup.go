package httpapi

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coder/agentapi/lib/logctx"
	"github.com/coder/agentapi/lib/termexec"
)

func SetupProcess(ctx context.Context, program string, programArgs ...string) (*termexec.Process, error) {
	logger := logctx.From(ctx)

	logger.Info(fmt.Sprintf("Running: %s %s", program, strings.Join(programArgs, " ")))

	process, err := termexec.StartProcess(ctx, termexec.StartProcessConfig{
		Program:        program,
		Args:           programArgs,
		TerminalWidth:  80,
		TerminalHeight: 1000,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Error starting process: %v", err))
		os.Exit(1)
	}

	// Handle SIGINT (Ctrl+C) and send it to the process
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		if err := process.Close(logger, 5*time.Second); err != nil {
			logger.Error("Error closing process", "error", err)
		}
	}()

	return process, nil
}
