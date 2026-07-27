package workflow

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/dagimg-dot/floww/internal/utils"
	"github.com/stretchr/testify/assert"
)

type mockWorkspaceManager struct {
	switchFunc   func(int) bool
	appendOffset int
	switchCalls  []int
}

func (m *mockWorkspaceManager) Switch(target int) bool {
	m.switchCalls = append(m.switchCalls, target)
	if m.switchFunc != nil {
		return m.switchFunc(target)
	}
	return true
}

func (m *mockWorkspaceManager) GetAppendBaseOffset() int {
	return m.appendOffset
}

type mockAppLauncher struct {
	launchFunc  func(App) (bool, error)
	launchCalls []App
}

func (m *mockAppLauncher) LaunchApp(app App) (bool, error) {
	m.launchCalls = append(m.launchCalls, app)
	if m.launchFunc != nil {
		return m.launchFunc(app)
	}
	return true, nil
}

type mockConfigManager struct {
	timingConfig  *utils.TimingConfig
	generalConfig *utils.GeneralConfig
}

func (m *mockConfigManager) GetTimingConfig() *utils.TimingConfig {
	return m.timingConfig
}

func (m *mockConfigManager) GetGeneralConfig() *utils.GeneralConfig {
	return m.generalConfig
}

func newTestWM(ws WorkspaceManager, al AppLauncher, cm ConfigManager) (*WorkflowManager, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	wm := NewWorkflowManager(ws, al, cm)
	wm.out = buf
	return wm, buf
}

func defaultTiming() *utils.TimingConfig {
	return &utils.TimingConfig{
		WorkspaceSwitchWait: 0,
		AppLaunchWait:       0,
		RespectAppWait:      true,
	}
}

func defaultGeneral() *utils.GeneralConfig {
	return &utils.GeneralConfig{
		ShowNotifications: false,
		WorkspaceBackend:  "auto",
	}
}

func newMockCfg(timing *utils.TimingConfig) *mockConfigManager {
	return &mockConfigManager{
		timingConfig:  timing,
		generalConfig: defaultGeneral(),
	}
}

// 1. Empty workflow — no workspaces, returns success.
func TestApply_EmptyWorkflow(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := &mockConfigManager{
		timingConfig:  defaultTiming(),
		generalConfig: defaultGeneral(),
	}
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{Workspaces: []Workspace{}}
	success := wm.Apply(data, false)

	assert.True(t, success)
	assert.Contains(t, buf.String(), "\033[32m✓ Workflow applied successfully\033[0m\n")
	assert.Empty(t, ws.switchCalls)
	assert.Empty(t, al.launchCalls)
}

// 2. Workflow with description — description printed to output.
func TestApply_WithDescription(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := &mockConfigManager{
		timingConfig:  defaultTiming(),
		generalConfig: defaultGeneral(),
	}
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Description: "Test workflow",
		Workspaces:  []Workspace{},
	}
	success := wm.Apply(data, false)

	assert.True(t, success)
	assert.Contains(t, buf.String(), "Workflow: Test workflow\n")
}

// 3. Workspace switch failure — continues to next workspace, returns false.
func TestApply_WorkspaceSwitchFailure(t *testing.T) {
	ws := &mockWorkspaceManager{
		switchFunc: func(target int) bool { return false },
	}
	al := &mockAppLauncher{}
	cm := &mockConfigManager{
		timingConfig:  defaultTiming(),
		generalConfig: defaultGeneral(),
	}
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{Target: 1, Apps: []App{{Name: "App1", Exec: "app1"}}},
		},
	}
	success := wm.Apply(data, false)

	assert.False(t, success)
	assert.Contains(t, buf.String(), "--> Switching to workspace 1...\n")
	assert.Contains(t, buf.String(), "\033[31mError: Failed to switch workspace 1\033[0m\n")
	assert.Contains(t, buf.String(), "\033[33m⚠ Workflow completed with errors\033[0m\n")
	assert.Equal(t, []int{1}, ws.switchCalls)
	assert.Empty(t, al.launchCalls)
}

