package screentracker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/xerrors"
)

type screenSnapshot struct {
	timestamp time.Time
	screen    string
}

type AgentIO interface {
	Write(data []byte) (int, error)
	ReadScreen() string
}

type ConversationConfig struct {
	AgentIO AgentIO
	// GetTime returns the current time
	GetTime func() time.Time
	// How often to take a snapshot for the stability check
	SnapshotInterval time.Duration
	// How long the screen should not change to be considered stable
	ScreenStabilityLength time.Duration
}

type ConversationRole string

const (
	ConversationRoleUser  ConversationRole = "user"
	ConversationRoleAgent ConversationRole = "agent"
)

type ConversationMessage struct {
	Message string
	Role    ConversationRole
	Time    time.Time
}

type Conversation struct {
	cfg ConversationConfig
	// How many stable snapshots are required to consider the screen stable
	stableSnapshotsThreshold    int
	snapshotBuffer              *RingBuffer[screenSnapshot]
	messages                    []ConversationMessage
	screenBeforeLastUserMessage string
	lock                        sync.Mutex
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
		messages: []ConversationMessage{
			{
				Message: "",
				Role:    ConversationRoleAgent,
				Time:    cfg.GetTime(),
			},
		},
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
				c.AddSnapshot(c.cfg.AgentIO.ReadScreen())
			}
		}
	}()
}

func FindNewMessage(oldScreen, newScreen string) string {
	oldLines := strings.Split(oldScreen, "\n")
	newLines := strings.Split(newScreen, "\n")
	oldLinesMap := make(map[string]bool)
	for _, line := range oldLines {
		oldLinesMap[line] = true
	}
	firstNonMatchingLine := len(newLines)
	lastNonMatchingLine := 0
	for i, line := range newLines {
		if !oldLinesMap[line] {
			firstNonMatchingLine = i
			break
		}
	}
	for i := len(newLines) - 1; i >= 0; i-- {
		if !oldLinesMap[newLines[i]] {
			lastNonMatchingLine = i
			break
		}
	}
	newSectionLines := newLines[firstNonMatchingLine : lastNonMatchingLine+1]

	// remove leading and trailing lines which are empty or have only whitespace
	startLine := 0
	endLine := len(newSectionLines) - 1
	for i := 0; i < len(newSectionLines); i++ {
		if strings.TrimSpace(newSectionLines[i]) != "" {
			startLine = i
			break
		}
	}
	for i := len(newSectionLines) - 1; i >= 0; i-- {
		if strings.TrimSpace(newSectionLines[i]) != "" {
			endLine = i
			break
		}
	}
	return strings.Join(newSectionLines[startLine:endLine+1], "\n")
}

// This function assumes that the caller holds the lock
func (c *Conversation) updateLastAgentMessage(screen string, timestamp time.Time) {
	agentMessage := FindNewMessage(c.screenBeforeLastUserMessage, screen)
	shouldCreateNewMessage := len(c.messages) == 0 || c.messages[len(c.messages)-1].Role == ConversationRoleUser
	conversationMessage := ConversationMessage{
		Message: agentMessage,
		Role:    ConversationRoleAgent,
		Time:    timestamp,
	}
	if shouldCreateNewMessage {
		c.messages = append(c.messages, conversationMessage)
	} else {
		c.messages[len(c.messages)-1] = conversationMessage
	}
}

func (c *Conversation) AddSnapshot(screen string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	snapshot := screenSnapshot{
		timestamp: c.cfg.GetTime(),
		screen:    screen,
	}
	c.snapshotBuffer.Add(snapshot)
	c.updateLastAgentMessage(screen, snapshot.timestamp)
}

type MessagePart interface {
	Do(writer AgentIO) error
	String() string
}

type MessagePartText struct {
	Content string
	Alias   string
	Hidden  bool
}

func (p MessagePartText) Do(writer AgentIO) error {
	_, err := writer.Write([]byte(p.Content))
	return err
}

func (p MessagePartText) String() string {
	if p.Hidden {
		return ""
	}
	if p.Alias != "" {
		return p.Alias
	}
	return p.Content
}

type MessagePartWait struct {
	Duration time.Duration
}

func (p MessagePartWait) Do(writer AgentIO) error {
	time.Sleep(p.Duration)
	return nil
}

func (p MessagePartWait) String() string {
	return ""
}

func PartsToString(parts ...MessagePart) string {
	var sb strings.Builder
	for _, part := range parts {
		sb.WriteString(part.String())
	}
	return sb.String()
}

func ExecuteParts(writer AgentIO, parts ...MessagePart) error {
	for _, part := range parts {
		if err := part.Do(writer); err != nil {
			return xerrors.Errorf("failed to write message part: %w", err)
		}
	}
	return nil
}

func (c *Conversation) SendMessage(messageParts ...MessagePart) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	screenBeforeMessage := c.cfg.AgentIO.ReadScreen()
	now := c.cfg.GetTime()
	c.updateLastAgentMessage(screenBeforeMessage, now)

	if err := ExecuteParts(c.cfg.AgentIO, messageParts...); err != nil {
		return xerrors.Errorf("failed to send message: %w", err)
	}

	c.screenBeforeLastUserMessage = screenBeforeMessage
	c.messages = append(c.messages, ConversationMessage{
		Message: PartsToString(messageParts...),
		Role:    ConversationRoleUser,
		Time:    now,
	})
	return nil
}

func (c *Conversation) Status() ConversationStatus {
	c.lock.Lock()
	defer c.lock.Unlock()

	// sanity checks
	if c.snapshotBuffer.Capacity() != c.stableSnapshotsThreshold {
		panic(fmt.Sprintf("snapshot buffer capacity %d is not equal to snapshot threshold %d. can't check stability", c.snapshotBuffer.Capacity(), c.stableSnapshotsThreshold))
	}
	if c.stableSnapshotsThreshold == 0 {
		panic("stable snapshots threshold is 0. can't check stability")
	}

	snapshots := c.snapshotBuffer.GetAll()
	if len(c.messages) > 0 && c.messages[len(c.messages)-1].Role == ConversationRoleUser {
		// if the last message is a user message then the snapshot loop hasn't
		// been triggered since the last user message, and we should assume
		// the screen is changing
		return ConversationStatusChanging
	}

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

func (c *Conversation) Messages() []ConversationMessage {
	c.lock.Lock()
	defer c.lock.Unlock()

	result := make([]ConversationMessage, len(c.messages))
	copy(result, c.messages)
	return result
}
