package workflow

import (
	"testing"

	"github.com/dagimg-dot/floww/internal/diagnostic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestValidateWorkflowDetailed_AccumulatesAllErrors(t *testing.T) {
	wf := &Workflow{
		Workspaces: []Workspace{
			{Target: -1, Apps: []App{{Name: "", Exec: "", Type: "invalid"}}},
			{Target: 1, Apps: []App{{Name: "good", Exec: "good"}}},
		},
		FinalWorkspace: ptr(-2),
	}
	diags := ValidateWorkflowDetailed("test", wf, nil)
	require.Len(t, diags, 5)
	assert.Contains(t, diags[0].Message, "final_workspace")
	assert.Contains(t, diags[1].Message, "'target' key")
	assert.Contains(t, diags[2].Message, "'name' key")
	assert.Contains(t, diags[3].Message, "'exec' key")
	assert.Contains(t, diags[4].Message, "'type' key")
}

func TestValidateWorkflowDetailed_NoLocatorPositionsAreZero(t *testing.T) {
	wf := &Workflow{
		Workspaces: []Workspace{{Target: -1, Apps: []App{{Name: "a", Exec: "a"}}}},
	}
	diags := ValidateWorkflowDetailed("test", wf, nil)
	require.Len(t, diags, 1)
	assert.Equal(t, diagnostic.Position{}, diags[0].Position)
}

type fakeLocator map[string]diagnostic.Position

func (f fakeLocator) Position(path string) (diagnostic.Position, bool) {
	pos, ok := f[path]
	return pos, ok
}

func TestValidateWorkflowDetailed_UsesLocatorPositions(t *testing.T) {
	loc := fakeLocator{
		"":                           {Line: 1, Column: 1},
		"final_workspace":            {Line: 2, Column: 10},
		"workspaces[0].target":       {Line: 3, Column: 13},
		"workspaces[0].apps[0].name": {Line: 5, Column: 15},
		"workspaces[0].apps[0].type": {Line: 7, Column: 15},
		"workspaces[0].apps[0]":      {Line: 5, Column: 7},
	}
	wf := &Workflow{
		Workspaces:     []Workspace{{Target: -1, Apps: []App{{Name: "", Exec: "", Type: "invalid"}}}},
		FinalWorkspace: ptr(-1),
	}
	diags := ValidateWorkflowDetailed("test", wf, loc)
	require.Len(t, diags, 5)
	assert.Equal(t, diagnostic.Position{Line: 2, Column: 10}, diags[0].Position)
	assert.Equal(t, diagnostic.Position{Line: 3, Column: 13}, diags[1].Position)
	assert.Equal(t, diagnostic.Position{Line: 5, Column: 15}, diags[2].Position)
	assert.Equal(t, diagnostic.Position{Line: 5, Column: 7}, diags[3].Position)
	assert.Equal(t, diagnostic.Position{Line: 7, Column: 15}, diags[4].Position)
}

func TestValidateWorkflowDetailed_MissingKeyFallsBackToItemPosition(t *testing.T) {
	loc := fakeLocator{
		"workspaces[0].apps[0]": {Line: 5, Column: 7},
	}
	wf := &Workflow{
		Workspaces: []Workspace{{Target: 0, Apps: []App{{Name: "", Exec: "xterm"}}}},
	}
	diags := ValidateWorkflowDetailed("test", wf, loc)
	require.Len(t, diags, 1)
	assert.Equal(t, diagnostic.Position{Line: 5, Column: 7}, diags[0].Position)
}

func TestValidateWorkflowDetailed_MissingWorkspacesUsesDocPosition(t *testing.T) {
	loc := fakeLocator{"": {Line: 4, Column: 1}}
	diags := ValidateWorkflowDetailed("test", &Workflow{Workspaces: nil}, loc)
	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Message, "missing the required 'workspaces' key")
	assert.Equal(t, diagnostic.Position{Line: 4, Column: 1}, diags[0].Position)
}

func TestValidateWorkflowDetailed_DefaultDocPosition(t *testing.T) {
	diags := ValidateWorkflowDetailed("test", &Workflow{Workspaces: nil}, nil)
	require.Len(t, diags, 1)
	assert.Equal(t, diagnostic.Position{Line: 1, Column: 1}, diags[0].Position)
}

func TestValidateWorkflowDetailed_EmptyData(t *testing.T) {
	diags := ValidateWorkflowDetailed("test", nil, nil)
	require.Len(t, diags, 1)
	assert.Contains(t, diags[0].Message, "empty or contains only null")
	assert.Equal(t, diagnostic.Position{Line: 1, Column: 1}, diags[0].Position)
}

func TestValidateWorkflowDetailed_AccumulatesTypeMutations(t *testing.T) {
	wf := &Workflow{
		Workspaces: []Workspace{{Target: 0, Apps: []App{{Name: "a", Exec: "a", Type: "flatpak"}}}},
	}
	diags := ValidateWorkflowDetailed("test", wf, nil)
	assert.Empty(t, diags)
	assert.Equal(t, "flatpak", wf.Workspaces[0].Apps[0].Type)
}
