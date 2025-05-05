import logging
import typer
from typing import Dict, Any
import time

from .workspace import WorkspaceManager
from .launcher import AppLauncher
from .errors import AppLaunchError
from .config import ConfigManager
from floww.utils import notify

logger = logging.getLogger(__name__)


class WorkflowManager:
    """Manages the application of workflows."""

    def __init__(self, show_notifications: bool = True):
        self.workspace_mgr = WorkspaceManager()
        self.app_launcher = AppLauncher()
        self.config_mgr = ConfigManager()
        self.show_notifications = (
            show_notifications
            and self.config_mgr.get_general_config()["show_notifications"]
        )

    def apply(self, workflow_data: Dict[str, Any], append: bool = False) -> bool:
        """
        Apply a workflow by switching workspaces and launching apps.

        Args:
            workflow_data: The parsed workflow dictionary

        Returns:
            bool: True if the workflow application process completed.
                  Note: Individual app launch failures are logged but don't cause
                  the entire workflow to return False unless workspace switching fails.
        Raises:
            WorkspaceError: If a workspace switch fails.
        """
        timing_config = self.config_mgr.get_timing_config()
        workspace_switch_wait = timing_config.get("workspace_switch_wait", 2)
        app_launch_wait = timing_config.get("app_launch_wait", 1)
        respect_app_wait = timing_config.get("respect_app_wait", True)

        if not workflow_data.get("workspaces"):
            logger.warning("Workflow contains no workspaces")

        if "description" in workflow_data:
            typer.echo(f"Workflow: {workflow_data['description']}")

        success = True
        num_workspaces = len(workflow_data.get("workspaces", []))
        total_workspaces = self.workspace_mgr.get_total_workspaces()

        for workspace_idx, workspace in enumerate(workflow_data.get("workspaces", [])):
            target = workspace["target"]
            apps = workspace.get("apps", [])

            if append:
                target += total_workspaces - 1

            typer.echo(f"--> Switching to workspace {target}...")
            if not self.workspace_mgr.switch(target):
                typer.secho(f"Error: Failed to switch workspace {target}", fg="red")
                success = False
                continue

            num_apps = len(apps)
            last_app_wait_to_apply = 0.0  # Initialize here to avoid UnboundLocalError

            for app_idx, app_config in enumerate(apps):
                app_name = app_config.get("name", app_config["exec"])
                typer.echo(f"    -> Launching {app_name}...")
                app_launched = False
                try:
                    app_launched = self.app_launcher.launch_app(app_config)
                    if not app_launched:
                        typer.secho(f"    ✗ Failed to launch {app_name}", fg="red")
                        success = False
                except AppLaunchError as e:
                    typer.secho(f"    ✗ Error launching {app_name}: {e}", fg="red")
                    success = False

                is_last_app_in_list = app_idx == num_apps - 1

                if app_launched:
                    current_app_wait = 0.0
                    app_wait_config = (
                        app_config.get("wait") if respect_app_wait else None
                    )

                    if app_wait_config is not None:
                        try:
                            wait_seconds = float(app_wait_config)
                            if wait_seconds >= 0:
                                current_app_wait = wait_seconds
                            else:
                                logger.warning(
                                    f"Invalid negative wait time ({app_wait_config}) for app '{app_name}', ignoring."
                                )
                        except (ValueError, TypeError):
                            logger.warning(
                                f"Invalid wait time ('{app_wait_config}') for app '{app_name}', ignoring."
                            )
                    elif not is_last_app_in_list:
                        current_app_wait = app_launch_wait

                    # Apply wait time after app launch, even for the last app in the last workspace
                    # but only if we're not at the very end of the workflow (last app in last workspace with no final_workspace)
                    should_skip_wait = (
                        is_last_app_in_list
                        and workspace_idx == num_workspaces - 1
                        and "final_workspace" not in workflow_data
                    )

                    if current_app_wait > 0 and not should_skip_wait:
                        typer.echo(
                            f"    ... Waiting {current_app_wait:.1f}s before next action..."
                        )
                        time.sleep(current_app_wait)

                    # Store the wait time if it was the last app in the list for potential use later
                    if is_last_app_in_list:
                        last_app_wait_to_apply = current_app_wait

            # Fix: Only apply wait if we're not at the last workspace
            if workspace_idx < num_workspaces - 1:
                final_wait = (
                    last_app_wait_to_apply
                    if last_app_wait_to_apply > 0
                    else workspace_switch_wait
                )

                if final_wait > 0:
                    wait_reason = (
                        "last app" if last_app_wait_to_apply > 0 else "workspace switch"
                    )
                    typer.echo(
                        f"    ... Waiting {final_wait:.1f}s (due to {wait_reason}) before next workspace..."
                    )
                    time.sleep(final_wait)

        # Store the last workspace's app wait for use with final_workspace
        last_workspace_app_wait = 0.0
        if num_workspaces > 0 and "final_workspace" in workflow_data:
            last_workspace = workflow_data["workspaces"][-1]
            last_apps = last_workspace.get("apps", [])
            if last_apps:
                last_app = last_apps[-1]
                app_wait_config = last_app.get("wait") if respect_app_wait else None
                if app_wait_config is not None:
                    try:
                        wait_seconds = float(app_wait_config)
                        if wait_seconds >= 0:
                            last_workspace_app_wait = wait_seconds
                    except (ValueError, TypeError):
                        pass

        # Switch to final workspace if specified
        if success and "final_workspace" in workflow_data:
            # Apply wait before switching to final workspace if needed
            final_wait = (
                last_workspace_app_wait
                if last_workspace_app_wait > 0
                else workspace_switch_wait
            )

            # Always apply a wait before switching to the final workspace,
            # using either the app's wait or the default workspace_switch_wait
            wait_reason = (
                "last app" if last_workspace_app_wait > 0 else "workspace switch"
            )
            typer.echo(
                f"    ... Waiting {final_wait:.1f}s (due to {wait_reason}) before final workspace..."
            )
            time.sleep(final_wait)

            final_workspace = workflow_data["final_workspace"]

            if append:
                final_workspace += total_workspaces - 1

            typer.echo(f"--> Switching to final workspace {final_workspace}...")
            if not self.workspace_mgr.switch(final_workspace):
                typer.secho(
                    f"Error: Failed to switch to final workspace {final_workspace}",
                    fg="red",
                )
                success = False

        if success:
            typer.secho("✓ Workflow applied successfully", fg="green")
            if self.show_notifications:
                notify("Workflow applied successfully")
        else:
            typer.secho("⚠ Workflow completed with errors", fg="yellow")
            if self.show_notifications:
                notify("Workflow completed with errors")

        logger.info("Workflow application process finished.")
        return success
