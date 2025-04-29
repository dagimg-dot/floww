import logging
import typer
from typing import Dict, Any
import time

from .workspace import WorkspaceManager
from .app_launcher import AppLauncher
from .errors import AppLaunchError
from .config import ConfigManager

logger = logging.getLogger(__name__)


class WorkflowManager:
    """Manages the application of workflows."""

    def __init__(self):
        self.workspace_mgr = WorkspaceManager()
        self.app_launcher = AppLauncher()
        self.config_mgr = ConfigManager()

    def apply(self, workflow_data: Dict[str, Any]) -> bool:
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

        for workspace_idx, workspace in enumerate(workflow_data.get("workspaces", [])):
            target = workspace["target"]
            apps = workspace.get("apps", [])

            typer.echo(f"--> Switching to workspace {target}...")
            if not self.workspace_mgr.switch(target):
                typer.secho(f"Error: Failed to switch workspace {target}", fg="red")
                success = False
                continue

            num_apps = len(apps)

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

                    if current_app_wait > 0 and not (
                        is_last_app_in_list and workspace_idx == num_workspaces - 1
                    ):
                        typer.echo(
                            f"    ... Waiting {current_app_wait:.1f}s before next action..."
                        )
                        time.sleep(current_app_wait)

                    # Store the wait time if it was the last app in the list for potential use later
                    if is_last_app_in_list:
                        last_app_wait_to_apply = current_app_wait

            is_last_workspace = workspace_idx == num_workspaces - 1

            # Only apply workspace wait if NOT the last workspace
            if not is_last_workspace:
                # Use the specific wait from the actual last app if it had one, otherwise use the global workspace wait
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

        if success:
            typer.secho("✓ Workflow applied successfully", fg="green")
        else:
            typer.secho("⚠ Workflow completed with errors", fg="yellow")

        logger.info("Workflow application process finished.")
        return success
