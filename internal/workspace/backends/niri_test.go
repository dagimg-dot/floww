package backends

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNiriBackend(t *testing.T) {
	b := NewNiriBackend()
	require.NotNil(t, b)
	assert.Equal(t, "niri", b.niriCmd)
}

func TestWorkflowTargetToIdx_OneBased(t *testing.T) {
	assert.Equal(t, 1, workflowTargetToIdx(1))
	assert.Equal(t, 2, workflowTargetToIdx(2))
	assert.Equal(t, 5, workflowTargetToIdx(5))
}

func TestWorkflowTargetToIdx_ClampsBelowOne(t *testing.T) {
	assert.Equal(t, 1, workflowTargetToIdx(0))
	assert.Equal(t, 1, workflowTargetToIdx(-1))
	assert.Equal(t, 1, workflowTargetToIdx(-5))
}

func TestFocusedWorkspaceIdx_FoundByIsFocused(t *testing.T) {
	ws := []niriWorkspace{
		{Idx: 1, IsFocused: false, IsActive: false},
		{Idx: 2, IsFocused: true, IsActive: false},
		{Idx: 3, IsFocused: false, IsActive: false},
	}
	idx, ok := focusedWorkspaceIdx(ws)
	assert.True(t, ok)
	assert.Equal(t, 2, idx)
}

func TestFocusedWorkspaceIdx_FoundByIsActive(t *testing.T) {
	ws := []niriWorkspace{
		{Idx: 1, IsFocused: false, IsActive: false},
		{Idx: 3, IsFocused: false, IsActive: true},
	}
	idx, ok := focusedWorkspaceIdx(ws)
	assert.True(t, ok)
	assert.Equal(t, 3, idx)
}

func TestFocusedWorkspaceIdx_ReturnsFirstMatch(t *testing.T) {
	// focusedWorkspaceIdx returns the first workspace where
	// IsFocused or IsActive is true — no preference between the two.
	ws := []niriWorkspace{
		{Idx: 5, IsFocused: false, IsActive: true},
		{Idx: 9, IsFocused: true, IsActive: false},
	}
	idx, ok := focusedWorkspaceIdx(ws)
	assert.True(t, ok)
	assert.Equal(t, 5, idx, "first matching workspace wins")
}

func TestFocusedWorkspaceIdx_NoneFocused(t *testing.T) {
	ws := []niriWorkspace{
		{Idx: 1, IsFocused: false, IsActive: false},
		{Idx: 2, IsFocused: false, IsActive: false},
	}
	_, ok := focusedWorkspaceIdx(ws)
	assert.False(t, ok)
}

func TestFocusedWorkspaceIdx_EmptyList(t *testing.T) {
	_, ok := focusedWorkspaceIdx([]niriWorkspace{})
	assert.False(t, ok)
}

func TestNiriIsAlreadyFocused_TrueWhenOnTarget(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo",
			`[{"id":0,"idx":1,"is_focused":true,"is_active":true}]`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewNiriBackend()
	assert.True(t, b.isAlreadyFocusedOnTarget(1))
}

func TestNiriIsAlreadyFocused_FalseWhenOnDifferentWorkspace(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo",
			`[{"id":0,"idx":1,"is_focused":true,"is_active":true},{"id":1,"idx":3,"is_focused":false,"is_active":false}]`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewNiriBackend()
	assert.False(t, b.isAlreadyFocusedOnTarget(3))
}

func TestNiriIsAlreadyFocused_FalseWhenJSONFails(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewNiriBackend()
	assert.False(t, b.isAlreadyFocusedOnTarget(1))
}

func TestNiriSwitch_SkipsWhenAlreadyOnTarget(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo",
			`[{"id":0,"idx":1,"is_focused":true,"is_active":true}]`)
	}
	defer func() { execCmd = savedExecCmd }()

	var runCmdCalled bool
	savedRunCmd := runCmd
	runCmd = func(string, ...string) bool {
		runCmdCalled = true
		return true
	}
	defer func() { runCmd = savedRunCmd }()

	b := NewNiriBackend()
	result := b.Switch(1)

	assert.True(t, result)
	assert.False(t, runCmdCalled)
}

func TestNiriSwitch_FallsThroughWhenNotOnTarget(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo",
			`[{"id":0,"idx":1,"is_focused":true,"is_active":true}]`)
	}
	defer func() { execCmd = savedExecCmd }()

	var recordedArgs []string
	savedRunCmd := runCmd
	runCmd = func(name string, args ...string) bool {
		recordedArgs = append([]string{name}, args...)
		return true
	}
	defer func() { runCmd = savedRunCmd }()

	b := NewNiriBackend()
	result := b.Switch(3)

	assert.True(t, result)
	assert.Equal(t, []string{"niri", "msg", "action", "focus-workspace", "3"}, recordedArgs)
}

func TestNiriSwitch_FallsThroughWhenCheckFails(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCmd = savedExecCmd }()

	var recordedArgs []string
	savedRunCmd := runCmd
	runCmd = func(name string, args ...string) bool {
		recordedArgs = append([]string{name}, args...)
		return true
	}
	defer func() { runCmd = savedRunCmd }()

	b := NewNiriBackend()
	result := b.Switch(2)

	assert.True(t, result)
	assert.Equal(t, []string{"niri", "msg", "action", "focus-workspace", "2"}, recordedArgs)
}

func TestNiriSwitch_ClampsTarget(t *testing.T) {
	var recordedArgs []string
	savedRunCmd := runCmd
	runCmd = func(name string, args ...string) bool {
		recordedArgs = append([]string{name}, args...)
		return true
	}
	defer func() { runCmd = savedRunCmd }()

	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewNiriBackend()
	b.Switch(0)

	assert.Equal(t, []string{"niri", "msg", "action", "focus-workspace", "1"}, recordedArgs)
}

func TestNiriSwitch_PropagatesFailure(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCmd = savedExecCmd }()

	savedRunCmd := runCmd
	runCmd = func(string, ...string) bool {
		return false
	}
	defer func() { runCmd = savedRunCmd }()

	b := NewNiriBackend()
	assert.False(t, b.Switch(5))
}

func TestNiriGetTotalWorkspaces_ReturnsCount(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo",
			`[{"id":0,"idx":1},{"id":1,"idx":2},{"id":2,"idx":3}]`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewNiriBackend()
	assert.Equal(t, 3, b.GetTotalWorkspaces())
}

func TestNiriGetTotalWorkspaces_EmptyList(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo", `[]`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewNiriBackend()
	assert.Equal(t, 1, b.GetTotalWorkspaces())
}

func TestNiriGetTotalWorkspaces_CommandFails(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewNiriBackend()
	assert.Equal(t, 0, b.GetTotalWorkspaces())
}

func TestNiriGetAppendBaseOffset(t *testing.T) {
	savedExecCmd := execCmd
	execCmd = func(string, ...string) *exec.Cmd {
		return exec.Command("echo",
			`[{"id":0,"idx":1},{"id":1,"idx":2},{"id":2,"idx":3},{"id":3,"idx":4},{"id":4,"idx":5}]`)
	}
	defer func() { execCmd = savedExecCmd }()

	b := NewNiriBackend()
	assert.Equal(t, 4, b.GetAppendBaseOffset())
}
