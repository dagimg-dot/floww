// Package workflow provides the WorkflowManager that applies workflows by
// switching workspaces and launching applications.
package workflow

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dagimg-dot/floww/internal/utils"
)

// ANSI color codes for terminal output (matching Python's typer.secho).
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// AppLauncher defines the interface for launching applications.
type AppLauncher interface {
	LaunchApp(app App) (bool, error)
}

// WorkspaceManager defines the interface for workspace operations.
type WorkspaceManager interface {
	Switch(target int) bool
	GetAppendBaseOffset() int
}

// ConfigManager defines the interface for reading configuration.
type ConfigManager interface {
	GetTimingConfig() *utils.TimingConfig
	GetGeneralConfig() *utils.GeneralConfig
}

// WorkflowManager manages the application of workflows.
type WorkflowManager struct {
	workspaceMgr      WorkspaceManager
	appLauncher       AppLauncher
	configMgr         ConfigManager
	showNotifications bool
	out               io.Writer
	sleepFn           func(time.Duration)
}

// NewWorkflowManager creates a new WorkflowManager with the given dependencies.
func NewWorkflowManager(ws WorkspaceManager, al AppLauncher, cm ConfigManager) *WorkflowManager {
	return &WorkflowManager{
		workspaceMgr:      ws,
		appLauncher:       al,
		configMgr:         cm,
		showNotifications: cm.GetGeneralConfig().ShowNotifications,
		out:               os.Stdout,
		sleepFn:           time.Sleep,
	}
}

// Apply applies a workflow by switching workspaces and launching apps.
// Individual app launch failures are logged but don't cause the entire
// workflow to return false — only workspace switch failures contribute
// to the final status.
func (wm *WorkflowManager) Apply(data *Workflow, append bool) bool {
	timing := wm.configMgr.GetTimingConfig()
	workspaceSwitchWait := timing.WorkspaceSwitchWait
	appLaunchWait := timing.AppLaunchWait
	respectAppWait := timing.RespectAppWait

	if data.Description != "" {
		_, _ = fmt.Fprintf(wm.out, "Workflow: %s\n", data.Description)
	}

	success := true
	numWorkspaces := len(data.Workspaces)
	appendBaseOffset := 0
	if append {
		appendBaseOffset = wm.workspaceMgr.GetAppendBaseOffset()
	}

	for workspaceIdx := range data.Workspaces {
		ws := &data.Workspaces[workspaceIdx]
		target := ws.Target

		if append {
			target += appendBaseOffset
		}

		_, _ = fmt.Fprintf(wm.out, "--> Switching to workspace %d...\n", target)
		if !wm.workspaceMgr.Switch(target) {
			_, _ = fmt.Fprintf(wm.out, "%sError: Failed to switch workspace %d%s\n", colorRed, target, colorReset)
			success = false
			continue
		}

		numApps := len(ws.Apps)
		lastAppWaitToApply := 0.0

		for appIdx := range ws.Apps {
			app := &ws.Apps[appIdx]
			appName := app.Name
			if appName == "" {
				appName = app.Exec
			}

			_, _ = fmt.Fprintf(wm.out, "    -> Launching %s...\n", appName)
			appLaunched := false

			launched, err := wm.appLauncher.LaunchApp(*app)
			switch {
			case err != nil:
				_, _ = fmt.Fprintf(wm.out, "    %s✗ Error launching %s: %s%s\n", colorRed, appName, err.Error(), colorReset)
				success = false
			case !launched:
				_, _ = fmt.Fprintf(wm.out, "    %s✗ Failed to launch %s%s\n", colorRed, appName, colorReset)
				success = false
			default:
				appLaunched = true
			}

			isLastAppInList := appIdx == numApps-1

			if appLaunched {
				currentAppWait := 0.0

				var appWaitConfig *float64
				if respectAppWait {
					appWaitConfig = app.Wait
				}

				if appWaitConfig != nil {
					waitSeconds := *appWaitConfig
					if waitSeconds >= 0 {
						currentAppWait = waitSeconds
					}
				} else if !isLastAppInList {
					currentAppWait = float64(appLaunchWait)
				}

				shouldSkipWait := isLastAppInList &&
					workspaceIdx == numWorkspaces-1 &&
					data.FinalWorkspace == nil

				if currentAppWait > 0 && !shouldSkipWait {
					_, _ = fmt.Fprintf(wm.out, "    ... Waiting %.1fs before next action...\n", currentAppWait)
					wm.sleepFn(time.Duration(currentAppWait * float64(time.Second)))
				}

				if isLastAppInList {
					lastAppWaitToApply = currentAppWait
				}
			}
		}

		if workspaceIdx < numWorkspaces-1 {
			finalWait := lastAppWaitToApply
			if lastAppWaitToApply <= 0 {
				finalWait = float64(workspaceSwitchWait)
			}

			if finalWait > 0 {
				waitReason := "last app"
				if lastAppWaitToApply <= 0 {
					waitReason = "workspace switch"
				}
				_, _ = fmt.Fprintf(wm.out, "    ... Waiting %.1fs (due to %s) before next workspace...\n", finalWait, waitReason)
				wm.sleepFn(time.Duration(finalWait * float64(time.Second)))
			}
		}
	}

	// Re-read the last workspace's last app wait config directly for use with
	// the final workspace (matching Python's behaviour of re-reading from config).
	lastWorkspaceAppWait := 0.0
	if numWorkspaces > 0 && data.FinalWorkspace != nil {
		lastWorkspace := &data.Workspaces[numWorkspaces-1]
		if len(lastWorkspace.Apps) > 0 {
			lastApp := &lastWorkspace.Apps[len(lastWorkspace.Apps)-1]
			var appWaitConfig *float64
			if respectAppWait {
				appWaitConfig = lastApp.Wait
			}
			if appWaitConfig != nil {
				waitSeconds := *appWaitConfig
				if waitSeconds >= 0 {
					lastWorkspaceAppWait = waitSeconds
				}
			}
		}
	}

	if success && data.FinalWorkspace != nil {
		finalWait := lastWorkspaceAppWait
		if lastWorkspaceAppWait <= 0 {
			finalWait = float64(workspaceSwitchWait)
		}

		waitReason := "last app"
		if lastWorkspaceAppWait <= 0 {
			waitReason = "workspace switch"
		}
		_, _ = fmt.Fprintf(wm.out, "    ... Waiting %.1fs (due to %s) before final workspace...\n", finalWait, waitReason)
		wm.sleepFn(time.Duration(finalWait * float64(time.Second)))

		finalWorkspace := *data.FinalWorkspace
		if append {
			finalWorkspace += appendBaseOffset
		}

		_, _ = fmt.Fprintf(wm.out, "--> Switching to final workspace %d...\n", finalWorkspace)
		if !wm.workspaceMgr.Switch(finalWorkspace) {
			_, _ = fmt.Fprintf(wm.out, "%sError: Failed to switch to final workspace %d%s\n", colorRed, finalWorkspace, colorReset)
			success = false
		}
	}

	if success {
		_, _ = fmt.Fprintf(wm.out, "%s✓ Workflow applied successfully%s\n", colorGreen, colorReset)
		if wm.showNotifications {
			utils.Notify("Workflow applied successfully")
		}
	} else {
		_, _ = fmt.Fprintf(wm.out, "%s⚠ Workflow completed with errors%s\n", colorYellow, colorReset)
		if wm.showNotifications {
			utils.Notify("Workflow completed with errors")
		}
	}

	return success
}
