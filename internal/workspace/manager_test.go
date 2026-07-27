package workspace

import (
	"testing"

	"github.com/dagimg-dot/floww/internal/config"
	"github.com/dagimg-dot/floww/internal/workspace/backends"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBackend struct {
	switchCallCount int
	switchTarget    int
	total           int
	appendOffset    int
}

func (m *mockBackend) Switch(target int) bool {
	m.switchCallCount++
	m.switchTarget = target
	return true
}

func (m *mockBackend) GetTotalWorkspaces() int { return m.total }

func (m *mockBackend) GetAppendBaseOffset() int { return m.appendOffset }

func TestNewWorkspaceManager(t *testing.T) {
	mock := &mockBackend{}
	mgr := NewWorkspaceManager(mock)
	require.NotNil(t, mgr)
	assert.Equal(t, mock, mgr.backend)
}

func TestNewWorkspaceManager_NilBackend(t *testing.T) {
	mgr := NewWorkspaceManager(nil)
	require.NotNil(t, mgr)
	assert.Nil(t, mgr.backend)
}

func TestWorkspaceManager_SwitchDelegates(t *testing.T) {
	mock := &mockBackend{}
	mgr := NewWorkspaceManager(mock)

	result := mgr.Switch(3)

	assert.True(t, result)
	assert.Equal(t, 1, mock.switchCallCount)
	assert.Equal(t, 3, mock.switchTarget)
}

func TestWorkspaceManager_SwitchDelegatesMultiple(t *testing.T) {
	mock := &mockBackend{}
	mgr := NewWorkspaceManager(mock)

	mgr.Switch(0)
	mgr.Switch(5)
	mgr.Switch(2)

	assert.Equal(t, 3, mock.switchCallCount)
	assert.Equal(t, 2, mock.switchTarget)
}

func TestWorkspaceManager_GetTotalWorkspacesDelegates(t *testing.T) {
	mock := &mockBackend{total: 7}
	mgr := NewWorkspaceManager(mock)
	assert.Equal(t, 7, mgr.GetTotalWorkspaces())
}

func TestWorkspaceManager_GetTotalWorkspaces_Zero(t *testing.T) {
	mock := &mockBackend{total: 0}
	mgr := NewWorkspaceManager(mock)
	assert.Equal(t, 0, mgr.GetTotalWorkspaces())
}

func TestWorkspaceManager_GetAppendBaseOffsetDelegates(t *testing.T) {
	mock := &mockBackend{appendOffset: 4}
	mgr := NewWorkspaceManager(mock)
	assert.Equal(t, 4, mgr.GetAppendBaseOffset())
}

func TestCreateBackend_ExplicitHyprland(t *testing.T) {
	be := CreateBackend("hyprland", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.HyprlandBackend)
	assert.True(t, ok)
}

func TestCreateBackend_ExplicitNiri(t *testing.T) {
	be := CreateBackend("niri", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.NiriBackend)
	assert.True(t, ok)
}

func TestCreateBackend_ExplicitWmctrl(t *testing.T) {
	be := CreateBackend("wmctrl", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.WmctrlBackend)
	assert.True(t, ok)
}

func TestCreateBackend_ExplicitEwmhFallsBackToWmctrl(t *testing.T) {
	// Invalidate DISPLAY so XGB cannot connect, forcing the EWMH→wmctrl fallback.
	t.Setenv("DISPLAY", ":999")
	be := CreateBackend("ewmh", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.WmctrlBackend)
	assert.True(t, ok)
}

func TestCreateBackend_AutoDetectHyprland(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "test-instance")
	be := CreateBackend("auto", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.HyprlandBackend)
	assert.True(t, ok)
}

func TestCreateBackend_AutoDetectNiriViaSocket(t *testing.T) {
	t.Setenv("NIRI_SOCKET", "/tmp/niri.sock")
	be := CreateBackend("auto", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.NiriBackend)
	assert.True(t, ok)
}

func TestCreateBackend_AutoDetectNiriViaDesktop(t *testing.T) {
	t.Setenv("XDG_CURRENT_DESKTOP", "niri")
	be := CreateBackend("auto", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.NiriBackend)
	assert.True(t, ok)
}

func TestCreateBackend_AutoDetectNiriViaDesktopCaseInsensitive(t *testing.T) {
	t.Setenv("XDG_CURRENT_DESKTOP", "Niri")
	be := CreateBackend("auto", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.NiriBackend)
	assert.True(t, ok)
}

