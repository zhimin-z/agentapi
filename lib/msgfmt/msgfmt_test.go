package msgfmt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeAndGetRuneLineMapping(t *testing.T) {
	msg := "Hello, World!\n \nTest.\n"
	normalizedMsg, lines, runeLineLocations := normalizeAndGetRuneLineMapping(msg)
	assert.Equal(t, normalizedMsg, "Hello,World!Test.")
	assert.Equal(t, []string{"Hello, World!", " ", "Test.", ""}, lines)
	assert.Equal(t, []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2}, runeLineLocations)
}

func TestFindUserInputStartIdx(t *testing.T) {
	t.Run("single-line-msg", func(t *testing.T) {
		prefix := "Hello, World!"
		userInput := "How are you doing?"
		suffix := "Good Good"
		msg := prefix + userInput + suffix
		userInputStartIdx := findUserInputStartIdx(msg, make([]int, len(msg)), userInput, make([]int, len(userInput)))
		assert.Equal(t, len(prefix), userInputStartIdx)
	})
	t.Run("truncated-user-input", func(t *testing.T) {
		prefix := "Hello, World!"
		userInput := "How are you doing?"
		suffix := "Good Good"
		// Only the first 6 runes of the user input are considered
		msg := prefix + "How ar" + suffix
		userInputStartIdx := findUserInputStartIdx(msg, make([]int, len(msg)), userInput, make([]int, len(userInput)))
		assert.Equal(t, len(prefix), userInputStartIdx)
	})
	t.Run("multi-line-msg", func(t *testing.T) {
		t.Run("long-first-line", func(t *testing.T) {
			firstLineBuilder := strings.Builder{}
			for range 100 {
				firstLineBuilder.WriteRune('-')
			}
			firstLine := firstLineBuilder.String()
			prefix := firstLine
			userInput := "How are you doing?"
			suffix := "Good Good"
			msg := prefix + userInput + suffix
			msgRuneLineMapping := make([]int, len(msg))
			for i := 100; i < len(msg); i++ {
				msgRuneLineMapping[i] = 1
			}
			userInputStartIdx := findUserInputStartIdx(msg, msgRuneLineMapping, userInput, make([]int, len(userInput)))
			assert.Equal(t, len(prefix), userInputStartIdx)
		})
		t.Run("short-leading-lines", func(t *testing.T) {
			prefixLines := []string{"a", "b", "c", "d", "e", "f"}
			prefix := strings.Join(prefixLines, "")
			userInput := "How are you doing?"
			suffix := "Good Good"
			msg := prefix + userInput + suffix
			msgRuneLineMapping := make([]int, len(msg))
			for i := range len(prefix) {
				msgRuneLineMapping[i] = i
			}
			for i := len(prefix); i < len(prefix)+len(userInput); i++ {
				msgRuneLineMapping[i] = len(prefixLines)
			}
			for i := len(prefix) + len(userInput); i < len(msg); i++ {
				msgRuneLineMapping[i] = len(prefixLines) + 1
			}
			userInputStartIdx := findUserInputStartIdx(msg, msgRuneLineMapping, userInput, make([]int, len(userInput)))
			assert.Equal(t, len(prefix), userInputStartIdx)
		})
	})
	t.Run("multi-line-user-input", func(t *testing.T) {
		t.Run("short-first-line", func(t *testing.T) {
			userInputLines := []string{"abc", "def"}
			userInput := strings.Join(userInputLines, "")
			prefix := "Hello, World!"
			suffix := "Good Good"
			// only the first line of input is considered
			msg := prefix + "abcxxx" + suffix
			userInputRuneLineMapping := []int{0, 0, 0, 1, 1, 1}
			userInputStartIdx := findUserInputStartIdx(msg, make([]int, len(msg)), userInput, userInputRuneLineMapping)
			assert.Equal(t, len(prefix), userInputStartIdx)
		})
	})
}

func TestFindUserInputEndIdx(t *testing.T) {
	t.Run("accurate-msg", func(t *testing.T) {
		prefix := "Hello, World!"
		userInput := "How are you doing?"
		suffix := "Good Good"
		msg := prefix + userInput + suffix
		userInputEndIdx := findUserInputEndIdx(len(prefix), msg, userInput)
		assert.Equal(t, suffix, msg[userInputEndIdx+1:])
	})
	t.Run("truncated-input", func(t *testing.T) {
		prefix := "Hello, World!"
		userInput := "How are you doing?"
		suffix := "...------"
		truncatedUserInput := userInput[:7]
		msg := prefix + truncatedUserInput + suffix
		userInputEndIdx := findUserInputEndIdx(len(prefix), msg, userInput)
		assert.Equal(t, len(prefix)+len(truncatedUserInput)-1, userInputEndIdx)
		assert.Equal(t, suffix, msg[userInputEndIdx+1:])
	})
	t.Run("broken-up-input", func(t *testing.T) {
		// This test simulates UI elements breaking up the user input.
		prefix := "Hello, World!"
		userInputParts := []string{"How ", "are ", "you ", "doing?"}
		userInput := strings.Join(userInputParts, "*|*")
		suffix := "...------"
		msg := prefix + userInput + suffix
		userInputEndIdx := findUserInputEndIdx(len(prefix), msg, userInput)
		assert.Equal(t, suffix, msg[userInputEndIdx+1:])
	})
}
