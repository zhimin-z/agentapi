package httpapi

import (
	mf "github.com/coder/agentapi/lib/msgfmt"
	st "github.com/coder/agentapi/lib/screentracker"
)

func formatPaste(message string) []st.MessagePart {
	return []st.MessagePart{
		// Bracketed paste mode start sequence
		st.MessagePartText{Content: "\x1b[200~", Hidden: true},
		st.MessagePartText{Content: message},
		// Bracketed paste mode end sequence
		st.MessagePartText{Content: "\x1b[201~", Hidden: true},
	}
}

func formatClaudeCodeMessage(message string) []st.MessagePart {
	parts := make([]st.MessagePart, 0)
	parts = append(parts, formatPaste(message)...)

	return parts
}

func FormatMessage(agentType mf.AgentType, message string) []st.MessagePart {
	message = mf.TrimWhitespace(message)
	// for now Claude Code formatting seems to also work for Goose and Aider
	// so we can use the same function for all three
	return formatClaudeCodeMessage(message)
}
