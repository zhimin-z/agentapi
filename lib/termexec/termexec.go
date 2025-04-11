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
	xp, err := xpty.New(args.TerminalWidth, args.TerminalHeight, true)
	if err != nil {
		return nil, err
	}
	execCmd := exec.Command(args.Program, args.Args...)
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

// Paste sends data enclosed in bracketed paste mode sequences to the process via the pseudo terminal.
func (p *Process) Paste(data []byte) error {
	// Bracketed paste mode escape sequences
	startSeq := []byte("\x1b[200~")
	endSeq := []byte("\x1b[201~")

	// janky hack: send a random character and then a backspace because otherwise
	// Claude Code echoes the startSeq back to the terminal.
	// This basically simulates a user typing and then removing the character.
	firstKeystrokes := []byte("x\b")

	payload := make([]byte, len(firstKeystrokes)+len(startSeq)+len(data)+len(endSeq))
	copy(payload, firstKeystrokes)
	copy(payload[len(firstKeystrokes):], startSeq)
	copy(payload[len(firstKeystrokes)+len(startSeq):], data)
	copy(payload[len(firstKeystrokes)+len(startSeq)+len(data):], endSeq)

	if _, err := p.Write(payload); err != nil {
		return xerrors.Errorf("failed to write paste data: %w", err)
	}

	// wait because Claude Code doesn't recognize "\r" as a command
	// to process the input if it's sent right away
	time.Sleep(50 * time.Millisecond)

	if _, err := p.Write([]byte("\r")); err != nil {
		return xerrors.Errorf("failed to write newline after paste: %w", err)
	}

	return nil
}

// Closecloses the process using a SIGINT signal or forcefully killing it if the process
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
	if _, err := p.execCmd.Process.Wait(); err != nil {
		return xerrors.Errorf("process exited with error: %w", err)
	}
	return nil
}
