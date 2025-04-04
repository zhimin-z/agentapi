package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ActiveState/termtest/xpty"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	outputFile string
)

func init() {
	runCmd.Flags().StringVarP(&outputFile, "output", "o", "", "File to save command output")
}

var runCmd = &cobra.Command{
	Use:   "run [program] [args...]",
	Short: "Run a program with the supplied arguments",
	Long:  `Run executes the specified program with any arguments provided.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		program := args[0]
		programArgs := args[1:]

		fmt.Printf("Running: %s %s\n", program, strings.Join(programArgs, " "))

		xp, _ := xpty.New(80, 1000, true)
		defer xp.Close()

		execCmd := exec.Command(program, programArgs...)
		if err := xp.StartProcessInTerminal(execCmd); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting process in terminal: %v\n", err)
			os.Exit(1)
		}
		execCmd.Stdin = os.Stdin

		snapshotFilePath := filepath.Join(".", fmt.Sprintf("snapshot-%s.log", time.Now().Format("20060102-150405")))
		snapshotFile2Path := filepath.Join(".", fmt.Sprintf("snapshot-%s-2.log", time.Now().Format("20060102-150405")))
		snapshotFile, err := os.Create(snapshotFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating snapshot file: %v\n", err)
			os.Exit(1)
		}
		defer snapshotFile.Close()
		snapshotFile2, err := os.Create(snapshotFile2Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating snapshot file: %v\n", err)
			os.Exit(1)
		}
		defer snapshotFile2.Close()

		go func() {
			// this will block until the command is finished
			if _, err := xp.WriteTo(snapshotFile2); err != nil && err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error writing snapshot to buf: %v\n", err)
				os.Exit(1)
			}
		}()

		go func() {
			fmt.Println("Starting snapshot loop")
			for {
				time.Sleep(100 * time.Millisecond)
				// Truncate file and seek to beginning to overwrite completely
				if err := snapshotFile.Truncate(0); err != nil {
					fmt.Fprintf(os.Stderr, "Error truncating snapshot file: %v\n", err)
					os.Exit(1)
				}
				if _, err := snapshotFile.Seek(0, 0); err != nil {
					fmt.Fprintf(os.Stderr, "Error seeking in snapshot file: %v\n", err)
					os.Exit(1)
				}
				if _, err := snapshotFile.Write([]byte(xp.State.String())); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing snapshot to file: %v\n", err)
					os.Exit(1)
				}
			}
		}()

		// Set stdin in raw mode
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting terminal to raw mode: %v\n", err)
			os.Exit(1)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		// Handle SIGINT (Ctrl+C) and send it to the process
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			for sig := range signalCh {
				// Forward the signal to the process
				if execCmd.Process != nil {
					execCmd.Process.Signal(sig)
				}
			}
		}()

		// Handle input to the process
		go func() {
			io.Copy(xp.TerminalInPipe(), os.Stdin)
		}()

		// Channel to receive the result of cmd.Wait()
		cmdDone := make(chan error, 1)
		go func() {
			cmdDone <- execCmd.Wait()
		}()

		// Channel to receive OS signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		fmt.Println("Waiting for command to finish...")

		select {
		case err := <-cmdDone:
			if err != nil {
				// Check if the error is due to the process being killed by a signal
				if exitErr, ok := err.(*exec.ExitError); ok {
					if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						// Don't print an error if killed by SIGTERM or SIGINT, as that's expected on signal
						if status.Signaled() && (status.Signal() == syscall.SIGTERM || status.Signal() == syscall.SIGINT) {
							fmt.Printf("Command terminated by signal: %s\n", status.Signal())
						} else {
							fmt.Fprintf(os.Stderr, "Command finished with error: %v\n", err)
						}
					} else {
						fmt.Fprintf(os.Stderr, "Command finished with error: %v\n", err)
					}
				} else {
					fmt.Fprintf(os.Stderr, "Command finished with error: %v\n", err)
				}
			} else {
				fmt.Println("Command finished successfully.")
			}
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal: %v. Cleaning up and exiting...\n", sig)
			if err := execCmd.Process.Signal(sig); err != nil {
				fmt.Fprintf(os.Stderr, "Error signaling command: %v\n", err)
				os.Exit(1)
			}
			if err := xp.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing terminal: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Terminal closed")
			<-cmdDone // Wait for the command to actually exit after signaling
			fmt.Println("Command finished")
		}

	},
}
