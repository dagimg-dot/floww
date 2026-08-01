// Package diagnostic provides position-aware, compiler-style diagnostics
// for floww validation, modeled on `niri validate` (miette).
package diagnostic

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ANSI color codes for terminal output.
const (
	colorRed   = "\033[31m"
	colorBold  = "\033[1m"
	colorCyan  = "\033[36m"
	colorDim   = "\033[2m"
	colorReset = "\033[0m"
)

// Position is a location in a source file. Line and Column are 1-based;
// a zero value means the position is unknown.
type Position struct {
	Line   int
	Column int
	Length int // best-effort token length for the caret underline; 0 = unknown
}

type Diagnostic struct {
	Message  string
	Position Position
}

// Path builds a schema path like "workspaces[0].apps[1].exec" from
// alternating string keys and int indices. The position index builders
// (internal/config) and the workflow validator share this grammar so
// paths always match.
func Path(parts ...any) string {
	var sb strings.Builder
	for _, part := range parts {
		switch v := part.(type) {
		case string:
			if sb.Len() > 0 {
				sb.WriteByte('.')
			}
			sb.WriteString(v)
		case int:
			fmt.Fprintf(&sb, "[%d]", v)
		}
	}
	return sb.String()
}

// Render writes diagnostics to w. With useColor and a known position, a
// niri/miette-style block shows the source line with a caret underline;
// otherwise a compact "file:line:col: error: message" (or "error: message"
// when the position is unknown).
func Render(w io.Writer, file string, source []byte, diags []Diagnostic, useColor bool) {
	for i, d := range diags {
		if i > 0 {
			_, _ = fmt.Fprintln(w)
		}
		if d.Position.Line == 0 {
			printPlainError(w, d.Message, useColor)
			continue
		}
		if !useColor {
			loc := fmt.Sprintf("%s:%d", file, d.Position.Line)
			if d.Position.Column > 0 {
				loc += fmt.Sprintf(":%d", d.Position.Column)
			}
			_, _ = fmt.Fprintf(w, "%s: error: %s\n", loc, d.Message)
			continue
		}
		renderBlock(w, file, source, d)
	}
}

func printPlainError(w io.Writer, message string, useColor bool) {
	if useColor {
		_, _ = fmt.Fprintf(w, "%s%s error: %s%s\n", colorBold, colorRed, message, colorReset)
	} else {
		_, _ = fmt.Fprintf(w, "error: %s\n", message)
	}
}

func renderBlock(w io.Writer, file string, source []byte, d Diagnostic) {
	line := sourceLine(source, d.Position.Line)
	gutterW := len(strconv.Itoa(d.Position.Line)) + 1
	number := fmt.Sprintf("%*d", gutterW, d.Position.Line)
	blank := strings.Repeat(" ", gutterW)

	header := fmt.Sprintf("%s:%d", file, d.Position.Line)
	if d.Position.Column > 0 {
		header += fmt.Sprintf(":%d", d.Position.Column)
	}

	_, _ = fmt.Fprintf(w, "%s%s error: %s%s\n", colorBold, colorRed, d.Message, colorReset)
	_, _ = fmt.Fprintf(w, "%s-->%s %s\n", colorCyan, colorReset, header)
	_, _ = fmt.Fprintf(w, "%s%s |%s %s\n", colorDim, number, colorReset, line)
	_, _ = fmt.Fprintf(w, "%s%s |%s %s%s%s%s\n",
		colorDim, blank, colorReset,
		strings.Repeat(" ", caretOffset(d, line)),
		colorRed, strings.Repeat("^", caretLength(d, line)), colorReset)
}

func caretOffset(d Diagnostic, line string) int {
	if d.Position.Column <= 1 {
		return 0
	}
	return min(d.Position.Column-1, len([]rune(line)))
}

func caretLength(d Diagnostic, line string) int {
	if d.Position.Length <= 0 {
		return 1
	}
	offset := caretOffset(d, line)
	return max(1, min(d.Position.Length, len([]rune(line))-offset))
}

func sourceLine(source []byte, line int) string {
	if line <= 0 {
		return ""
	}
	lines := strings.Split(string(source), "\n")
	if line > len(lines) {
		return ""
	}
	return lines[line-1]
}
