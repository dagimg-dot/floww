import logging
import typer
from typing import Dict, Any
import time  # Import the time module

from .workspace import WorkspaceManager
from .app_launcher import AppLauncher
from .errors import AppLaunchError

logger = logging.getLogger(__name__)


class WorkflowManager:
    """Manages the application of workflows."""

    def __init__(self):
        self.workspace_mgr = WorkspaceManager()
        self.app_launcher = AppLauncher()

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
        if not workflow_data.get("workspaces"):
            logger.warning("Workflow contains no workspaces")

        if "description" in workflow_data:
            typer.echo(f"Workflow: {workflow_data['description']}")

        success = True

        for workspace_index, workspace in enumerate(
            workflow_data.get("workspaces", [])
        ):
            target = workspace["target"]
            apps = workspace.get("apps", [])

            typer.echo(f"--> Switching to workspace {target}...")
            if not self.workspace_mgr.switch(target):
                typer.secho(f"Error: Failed to switch workspace {target}", fg="red")
                success = False
                continue

            for app_index, app_config in enumerate(apps):
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

                # Check for wait time after launching the app
                wait_time = app_config.get("wait", None)
                # Only wait if the app launch didn't immediately fail
                if app_launched and wait_time is not None:
                    try:
                        wait_seconds = float(wait_time)
                        if wait_seconds > 0:
                            typer.echo(
                                f"    ... Waiting {wait_seconds:.1f}s before next action..."
                            )
                            time.sleep(wait_seconds)
                        elif wait_seconds < 0:
                            logger.warning(
                                f"Invalid negative wait time ({wait_seconds}s) for app '{app_name}', ignoring."
                            )
                    except (ValueError, TypeError):
                        logger.warning(
                            f"Invalid wait time ('{wait_time}') for app '{app_name}', ignoring."
                        )

        if success:
            typer.secho("✓ Workflow applied successfully", fg="green")
        else:
            typer.secho("⚠ Workflow completed with errors", fg="yellow")

        logger.info("Workflow application process finished.")
        return success