func TestCreateBackend_AutoDetectFallsToWmctrl(t *testing.T) {
	// Clear all compositor env vars and invalidate DISPLAY so EWMH fails too.
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "")
	t.Setenv("NIRI_SOCKET", "")
	t.Setenv("XDG_CURRENT_DESKTOP", "")
	t.Setenv("DISPLAY", ":999")
	be := CreateBackend("auto", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.WmctrlBackend)
	assert.True(t, ok)
}

func TestCreateBackend_HyprlandTakesPriorityOverNiri(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "test-instance")
	t.Setenv("NIRI_SOCKET", "/tmp/niri.sock")
	be := CreateBackend("auto", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.HyprlandBackend)
	assert.True(t, ok)
}

func TestCreateBackend_EmptyNameUsesConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := config.NewConfigManager(dir)
	// Default config has WorkspaceBackend: "auto".
	// Override env to force deterministic fallback to wmctrl.
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "")
	t.Setenv("NIRI_SOCKET", "")
	t.Setenv("XDG_CURRENT_DESKTOP", "")
	t.Setenv("DISPLAY", ":999")
	be := CreateBackend("", cfg)
	require.NotNil(t, be)
	_, ok := be.(*backends.WmctrlBackend)
	assert.True(t, ok)
}

func TestCreateBackend_UnknownNameFallsBackToAuto(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "")
	t.Setenv("NIRI_SOCKET", "")
	t.Setenv("XDG_CURRENT_DESKTOP", "")
	t.Setenv("DISPLAY", ":999")
	be := CreateBackend("nonexistent-backend", nil)
	require.NotNil(t, be)
	_, ok := be.(*backends.WmctrlBackend)
	assert.True(t, ok)
}

func TestNormalizeBackendName_Lowercases(t *testing.T) {
	assert.Equal(t, "hyprland", normalizeBackendName("HYPRLAND"))
	assert.Equal(t, "niri", normalizeBackendName("NIRI"))
	assert.Equal(t, "ewmh", normalizeBackendName("EWMH"))
	assert.Equal(t, "wmctrl", normalizeBackendName("WMCTRL"))
}

func TestNormalizeBackendName_Trims(t *testing.T) {
	assert.Equal(t, "hyprland", normalizeBackendName("  hyprland  "))
	assert.Equal(t, "niri", normalizeBackendName("\tniri\n"))
}

func TestNormalizeBackendName_ValidNames(t *testing.T) {
	for _, name := range []string{"auto", "hyprland", "niri", "ewmh", "wmctrl"} {
		assert.Equal(t, name, normalizeBackendName(name))
	}
}

func TestNormalizeBackendName_UnknownDefaultsToAuto(t *testing.T) {
	for _, name := range []string{"kde", "gnome", "sway", "", "awesome"} {
		assert.Equal(t, "auto", normalizeBackendName(name), "name: %q", name)
	}
}

func TestIsHyprlandSession(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "")
	assert.False(t, isHyprlandSession())

	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "abc123")
	assert.True(t, isHyprlandSession())
}

func TestIsNiriSession_ViaSocket(t *testing.T) {
	t.Setenv("NIRI_SOCKET", "")
	t.Setenv("XDG_CURRENT_DESKTOP", "")
	assert.False(t, isNiriSession())

	t.Setenv("NIRI_SOCKET", "/tmp/niri.sock")
	assert.True(t, isNiriSession())
}

func TestIsNiriSession_ViaDesktop(t *testing.T) {
	t.Setenv("NIRI_SOCKET", "")

	t.Setenv("XDG_CURRENT_DESKTOP", "niri")
	assert.True(t, isNiriSession())

	t.Setenv("XDG_CURRENT_DESKTOP", "Niri")
	assert.True(t, isNiriSession())

	t.Setenv("XDG_CURRENT_DESKTOP", "gnome")
	assert.False(t, isNiriSession())
}

func TestDetectAuto_HyprlandTakesPriority(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "test")
	t.Setenv("NIRI_SOCKET", "/tmp/niri.sock")
	be := detectAuto()
	_, ok := be.(*backends.HyprlandBackend)
	assert.True(t, ok)
}

func TestDetectAuto_NiriWhenNoHyprland(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "")
	t.Setenv("NIRI_SOCKET", "/tmp/niri.sock")
	be := detectAuto()
	_, ok := be.(*backends.NiriBackend)
	assert.True(t, ok)
}

func TestDetectAuto_FallsToWmctrl(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "")
	t.Setenv("NIRI_SOCKET", "")
	t.Setenv("XDG_CURRENT_DESKTOP", "")
	t.Setenv("DISPLAY", ":999")
	be := detectAuto()
	_, ok := be.(*backends.WmctrlBackend)
	assert.True(t, ok)
}
