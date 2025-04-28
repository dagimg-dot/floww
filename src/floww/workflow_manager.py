import time
import logging
import typer
from typing import Dict, Any

from .workspace import WorkspaceManager
from .app_launcher import AppLauncher
from .errors import WorkspaceError, AppLaunchError

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
        description = workflow_data.get("description", "")
        if description:
            typer.echo(f"Workflow: {description}")
            logger.info(f"Applying workflow: {description}")

        workspaces = workflow_data.get("workspaces", [])
        if not workspaces:
            typer.echo(
                typer.style("Warning:", fg=typer.colors.YELLOW)
                + " Workflow contains no workspaces"
            )
            logger.warning("Workflow contains no workspaces")
            return True

        overall_success = True
        for workspace in workspaces:
            target = workspace.get("target")

            typer.echo(f"--> Switching to workspace {target}...")
            logger.info(f"Switching to workspace {target}")

            try:
                self.workspace_mgr.switch(target)
            except WorkspaceError as e:
                raise e

            time.sleep(0.2)

            apps = workspace.get("apps", [])
            if not apps:
                logger.info(f"No apps defined for workspace {target}")
                continue

            for app in apps:
                app_name = app.get("name", app.get("exec", "Unknown App"))
                typer.echo(f"    -> Launching {app_name}...")
                logger.info(f"Launching app: {app_name}")

                try:
                    self.app_launcher.launch_app(app)
                except AppLaunchError as e:
                    err_msg = f"Failed to launch '{app_name}': {e}"
                    typer.echo(
                        typer.style(f"      Error: {err_msg}", fg=typer.colors.RED)
                    )
                    logger.error(err_msg)
                    overall_success = False

                time.sleep(0.1)

        if overall_success:
            typer.echo(
                typer.style("✓ Workflow applied successfully", fg=typer.colors.GREEN)
            )
        else:
            typer.echo(
                typer.style(
                    "⚠ Workflow applied, but some apps failed to launch.",
                    fg=typer.colors.YELLOW,
                )
            )

        logger.info("Workflow application process finished.")
        return True
