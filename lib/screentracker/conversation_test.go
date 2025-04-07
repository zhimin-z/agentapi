package screentracker_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	st "github.com/coder/openagent/lib/screentracker"
)

type statusTestStep struct {
	snapshot string
	status   st.ConversationStatus
}
type statusTestParams struct {
	cfg   st.ConversationConfig
	steps []statusTestStep
}

func statusTest(t *testing.T, params statusTestParams) {
	ctx := context.Background()
	t.Run(fmt.Sprintf("interval-%s,stability_length-%s", params.cfg.SnapshotInterval, params.cfg.ScreenStabilityLength), func(t *testing.T) {
		if params.cfg.GetTime == nil {
			params.cfg.GetTime = func() time.Time { return time.Now() }
		}
		c := st.NewConversation(ctx, params.cfg)
		assert.Equal(t, st.ConversationStatusInitializing, c.Status())

		for i, step := range params.steps {
			c.AddSnapshot(step.snapshot)
			assert.Equal(t, step.status, c.Status(), "step %d", i)
		}
	})
}

func TestConversation(t *testing.T) {
	changing := st.ConversationStatusChanging
	stable := st.ConversationStatusStable
	initializing := st.ConversationStatusInitializing

	statusTest(t, statusTestParams{
		cfg: st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			// stability threshold: 3
		},
		steps: []statusTestStep{
			{snapshot: "1", status: initializing},
			{snapshot: "1", status: initializing},
			{snapshot: "1", status: stable},
			{snapshot: "1", status: stable},
			{snapshot: "2", status: changing},
		},
	})

	statusTest(t, statusTestParams{
		cfg: st.ConversationConfig{
			SnapshotInterval:      2 * time.Second,
			ScreenStabilityLength: 3 * time.Second,
			// stability threshold: 3
		},
		steps: []statusTestStep{
			{snapshot: "1", status: initializing},
			{snapshot: "1", status: initializing},
			{snapshot: "1", status: stable},
			{snapshot: "1", status: stable},
			{snapshot: "2", status: changing},
			{snapshot: "2", status: changing},
			{snapshot: "2", status: stable},
			{snapshot: "2", status: stable},
			{snapshot: "2", status: stable},
		},
	})

	statusTest(t, statusTestParams{
		cfg: st.ConversationConfig{
			SnapshotInterval:      6 * time.Second,
			ScreenStabilityLength: 14 * time.Second,
			// stability threshold: 4
		},
		steps: []statusTestStep{
			{snapshot: "1", status: initializing},
			{snapshot: "1", status: initializing},
			{snapshot: "1", status: initializing},
			{snapshot: "1", status: stable},
			{snapshot: "1", status: stable},
			{snapshot: "1", status: stable},
			{snapshot: "2", status: changing},
			{snapshot: "2", status: changing},
			{snapshot: "2", status: changing},
			{snapshot: "2", status: stable},
		},
	})
}

func TestMessages(t *testing.T) {
	now := time.Now()
	agentMsg := func(msg string) st.ConversationMessage {
		return st.ConversationMessage{
			Message: msg,
			Role:    st.ConversationRoleAgent,
			Time:    now,
		}
	}
	userMsg := func(msg string) st.ConversationMessage {
		return st.ConversationMessage{
			Message: msg,
			Role:    st.ConversationRoleUser,
			Time:    now,
		}
	}
	t.Run("messages are copied", func(t *testing.T) {
		c := st.NewConversation(context.Background(), st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			GetTime:               func() time.Time { return now },
		})
		messages := c.Messages()
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(""),
		}, messages)

		messages[0].Message = "modification"

		assert.Equal(t, []st.ConversationMessage{
			agentMsg(""),
		}, c.Messages())
	})

	t.Run("tracking messages", func(t *testing.T) {
		screen := struct {
			content string
		}{
			content: "",
		}

		c := st.NewConversation(context.Background(), st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			GetTime:               func() time.Time { return now },
			GetScreen:             func() string { return screen.content },
			SendMessage:           func(msg string) error { return nil },
		})
		// agent message is recorded when the first snapshot is added
		c.AddSnapshot("1")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg("1"),
		}, c.Messages())

		// agent message is updated when the screen changes
		c.AddSnapshot("2")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg("2"),
		}, c.Messages())

		// user message is recorded
		screen.content = "2"
		assert.NoError(t, c.SendMessage("3"))
		assert.Equal(t, []st.ConversationMessage{
			agentMsg("2"),
			userMsg("3"),
		}, c.Messages())

		// agent message is added after a user message
		c.AddSnapshot("4")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg("2"),
			userMsg("3"),
			agentMsg("4"),
		}, c.Messages())

		// agent message is updated when the screen changes before a user message
		screen.content = "5"
		assert.NoError(t, c.SendMessage("6"))
		assert.Equal(t, []st.ConversationMessage{
			agentMsg("2"),
			userMsg("3"),
			agentMsg("5"),
			userMsg("6"),
		}, c.Messages())

		// conversation status is changing right after a user message
		c.AddSnapshot("7")
		c.AddSnapshot("7")
		c.AddSnapshot("7")
		assert.Equal(t, st.ConversationStatusStable, c.Status())
		screen.content = "7"
		assert.NoError(t, c.SendMessage("8"))
		assert.Equal(t, []st.ConversationMessage{
			agentMsg("2"),
			userMsg("3"),
			agentMsg("5"),
			userMsg("6"),
			agentMsg("7"),
			userMsg("8"),
		}, c.Messages())
		assert.Equal(t, st.ConversationStatusChanging, c.Status())

		// conversation status is back to stable after a snapshot that
		// doesn't change the screen
		c.AddSnapshot("7")
		assert.Equal(t, st.ConversationStatusStable, c.Status())
	})

	t.Run("tracking messages overlap", func(t *testing.T) {
		screen := struct {
			content string
		}{
			content: "",
		}
		c := st.NewConversation(context.Background(), st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			GetTime:               func() time.Time { return now },
			GetScreen:             func() string { return screen.content },
			SendMessage:           func(msg string) error { return nil },
		})

		// common overlap between screens is removed after a user message
		c.AddSnapshot("1")
		screen.content = "1"
		c.SendMessage("2")
		c.AddSnapshot("13")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg("1"),
			userMsg("2"),
			agentMsg("3"),
		}, c.Messages())

		screen.content = "13x"
		c.SendMessage("4")
		c.AddSnapshot("13x5")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg("1"),
			userMsg("2"),
			agentMsg("3x"),
			userMsg("4"),
			agentMsg("5"),
		}, c.Messages())
	})
}