// 4. Launch apps for workspace — all apps launched.
func TestApply_LaunchApps(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := &mockConfigManager{
		timingConfig:  defaultTiming(),
		generalConfig: defaultGeneral(),
	}
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{
				Target: 1,
				Apps: []App{
					{Name: "App1", Exec: "app1"},
					{Name: "App2", Exec: "app2", Args: []string{"--flag"}},
				},
			},
		},
	}
	success := wm.Apply(data, false)

	assert.True(t, success)
	assert.Contains(t, buf.String(), "\033[32m✓ Workflow applied successfully\033[0m\n")
	assert.Equal(t, []int{1}, ws.switchCalls)
	assert.Len(t, al.launchCalls, 2)
	assert.Equal(t, "App1", al.launchCalls[0].Name)
	assert.Equal(t, "app1", al.launchCalls[0].Exec)
	assert.Equal(t, "App2", al.launchCalls[1].Name)
	assert.Equal(t, "app2", al.launchCalls[1].Exec)
	assert.Equal(t, []string{"--flag"}, al.launchCalls[1].Args)
}

// 5. App launch failure — continues to next app, returns false.
func TestApply_AppLaunchFailure(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{
		launchFunc: func(app App) (bool, error) { return false, nil },
	}
	cm := &mockConfigManager{
		timingConfig:  defaultTiming(),
		generalConfig: defaultGeneral(),
	}
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{
				Target: 1,
				Apps: []App{
					{Name: "App1", Exec: "app1"},
					{Name: "App2", Exec: "app2"},
				},
			},
		},
	}
	success := wm.Apply(data, false)

	assert.False(t, success)
	assert.Contains(t, buf.String(), "    \033[31m✗ Failed to launch App1\033[0m\n")
	assert.Contains(t, buf.String(), "    \033[31m✗ Failed to launch App2\033[0m\n")
	assert.Contains(t, buf.String(), "\033[33m⚠ Workflow completed with errors\033[0m\n")
	assert.Equal(t, []int{1}, ws.switchCalls)
	assert.Len(t, al.launchCalls, 2)
}

// 5b. App launch error with err != nil — error formatted differently.
func TestApply_AppLaunchError(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{
		launchFunc: func(app App) (bool, error) {
			return false, fmt.Errorf("exec not found")
		},
	}
	cm := &mockConfigManager{
		timingConfig:  defaultTiming(),
		generalConfig: defaultGeneral(),
	}
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{
				Target: 1,
				Apps: []App{
					{Name: "App1", Exec: "app1"},
				},
			},
		},
	}
	success := wm.Apply(data, false)

	assert.False(t, success)
	assert.Contains(t, buf.String(), "    \033[31m✗ Error launching App1: exec not found\033[0m\n")
}

// 6. Multiple workspaces — switches to each.
func TestApply_MultipleWorkspaces(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := &mockConfigManager{
		timingConfig:  defaultTiming(),
		generalConfig: defaultGeneral(),
	}
	wm, _ := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{Target: 1, Apps: []App{{Name: "App1", Exec: "app1"}}},
			{Target: 2, Apps: []App{{Name: "App2", Exec: "app2"}}},
		},
	}
	success := wm.Apply(data, false)

	assert.True(t, success)
	assert.Equal(t, []int{1, 2}, ws.switchCalls)
	assert.Len(t, al.launchCalls, 2)
}

// 7. App-specific wait — uses app's wait value when respect_app_wait=true.
func TestApply_AppSpecificWait(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := newMockCfg(&utils.TimingConfig{
		WorkspaceSwitchWait: 0,
		AppLaunchWait:       0,
		RespectAppWait:      true,
	})
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{
				Target: 1,
				Apps: []App{
					{Name: "App1", Exec: "app1", Wait: ptr(0.1)},
					{Name: "App2", Exec: "app2"},
				},
			},
		},
	}
	success := wm.Apply(data, false)
	assert.True(t, success)

	output := buf.String()
	assert.Contains(t, output, "    ... Waiting 0.1s before next action...\n")
	assert.Equal(t, 1, strings.Count(output, "... Waiting"))
}

// 8. Default launch wait — uses app_launch_wait for non-last apps.
func TestApply_DefaultLaunchWait(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := newMockCfg(&utils.TimingConfig{
		WorkspaceSwitchWait: 0,
		AppLaunchWait:       1,
		RespectAppWait:      true,
	})
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{
				Target: 1,
				Apps: []App{
					{Name: "App1", Exec: "app1"},
					{Name: "App2", Exec: "app2"},
				},
			},
		},
	}
	success := wm.Apply(data, false)
	assert.True(t, success)

	output := buf.String()
	assert.Contains(t, output, "    ... Waiting 1.0s before next action...\n")
	assert.Equal(t, 1, strings.Count(output, "... Waiting"))
}

