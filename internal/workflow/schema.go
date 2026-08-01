package workflow

import (
	"fmt"

	"github.com/dagimg-dot/floww/internal/diagnostic"
)

// App represents an application to launch on a workspace.
type App struct {
	Name string   `yaml:"name" json:"name" toml:"name"`
	Exec string   `yaml:"exec" json:"exec" toml:"exec"`
	Type string   `yaml:"type" json:"type" toml:"type"`
	Args []string `yaml:"args" json:"args" toml:"args"`
	Wait *float64 `yaml:"wait" json:"wait" toml:"wait"`
}

// Workspace represents a workspace with apps to launch.
type Workspace struct {
	Target int   `yaml:"target" json:"target" toml:"target"`
	Apps   []App `yaml:"apps" json:"apps" toml:"apps"`
}

// Workflow represents a complete workflow configuration.
type Workflow struct {
	Description    string      `yaml:"description" json:"description" toml:"description"`
	Workspaces     []Workspace `yaml:"workspaces" json:"workspaces" toml:"workspaces"`
	FinalWorkspace *int        `yaml:"final_workspace" json:"final_workspace" toml:"final_workspace"`
}

// WorkflowSchemaError represents a workflow schema validation error.
type WorkflowSchemaError struct {
	Message string
}

func (e *WorkflowSchemaError) Error() string {
	return e.Message
}

// WorkflowNotFoundError represents a workflow not found error.
type WorkflowNotFoundError struct {
	Message string
}

func (e *WorkflowNotFoundError) Error() string {
	return e.Message
}

// Locator resolves schema paths (e.g. "workspaces[0].apps[1].exec", built
// with diagnostic.Path) to source positions. The path "" is the document
// start. Implementations may return false for unknown paths.
type Locator interface {
	Position(path string) (diagnostic.Position, bool)
}

// ValidateWorkflow validates a workflow against the expected schema.
// It returns nil if the workflow is valid, or a *WorkflowSchemaError
// describing the first validation failure encountered.
func ValidateWorkflow(name string, data *Workflow) error {
	diags := ValidateWorkflowDetailed(name, data, nil)
	if len(diags) == 0 {
		return nil
	}
	return &WorkflowSchemaError{Message: diags[0].Message}
}

// ValidateWorkflowDetailed validates a workflow, accumulating every schema
// violation as a Diagnostic (compiler-style, not first-failure-only).
// Positions come from loc when available; the zero Position means the
// location is unknown and the diagnostic renders message-only.
func ValidateWorkflowDetailed(name string, data *Workflow, loc Locator) []diagnostic.Diagnostic {
	var diags []diagnostic.Diagnostic

	pos := func(path string) diagnostic.Position {
		if loc == nil {
			return diagnostic.Position{}
		}
		p, _ := loc.Position(path)
		return p
	}
	docPos := pos("")
	if docPos.Line == 0 {
		docPos = diagnostic.Position{Line: 1, Column: 1}
	}
	posOr := func(path, fallback string) diagnostic.Position {
		p := pos(path)
		if p.Line == 0 {
			p = pos(fallback)
		}
		return p
	}
	add := func(path, message string) {
		diags = append(diags, diagnostic.Diagnostic{Message: message, Position: pos(path)})
	}

	if data == nil {
		return []diagnostic.Diagnostic{{
			Message:  fmt.Sprintf("Workflow file '%s' is empty or contains only null.", name),
			Position: docPos,
		}}
	}

	if len(data.Workspaces) == 0 {
		diags = append(diags, diagnostic.Diagnostic{
			Message:  fmt.Sprintf("Workflow '%s' is missing the required 'workspaces' key.", name),
			Position: docPos,
		})
	}

	if data.FinalWorkspace != nil && *data.FinalWorkspace < 0 {
		add("final_workspace",
			fmt.Sprintf("The 'final_workspace' key in workflow '%s' must be an integer greater than or equal to 0.", name))
	}

	for i := range data.Workspaces {
		ws := &data.Workspaces[i]

		// ws_id: always use target format since Target is always present as int
		wsID := fmt.Sprintf("workspace target '%d' (index %d)", ws.Target, i)

		if ws.Target < 0 {
			add(diagnostic.Path("workspaces", i, "target"),
				fmt.Sprintf("The 'target' key for %s must be an integer greater than or equal to 0.", wsID))
		}

		if ws.Apps == nil {
			add(diagnostic.Path("workspaces", i),
				fmt.Sprintf("Workspace definition for %s is missing the required 'apps' key.", wsID))
			continue
		}

		for j := range ws.Apps {
			app := &ws.Apps[j]

			appID := fmt.Sprintf("app index %d in %s", j, wsID)
			if app.Name != "" {
				appID = fmt.Sprintf("app '%s' (%s)", app.Name, appID)
			}

			namePath := diagnostic.Path("workspaces", i, "apps", j, "name")
			if app.Name == "" {
				diags = append(diags, diagnostic.Diagnostic{
					Message:  fmt.Sprintf("App definition for %s is missing the required 'name' key.", appID),
					Position: posOr(namePath, diagnostic.Path("workspaces", i, "apps", j)),
				})
			}

			execPath := diagnostic.Path("workspaces", i, "apps", j, "exec")
			if app.Exec == "" {
				diags = append(diags, diagnostic.Diagnostic{
					Message:  fmt.Sprintf("App definition for %s is missing the required 'exec' key.", appID),
					Position: posOr(execPath, diagnostic.Path("workspaces", i, "apps", j)),
				})
			}

			appType := app.Type
			if appType == "" {
				appType = "binary"
			}
			if appType != "binary" && appType != "flatpak" && appType != "snap" {
				diags = append(diags, diagnostic.Diagnostic{
					Message:  fmt.Sprintf("The 'type' key for %s must be one of 'binary', 'flatpak', 'snap', but got '%s'.", appID, appType),
					Position: pos(diagnostic.Path("workspaces", i, "apps", j, "type")),
				})
			}
			app.Type = appType
		}
	}

	return diags
}
