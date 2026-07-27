package init

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

func TestInit_CreatesDirectoriesAndConfig(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{})

	err := Command.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Initialized config at")

	flowwDir := filepath.Join(dir, "floww")
	workflowsDir := filepath.Join(flowwDir, "workflows")
	configFile := filepath.Join(flowwDir, "config.yaml")

	info, err := os.Stat(flowwDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	info, err = os.Stat(workflowsDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	info, err = os.Stat(configFile)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
}

func TestInit_WithExample(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"--example"})

	err := Command.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Initialized config at")

	examplePath := filepath.Join(dir, "floww", "workflows", "example.yaml")
	info, err := os.Stat(examplePath)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
}

func TestInit_WithExampleJSON(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	buf := new(bytes.Buffer)
	Command.SetOut(buf)
	Command.SetArgs([]string{"--example", "--type", "json"})

	err := Command.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Initialized config at")

	examplePath := filepath.Join(dir, "floww", "workflows", "example.json")
	info, err := os.Stat(examplePath)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
}

func TestInit_Idempotent(t *testing.T) {
	resetCmd()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	for i := 0; i < 2; i++ {
		buf := new(bytes.Buffer)
		Command.SetOut(buf)
		Command.SetArgs([]string{})
		err := Command.Execute()
		require.NoError(t, err)
	}
}
