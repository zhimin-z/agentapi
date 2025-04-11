package screentracker_test

import (
	"context"
	"embed"
	"fmt"
	"path"
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

type testAgent struct {
	st.AgentIO
	screen string
}

func (a *testAgent) ReadScreen() string {
	return a.screen
}

func (a *testAgent) Write(data []byte) (int, error) {
	return 0, nil
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
			AgentIO: &testAgent{
				screen: "1",
			},
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
	agentMsg := func(id int, msg string) st.ConversationMessage {
		return st.ConversationMessage{
			Id:      id,
			Message: msg,
			Role:    st.ConversationRoleAgent,
			Time:    now,
		}
	}
	userMsg := func(id int, msg string) st.ConversationMessage {
		return st.ConversationMessage{
			Id:      id,
			Message: msg,
			Role:    st.ConversationRoleUser,
			Time:    now,
		}
	}
	sendMsg := func(c *st.Conversation, msg string) error {
		return c.SendMessage(st.MessagePartText{Content: msg})
	}
	t.Run("messages are copied", func(t *testing.T) {
		c := st.NewConversation(context.Background(), st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			GetTime:               func() time.Time { return now },
		})
		messages := c.Messages()
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, ""),
		}, messages)

		messages[0].Message = "modification"

		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, ""),
		}, c.Messages())
	})

	t.Run("whitespace-padding", func(t *testing.T) {
		c := st.NewConversation(context.Background(), st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			GetTime:               func() time.Time { return now },
		})
		for _, msg := range []string{"123 ", " 123", "123\t\t", "\n123", "123\n\t", " \t123\n\t"} {
			err := c.SendMessage(st.MessagePartText{Content: msg})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "message must be trimmed of leading and trailing whitespace")
		}
	})

	t.Run("no-change-no-message-update", func(t *testing.T) {
		nowWrapper := struct {
			time.Time
		}{
			Time: now,
		}
		c := st.NewConversation(context.Background(), st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			GetTime:               func() time.Time { return nowWrapper.Time },
		})
		c.AddSnapshot("1")
		msgs := c.Messages()
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "1"),
		}, msgs)
		nowWrapper.Time = nowWrapper.Time.Add(1 * time.Second)
		c.AddSnapshot("1")
		assert.Equal(t, msgs, c.Messages())
	})

	t.Run("tracking messages", func(t *testing.T) {
		agent := &testAgent{}
		c := st.NewConversation(context.Background(), st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			GetTime:               func() time.Time { return now },
			AgentIO:               agent,
		})
		// agent message is recorded when the first snapshot is added
		c.AddSnapshot("1")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "1"),
		}, c.Messages())

		// agent message is updated when the screen changes
		c.AddSnapshot("2")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "2"),
		}, c.Messages())

		// user message is recorded
		agent.screen = "2"
		assert.NoError(t, sendMsg(c, "3"))
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "2"),
			userMsg(1, "3"),
		}, c.Messages())

		// agent message is added after a user message
		c.AddSnapshot("4")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "2"),
			userMsg(1, "3"),
			agentMsg(2, "4"),
		}, c.Messages())

		// agent message is updated when the screen changes before a user message
		agent.screen = "5"
		assert.NoError(t, sendMsg(c, "6"))
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "2"),
			userMsg(1, "3"),
			agentMsg(2, "5"),
			userMsg(3, "6"),
		}, c.Messages())

		// conversation status is changing right after a user message
		c.AddSnapshot("7")
		c.AddSnapshot("7")
		c.AddSnapshot("7")
		assert.Equal(t, st.ConversationStatusStable, c.Status())
		agent.screen = "7"
		assert.NoError(t, sendMsg(c, "8"))
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "2"),
			userMsg(1, "3"),
			agentMsg(2, "5"),
			userMsg(3, "6"),
			agentMsg(4, "7"),
			userMsg(5, "8"),
		}, c.Messages())
		assert.Equal(t, st.ConversationStatusChanging, c.Status())

		// conversation status is back to stable after a snapshot that
		// doesn't change the screen
		c.AddSnapshot("7")
		assert.Equal(t, st.ConversationStatusStable, c.Status())
	})

	t.Run("tracking messages overlap", func(t *testing.T) {
		agent := &testAgent{}
		c := st.NewConversation(context.Background(), st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			GetTime:               func() time.Time { return now },
			AgentIO:               agent,
		})

		// common overlap between screens is removed after a user message
		c.AddSnapshot("1")
		agent.screen = "1"
		assert.NoError(t, sendMsg(c, "2"))
		c.AddSnapshot("1\n3")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "1"),
			userMsg(1, "2"),
			agentMsg(2, "3"),
		}, c.Messages())

		agent.screen = "1\n3x"
		assert.NoError(t, sendMsg(c, "4"))
		c.AddSnapshot("1\n3x\n5")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "1"),
			userMsg(1, "2"),
			agentMsg(2, "3x"),
			userMsg(3, "4"),
			agentMsg(4, "5"),
		}, c.Messages())
	})

	t.Run("format-message", func(t *testing.T) {
		agent := &testAgent{}
		c := st.NewConversation(context.Background(), st.ConversationConfig{
			SnapshotInterval:      1 * time.Second,
			ScreenStabilityLength: 2 * time.Second,
			GetTime:               func() time.Time { return now },
			AgentIO:               agent,
			FormatMessage: func(message string, userInput string) string {
				return message + " " + userInput
			},
		})
		agent.screen = "1"
		assert.NoError(t, sendMsg(c, "2"))
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "1 "),
			userMsg(1, "2"),
		}, c.Messages())
		agent.screen = "x"
		c.AddSnapshot("x")
		assert.Equal(t, []st.ConversationMessage{
			agentMsg(0, "1 "),
			userMsg(1, "2"),
			agentMsg(2, "x 2"),
		}, c.Messages())
	})
}

