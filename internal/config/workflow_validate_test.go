package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/dagimg-dot/floww/internal/diagnostic"
	"github.com/dagimg-dot/floww/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager(t *testing.T, files map[string]string) *ConfigManager {
	t.Helper()
	dir := t.TempDir()
	flowwDir := filepath.Join(dir, "floww")
	require.NoError(t, os.MkdirAll(filepath.Join(flowwDir, "workflows"), 0750))
	require.NoError(t, os.WriteFile(filepath.Join(flowwDir, "config.yaml"), []byte("general: {}\ntiming: {}\n"), 0600))
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(flowwDir, "workflows", name), []byte(content), 0600))
	}
	return NewConfigManager(filepath.Join(dir, "floww"))
}

func validateYAML(t *testing.T, cm *ConfigManager, name string) *ValidationResult {
	t.Helper()
	path, err := cm.ResolveWorkflowPath(name)
	require.NoError(t, err)
	res, err := cm.ValidateWorkflowFile(path)
	require.NoError(t, err)
	return res
}

func TestValidateWorkflowFile_Valid(t *testing.T) {
	// ------------
	cm := newTestManager(t, map[string]string{"good.yaml": `description: "x"
workspaces:
  - target: 1
    apps:
      - name: term
        exec: xterm
        type: binary
`})
	res := validateYAML(t, cm, "good")
	assert.Empty(t, res.Diagnostics)
}

func TestValidateWorkflowFile_Formats(t *testing.T) {
	// ------------
	jsonWF := `{"description": "x", "workspaces": [{"target": 1, "apps": [{"name": "t", "exec": "e"}]}]}`
	tomlWF := "[[workspaces]]\ntarget = 1\n[[workspaces.apps]]\nname = \"t\"\nexec = \"e\"\n"
	cm := newTestManager(t, map[string]string{
		"a.json": jsonWF,
		"b.toml": tomlWF,
	})
	for _, name := range []string{"a", "b"} {
		res := validateYAML(t, cm, name)
		assert.Empty(t, res.Diagnostics, name)
	}
}

func TestValidateWorkflowFile_YAMLSyntaxError(t *testing.T) {
	// ------------
	cm := newTestManager(t, map[string]string{"bad.yaml": "workspaces:\n  - target: [1, 2\n"})
	res := validateYAML(t, cm, "bad")
	require.Len(t, res.Diagnostics, 1)
	assert.Equal(t, 1, res.Diagnostics[0].Position.Line)
	assert.Contains(t, res.Diagnostics[0].Message, "yaml: line 1")
}

func TestValidateWorkflowFile_YAMLTypeErrorsAccumulate(t *testing.T) {
	// ------------
	cm := newTestManager(t, map[string]string{"bad.yaml": "workspaces:\n  - target: abc\n  - target: def\n"})
	res := validateYAML(t, cm, "bad")
	require.Len(t, res.Diagnostics, 2)
	assert.Equal(t, 2, res.Diagnostics[0].Position.Line)
	assert.Equal(t, 3, res.Diagnostics[1].Position.Line)
	assert.Contains(t, res.Diagnostics[0].Message, "cannot unmarshal")
}

func TestValidateWorkflowFile_JSONSyntaxError(t *testing.T) {
	// ------------
	cm := newTestManager(t, map[string]string{"bad.json": "{\"workspaces\": [}"})
	res := validateYAML(t, cm, "bad")
	require.Len(t, res.Diagnostics, 1)
	assert.Equal(t, 1, res.Diagnostics[0].Position.Line)
	assert.Contains(t, res.Diagnostics[0].Message, "invalid character")
}

func TestValidateWorkflowFile_JSONTypeError(t *testing.T) {
	// ------------
	cm := newTestManager(t, map[string]string{"bad.json": "{\"workspaces\": [{\"target\": \"abc\"}]}"})
	res := validateYAML(t, cm, "bad")
	require.Len(t, res.Diagnostics, 1)
	d := res.Diagnostics[0]
	assert.Equal(t, 1, d.Position.Line)
	assert.Equal(t, 28, d.Position.Column)
	assert.Contains(t, d.Message, "cannot unmarshal string into int")
	assert.Contains(t, d.Message, "workspaces.target")
}

func TestValidateWorkflowFile_TOMLSyntaxError(t *testing.T) {
	// ------------
	cm := newTestManager(t, map[string]string{"bad.toml": "[[workspaces]]\ntarget = [1,\n"})
	res := validateYAML(t, cm, "bad")
	require.Len(t, res.Diagnostics, 1)
	d := res.Diagnostics[0]
	assert.Equal(t, diagnostic.Position{Line: 2, Column: 13}, d.Position)
}

func TestValidateWorkflowFile_TOMLTypeError(t *testing.T) {
	// ------------
	cm := newTestManager(t, map[string]string{"bad.toml": "[[workspaces]]\ntarget = \"abc\"\n"})
	res := validateYAML(t, cm, "bad")
	require.Len(t, res.Diagnostics, 1)
	d := res.Diagnostics[0]
	assert.Equal(t, diagnostic.Position{Line: 2, Column: 10}, d.Position)
	assert.Contains(t, d.Message, "cannot decode TOML string")
}

