package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/spf13/cobra"
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

		// Channel to signal when the goroutine is done
		done := make(chan struct{})

		// Create a buffer for chunk-wise reading
		buffer := make([]byte, 1024)

		// Read from pty in small chunks to ensure continuous writing to file
		go func() {
			defer close(done) // Signal that we're done when the goroutine exits
			
			for {
				n, err := ptmx.Read(buffer)
				if err != nil {
					if err != io.EOF {
						fmt.Fprintf(os.Stderr, "Error reading from pty: %v\n", err)
					}
					break
				}
				
				if n > 0 {
					// Write the chunk to our multiwriter (stdout, buffer, and file)
					_, err = mw.Write(buffer[:n])
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
						break
					}
					
					// Explicitly flush the file to ensure continuous writing
					f.Sync()
				}
			}
		}()

		// Wait for command to finish
		cmdErr := execCmd.Wait()
		
		// Wait for the goroutine to finish processing all output
		<-done
		
		if cmdErr != nil {
			fmt.Fprintf(os.Stderr, "Command exited with error: %v\n", cmdErr)
			os.Exit(1)
		}

		fmt.Printf("\nCommand completed successfully. Output captured to %s\n", outputFile)
	},
}