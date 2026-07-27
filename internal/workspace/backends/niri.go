package backends

import (
	"encoding/json"
	"log/slog"
	"strconv"
)

// NiriBackend controls Niri workspaces via “niri msg“.
type NiriBackend struct {
	niriCmd string
}

// NewNiriBackend returns a NiriBackend that uses the given niri command
// (default "niri").
func NewNiriBackend() *NiriBackend {
	return &NiriBackend{
		niriCmd: "niri",
	}
}

// niriWorkspace represents a single workspace in niri's JSON output.
type niriWorkspace struct {
	ID        int  `json:"id"`
	Idx       int  `json:"idx"`
	IsFocused bool `json:"is_focused"`
	IsActive  bool `json:"is_active"`
}

// workflowTargetToIdx converts a workflow target to niri's 1-based idx,
// clamping values below 1 to 1.
func workflowTargetToIdx(target int) int {
	if target < 1 {
		return 1
	}
	return target
}

// fetchWorkspacesJSON runs "niri msg -j workspaces" and returns the parsed
// list of workspaces.
func (b *NiriBackend) fetchWorkspacesJSON() ([]niriWorkspace, error) {
	cmd := execCmd(b.niriCmd, "msg", "-j", "workspaces")
	out, err := cmd.Output()
	if err != nil {
		slog.Debug("niri workspaces fetch failed", "error", err)
		return nil, err
	}

	var workspaces []niriWorkspace
	if err := json.Unmarshal(out, &workspaces); err != nil {
		slog.Debug("niri workspaces parse failed", "error", err)
		return nil, err
	}
	return workspaces, nil
}

// focusedWorkspaceIdx returns the idx of the focused (or active) workspace.
func focusedWorkspaceIdx(workspaces []niriWorkspace) (int, bool) {
	for _, ws := range workspaces {
		if ws.IsFocused || ws.IsActive {
			return ws.Idx, true
		}
	}
	return 0, false
}

// isAlreadyFocusedOnTarget checks whether the currently focused workspace
// has the same idx as the desired target. Returns false when the workspace
// state cannot be determined, causing Switch to fall through to
// focus-workspace.
func (b *NiriBackend) isAlreadyFocusedOnTarget(target int) bool {
	workspaces, err := b.fetchWorkspacesJSON()
	if err != nil {
		return false
	}
	desired := workflowTargetToIdx(target)
	cur, ok := focusedWorkspaceIdx(workspaces)
	if !ok {
		return false
	}
	return cur == desired
}

// Switch focuses the given workspace (1-based). It skips the
// focus-workspace action when already on the target workspace
// (auto-back-and-forth guard).
func (b *NiriBackend) Switch(target int) bool {
	if b.isAlreadyFocusedOnTarget(target) {
		slog.Debug("niri: already on workspace", "idx", workflowTargetToIdx(target))
		return true
	}

	idx := workflowTargetToIdx(target)
	return runCmd(b.niriCmd, "msg", "action", "focus-workspace", strconv.Itoa(idx))
}

// GetTotalWorkspaces returns the number of niri workspaces on the focused
// output. Returns 1 when the list is empty (niri always has at least one
// workspace) and 0 when the query fails.
func (b *NiriBackend) GetTotalWorkspaces() int {
	workspaces, err := b.fetchWorkspacesJSON()
	if err != nil {
		return 0
	}
	if len(workspaces) == 0 {
		return 1
	}
	return len(workspaces)
}

// GetAppendBaseOffset returns max(0, total-1), the default offset for
// append mode.
func (b *NiriBackend) GetAppendBaseOffset() int {
	return AppendBaseOffset(b.GetTotalWorkspaces())
}
