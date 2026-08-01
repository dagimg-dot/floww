package config

import (
	"testing"
	"time"

	"github.com/dagimg-dot/floww/internal/diagnostic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func parseYAMLNode(t *testing.T, content string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(content), &doc))
	return &doc
}

func TestPositions_YAMLBasic(t *testing.T) {
	// ------------
	content := `description: "web dev"
workspaces:
  - target: 1
    apps:
      - name: term
        exec: xterm
        type: invalid
      - name: editor
        exec: code
  - target: 3
    apps:
      - name: browser
        exec: firefox
final_workspace: 0
`
	p := buildPositionsFromYAML(parseYAMLNode(t, content))

	pos, ok := p.Position("")
	require.True(t, ok)
	assert.Equal(t, diagnostic.Position{Line: 1, Column: 1}, pos)

	pos, _ = p.Position("description")
	assert.Equal(t, diagnostic.Position{Line: 1, Column: 14, Length: 7}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "target"))
	assert.Equal(t, diagnostic.Position{Line: 3, Column: 13, Length: 1}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "apps", 0, "type"))
	assert.Equal(t, diagnostic.Position{Line: 7, Column: 15, Length: 7}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 1, "apps", 0, "exec"))
	assert.Equal(t, diagnostic.Position{Line: 13, Column: 15, Length: 7}, pos)

	pos, _ = p.Position("final_workspace")
	assert.Equal(t, diagnostic.Position{Line: 14, Column: 18, Length: 1}, pos)

	// item positions (used for missing-key errors)
	pos, _ = p.Position(diagnostic.Path("workspaces", 0))
	assert.Equal(t, diagnostic.Position{Line: 3, Column: 5}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "apps", 1))
	assert.Equal(t, diagnostic.Position{Line: 8, Column: 9}, pos)
}

func TestPositions_YAMLDuplicateKeysLastWins(t *testing.T) {
	// ------------
	content := "workspaces:\n  - target: 1\n    target: 2\n    apps: []\n"
	p := buildPositionsFromYAML(parseYAMLNode(t, content))

	pos, ok := p.Position(diagnostic.Path("workspaces", 0, "target"))
	require.True(t, ok)
	assert.Equal(t, 3, pos.Line)
}

func TestPositions_YAMLAliasAndMerge(t *testing.T) {
	// ------------
	content := `cmd: &c xterm
base: &b
  apps:
    - name: n
      exec: *c
ws:
  target: 0
  <<: *b
`
	p := buildPositionsFromYAML(parseYAMLNode(t, content))

	// alias value points at the anchor definition
	pos, ok := p.Position(diagnostic.Path("ws", "apps", 0, "exec"))
	require.True(t, ok)
	assert.Equal(t, 1, pos.Line)

	// merge key content recorded at the merging path with anchor positions
	pos, ok = p.Position(diagnostic.Path("ws", "apps", 0, "name"))
	require.True(t, ok)
	assert.Equal(t, 4, pos.Line)
}

func TestPositions_YAMLAliasedSubtreeIsIndexed(t *testing.T) {
	// ------------
	content := `apps: &apps
  - name: n
    exec: e
ws:
  target: 0
  apps: *apps
`
	p := buildPositionsFromYAML(parseYAMLNode(t, content))

	// paths inside the aliased subtree resolve at the anchor definition
	pos, ok := p.Position(diagnostic.Path("ws", "apps", 0, "name"))
	require.True(t, ok)
	assert.Equal(t, diagnostic.Position{Line: 2, Column: 11, Length: 1}, pos)

	pos, ok = p.Position(diagnostic.Path("ws", "apps", 0, "exec"))
	require.True(t, ok)
	assert.Equal(t, diagnostic.Position{Line: 3, Column: 11, Length: 1}, pos)
}

func TestPositions_YAMLCyclesTerminate(t *testing.T) {
	// ------------
	// self-referential anchors must not hang the walker. Inputs yaml.v3
	// rejects at parse time (forward references) are skipped — there is
	// no node tree to walk.
	contents := []string{
		"ws: &x\n  - *x\n",
		"x: &x\n  <<: *x\n",
		"x: &x\n  b: *x\n",
		"x: &x\n  b:\n    - *x\n",
	}
	for _, content := range contents {
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
			continue
		}
		done := make(chan bool, 1)
		go func() {
			buildPositionsFromYAML(&doc)
			done <- true
		}()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("walker did not terminate for %q", content)
		}
	}
}

