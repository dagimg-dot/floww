package edit

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestEdit_NonexistentWorkflow(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	setupTest(t, dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"nonexistent"})

	err := Command.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEdit_ExistingWorkflow(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	wfDir := setupTest(t, dir)

	wfPath := filepath.Join(wfDir, "myflow.yaml")
	err := os.WriteFile(wfPath, []byte("description: test\nworkspaces: []\n"), 0600)
	require.NoError(t, err)

	t.Setenv("EDITOR", "true")

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"myflow"})

	err = Command.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Opening workflow")
}
