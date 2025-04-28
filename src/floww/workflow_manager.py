import logging
import typer
from typing import Dict, Any

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

        for workspace in workflow_data.get("workspaces", []):
            target = workspace["target"]
            apps = workspace.get("apps", [])

            typer.echo(f"--> Switching to workspace {target}...")
            if not self.workspace_mgr.switch(target):
                typer.secho("Error: Failed to switch workspaces", fg="red")
                success = False
                continue  # Don't try to launch apps if workspace switch failed

            for app_config in apps:
                app_name = app_config.get("name", app_config["exec"])
                typer.echo(f"    -> Launching {app_name}...")
                try:
                    if not self.app_launcher.launch_app(app_config):
                        typer.secho(f"Failed to launch {app_name}", fg="red")
                        success = False
                except AppLaunchError as e:
                    typer.secho(str(e), fg="red")
                    success = False

        if success:
            typer.secho("✓ Workflow applied successfully", fg="green")
        else:
            typer.secho("⚠ Workflow completed with errors", fg="yellow")

        logger.info("Workflow application process finished.")
        return success
