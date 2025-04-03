package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [program] [args...]",
	Short: "Run a program with the supplied arguments",
	Long:  `Run executes the specified program with any arguments provided.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		program := args[0]
		programArgs := args[1:]

		fmt.Printf("Running: %s %s\n", program, strings.Join(programArgs, " "))

		execCmd := exec.Command(program, programArgs...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.Stdin = os.Stdin

		err := execCmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running command: %v\n", err)
			os.Exit(1)
		}
	},
}