//go:embed testdata
var testdataDir embed.FS

func TestFindNewMessage(t *testing.T) {
	assert.Equal(t, "", st.FindNewMessage("123456", "123456"))
	assert.Equal(t, "1234567", st.FindNewMessage("123456", "1234567"))
	assert.Equal(t, "42", st.FindNewMessage("123", "123\n  \n \n \n42"))
	assert.Equal(t, "12342", st.FindNewMessage("123", "12342\n   \n \n \n"))
	assert.Equal(t, "42", st.FindNewMessage("123", "123\n  \n \n \n42\n   \n \n \n"))
	assert.Equal(t, "42", st.FindNewMessage("89", "42"))

	dir := "testdata/diff"
	cases, err := testdataDir.ReadDir(dir)
	assert.NoError(t, err)
	for _, c := range cases {
		t.Run(c.Name(), func(t *testing.T) {
			before, err := testdataDir.ReadFile(path.Join(dir, c.Name(), "before.txt"))
			assert.NoError(t, err)
			after, err := testdataDir.ReadFile(path.Join(dir, c.Name(), "after.txt"))
			assert.NoError(t, err)
			expected, err := testdataDir.ReadFile(path.Join(dir, c.Name(), "expected.txt"))
			assert.NoError(t, err)
			assert.Equal(t, string(expected), st.FindNewMessage(string(before), string(after)))
		})
	}
}

func TestPartsToString(t *testing.T) {
	assert.Equal(t, "123", st.PartsToString(st.MessagePartText{Content: "123"}))
	assert.Equal(t,
		"123",
		st.PartsToString(
			st.MessagePartText{Content: "1"},
			st.MessagePartText{Content: "2"},
			st.MessagePartText{Content: "3"},
		),
	)
	assert.Equal(t,
		"123",
		st.PartsToString(
			st.MessagePartText{Content: "1"},
			st.MessagePartText{Content: "x", Hidden: true},
			st.MessagePartText{Content: "2"},
			st.MessagePartText{Content: "3"},
			st.MessagePartText{Content: "y", Hidden: true},
		),
	)
	assert.Equal(t,
		"123",
		st.PartsToString(
			st.MessagePartText{Content: "1"},
			st.MessagePartWait{Duration: 1 * time.Second},
			st.MessagePartText{Content: "2"},
			st.MessagePartWait{Duration: 1 * time.Second},
			st.MessagePartText{Content: "3"},
			st.MessagePartWait{Duration: 1 * time.Second},
		),
	)
	assert.Equal(t,
		"ab",
		st.PartsToString(
			st.MessagePartText{Content: "1", Alias: "a"},
			st.MessagePartWait{Duration: 1 * time.Second},
			st.MessagePartText{Content: "2", Alias: "b"},
			st.MessagePartText{Content: "3", Alias: "c", Hidden: true},
		),
	)
}

func TestConversationFormatMessage(t *testing.T) {
	agent := &testAgent{}

	now := time.Now()
	c := st.NewConversation(context.Background(), st.ConversationConfig{
		SnapshotInterval:      1 * time.Second,
		ScreenStabilityLength: 2 * time.Second,
		GetTime:               func() time.Time { return now },
		AgentIO:               agent,
		FormatMessage: func(message string, userInput string) string {
			return "formatted"
		},
	})
	assert.Equal(t, []st.ConversationMessage{
		{
			Id:      0,
			Message: "",
			Role:    st.ConversationRoleAgent,
			Time:    now,
		},
	}, c.Messages())

}
