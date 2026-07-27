package workflow

import "fmt"

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

// ValidateWorkflow validates a workflow against the expected schema.
// It returns nil if the workflow is valid, or a *WorkflowSchemaError describing
// the first validation failure encountered.
func ValidateWorkflow(name string, data *Workflow) error {
	if data == nil {
		return &WorkflowSchemaError{
			Message: fmt.Sprintf("Workflow file '%s' is empty or contains only null.", name),
		}
	}

	if len(data.Workspaces) == 0 {
		return &WorkflowSchemaError{
			Message: fmt.Sprintf("Workflow '%s' is missing the required 'workspaces' key.", name),
		}
	}

	if data.FinalWorkspace != nil {
		if *data.FinalWorkspace < 0 {
			return &WorkflowSchemaError{
				Message: fmt.Sprintf("The 'final_workspace' key in workflow '%s' must be an integer greater than or equal to 0.", name),
			}
		}
	}

	for i := range data.Workspaces {
		ws := &data.Workspaces[i]

		// ws_id: always use target format since Target is always present as int
		wsID := fmt.Sprintf("workspace target '%d' (index %d)", ws.Target, i)

		if ws.Target < 0 {
			return &WorkflowSchemaError{
				Message: fmt.Sprintf("The 'target' key for %s must be an integer greater than or equal to 0.", wsID),
			}
		}

		if ws.Apps == nil {
			return &WorkflowSchemaError{
				Message: fmt.Sprintf("Workspace definition for %s is missing the required 'apps' key.", wsID),
			}
		}

		for j := range ws.Apps {
			app := &ws.Apps[j]

			appID := fmt.Sprintf("app index %d in %s", j, wsID)
			if app.Name != "" {
				appID = fmt.Sprintf("app '%s' (%s)", app.Name, appID)
			}

			if app.Name == "" {
				return &WorkflowSchemaError{
					Message: fmt.Sprintf("App definition for %s is missing the required 'name' key.", appID),
				}
			}

			if app.Exec == "" {
				return &WorkflowSchemaError{
					Message: fmt.Sprintf("App definition for %s is missing the required 'exec' key.", appID),
				}
			}

			appType := app.Type
			if appType == "" {
				appType = "binary"
			}
			if appType != "binary" && appType != "flatpak" && appType != "snap" {
				return &WorkflowSchemaError{
					Message: fmt.Sprintf("The 'type' key for %s must be one of 'binary', 'flatpak', 'snap', but got '%s'.", appID, appType),
				}
			}
			app.Type = appType
		}
	}

	return nil
}
