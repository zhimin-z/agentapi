package screentracker

import (
	"context"
	"fmt"
	"time"
)

type screenSnapshot struct {
	timestamp time.Time
	screen    string
}

type ConversationConfig struct {
	GetScreen                func() string
	SnapshotBufferSize       int
	SnapshotInterval         time.Duration
	ScreenStabilityThreshold time.Duration
}

type Conversation struct {
	cfg            ConversationConfig
	snapshotBuffer *RingBuffer[screenSnapshot]
}

type ConversationStatus string

const (
	ConversationStatusChanging ConversationStatus = "changing"
	ConversationStatusStable   ConversationStatus = "stable"
)

func NewConversation(ctx context.Context, cfg ConversationConfig) *Conversation {
	c := &Conversation{
		cfg:            cfg,
		snapshotBuffer: NewRingBuffer[screenSnapshot](cfg.SnapshotBufferSize),
	}
	return c
}

func (c *Conversation) StartSnapshotLoop(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(c.cfg.SnapshotInterval):
				c.AddSnapshot(c.cfg.GetScreen())
			}
		}
	}()
}

func (c *Conversation) AddSnapshot(screen string) {
	c.snapshotBuffer.Add(screenSnapshot{
		timestamp: time.Now(),
		screen:    screen,
	})
}

func (c *Conversation) Status() ConversationStatus {
	// Check whether the snapshots have changed within
	// the screenStabilityThreshold
	snapshots := c.snapshotBuffer.GetAll()

	if len(snapshots) == 0 {
		return ConversationStatusStable
	}

	a := c.cfg.ScreenStabilityThreshold.Milliseconds()
	b := c.cfg.SnapshotInterval.Milliseconds()

	ticks := int(a / b)
	if a%b != 0 {
		ticks++
	}

	if c.cfg.SnapshotBufferSize < ticks+1 {
		panic(fmt.Sprintf("snapshot buffer size %d is less than ticks %d. can't check stability", c.cfg.SnapshotBufferSize, ticks))
	}

	if len(snapshots) < ticks+1 {
		return ConversationStatusChanging
	}

	latestSnapshot := snapshots[len(snapshots)-1]
	for i := 1; i <= ticks; i++ {
		if snapshots[len(snapshots)-1-i].screen != latestSnapshot.screen {
			return ConversationStatusChanging
		}
	}
	return ConversationStatusStable
}
