import os
import yaml
from pathlib import Path
from typing import Dict, List, Any
import logging

logger = logging.getLogger(__name__)


class WorkflowSchemaError(ValueError):
    """Custom exception for workflow validation errors."""

    pass


class ConfigManager:
    """Handles loading and initializing configuration and workflows."""

    def __init__(self, config_path: Path = None):
        # Allow override, else respect XDG_CONFIG_HOME or default to ~/.config
        if config_path:
            self.config_path = config_path
        else:
            xdg_config = os.environ.get("XDG_CONFIG_HOME")
            logger.debug(f"XDG_CONFIG_HOME: {xdg_config}")
            base = Path(xdg_config if xdg_config else Path.home() / ".config")
            logger.debug(f"Base config path: {base}")
            self.config_path = base / "floww"
            logger.debug(f"Final config path: {self.config_path}")
        self.config_file = self.config_path / "config.yaml"
        self.workflows_dir = self.config_path / "workflows"

    def init(self, create_example: bool = False):
        """Create config directory, default config.yaml, and workflows directory."""
        self.config_path.mkdir(parents=True, exist_ok=True)
        if not self.config_file.exists():
            default_content = {"workspaces": {}}
            with open(self.config_file, "w") as f:
                yaml.dump(default_content, f, default_flow_style=False)
        self.workflows_dir.mkdir(exist_ok=True)

        if create_example:
            sample_workflow_path = self.workflows_dir / "example.yaml"
            if not any(self.workflows_dir.glob("*.yaml")):
                sample_content = {
                    "description": "An example workflow.",
                    "workspaces": [
                        {
                            "target": 1,
                            "apps": [
                                {
                                    "name": "Editor",
                                    "exec": "gedit",
                                    "args": ["~/Notes/scratch.txt"],
                                },
                                {"name": "Terminal", "exec": "gnome-terminal"},
                            ],
                        },
                        {
                            "target": 2,
                            "apps": [
                                {
                                    "name": "Browser",
                                    "exec": "firefox",
                                    "args": ["https://duckduckgo.com"],
                                }
                            ],
                        },
                    ],
                }
                with open(sample_workflow_path, "w") as f:
                    yaml.dump(
                        sample_content, f, default_flow_style=False, sort_keys=False
                    )

    def list_workflow_names(self) -> List[str]:
        """Return list of workflow file names (without extension), sorted."""
        if not self.workflows_dir.is_dir():
            return []
        return sorted([p.stem for p in self.workflows_dir.glob("*.yaml")])

    def load_workflow(self, workflow_name: str) -> Dict[str, Any]:
        """
        Loads and validates a workflow YAML file from the workflows directory.

        Args:
            workflow_name: The name of the workflow file (without the .yaml extension).

        Returns:
            A dictionary representing the parsed and validated workflow.

        Raises:
            FileNotFoundError: If the workflow file doesn't exist.
            yaml.YAMLError: If the YAML is invalid.
            WorkflowSchemaError: If the workflow doesn't adhere to the expected schema.
        """
        workflow_file = self.workflows_dir / f"{workflow_name}.yaml"

        if not workflow_file.is_file():
            raise FileNotFoundError(
                f"Workflow '{workflow_name}' not found at: {workflow_file}"
            )

        try:
            with open(workflow_file, "r") as f:
                workflow_data = yaml.safe_load(f)
        except yaml.YAMLError as e:
            raise yaml.YAMLError(f"Error parsing YAML file {workflow_file}: {e}")
        except Exception as e:
            raise IOError(f"Could not read workflow file {workflow_file}: {e}")

        if workflow_data is None:
            raise WorkflowSchemaError(
                f"Workflow file '{workflow_name}' is empty or contains only null."
            )

        if not isinstance(workflow_data, dict):
            raise WorkflowSchemaError(
                f"Workflow '{workflow_name}' content must be a dictionary (YAML mapping)."
            )

        if "workspaces" not in workflow_data:
            raise WorkflowSchemaError(
                f"Workflow '{workflow_name}' is missing the required 'workspaces' key."
            )
        if not isinstance(workflow_data["workspaces"], list):
            raise WorkflowSchemaError(
                f"The 'workspaces' key in workflow '{workflow_name}' must contain a list."
            )

        for i, ws in enumerate(workflow_data["workspaces"]):
            ws_id = f"workspace index {i}"
            if isinstance(ws, dict) and "target" in ws:
                ws_id = f"workspace target '{ws['target']}' (index {i})"

            if not isinstance(ws, dict):
                raise WorkflowSchemaError(
                    f"Item at {ws_id} in 'workspaces' must be a dictionary."
                )
            if "target" not in ws:
                raise WorkflowSchemaError(
                    f"Item at {ws_id} is missing the required 'target' key."
                )

            if "apps" not in ws:
                raise WorkflowSchemaError(
                    f"Workspace definition for {ws_id} is missing the required 'apps' key."
                )
            if not isinstance(ws["apps"], list):
                raise WorkflowSchemaError(
                    f"The 'apps' key for {ws_id} must contain a list."
                )

            for j, app in enumerate(ws["apps"]):
                app_id = f"app index {j} in {ws_id}"
                app_name = app.get("name", "")
                if app_name:
                    app_id = f"app '{app_name}' ({app_id})"

                if not isinstance(app, dict):
                    raise WorkflowSchemaError(f"Item at {app_id} must be a dictionary.")
                if "exec" not in app:
                    raise WorkflowSchemaError(
                        f"App definition for {app_id} is missing the required 'exec' key."
                    )
                if not isinstance(app.get("exec"), str) or not app["exec"].strip():
                    raise WorkflowSchemaError(
                        f"The 'exec' key for {app_id} must be a non-empty string."
                    )

                if "args" in app and not isinstance(app["args"], list):
                    raise WorkflowSchemaError(
                        f"The 'args' key for {app_id}, if present, must be a list."
                    )

                app_type = app.get("type", "binary")
                if not isinstance(app_type, str):
                    raise WorkflowSchemaError(
                        f"The 'type' key for {app_id}, if present, must be a string."
                    )
                if app_type not in ["binary", "flatpak", "snap"]:
                    raise WorkflowSchemaError(
                        f"The 'type' key for {app_id} must be one of 'binary', 'flatpak', 'snap', but got '{app_type}'."
                    )
                app["type"] = app_type  # Ensure type is set even if defaulted

        description = workflow_data.get("description")
        if description is not None and not isinstance(description, str):
            raise WorkflowSchemaError(
                f"The 'description' in workflow '{workflow_name}', if present, must be a string."
            )

        return workflow_data

    # Keep the original method name for backward compatibility
    def list_workflows(self) -> List[str]:
        """Return list of workflow file names (without extension)."""
        return self.list_workflow_names()
