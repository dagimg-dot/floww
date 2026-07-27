package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidWorkspaceBackends_ContainsExpected(t *testing.T) {
	assert.True(t, ValidWorkspaceBackends["auto"])
	assert.True(t, ValidWorkspaceBackends["hyprland"])
	assert.True(t, ValidWorkspaceBackends["niri"])
	assert.True(t, ValidWorkspaceBackends["ewmh"])
	assert.True(t, ValidWorkspaceBackends["wmctrl"])
}

func TestValidWorkspaceBackends_RejectsInvalid(t *testing.T) {
	assert.False(t, ValidWorkspaceBackends["invalid"])
	assert.False(t, ValidWorkspaceBackends[""])
	assert.False(t, ValidWorkspaceBackends["kde"])
	assert.False(t, ValidWorkspaceBackends["gnome"])
}

func TestValidWorkspaceBackends_Count(t *testing.T) {
	assert.Len(t, ValidWorkspaceBackends, 5)
}

func TestDefaultConfigValues_General(t *testing.T) {
	assert.True(t, DefaultConfigValues.General.ShowNotifications)
	assert.Equal(t, "auto", DefaultConfigValues.General.WorkspaceBackend)
}

func TestDefaultConfigValues_Timing(t *testing.T) {
	assert.Equal(t, float64(3), DefaultConfigValues.Timing.WorkspaceSwitchWait)
	assert.Equal(t, float64(1), DefaultConfigValues.Timing.AppLaunchWait)
	assert.True(t, DefaultConfigValues.Timing.RespectAppWait)
}

func TestDefaultConfigValues_AllFieldsSet(t *testing.T) {
	cfg := DefaultConfigValues
	assert.NotEmpty(t, cfg.General.WorkspaceBackend)
	assert.Greater(t, cfg.Timing.WorkspaceSwitchWait, float64(0))
	assert.Greater(t, cfg.Timing.AppLaunchWait, float64(0))
}

func TestDefaultConfig_Structure(t *testing.T) {
	cfg := DefaultConfigValues
	// GeneralConfig fields
	_ = cfg.General.ShowNotifications
	_ = cfg.General.WorkspaceBackend
	// TimingConfig fields
	_ = cfg.Timing.WorkspaceSwitchWait
	_ = cfg.Timing.AppLaunchWait
	_ = cfg.Timing.RespectAppWait
	// If any of these fields were renamed or removed, this test wouldn't compile
}

func TestFileTypeConstants(t *testing.T) {
	assert.Equal(t, FileType("yaml"), FileTypeYAML)
	assert.Equal(t, FileType("json"), FileTypeJSON)
	assert.Equal(t, FileType("toml"), FileTypeTOML)
}
