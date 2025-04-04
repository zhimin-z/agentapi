package httpapi

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/coder/openagent/lib/logctx"
	"github.com/coder/openagent/lib/termexec"
)

func SetupProcess(ctx context.Context, program string, programArgs ...string) (*termexec.Process, func(), error) {
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

	snapshotFilePath := filepath.Join(".", fmt.Sprintf("snapshot-%s.log", time.Now().Format("20060102-150405")))
	snapshotFile, err := os.Create(snapshotFilePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating snapshot file: %v", err))
		os.Exit(1)
	}
	cleanup := func() {
		if err := snapshotFile.Close(); err != nil {
			logger.Error(fmt.Sprintf("Error closing snapshot file: %v", err))
		}
	}

	go func() {
		logger.Info("Starting snapshot loop")
		for {
			time.Sleep(100 * time.Millisecond)
			// Truncate file and seek to beginning to overwrite completely
			if err := snapshotFile.Truncate(0); err != nil {
				logger.Error("Error truncating snapshot file", "error", err)
				os.Exit(1)
			}
			if _, err := snapshotFile.Seek(0, 0); err != nil {
				logger.Error("Error seeking in snapshot file", "error", err)
				os.Exit(1)
			}
			if _, err := snapshotFile.Write([]byte(process.ReadScreen())); err != nil {
				logger.Error("Error writing snapshot to file", "error", err)
				os.Exit(1)
			}
		}
	}()

	// Handle SIGINT (Ctrl+C) and send it to the process
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalCh
		if err := process.Close(logger, 5*time.Second); err != nil {
			logger.Error("Error closing process", "error", err)
		}
	}()

	return process, cleanup, nil
}
