package diagnostic

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPath(t *testing.T) {
	// ------------
	assert.Equal(t, "workspaces", Path("workspaces"))
	assert.Equal(t, "workspaces[0]", Path("workspaces", 0))
	assert.Equal(t, "workspaces[0].apps[1].exec", Path("workspaces", 0, "apps", 1, "exec"))
	assert.Equal(t, "final_workspace", Path("final_workspace"))
	assert.Equal(t, "", Path())
}

func TestRender_NoPosition(t *testing.T) {
	// ------------
	buf := new(bytes.Buffer)
	Render(buf, "f.yaml", []byte(""), []Diagnostic{{Message: "boom"}}, false)
	assert.Equal(t, "error: boom\n", buf.String())
}

func TestRender_CompactWithPosition(t *testing.T) {
	// ------------
	buf := new(bytes.Buffer)
	Render(buf, "/cfg/f.yaml", []byte(""), []Diagnostic{
		{Message: "bad target", Position: Position{Line: 3, Column: 5}},
	}, false)
	assert.Equal(t, "/cfg/f.yaml:3:5: error: bad target\n", buf.String())
}

func TestRender_CompactLineWithoutColumn(t *testing.T) {
	// ------------
	buf := new(bytes.Buffer)
	Render(buf, "f.yaml", nil, []Diagnostic{
		{Message: "no col", Position: Position{Line: 2}},
	}, false)
	assert.Equal(t, "f.yaml:2: error: no col\n", buf.String())
}

func TestRender_CompactMultiple(t *testing.T) {
	// ------------
	buf := new(bytes.Buffer)
	Render(buf, "f.yaml", nil, []Diagnostic{
		{Message: "one", Position: Position{Line: 2, Column: 1}},
		{Message: "two", Position: Position{Line: 9, Column: 4}},
	}, false)
	assert.Equal(t, "f.yaml:2:1: error: one\n\nf.yaml:9:4: error: two\n", buf.String())
}

func TestRender_BlockWithColor(t *testing.T) {
	// ------------
	source := []byte("description: x\nworkspaces:\n  - target: 1\n    apps:\n      - name: term\n        exec: xterm\n        type: invalid\n")
	buf := new(bytes.Buffer)
	Render(buf, "f.yaml", source, []Diagnostic{
		{Message: "bad type", Position: Position{Line: 7, Column: 9, Length: 7}},
	}, true)

	out := buf.String()
	assert.Contains(t, out, "error: bad type")
	assert.Contains(t, out, "f.yaml:7:9")
	assert.Contains(t, out, " 7 |")
	assert.Contains(t, out, "        type: invalid")
	assert.Contains(t, out, "^^^^^^^")
	assert.NotContains(t, out, " 8 |")
}

func TestRender_BlockCaretClamped(t *testing.T) {
	// ------------
	// Length beyond the line end must be clamped; missing column starts at 1.
	source := []byte("a: 1\n")
	buf := new(bytes.Buffer)
	Render(buf, "f.yaml", source, []Diagnostic{
		{Message: "long", Position: Position{Line: 1, Column: 1, Length: 99}},
	}, true)
	assert.Contains(t, buf.String(), " 1 |")
	assert.Contains(t, buf.String(), "a: 1")
	assert.Contains(t, buf.String(), "^^^^")

	buf2 := new(bytes.Buffer)
	Render(buf2, "f.yaml", source, []Diagnostic{
		{Message: "no col", Position: Position{Line: 1}},
	}, true)
	assert.Contains(t, buf2.String(), " 1 |")
	assert.Contains(t, buf2.String(), "f.yaml:1")
	assert.NotContains(t, buf2.String(), "f.yaml:1:")
}

func TestRender_BlockLineOutOfRange(t *testing.T) {
	// ------------
	buf := new(bytes.Buffer)
	Render(buf, "f.yaml", []byte("a: 1\n"), []Diagnostic{
		{Message: "ghost", Position: Position{Line: 99, Column: 1}},
	}, true)
	assert.Contains(t, buf.String(), "error: ghost")
	assert.Contains(t, buf.String(), "f.yaml:99:1")
}

func TestRender_LineAndColumnOneBased(t *testing.T) {
	// ------------
	source := []byte("one\ntwo\nthree\n")
	buf := new(bytes.Buffer)
	Render(buf, "f", source, []Diagnostic{
		{Message: "m", Position: Position{Line: 2, Column: 3, Length: 1}},
	}, true)
	lines := strings.Split(stripANSI(buf.String()), "\n")
	require.Equal(t, " 2 | two", lines[2])
	assert.Equal(t, "   |   ^", lines[3])
}

func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}
