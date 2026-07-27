package apply

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/dagimg-dot/floww/internal/config"
	"github.com/dagimg-dot/floww/internal/workflow"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWFManager satisfies the workflowApplier interface for testing.
type mockWFManager struct {
	applyResult    bool
	capturedData   *workflow.Workflow
	capturedAppend bool
}

func (m *mockWFManager) Apply(data *workflow.Workflow, append bool) bool {
	m.capturedData = data
	m.capturedAppend = append
	return m.applyResult
}

// resetCmd resets command state to prevent leakage between tests.
func resetCmd() {
	Command.SetArgs([]string{})
	Command.SetOut(nil)
	Command.SetErr(nil)
	Command.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
}

func setupTest(t *testing.T, dir string) string {
	t.Helper()
	flowwDir := filepath.Join(dir, "floww")
	workflowsDir := filepath.Join(flowwDir, "workflows")
	configFile := filepath.Join(flowwDir, "config.yaml")

	err := os.MkdirAll(workflowsDir, 0750)
	require.NoError(t, err)

	err = os.WriteFile(configFile, []byte("general: {}\ntiming: {}\n"), 0600)
	require.NoError(t, err)

	return workflowsDir
}

func createWorkflowFile(t *testing.T, wfDir, name string) {
	t.Helper()
	content := `description: "Example workflow"
workspaces:
  - target: 1
    apps:
      - name: Terminal
        exec: gnome-terminal
`
	err := os.WriteFile(filepath.Join(wfDir, name), []byte(content), 0600)
	require.NoError(t, err)
}

func setMockWFManager(t *testing.T, mock *mockWFManager) {
	t.Helper()
	old := wfManagerFactory
	wfManagerFactory = func(_ *config.ConfigManager, _ workflow.WorkspaceManager, _ workflow.AppLauncher) workflowApplier {
		return mock
	}
	t.Cleanup(func() { wfManagerFactory = old })
}

func TestApply_WorkflowByName(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)
	createWorkflowFile(t, wfDir, "example.yaml")

	mock := &mockWFManager{applyResult: true}
	setMockWFManager(t, mock)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"example"})

	err := Command.Execute()
	require.NoError(t, err)
	assert.False(t, mock.capturedAppend, "append should be false by default")
}

func TestApply_WorkflowFromFile(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupTest(t, dir)

	extFile := filepath.Join(dir, "external.yaml")
	content := `description: "External workflow"
workspaces:
  - target: 2
    apps:
      - name: Browser
        exec: firefox
`
	err := os.WriteFile(extFile, []byte(content), 0600)
	require.NoError(t, err)

	mock := &mockWFManager{applyResult: true}
	setMockWFManager(t, mock)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"--file", extFile})

	err = Command.Execute()
	require.NoError(t, err)
}

func TestApply_FromFileWithoutInit(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	extFile := filepath.Join(dir, "external.yaml")
	content := `description: "External workflow"
workspaces:
  - target: 2
    apps:
      - name: Browser
        exec: firefox
`
	err := os.WriteFile(extFile, []byte(content), 0600)
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "floww"))
	assert.True(t, os.IsNotExist(err), "floww config should NOT exist before init")

	mock := &mockWFManager{applyResult: true}
	setMockWFManager(t, mock)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"--file", extFile})

	err = Command.Execute()
	require.NoError(t, err, "--file should work without floww init")
}

func TestApply_NonexistentFile(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupTest(t, dir)

	nonexistentFile := filepath.Join(dir, "nonexistent.yaml")

	mock := &mockWFManager{applyResult: true}
	setMockWFManager(t, mock)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"--file", nonexistentFile})

	err := Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestApply_WorkflowFailure(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)
	createWorkflowFile(t, wfDir, "example.yaml")

	mock := &mockWFManager{applyResult: false}
	setMockWFManager(t, mock)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"example"})

	err := Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}

func TestApply_AppendFlag(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)
	createWorkflowFile(t, wfDir, "example.yaml")

	mock := &mockWFManager{applyResult: true}
	setMockWFManager(t, mock)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"example", "--append"})

	err := Command.Execute()
	require.NoError(t, err)
	assert.True(t, mock.capturedAppend, "append should be true")
}