func TestValidateWorkflowFile_SchemaErrorsAccumulate(t *testing.T) {
	// ------------
	content := `workspaces:
  - target: 1
    apps:
      - name: ""
        exec: xterm
        type: invalid
  - target: -2
    apps:
      - name: browser
        exec: firefox
`
	cm := newTestManager(t, map[string]string{"bad.yaml": content})
	res := validateYAML(t, cm, "bad")

	require.Len(t, res.Diagnostics, 3)
	// order follows the validation pass: name, type, then next workspace's target
	assert.Contains(t, res.Diagnostics[0].Message, "missing the required 'name' key")
	assert.Contains(t, res.Diagnostics[1].Message, "must be one of 'binary', 'flatpak', 'snap'")
	assert.Contains(t, res.Diagnostics[2].Message, "must be an integer greater than or equal to 0")

	// positions: name empty value, type value, target value
	assert.Equal(t, 4, res.Diagnostics[0].Position.Line)
	assert.Equal(t, 6, res.Diagnostics[1].Position.Line)
	assert.Equal(t, 7, res.Diagnostics[2].Position.Line)
}

func TestValidateWorkflowFile_MissingAppsPointsAtWorkspace(t *testing.T) {
	// ------------
	content := "description: x\nworkspaces:\n  - target: 1\n"
	cm := newTestManager(t, map[string]string{"bad.yaml": content})
	res := validateYAML(t, cm, "bad")

	require.Len(t, res.Diagnostics, 1)
	assert.Contains(t, res.Diagnostics[0].Message, "missing the required 'apps' key")
	assert.Equal(t, diagnostic.Position{Line: 3, Column: 5}, res.Diagnostics[0].Position)
}

func TestValidateWorkflowFile_EmptyAndNull(t *testing.T) {
	// ------------
	for _, content := range []string{"", "null\n"} {
		cm := newTestManager(t, map[string]string{"empty.yaml": content})
		res := validateYAML(t, cm, "empty")
		require.Len(t, res.Diagnostics, 1)
		assert.Contains(t, res.Diagnostics[0].Message, "missing the required 'workspaces' key")
		assert.Equal(t, diagnostic.Position{Line: 1, Column: 1}, res.Diagnostics[0].Position)
	}
}

func TestValidateWorkflowFile_TOMLSchemaPositions(t *testing.T) {
	// ------------
	content := `[[workspaces]]
target = 1
[[workspaces.apps]]
name = "term"
exec = "xterm"
type = "invalid"
`
	cm := newTestManager(t, map[string]string{"bad.toml": content})
	res := validateYAML(t, cm, "bad")

	require.Len(t, res.Diagnostics, 1)
	assert.Contains(t, res.Diagnostics[0].Message, "must be one of 'binary'")
	assert.Equal(t, diagnostic.Position{Line: 6, Column: 8, Length: 9}, res.Diagnostics[0].Position)
}

func TestValidateWorkflowFile_JSONSchemaPositions(t *testing.T) {
	// ------------
	content := `{"workspaces": [{"target": -1, "apps": [{"name": "t", "exec": "e"}]}]}`
	cm := newTestManager(t, map[string]string{"bad.json": content})
	res := validateYAML(t, cm, "bad")

	require.Len(t, res.Diagnostics, 1)
	assert.Contains(t, res.Diagnostics[0].Message, "must be an integer greater than or equal to 0")
	assert.Equal(t, 1, res.Diagnostics[0].Position.Line)
	assert.Equal(t, 28, res.Diagnostics[0].Position.Column)
}

func TestJSONDiagnostic_MalformedInputsDoNotPanic(t *testing.T) {
	// ------------
	// heavily malformed JSON must never panic the offset scanner
	contents := []string{
		`{"workspaces": [{"target": }]}`,
		`{"workspaces": [{"target": "a"`,
		`{"workspaces": [{"target": 1`,
		`{"workspaces": "str", "x": [{"a": [1,2]}]}`,
		`{"target": "s"}`,
		``,
		`{]`,
		`"just a string"`,
		`[1, 2, 3]`,
	}
	for _, content := range contents {
		var wf workflow.Workflow
		if err := json.Unmarshal([]byte(content), &wf); err != nil {
			_ = jsonDiagnostic([]byte(content), err)
		}
	}
}

func TestValidateWorkflowFile_NotFound(t *testing.T) {
	// ------------
	cm := newTestManager(t, nil)
	_, err := cm.ValidateWorkflowFile(filepath.Join(cm.WorkflowsDir(), "ghost.yaml"))
	var wnfe *WorkflowNotFoundError
	assert.True(t, errors.As(err, &wnfe))
}

func TestValidateWorkflowFile_UnsupportedFormat(t *testing.T) {
	// ------------
	cm := newTestManager(t, nil)
	path := filepath.Join(cm.WorkflowsDir(), "bad.ini")
	require.NoError(t, os.WriteFile(path, []byte("x=1\n"), 0600))
	_, err := cm.ValidateWorkflowFile(path)
	assert.ErrorContains(t, err, "unsupported configuration format")
}

func TestValidateWorkflowFile_AppliesBinaryDefault(t *testing.T) {
	// ------------
	cm := newTestManager(t, map[string]string{"d.yaml": `workspaces:
  - target: 1
    apps:
      - name: term
        exec: xterm
`})
	res := validateYAML(t, cm, "d")
	assert.Empty(t, res.Diagnostics)
}

func TestValidateWorkflowFile_ExplicitValidType(t *testing.T) {
	// ------------
	cm := newTestManager(t, map[string]string{"d.yaml": `workspaces:
  - target: 1
    apps:
      - name: term
        exec: xterm
        type: flatpak
`})
	res := validateYAML(t, cm, "d")
	assert.Empty(t, res.Diagnostics)
}
