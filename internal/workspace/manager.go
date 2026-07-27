package workspace

import "github.com/dagimg-dot/floww/internal/workspace/backends"

// WorkspaceManager delegates workspace operations to a compositor-specific backend.
type WorkspaceManager struct {
	backend backends.WorkspaceBackend
}

// NewWorkspaceManager creates a WorkspaceManager with the given backend.
func NewWorkspaceManager(backend backends.WorkspaceBackend) *WorkspaceManager {
	return &WorkspaceManager{backend: backend}
}

// Switch to workspace target. Returns true on success.
func (m *WorkspaceManager) Switch(target int) bool {
	return m.backend.Switch(target)
}

// GetTotalWorkspaces returns the backend-specific notion of workspace count.
func (m *WorkspaceManager) GetTotalWorkspaces() int {
	return m.backend.GetTotalWorkspaces()
}

// GetAppendBaseOffset returns the offset added to workflow targets when append is true.
func (m *WorkspaceManager) GetAppendBaseOffset() int {
	return m.backend.GetAppendBaseOffset()
}
