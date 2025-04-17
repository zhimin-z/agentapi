package termexec

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/ActiveState/termtest/xpty"
	"github.com/coder/agentapi/lib/logctx"
	"golang.org/x/xerrors"
)

type Process struct {
	xp      *xpty.Xpty
	execCmd *exec.Cmd
}

type StartProcessConfig struct {
	Program        string
	Args           []string
	TerminalWidth  uint16
	TerminalHeight uint16
}

func StartProcess(ctx context.Context, args StartProcessConfig) (*Process, error) {
	logger := logctx.From(ctx)
	xp, err := xpty.New(args.TerminalWidth, args.TerminalHeight, false)
	if err != nil {
		return nil, err
	}
	execCmd := exec.Command(args.Program, args.Args...)
	// vt100 is the terminal type that the vt10x library emulates.
	// Setting this signals to the process that it should only use compatible
	// escape sequences.
	execCmd.Env = append(os.Environ(), "TERM=vt100")
	if err := xp.StartProcessInTerminal(execCmd); err != nil {
		return nil, err
	}

	go func() {
		for {
			// calling ReadRune updates the terminal state. without it,
			// xp.State will always return an empty string
			if _, _, err := xp.ReadRune(); err != nil {
				if err != io.EOF {
					logger.Error("Error reading from pseudo terminal", "error", err)
				}
				// TODO: handle this error better. if this happens, the terminal
				// state will never be updated anymore and the process will appear
				// unresponsive.
				return
			}
		}
	}()

	return &Process{xp: xp, execCmd: execCmd}, nil
}

func (p *Process) Signal(sig os.Signal) error {
	return p.execCmd.Process.Signal(sig)
}

// ReadScreen returns the contents of the terminal window.
func (p *Process) ReadScreen() string {
	return p.xp.State.String()
}

// Write sends input to the process via the pseudo terminal.
func (p *Process) Write(data []byte) (int, error) {
	return p.xp.TerminalInPipe().Write(data)
}

// Close closes the process using a SIGINT signal or forcefully killing it if the process
// does not exit after the timeout. It then closes the pseudo terminal.
func (p *Process) Close(logger *slog.Logger, timeout time.Duration) error {
	logger.Info("Closing process")
	if err := p.execCmd.Process.Signal(os.Interrupt); err != nil {
		return xerrors.Errorf("failed to send SIGINT to process: %w", err)
	}

	exited := make(chan error, 1)
	go func() {
		_, err := p.execCmd.Process.Wait()
		exited <- err
		close(exited)
	}()

	var exitErr error
	select {
	case <-time.After(timeout):
		if err := p.execCmd.Process.Kill(); err != nil {
			exitErr = xerrors.Errorf("failed to forcefully kill the process: %w", err)
		}
		// don't wait for the process to exit to avoid hanging indefinitely
		// if the process never exits
	case err := <-exited:
		var pathErr *os.SyscallError
		// ECHILD is expected if the process has already exited
		if err != nil && !(errors.As(err, &pathErr) && pathErr.Err == syscall.ECHILD) {
			exitErr = xerrors.Errorf("process exited with error: %w", err)
		}
	}
	if err := p.xp.Close(); err != nil {
		return xerrors.Errorf("failed to close pseudo terminal: %w, exitErr: %w", err, exitErr)
	}
	return exitErr
}

// Wait waits for the process to exit.
func (p *Process) Wait() error {
	state, err := p.execCmd.Process.Wait()
	if err != nil {
		return xerrors.Errorf("process exited with error: %w", err)
	}
	if state.ExitCode() != 0 {
		return xerrors.Errorf("non-zero exit code: %d", state.ExitCode())
	}
	return nil
}
