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
	// GetScreen returns the current screen snapshot
	GetScreen func() string
	// How often to take a snapshot for the stability check
	SnapshotInterval time.Duration
	// How long the screen should not change to be considered stable
	ScreenStabilityLength time.Duration
}

type Conversation struct {
	cfg ConversationConfig
	// How many stable snapshots are required to consider the screen stable
	stableSnapshotsThreshold int
	snapshotBuffer           *RingBuffer[screenSnapshot]
}

type ConversationStatus string

const (
	ConversationStatusChanging     ConversationStatus = "changing"
	ConversationStatusStable       ConversationStatus = "stable"
	ConversationStatusInitializing ConversationStatus = "initializing"
)

func getStableSnapshotsThreshold(cfg ConversationConfig) int {
	length := cfg.ScreenStabilityLength.Milliseconds()
	interval := cfg.SnapshotInterval.Milliseconds()
	threshold := int(length / interval)
	if length%interval != 0 {
		threshold++
	}
	return threshold + 1
}

func NewConversation(ctx context.Context, cfg ConversationConfig) *Conversation {
	threshold := getStableSnapshotsThreshold(cfg)
	c := &Conversation{
		cfg:                      cfg,
		stableSnapshotsThreshold: threshold,
		snapshotBuffer:           NewRingBuffer[screenSnapshot](threshold),
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
	// sanity checks
	if c.snapshotBuffer.Capacity() != c.stableSnapshotsThreshold {
		panic(fmt.Sprintf("snapshot buffer capacity %d is not equal to snapshot threshold %d. can't check stability", c.snapshotBuffer.Capacity(), c.stableSnapshotsThreshold))
	}
	if c.stableSnapshotsThreshold == 0 {
		panic("stable snapshots threshold is 0. can't check stability")
	}

	snapshots := c.snapshotBuffer.GetAll()
	if len(snapshots) != c.stableSnapshotsThreshold {
		return ConversationStatusInitializing
	}

	for i := 1; i < len(snapshots); i++ {
		if snapshots[0].screen != snapshots[i].screen {
			return ConversationStatusChanging
		}
	}
	return ConversationStatusStable
}
