package util

import (
	"context"
	"time"

	"golang.org/x/xerrors"
)

type WaitTimeout struct {
	Timeout     time.Duration
	MinInterval time.Duration
	MaxInterval time.Duration
	InitialWait bool
}

var WaitTimedOut = xerrors.New("timeout waiting for condition")

// WaitFor waits for a condition to be true or the timeout to expire.
// It will wait for the condition to be true with exponential backoff.
func WaitFor(ctx context.Context, timeout WaitTimeout, condition func() (bool, error)) error {
	minInterval := timeout.MinInterval
	maxInterval := timeout.MaxInterval
	timeoutDuration := timeout.Timeout
	if minInterval == 0 {
		minInterval = 10 * time.Millisecond
	}
	if maxInterval == 0 {
		maxInterval = 500 * time.Millisecond
	}
	if timeoutDuration == 0 {
		timeoutDuration = 10 * time.Second
	}
	timeoutAfter := time.After(timeoutDuration)

	if minInterval > maxInterval {
		return xerrors.Errorf("minInterval is greater than maxInterval")
	}

	interval := minInterval
	if timeout.InitialWait {
		time.Sleep(interval)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeoutAfter:
			return WaitTimedOut
		default:
			ok, err := condition()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
			time.Sleep(interval)
			interval = min(interval*2, maxInterval)
		}
	}
}
