package msgfmt

import "strings"

const whiteSpaceChars = " \t\n\r\f\v"

// Normalize the string to remove any whitespace.
// Remember in which line each rune is located.
// Return the normalized string, the lines, and the rune to line location mapping.
func normalizeAndGetRuneLineMapping(msgRaw string) (string, []string, []int) {
	msgBuilder := strings.Builder{}
	msgLines := strings.Split(msgRaw, "\n")
	msgRuneLineLocations := []int{}
	for lineIdx, line := range msgLines {
		for _, r := range line {
			if !strings.ContainsRune(whiteSpaceChars, r) {
				msgBuilder.WriteRune(r)
				msgRuneLineLocations = append(msgRuneLineLocations, lineIdx)
			}
		}
	}
	return msgBuilder.String(), msgLines, msgRuneLineLocations
}

// Find where the user input starts in the message
func findUserInputStartIdx(msg string, msgRuneLineLocations []int, userInput string, userInputLineLocations []int) int {
	// We take up to 6 runes from the first line of the user input
	// and search for it in the message. 6 is arbitrary.
	// We only look at the first line to avoid running into user input
	// being broken up by UI elements.
	maxUserInputPrefixLen := 6
	userInputPrefixLen := -1
	for i, lineIdx := range userInputLineLocations {
		if lineIdx > 0 {
			break
		}
		if i >= maxUserInputPrefixLen {
			break
		}
		userInputPrefixLen = i + 1
	}
	if userInputPrefixLen == -1 {
		return -1
	}
	userInputPrefix := userInput[:userInputPrefixLen]

	// We'll only search the first 5 lines or 25 runes of the message,
	// whichever has more runes. This number is arbitrary. The intuition
	// is that user input is echoed back at the start of the message. The first
	// line or two may contain some UI elements.
	msgPrefixLen := 0
	msgPrefix := ""
	for i, lineIdx := range msgRuneLineLocations {
		if lineIdx > 5 {
			break
		}
		msgPrefixLen = i + 1
	}
	defaultRunesFromMsg := 25
	if msgPrefixLen < defaultRunesFromMsg {
		msgPrefixLen = defaultRunesFromMsg
	}
	msgPrefix = msg[:msgPrefixLen]

	// Search for the user input prefix in the message prefix
	return strings.Index(msgPrefix, userInputPrefix)
}

// Find where the user input ends in the message. Returns the index of the last rune
// of the user input in the message.
func findUserInputEndIdx(userInputStartIdx int, msg string, userInput string) int {
	userInputIdx := 0
	msgIdx := userInputStartIdx
OuterLoop:
	for {
		if userInputIdx >= len(userInput) {
			break
		}
		if msgIdx >= len(msg) {
			break
		}
		if userInput[userInputIdx] == msg[msgIdx] {
			userInputIdx++
			msgIdx++
			continue
		}
		// If we haven't found a match, we'll search the next 5 runes of the message.
		// If we can't find a match, we'll assume the echoed user input was truncated.
		// 5 is arbitrary.
		for i := 1; i <= 5; i++ {
			if msgIdx+i >= len(msg) {
				break
			}
			if userInput[userInputIdx] == msg[msgIdx+i] {
				userInputIdx++
				msgIdx = msgIdx + i
				continue OuterLoop
			}
		}
		break
	}
	return msgIdx - 1
}

// RemoveUserInput removes the user input from the message.
// Goose, Aider, and Claude Code echo back the user's input to
// make it visible in the terminal. This function makes a best effort
// attempt to remove it.
func RemoveUserInput(msgRaw string, userInputRaw string) string {
	msg, msgLines, msgRuneLineLocations := normalizeAndGetRuneLineMapping(msgRaw)
	userInput, _, userInputLineLocations := normalizeAndGetRuneLineMapping(userInputRaw)
	userInputStartIdx := findUserInputStartIdx(msg, msgRuneLineLocations, userInput, userInputLineLocations)

	if userInputStartIdx == -1 {
		// The user input prefix was not found in the message prefix
		// Return the original message
		return msgRaw
	}

	userInputEndIdx := findUserInputEndIdx(userInputStartIdx, msg, userInput)

	// Return the original message starting with the first line
	// that doesn't contain the echoed user input.
	lastUserInputLineIdx := msgRuneLineLocations[userInputEndIdx]
	return strings.Join(msgLines[lastUserInputLineIdx+1:], "\n")
}

func FormatClaudeMessage(message string) string {
	return strings.TrimSpace(message)
}
