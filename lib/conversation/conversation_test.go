package conversation_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hugodutka/openagent/lib/conversation"
)

type statusTestStep struct {
	snapshot       string
	expectedStatus conversation.ConversationStatus
}
type statusTestParams struct {
	cfg   conversation.ConversationConfig
	steps []statusTestStep
}

func statusTest(t *testing.T, params statusTestParams) {
	ctx := context.Background()
	c := conversation.NewConversation(ctx, params.cfg)
	assert.Equal(t, conversation.ConversationStatusStable, c.Status())

	for i, step := range params.steps {
		c.AddSnapshot(step.snapshot)
		assert.Equal(t, step.expectedStatus, c.Status(), "step %d", i)
	}
}

func TestConversation(t *testing.T) {
	t.Run("buffer size 5, interval 1s, threshold 2s", func(t *testing.T) {
		statusTest(t, statusTestParams{
			cfg: conversation.ConversationConfig{
				GetScreen: func() string {
					return ""
				},
				SnapshotBufferSize:       5,
				SnapshotInterval:         1 * time.Second,
				ScreenStabilityThreshold: 2 * time.Second,
			},
			steps: []statusTestStep{
				{snapshot: "Hello, world!", expectedStatus: conversation.ConversationStatusChanging},
				{snapshot: "Hello, world!", expectedStatus: conversation.ConversationStatusChanging},
				{snapshot: "Hello, world!", expectedStatus: conversation.ConversationStatusStable},
				{snapshot: "Hello, world!", expectedStatus: conversation.ConversationStatusStable},
				{snapshot: "What's up?", expectedStatus: conversation.ConversationStatusChanging},
			},
		})
	})
}
