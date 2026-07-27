package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCommand_ExistingCommand(t *testing.T) {
	ok := RunCommand("echo", "hello")
	assert.True(t, ok)
}

func TestRunCommand_NonexistentCommand(t *testing.T) {
	ok := RunCommand("this-command-does-not-exist-99999")
	assert.False(t, ok)
}

func TestRunCommand_WithArgs(t *testing.T) {
	ok := RunCommand("echo", "hello", "world")
	assert.True(t, ok)
}

func TestRunCommand_FailingCommand(t *testing.T) {
	// `false` exits with non-zero status
	ok := RunCommand("false")
	assert.False(t, ok)
}

func TestNotify_DoesNotPanic(t *testing.T) {
	// Notify should never panic regardless of whether notify-send is available
	Notify("test-message")
}

func TestNotify_EmptyMessage(t *testing.T) {
	Notify("")
}