// 9. Workspace switch wait — uses workspace_switch_wait between workspaces.
func TestApply_WorkspaceSwitchWait(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := newMockCfg(&utils.TimingConfig{
		WorkspaceSwitchWait: 1,
		AppLaunchWait:       0,
		RespectAppWait:      true,
	})
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{Target: 1, Apps: []App{{Name: "App1", Exec: "app1"}}},
			{Target: 2, Apps: []App{{Name: "App2", Exec: "app2"}}},
		},
	}
	success := wm.Apply(data, false)
	assert.True(t, success)

	output := buf.String()
	assert.Contains(t, output, "    ... Waiting 1.0s (due to workspace switch) before next workspace...\n")
	assert.Equal(t, 1, strings.Count(output, "... Waiting"))
}

// 10. respect_app_wait=false — ignores app wait values, uses app_launch_wait instead.
func TestApply_RespectAppWaitDisabled(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := newMockCfg(&utils.TimingConfig{
		WorkspaceSwitchWait: 0,
		AppLaunchWait:       1,
		RespectAppWait:      false,
	})
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{
				Target: 1,
				Apps: []App{
					{Name: "App1", Exec: "app1", Wait: ptr(0.3)},
					{Name: "App2", Exec: "app2"},
				},
			},
		},
	}
	success := wm.Apply(data, false)
	assert.True(t, success)

	output := buf.String()
	assert.Contains(t, output, "    ... Waiting 1.0s before next action...\n")
	assert.NotContains(t, output, "Waiting 0.3s")
	assert.Equal(t, 1, strings.Count(output, "... Waiting"))
}

// 11. Last app wait between workspaces — last app's wait becomes workspace switch wait.
func TestApply_LastAppWaitBetweenWorkspaces(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := newMockCfg(&utils.TimingConfig{
		WorkspaceSwitchWait: 3,
		AppLaunchWait:       0,
		RespectAppWait:      true,
	})
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{
				Target: 1,
				Apps: []App{
					{Name: "App1", Exec: "app1"},
					{Name: "App2", Exec: "app2", Wait: ptr(0.1)},
				},
			},
			{Target: 2, Apps: []App{{Name: "App3", Exec: "app3"}}},
		},
	}
	success := wm.Apply(data, false)
	assert.True(t, success)

	output := buf.String()
	assert.Contains(t, output, "    ... Waiting 0.1s before next action...\n")
	assert.Contains(t, output, "    ... Waiting 0.1s (due to last app) before next workspace...\n")
	assert.Equal(t, 2, strings.Count(output, "... Waiting"))
}

// 12. Final workspace with wait — switches to final_workspace after app-specific wait.
func TestApply_FinalWorkspaceWithWait(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := newMockCfg(&utils.TimingConfig{
		WorkspaceSwitchWait: 1,
		AppLaunchWait:       0,
		RespectAppWait:      true,
	})
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{Target: 1, Apps: []App{{Name: "App1", Exec: "app1"}}},
			{
				Target: 2,
				Apps:   []App{{Name: "App2", Exec: "app2", Wait: ptr(0.1)}},
			},
		},
		FinalWorkspace: ptr(0),
	}
	success := wm.Apply(data, false)
	assert.True(t, success)

	output := buf.String()
	assert.Contains(t, output, "    ... Waiting 1.0s (due to workspace switch) before next workspace...\n")
	assert.Contains(t, output, "    ... Waiting 0.1s before next action...\n")
	assert.Contains(t, output, "    ... Waiting 0.1s (due to last app) before final workspace...\n")
	assert.Contains(t, output, "--> Switching to final workspace 0...\n")
	assert.Equal(t, 3, strings.Count(output, "... Waiting"))
	assert.Equal(t, []int{1, 2, 0}, ws.switchCalls)
}

