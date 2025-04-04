package conversation_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hugodutka/openagent/lib/conversation"
)

type statusTestStep struct {
	snapshot string
	status   conversation.ConversationStatus
}
type statusTestParams struct {
	cfg   conversation.ConversationConfig
	steps []statusTestStep
}

func statusTest(t *testing.T, params statusTestParams) {
	ctx := context.Background()
	t.Run(fmt.Sprintf("buffer_size-%d,interval-%s,threshold-%s", params.cfg.SnapshotBufferSize, params.cfg.SnapshotInterval, params.cfg.ScreenStabilityThreshold), func(t *testing.T) {
		c := conversation.NewConversation(ctx, params.cfg)
		assert.Equal(t, conversation.ConversationStatusStable, c.Status())

		for i, step := range params.steps {
			c.AddSnapshot(step.snapshot)
			assert.Equal(t, step.status, c.Status(), "step %d", i)
		}
	})
}

func TestConversation(t *testing.T) {
	changing := conversation.ConversationStatusChanging
	stable := conversation.ConversationStatusStable

	statusTest(t, statusTestParams{
		cfg: conversation.ConversationConfig{
			SnapshotBufferSize:       5,
			SnapshotInterval:         1 * time.Second,
			ScreenStabilityThreshold: 2 * time.Second,
			// ticks: 2
		},
		steps: []statusTestStep{
			{snapshot: "1", status: changing},
			{snapshot: "1", status: changing},
			{snapshot: "1", status: stable},
			{snapshot: "1", status: stable},
			{snapshot: "2", status: changing},
		},
	})

	statusTest(t, statusTestParams{
		cfg: conversation.ConversationConfig{
			SnapshotBufferSize:       5,
			SnapshotInterval:         2 * time.Second,
			ScreenStabilityThreshold: 3 * time.Second,
			// ticks: 2
		},
		steps: []statusTestStep{
			{snapshot: "1", status: changing},
			{snapshot: "1", status: changing},
			{snapshot: "1", status: stable},
			{snapshot: "1", status: stable},
			{snapshot: "2", status: changing},
		},
	})

	statusTest(t, statusTestParams{
		cfg: conversation.ConversationConfig{
			SnapshotBufferSize:       5,
			SnapshotInterval:         6 * time.Second,
			ScreenStabilityThreshold: 14 * time.Second,
			// ticks: 3
		},
		steps: []statusTestStep{
			{snapshot: "1", status: changing},
			{snapshot: "1", status: changing},
			{snapshot: "1", status: changing},
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
