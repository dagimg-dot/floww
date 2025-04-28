import time
import logging
import typer
from typing import Dict, Any

from .workspace import WorkspaceManager
from .app_launcher import AppLauncher

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
            bool: True if the workflow was applied successfully
        """
        description = workflow_data.get("description", "")
        if description:
            typer.echo(f"Workflow description: {description}")
            logger.info(f"Applying workflow: {description}")

        workspaces = workflow_data.get("workspaces", [])
        if not workspaces:
            typer.echo("Warning: Workflow contains no workspaces")
            logger.warning("Workflow contains no workspaces")
            return True  

        for workspace in workspaces:
            target = workspace.get("target")

            typer.echo(f"Switching to workspace {target}...")
            logger.info(f"Switching to workspace {target}")

            success = self.workspace_mgr.switch(target)
            if not success:
                error_msg = f"Failed to switch to workspace {target}"
                typer.echo(error_msg)
                logger.error(error_msg)
                return False

            time.sleep(0.5)

            apps = workspace.get("apps", [])
            if not apps:
                typer.echo(f"No apps defined for workspace {target}")
                logger.info(f"No apps defined for workspace {target}")
                continue

            for app in apps:
                app_name = app.get("name", app["exec"])
                typer.echo(f"Launching {app_name}...")
                logger.info(f"Launching app: {app_name}")

                success = self.app_launcher.launch_app(app)
                if not success:
                    error_msg = f"Failed to launch {app_name}"
                    typer.echo(error_msg)
                    logger.error(error_msg)

                time.sleep(0.2)

        typer.echo("Workflow applied successfully")
        logger.info("Workflow applied successfully")
        return True