func TestPositions_JSONViaYAML(t *testing.T) {
	// ------------
	content := `{"workspaces": [{"target": 1, "apps": [{"name": "term", "exec": "x", "type": "bad"}]}]}`
	p := buildPositionsFromYAML(parseYAMLNode(t, content))

	pos, ok := p.Position(diagnostic.Path("workspaces", 0, "target"))
	require.True(t, ok)
	assert.Equal(t, diagnostic.Position{Line: 1, Column: 28, Length: 1}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "apps", 0, "type"))
	assert.Equal(t, diagnostic.Position{Line: 1, Column: 78, Length: 3}, pos)
}

func TestPositions_EmptyAndNull(t *testing.T) {
	// ------------
	for _, content := range []string{"", "null\n"} {
		p := buildPositionsFromYAML(parseYAMLNode(t, content))
		pos, ok := p.Position("")
		require.True(t, ok)
		assert.Equal(t, diagnostic.Position{Line: 1, Column: 1}, pos)
		_, ok = p.Position("workspaces")
		assert.False(t, ok)
	}
}

func TestPositions_TOMLArrayTables(t *testing.T) {
	// ------------
	content := `# comment
[[workspaces]]
target = 1
[[workspaces.apps]]
name = "term"
exec = "xterm"
type = "bad"
[[workspaces]]
target = 2
`
	p := buildPositionsFromTOML([]byte(content))

	pos, ok := p.Position("")
	require.True(t, ok)
	assert.Equal(t, diagnostic.Position{Line: 1, Column: 1}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0))
	assert.Equal(t, diagnostic.Position{Line: 2, Column: 3, Length: 10}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "target"))
	assert.Equal(t, diagnostic.Position{Line: 3, Column: 10, Length: 1}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "apps", 0))
	assert.Equal(t, diagnostic.Position{Line: 4, Column: 3, Length: 10}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "apps", 0, "name"))
	assert.Equal(t, diagnostic.Position{Line: 5, Column: 8, Length: 6}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "apps", 0, "type"))
	assert.Equal(t, diagnostic.Position{Line: 7, Column: 8, Length: 5}, pos)

	// second [[workspaces]] element resolves independently
	pos, _ = p.Position(diagnostic.Path("workspaces", 1))
	assert.Equal(t, diagnostic.Position{Line: 8, Column: 3, Length: 10}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 1, "target"))
	assert.Equal(t, diagnostic.Position{Line: 9, Column: 10, Length: 1}, pos)

	_, ok = p.Position(diagnostic.Path("workspaces", 0, "apps", 1))
	assert.False(t, ok)
}

func TestPositions_TOMLInlineArray(t *testing.T) {
	// ------------
	content := `[[workspaces]]
target = 1
apps = [{ name = "a", exec = "x" }, { name = "b", exec = "y" }]
`
	p := buildPositionsFromTOML([]byte(content))

	pos, _ := p.Position(diagnostic.Path("workspaces", 0, "apps"))
	assert.Equal(t, diagnostic.Position{Line: 3, Column: 9, Length: 1}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "apps", 0, "name"))
	assert.Equal(t, diagnostic.Position{Line: 3, Column: 18, Length: 3}, pos)

	pos, _ = p.Position(diagnostic.Path("workspaces", 0, "apps", 1, "exec"))
	assert.Equal(t, diagnostic.Position{Line: 3, Column: 58, Length: 3}, pos)
}

func TestPositions_TOMLSimpleKeys(t *testing.T) {
	// ------------
	content := "final_workspace = 0\ndescription = \"d\"\n"
	p := buildPositionsFromTOML([]byte(content))

	pos, ok := p.Position("final_workspace")
	require.True(t, ok)
	assert.Equal(t, diagnostic.Position{Line: 1, Column: 19, Length: 1}, pos)

	pos, _ = p.Position("description")
	assert.Equal(t, diagnostic.Position{Line: 2, Column: 15, Length: 3}, pos)
}
