package httpapi

import (
	"time"

	st "github.com/coder/openagent/lib/screentracker"
)

func FormatClaudeCodeMessage(message string) []st.MessagePart {
	return []st.MessagePart{
		// janky hack: send a random character and then a backspace because otherwise
		// Claude Code echoes the startSeq back to the terminal.
		// This basically simulates a user typing and then removing the character.
		st.MessagePartText{Content: "x\b", Hidden: true},
		// Bracketed paste mode start sequence
		st.MessagePartText{Content: "\x1b[200~", Hidden: true},
		st.MessagePartText{Content: message},
		// Bracketed paste mode end sequence
		st.MessagePartText{Content: "\x1b[201~", Hidden: true},
		// wait because Claude Code doesn't recognize "\r" as a command
		// to process the input if it's sent right away
		st.MessagePartWait{Duration: 50 * time.Millisecond},
		st.MessagePartText{Content: "\r", Hidden: true},
	}
}
