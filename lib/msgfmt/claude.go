package msgfmt

import (
	"strings"
)

func removeClaudeMessageBox(msg string) string {
	lines := strings.Split(msg, "\n")
	lastLine := func() string {
		if len(lines) > 0 {
			return lines[len(lines)-1]
		}
		return ""
	}
	trimmedLastLine := func() string {
		return strings.TrimSpace(lastLine())
	}
	popLine := func() {
		if len(lines) > 0 {
			lines = lines[:len(lines)-1]
		}
	}

	if strings.HasPrefix(trimmedLastLine(), "? for shortcuts") {
		popLine()
	}
	msgBoxEdge := "───────────────"
	if strings.Contains(trimmedLastLine(), msgBoxEdge) {
		popLine()
	}
	if strings.Contains(trimmedLastLine(), ">") {
		popLine()
	}
	if strings.Contains(trimmedLastLine(), msgBoxEdge) {
		popLine()
	}

	return strings.Join(lines, "\n")
}

func formatClaudeMessage(message string, userInput string) string {
	message = RemoveUserInput(message, userInput)
	message = removeClaudeMessageBox(message)
	message = trimEmptyLines(message)
	return message
}

func formatGooseMessage(message string, userInput string) string {
	// The current formatClaudeMessage implementation is so generic
	// that it works with both Goose and Aider too.
	return formatClaudeMessage(message, userInput)
}

func formatAiderMessage(message string, userInput string) string {
	return formatClaudeMessage(message, userInput)
}

func formatCustomMessage(message string, userInput string) string {
	return formatClaudeMessage(message, userInput)
}
