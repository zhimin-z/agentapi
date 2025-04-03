package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
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

		// Create command
		execCmd := exec.Command(program, programArgs...)

		// Create pty
		ptmx, err := pty.Start(execCmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting pty: %v\n", err)
			os.Exit(1)
		}
		defer ptmx.Close()

		// Buffer to store the output
		var outputBuffer bytes.Buffer

		// Determine output file path
		if outputFile == "" {
			// Generate default output file name based on timestamp
			timestamp := time.Now().Format("20060102-150405")
			outputFile = filepath.Join(".", fmt.Sprintf("output-%s.log", timestamp))
		}

		// Create or open the output file
		f, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		fmt.Printf("Saving output to: %s\n", outputFile)

		// Use MultiWriter to write to both stdout, buffer and file
		mw := io.MultiWriter(os.Stdout, &outputBuffer, f)

		// Handle window size changes
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		go func() {
			for range ch {
				if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
					fmt.Fprintf(os.Stderr, "Error resizing pty: %v\n", err)
				}
			}
		}()
		// Initial resize
		ch <- syscall.SIGWINCH

		// Set stdin in raw mode
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting terminal to raw mode: %v\n", err)
			os.Exit(1)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)

		// Channel to signal when the command is done
		cmdDone := make(chan struct{})

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

		// Handle output from the process
		go func() {
			io.Copy(mw, ptmx)
			close(cmdDone)
		}()

		// Handle input to the process
		go func() {
			io.Copy(ptmx, os.Stdin)
			// Don't close anything here - the process might still be running
		}()

		// Wait for command to finish
		cmdErr := execCmd.Wait()
		
		// Wait for all output to be processed
		<-cmdDone
		
		// Stop handling signals
		signal.Stop(ch)
		signal.Stop(signalCh)
		close(ch)
		close(signalCh)

		// Restore terminal before printing final message
		term.Restore(int(os.Stdin.Fd()), oldState)
		
		if cmdErr != nil {
			fmt.Fprintf(os.Stderr, "\nCommand exited with error: %v\n", cmdErr)
			os.Exit(1)
		}

		fmt.Printf("\nCommand completed successfully. Output captured to %s\n", outputFile)
	},
}