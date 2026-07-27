package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}

func TestValidateWorkflow_NilData(t *testing.T) {
	err := ValidateWorkflow("test", nil)
	assert.Error(t, err)
	assert.Equal(t,
		"Workflow file 'test' is empty or contains only null.",
		err.Error(),
	)
}

func TestValidateWorkflow_MissingWorkspaces(t *testing.T) {
	tests := []struct {
		name string
		wf   *Workflow
	}{
		{"nil workspaces", &Workflow{Workspaces: nil}},
		{"empty workspaces", &Workflow{Workspaces: []Workspace{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkflow("test", tt.wf)
			assert.Error(t, err)
			assert.Equal(t,
				"Workflow 'test' is missing the required 'workspaces' key.",
				err.Error(),
			)
		})
	}
}

func TestValidateWorkflow_InvalidFinalWorkspace(t *testing.T) {
	err := ValidateWorkflow("test", &Workflow{
		Workspaces:     []Workspace{{Target: 0, Apps: []App{{Name: "a", Exec: "a"}}}},
		FinalWorkspace: ptr(-1),
	})
	assert.Error(t, err)
	assert.Equal(t,
		"The 'final_workspace' key in workflow 'test' must be an integer greater than or equal to 0.",
		err.Error(),
	)
}

func TestValidateWorkflow_ValidFinalWorkspace(t *testing.T) {
	err := ValidateWorkflow("test", &Workflow{
		Workspaces:     []Workspace{{Target: 0, Apps: []App{{Name: "a", Exec: "a"}}}},
		FinalWorkspace: ptr(0),
	})
	assert.NoError(t, err)

	err = ValidateWorkflow("test", &Workflow{
		Workspaces:     []Workspace{{Target: 0, Apps: []App{{Name: "a", Exec: "a"}}}},
		FinalWorkspace: ptr(5),
	})
	assert.NoError(t, err)
}

func TestValidateWorkflow_NegativeTarget(t *testing.T) {
	err := ValidateWorkflow("test", &Workflow{
		Workspaces: []Workspace{{Target: -1, Apps: []App{{Name: "a", Exec: "a"}}}},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"The 'target' key for workspace target '-1' (index 0) must be an integer greater than or equal to 0.",
		err.Error(),
	)
}

func TestValidateWorkflow_NilApps(t *testing.T) {
	err := ValidateWorkflow("test", &Workflow{
		Workspaces: []Workspace{{Target: 0, Apps: nil}},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"Workspace definition for workspace target '0' (index 0) is missing the required 'apps' key.",
		err.Error(),
	)
}

func TestValidateWorkflow_EmptyAppName(t *testing.T) {
	err := ValidateWorkflow("test", &Workflow{
		Workspaces: []Workspace{{Target: 0, Apps: []App{{Name: "", Exec: "xterm"}}}},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"App definition for app index 0 in workspace target '0' (index 0) is missing the required 'name' key.",
		err.Error(),
	)
}

func TestValidateWorkflow_EmptyAppExec(t *testing.T) {
	err := ValidateWorkflow("test", &Workflow{
		Workspaces: []Workspace{{Target: 0, Apps: []App{{Name: "term", Exec: ""}}}},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"App definition for app 'term' (app index 0 in workspace target '0' (index 0)) is missing the required 'exec' key.",
		err.Error(),
	)
}

func TestValidateWorkflow_InvalidAppType(t *testing.T) {
	err := ValidateWorkflow("test", &Workflow{
		Workspaces: []Workspace{{Target: 0, Apps: []App{{Name: "term", Exec: "xterm", Type: "invalid"}}}},
	})
	assert.Error(t, err)
	assert.Equal(t,
		"The 'type' key for app 'term' (app index 0 in workspace target '0' (index 0)) must be one of 'binary', 'flatpak', 'snap', but got 'invalid'.",
		err.Error(),
	)
}

func TestValidateWorkflow_DefaultTypeIsBinary(t *testing.T) {
	wf := &Workflow{
		Workspaces: []Workspace{{
			Target: 0,
			Apps:   []App{{Name: "term", Exec: "xterm"}},
		}},
	}
	err := ValidateWorkflow("test", wf)
	assert.NoError(t, err)
	assert.Equal(t, "binary", wf.Workspaces[0].Apps[0].Type)
}

func TestValidateWorkflow_ValidTypePreserved(t *testing.T) {
	for _, typ := range []string{"binary", "flatpak", "snap"} {
		t.Run(typ, func(t *testing.T) {
			wf := &Workflow{
				Workspaces: []Workspace{{
					Target: 0,
					Apps:   []App{{Name: "term", Exec: "xterm", Type: typ}},
				}},
			}
			err := ValidateWorkflow("test", wf)
			assert.NoError(t, err)
			assert.Equal(t, typ, wf.Workspaces[0].Apps[0].Type)
		})
	}
}

func TestValidateWorkflow_WithDescription(t *testing.T) {
	wf := &Workflow{
		Description: "My workflow",
		Workspaces:  []Workspace{{Target: 0, Apps: []App{{Name: "term", Exec: "xterm"}}}},
	}
	err := ValidateWorkflow("test", wf)
	assert.NoError(t, err)
}

func TestValidateWorkflow_WithWaitAndArgs(t *testing.T) {
	wf := &Workflow{
		Workspaces: []Workspace{{
			Target: 0,
			Apps:   []App{{Name: "term", Exec: "xterm", Args: []string{"--flag", "val"}, Wait: ptr(1.5)}},
		}},
	}
	err := ValidateWorkflow("test", wf)
	assert.NoError(t, err)
}

func TestValidateWorkflow_MultipleWorkspacesAndApps(t *testing.T) {
	wf := &Workflow{
		Description: "multi workspace",
		Workspaces: []Workspace{
			{Target: 0, Apps: []App{{Name: "term", Exec: "xterm"}, {Name: "editor", Exec: "code", Args: []string{"."}}}},
			{Target: 2, Apps: []App{{Name: "browser", Exec: "firefox", Type: "flatpak"}}},
		},
		FinalWorkspace: ptr(0),
	}
	err := ValidateWorkflow("test", wf)
	assert.NoError(t, err)
}

func TestValidateWorkflow_ErrorOnFirstFailure(t *testing.T) {
	wf := &Workflow{
		Workspaces: []Workspace{
			{Target: 0, Apps: []App{{Name: "", Exec: ""}}},
			{Target: 1, Apps: []App{{Name: "good", Exec: "good"}}},
		},
	}
	err := ValidateWorkflow("test", wf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestValidateWorkflow_WorkflowNotFoundError(t *testing.T) {
	err := &WorkflowNotFoundError{Message: "Workflow 'foo' not found"}
	assert.Error(t, err)
	assert.Equal(t, "Workflow 'foo' not found", err.Error())
}

func TestValidateWorkflow_WorkflowSchemaError(t *testing.T) {
	err := &WorkflowSchemaError{Message: "schema error"}
	assert.Error(t, err)
	assert.Equal(t, "schema error", err.Error())
}