// 13. Default workspace_switch_wait for final workspace — uses default when no app wait.
func TestApply_DefaultWaitForFinalWorkspace(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := newMockCfg(&utils.TimingConfig{
		WorkspaceSwitchWait: 1,
		AppLaunchWait:       0,
		RespectAppWait:      true,
	})
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{Target: 1, Apps: []App{{Name: "App1", Exec: "app1"}}},
			{Target: 2, Apps: []App{{Name: "App2", Exec: "app2"}}},
		},
		FinalWorkspace: ptr(0),
	}
	success := wm.Apply(data, false)
	assert.True(t, success)

	output := buf.String()
	assert.Contains(t, output, "    ... Waiting 1.0s (due to workspace switch) before next workspace...\n")
	assert.Contains(t, output, "    ... Waiting 1.0s (due to workspace switch) before final workspace...\n")
	assert.Equal(t, 2, strings.Count(output, "... Waiting"))
	assert.Equal(t, []int{1, 2, 0}, ws.switchCalls)
}

// 14. Append mode offset — adds GetAppendBaseOffset() to targets.
func TestApply_AppendModeOffset(t *testing.T) {
	ws := &mockWorkspaceManager{
		appendOffset: 4,
	}
	al := &mockAppLauncher{}
	cm := &mockConfigManager{
		timingConfig:  defaultTiming(),
		generalConfig: defaultGeneral(),
	}
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{Target: 1, Apps: []App{{Name: "App1", Exec: "app1"}}},
			{Target: 2, Apps: []App{{Name: "App2", Exec: "app2"}}},
		},
	}
	success := wm.Apply(data, true)

	assert.True(t, success)
	assert.Equal(t, []int{5, 6}, ws.switchCalls)
	assert.Len(t, al.launchCalls, 2)
	assert.Contains(t, buf.String(), "--> Switching to workspace 5...\n")
	assert.Contains(t, buf.String(), "--> Switching to workspace 6...\n")
}

// 14b. Append mode with final workspace — final workspace is also offset.
func TestApply_AppendModeWithFinalWorkspace(t *testing.T) {
	ws := &mockWorkspaceManager{
		appendOffset: 4,
	}
	al := &mockAppLauncher{}
	cm := &mockConfigManager{
		timingConfig:  defaultTiming(),
		generalConfig: defaultGeneral(),
	}
	wm, _ := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{Target: 1, Apps: []App{{Name: "App1", Exec: "app1"}}},
		},
		FinalWorkspace: ptr(0),
	}
	success := wm.Apply(data, true)

	assert.True(t, success)
	assert.Equal(t, []int{5, 4}, ws.switchCalls)
}

// 15. Notification on success — calls Notify when showNotifications=true.
func TestApply_NotificationOnSuccess(t *testing.T) {
	ws := &mockWorkspaceManager{}
	al := &mockAppLauncher{}
	cm := &mockConfigManager{
		timingConfig: defaultTiming(),
		generalConfig: &utils.GeneralConfig{
			ShowNotifications: true,
			WorkspaceBackend:  "auto",
		},
	}
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{Target: 1, Apps: []App{{Name: "App1", Exec: "app1"}}},
		},
	}
	success := wm.Apply(data, false)

	assert.True(t, success)
	// Notify is called internally but cannot be mocked at package level.
	// The success output confirms the correct code path was taken.
	assert.Contains(t, buf.String(), "\033[32m✓ Workflow applied successfully\033[0m\n")
	assert.True(t, wm.showNotifications)
}

// 16. Notification on partial failure — calls Notify with error message.
func TestApply_NotificationOnPartialFailure(t *testing.T) {
	ws := &mockWorkspaceManager{
		switchFunc: func(target int) bool { return false },
	}
	al := &mockAppLauncher{}
	cm := &mockConfigManager{
		timingConfig: defaultTiming(),
		generalConfig: &utils.GeneralConfig{
			ShowNotifications: true,
			WorkspaceBackend:  "auto",
		},
	}
	wm, buf := newTestWM(ws, al, cm)

	data := &Workflow{
		Workspaces: []Workspace{
			{Target: 1, Apps: []App{{Name: "App1", Exec: "app1"}}},
		},
	}
	success := wm.Apply(data, false)

	assert.False(t, success)
	assert.Contains(t, buf.String(), "\033[33m⚠ Workflow completed with errors\033[0m\n")
	assert.True(t, wm.showNotifications)
}
