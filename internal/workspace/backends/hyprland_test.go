package backends

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHyprlandBackend(t *testing.T) {
	b := NewHyprlandBackend()
	require.NotNil(t, b)
	assert.Equal(t, "hyprctl", b.hyprctlCmd)
}

func TestHyprlandSwitch_CorrectCommand(t *testing.T) {
	var recordedName string
	var recordedArgs []string

	savedRunCmd := runCmd
	runCmd = func(name string, args ...string) bool {
		recordedName = name
		recordedArgs = args
		return true
	}
	defer func() { runCmd = savedRunCmd }()

	b := NewHyprlandBackend()
	result := b.Switch(3)

	assert.True(t, result)
	assert.Equal(t, "hyprctl", recordedName)
	assert.Equal(t, []string{"dispatch", "workspace", "3"}, recordedArgs)
}

func TestHyprlandSwitch_ZeroTarget(t *testing.T) {
	var recordedArgs []string

	savedRunCmd := runCmd
	runCmd = func(name string, args ...string) bool {
		recordedArgs = args
		return true
	}
	defer func() { runCmd = savedRunCmd }()

	b := NewHyprlandBackend()
	result := b.Switch(0)

	assert.True(t, result)
	assert.Equal(t, []string{"dispatch", "workspace", "0"}, recordedArgs)
}

func TestHyprlandSwitch_PropagatesFailure(t *testing.T) {
	savedRunCmd := runCmd
	runCmd = func(name string, args ...string) bool {
		return false
	}
	defer func() { runCmd = savedRunCmd }()

	b := NewHyprlandBackend()
	assert.False(t, b.Switch(99))
}

func TestHyprlandGetTotalWorkspaces_ParsesMaxID(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo", `[{"id":0},{"id":1},{"id":5}]`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewHyprlandBackend()
	result := b.GetTotalWorkspaces()
	assert.Equal(t, 5, result)
}

func TestHyprlandGetTotalWorkspaces_EmptyList(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo", `[]`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewHyprlandBackend()
	result := b.GetTotalWorkspaces()
	assert.Equal(t, 1, result)
}

func TestHyprlandGetTotalWorkspaces_InvalidJSON(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo", "not-json")
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewHyprlandBackend()
	result := b.GetTotalWorkspaces()
	assert.Equal(t, 0, result)
}

func TestHyprlandGetTotalWorkspaces_CommandError(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewHyprlandBackend()
	result := b.GetTotalWorkspaces()
	assert.Equal(t, 0, result)
}

func TestHyprlandGetAppendBaseOffset(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo", `[{"id":0},{"id":1},{"id":2},{"id":3},{"id":4}]`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewHyprlandBackend()
	assert.Equal(t, 3, b.GetAppendBaseOffset())
}

func TestAppendBaseOffset(t *testing.T) {
	tests := []struct {
		total    int
		expected int
	}{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 2},
		{5, 4},
		{10, 9},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, AppendBaseOffset(tt.total),
			"AppendBaseOffset(%d)", tt.total)
	}
}
