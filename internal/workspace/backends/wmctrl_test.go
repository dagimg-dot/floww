package backends

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWmctrlBackend(t *testing.T) {
	b := NewWmctrlBackend()
	require.NotNil(t, b)
	assert.Equal(t, "wmctrl", b.wmctrlCmd)
}

func TestWmctrlSwitch_CorrectCommand(t *testing.T) {
	var recordedName string
	var recordedArgs []string

	savedRunCmd := runCmd
	runCmd = func(name string, args ...string) bool {
		recordedName = name
		recordedArgs = args
		return true
	}
	defer func() { runCmd = savedRunCmd }()

	b := NewWmctrlBackend()
	result := b.Switch(3)

	assert.True(t, result)
	assert.Equal(t, "wmctrl", recordedName)
	assert.Equal(t, []string{"-s", "3"}, recordedArgs)
}

func TestWmctrlSwitch_PropagatesFailure(t *testing.T) {
	savedRunCmd := runCmd
	runCmd = func(string, ...string) bool {
		return false
	}
	defer func() { runCmd = savedRunCmd }()

	b := NewWmctrlBackend()
	assert.False(t, b.Switch(0))
}

func TestWmctrlGetTotalWorkspaces_ParsesOutput(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("sh", "-c",
			`printf '0  * DG: 3840x1080  VP: 0,0  WA: 0,0 1920x1080  1\n1  - DG: 3840x1080  VP: N/A  WA: 0,0 0,0  N/A\n'`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewWmctrlBackend()
	assert.Equal(t, 2, b.GetTotalWorkspaces())
}

func TestWmctrlGetTotalWorkspaces_SingleWorkspace(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo", "0  * DG: 3840x1080  VP: 0,0  WA: 0,0 1920x1080  1")
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewWmctrlBackend()
	assert.Equal(t, 1, b.GetTotalWorkspaces())
}

func TestWmctrlGetTotalWorkspaces_CommandFails(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewWmctrlBackend()
	assert.Equal(t, 0, b.GetTotalWorkspaces())
}

func TestWmctrlGetAppendBaseOffset(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("sh", "-c",
			`printf '0\n1\n2\n3\n4\n5\n'`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewWmctrlBackend()
	assert.Equal(t, 5, b.GetAppendBaseOffset())
}
