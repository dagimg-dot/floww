package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dagimg-dot/floww/internal/diagnostic"
	"github.com/dagimg-dot/floww/internal/workflow"
	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// ValidationResult is the outcome of validating a workflow file: the raw
// source (for excerpt rendering) and any diagnostics found.
type ValidationResult struct {
	Source      []byte
	Diagnostics []diagnostic.Diagnostic
}

// ValidateWorkflowFile fully validates a workflow file at path, decoding it
// directly per format (no JSON round-trip) so parse, type, and schema errors
// all carry real source positions. Errors are accumulated, not fail-fast.
//
// An error is returned only for infrastructure failures (missing file,
// unsupported extension); validation findings are returned as diagnostics.
func (cm *ConfigManager) ValidateWorkflowFile(path string) (*ValidationResult, error) {
	if !cm.loader.IsSupportedFormat(path) {
		return nil, fmt.Errorf("unsupported configuration format: %s", filepath.Ext(path))
	}

	data, err := os.ReadFile(path) //nolint:gosec // Intentional file open
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &WorkflowNotFoundError{
				ConfigError: ConfigError{
					FlowwError: FlowwError{
						Msg: fmt.Sprintf("Workflow file '%s' not found", path),
					},
				},
			}
		}
		return nil, err
	}

	var wf *workflow.Workflow
	var positions *Positions
	var diags []diagnostic.Diagnostic

	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		wf, positions, diags = parseWorkflowYAML(data)
	case ".json":
		wf, positions, diags = parseWorkflowJSON(data)
	case ".toml":
		wf, positions, diags = parseWorkflowTOML(data)
	}

	if len(diags) > 0 {
		return &ValidationResult{Source: data, Diagnostics: diags}, nil
	}

	schemaDiags := workflow.ValidateWorkflowDetailed(path, wf, positions)
	return &ValidationResult{Source: data, Diagnostics: schemaDiags}, nil
}

// parseWorkflowYAML decodes a YAML workflow directly into the typed struct
// and builds the position index. Parse failures yield one diagnostic; type
// mismatches yield one diagnostic per offending value (yaml.TypeError
// accumulates), each with the error line.
func parseWorkflowYAML(data []byte) (*workflow.Workflow, *Positions, []diagnostic.Diagnostic) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, nil, []diagnostic.Diagnostic{yamlSyntaxDiagnostic(err)}
	}

	var wf workflow.Workflow
	if err := doc.Decode(&wf); err != nil {
		var te *yaml.TypeError
		if errors.As(err, &te) {
			diags := make([]diagnostic.Diagnostic, 0, len(te.Errors))
			for _, e := range te.Errors {
				line := yamlLineNumber(e)
				diags = append(diags, diagnostic.Diagnostic{
					Message:  e,
					Position: diagnostic.Position{Line: line},
				})
			}
			return nil, nil, diags
		}
		return nil, nil, []diagnostic.Diagnostic{{Message: err.Error()}}
	}

	return &wf, buildPositionsFromYAML(&doc), nil
}

var yamlLineRe = regexp.MustCompile(`line (\d+)`)

func yamlSyntaxDiagnostic(err error) diagnostic.Diagnostic {
	return diagnostic.Diagnostic{Message: err.Error(), Position: yamlPosition(err.Error())}
}

func yamlLineNumber(msg string) int {
	return yamlPosition(msg).Line
}

func yamlPosition(msg string) diagnostic.Position {
	pos := diagnostic.Position{}
	if m := yamlLineRe.FindStringSubmatch(msg); m != nil {
		if line, err := strconv.Atoi(m[1]); err == nil {
			pos.Line = line
		}
	}
	return pos
}

// parseWorkflowJSON decodes a JSON workflow directly into the typed struct.
// Positions for parse and type errors come from the byte offset in the
// original file; the position index for schema checks is built by parsing
// the JSON as YAML (YAML 1.2 is a superset of JSON).
func parseWorkflowJSON(data []byte) (*workflow.Workflow, *Positions, []diagnostic.Diagnostic) {
	var wf workflow.Workflow
	if err := json.Unmarshal(data, &wf); err != nil {
		return nil, nil, []diagnostic.Diagnostic{jsonDiagnostic(data, err)}
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return &wf, nil, nil
	}
	return &wf, buildPositionsFromYAML(&doc), nil
}

func jsonDiagnostic(data []byte, err error) diagnostic.Diagnostic {
	var se *json.SyntaxError
	var te *json.UnmarshalTypeError
	switch {
	case errors.As(err, &se):
		line, col := offsetToLineCol(data, se.Offset)
		return diagnostic.Diagnostic{
			Message:  strings.TrimPrefix(err.Error(), "json: "),
			Position: diagnostic.Position{Line: line, Column: col},
		}
	case errors.As(err, &te):
		start := startOfJSONScalar(data, te.Offset)
		line, col := offsetToLineCol(data, int64(start))
		return diagnostic.Diagnostic{
			Message:  fmt.Sprintf("cannot unmarshal %s into %s (field %s)", te.Value, te.Type, te.Field),
			Position: diagnostic.Position{Line: line, Column: col},
		}
	default:
		return diagnostic.Diagnostic{Message: err.Error()}
	}
}

// offsetToLineCol converts a byte offset into 1-based line/column, counting
// runes so columns match other parsers.
func offsetToLineCol(data []byte, offset int64) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > int64(len(data)) {
		offset = int64(len(data))
	}
	line, col := 1, 1
	for _, r := range string(data[:offset]) {
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// startOfJSONScalar scans backward from the end of a JSON scalar token to its
// start, so the caret points at the value rather than its trailing quote.
func startOfJSONScalar(data []byte, end int64) int {
	i := int(end)
	for i > 0 && isJSONSpace(data[i-1]) {
		i--
	}
	if i == 0 {
		return i
	}
	switch data[i-1] {
	case '"':
		j := i - 2
		for j >= 0 {
			if data[j] == '"' {
				slashes := 0
				for k := j; k > 0 && data[k-1] == '\\'; k-- {
					slashes++
				}
				if slashes%2 == 0 {
					return j
				}
			}
			j--
		}
		return i
	case 'e', 'E', '+', '-', '.', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		j := i - 1
		for j > 0 && isJSONNumberChar(data[j-1]) {
			j--
		}
		return j
	default:
		j := i - 1
		for j > 0 && data[j-1] >= 'a' && data[j-1] <= 'z' {
			j--
		}
		return j
	}
}

func isJSONSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isJSONNumberChar(b byte) bool {
	return b == 'e' || b == 'E' || b == '+' || b == '-' || b == '.' ||
		(b >= '0' && b <= '9')
}

// parseWorkflowTOML decodes a TOML workflow directly into the typed struct.
// go-toml's DecodeError carries the exact position of parse and type errors.
func parseWorkflowTOML(data []byte) (*workflow.Workflow, *Positions, []diagnostic.Diagnostic) {
	var wf workflow.Workflow
	if err := toml.Unmarshal(data, &wf); err != nil {
		var de *toml.DecodeError
		if errors.As(err, &de) {
			row, col := de.Position()
			return nil, nil, []diagnostic.Diagnostic{{
				Message:  de.Error(),
				Position: diagnostic.Position{Line: row, Column: col},
			}}
		}
		return nil, nil, []diagnostic.Diagnostic{{Message: err.Error()}}
	}
	return &wf, buildPositionsFromTOML(data), nil
}
