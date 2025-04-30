import os
import yaml
from pathlib import Path
from typing import Dict, List, Any
import logging

from .errors import (
    ConfigError,
    WorkflowNotFoundError,
    WorkflowSchemaError,
    ConfigLoadError,
)
from .loader import ConfigLoader
from floww.utils import Singleton, DEFAULT_CONFIG, SAMPLE_WORKFLOW_CONTENT

logger = logging.getLogger(__name__)


class ConfigManager(metaclass=Singleton):
    """Handles loading and initializing configuration and workflows."""

    default_conf = DEFAULT_CONFIG

    def __init__(self, config_path: Path = None):
        self.config_loader = ConfigLoader()

        if config_path:
            self.config_path = config_path
        else:
            # Determine base config directory respecting XDG standard
            xdg_config_home = os.environ.get("XDG_CONFIG_HOME")
            base_dir = (
                Path(xdg_config_home) if xdg_config_home else Path.home() / ".config"
            )
            self.config_path = base_dir / "floww"
            logger.debug(f"Using default config path: {self.config_path}")

        # TODO: Make this configurable for different file formats
        self.config_file = self.config_path / "config.yaml"
        self.workflows_dir = self.config_path / "workflows"

        self.config = self._load_and_merge_config()

    def _load_and_merge_config(self) -> Dict[str, Any]:
        """Loads user config and merges it with defaults."""
        user_config = self._load_main_config_file()

        logger.debug(f"User configuration: {user_config}")

        merged_config = yaml.safe_load(yaml.dump(self.default_conf))

        if isinstance(user_config.get("timing"), dict):
            user_timing = user_config["timing"]

            for key, default_value in self.default_conf["timing"].items():
                if key in user_timing:
                    value = user_timing[key]
                    if key in ["workspace_switch_wait", "app_launch_wait"]:
                        try:
                            float_value = float(value)
                            if float_value >= 0:
                                merged_config["timing"][key] = float_value
                            else:
                                logger.warning(
                                    f"Invalid negative timing value for '{key}' ({value}). Using default ({default_value})."
                                )
                        except (ValueError, TypeError):
                            logger.warning(
                                f"Invalid non-numeric timing value for '{key}' ('{value}'). Using default ({default_value})."
                            )
                    elif key == "respect_app_wait":
                        if isinstance(value, bool):
                            merged_config["timing"][key] = value
                        else:
                            logger.warning(
                                f"Invalid non-boolean value for '{key}' ('{value}'). Using default ({default_value})."
                            )

        logger.debug(f"Final merged configuration: {merged_config}")
        return merged_config

    def _load_main_config_file(
        self,
    ) -> Dict[str, Any]:
        """Loads the main config file. Returns empty dict if not found or invalid.

        Supports multiple formats (YAML, JSON, TOML) using ConfigLoader.
        """
        config_base = self.config_path / "config"
        config_files = [
            config_base.with_suffix(ext)
            for ext in self.config_loader.get_supported_formats()
        ]

        # Prioritize the default config.yaml if it exists
        if self.config_file.is_file():
            config_files.insert(0, self.config_file)

        for file_path in config_files:
            if file_path.is_file():
                try:
                    config_data = self.config_loader.load(file_path)
                    if not isinstance(config_data, dict):
                        logger.warning(
                            f"Invalid format in config file {file_path}: Expected a dictionary (mapping), found {type(config_data)}. Ignoring."
                        )
                        continue

                    logger.debug(f"Successfully loaded config from {file_path}")
                    return config_data
                except ConfigLoadError as e:
                    logger.error(f"Error loading config file {file_path}: {e}")
                except Exception as e:
                    logger.error(
                        f"An unexpected error occurred loading config {file_path}: {e}"
                    )

        logger.debug(f"No valid config file found in {self.config_path}")
        return {}

    def init(self, create_example: bool = False, file_type: str = "yaml"):
        """Create config directory, default config.yaml, and workflows directory.

        Args:
            create_example: Whether to create an example workflow file
            file_type: File format for the example workflow (yaml, json, toml)
        """
        try:
            self.config_path.mkdir(parents=True, exist_ok=True)

            if not self.config_file.exists():
                default_content = {}
                self.config_loader.save(default_content, self.config_file)

            self.workflows_dir.mkdir(exist_ok=True)

            if create_example:
                supported_formats = [
                    ext.lstrip(".")
                    for ext in self.config_loader.get_supported_formats()
                ]

                if file_type not in supported_formats:
                    logger.warning(
                        f"Unsupported file type '{file_type}', using yaml instead"
                    )
                    file_type = "yaml"

                sample_workflow_path = self.workflows_dir / f"example.{file_type}"

                if not any(self.workflows_dir.glob("example.*")):
                    try:
                        self.config_loader.save(
                            SAMPLE_WORKFLOW_CONTENT, sample_workflow_path
                        )
                        logger.debug(
                            f"Created example workflow in {file_type} format at {sample_workflow_path}"
                        )
                    except ConfigLoadError as e:
                        logger.error(f"Failed to create example workflow: {e}")

        except OSError as e:
            raise ConfigError(f"Failed to initialize config directory: {e}") from e
        except Exception as e:
            raise ConfigError(
                f"An unexpected error occurred during initialization: {e}"
            ) from e

    def list_workflow_names(self) -> List[str]:
        """Return list of workflow file names (without extension), sorted."""
        if not self.workflows_dir.is_dir():
            return []
        try:
            workflows = set()
            for ext in self.config_loader.get_supported_formats():
                workflows.update([p.stem for p in self.workflows_dir.glob(f"*{ext}")])
            return sorted(list(workflows))
        except OSError as e:
            raise ConfigError(
                f"Failed to list workflows in {self.workflows_dir}: {e}"
            ) from e

    def validate_workflow(
        self, workflow_name: str, workflow_data: Dict[str, Any]
    ) -> None:
        """
        Validates a workflow data against the expected schema.

        Args:
            workflow_name: The name of the workflow (used for error messages).
            workflow_data: The workflow data to validate.

        Raises:
            WorkflowSchemaError: If the workflow doesn't adhere to the expected schema.
        """
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

        if "final_workspace" in workflow_data:
            final_workspace = workflow_data["final_workspace"]
            if not isinstance(final_workspace, int) or final_workspace < 0:
                raise WorkflowSchemaError(
                    f"The 'final_workspace' key in workflow '{workflow_name}' must be an integer greater than or equal to 0."
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

            if not isinstance(ws["target"], int) or ws["target"] < 0:
                raise WorkflowSchemaError(
                    f"The 'target' key for {ws_id} must be an integer greater than or equal to 0."
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

                if "name" not in app:
                    raise WorkflowSchemaError(
                        f"App definition for {app_id} is missing the required 'name' key."
                    )

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
                app["type"] = app_type

        description = workflow_data.get("description")
        if description is not None and not isinstance(description, str):
            raise WorkflowSchemaError(
                f"The 'description' in workflow '{workflow_name}', if present, must be a string."
            )

    def load_workflow(
        self, workflow_name: str, is_direct_load: bool = False
    ) -> Dict[str, Any]:
        """
        Loads and validates a workflow file from the workflows directory or from a specific file path.

        Supports multiple formats (YAML, JSON, TOML) using ConfigLoader.

        Args:
            workflow_name: The name of the workflow file (without extension) or the full file path if is_direct_load is True.
            is_direct_load: If True, workflow_name is treated as a direct file path.

        Returns:
            A dictionary representing the parsed and validated workflow.

        Raises:
            WorkflowNotFoundError: If the workflow file doesn't exist.
            ConfigError: If the file is invalid or cannot be read.
            WorkflowSchemaError: If the workflow doesn't adhere to the expected schema.
        """
        workflow_file = None

        if is_direct_load:
            workflow_file = Path(workflow_name)
            if not workflow_file.is_file():
                raise WorkflowNotFoundError(f"Workflow file not found: {workflow_name}")

            display_name = workflow_file.stem
        else:
            workflow_base = self.workflows_dir / workflow_name
            possible_files = [
                workflow_base.with_suffix(ext)
                for ext in self.config_loader.get_supported_formats()
            ]

            for file_path in possible_files:
                if file_path.is_file():
                    workflow_file = file_path
                    break

            if not workflow_file:
                extensions = ", ".join(self.config_loader.get_supported_formats())
                raise WorkflowNotFoundError(
                    f"Workflow '{workflow_name}' not found with any supported format ({extensions})"
                )

            display_name = workflow_name

        try:
            workflow_data = self.config_loader.load(workflow_file)
        except ConfigLoadError as e:
            raise ConfigError(f"Error loading workflow '{display_name}': {e}") from e
        except Exception as e:
            raise ConfigError(
                f"An unexpected error occurred loading workflow '{display_name}': {e}"
            ) from e

        self.validate_workflow(display_name, workflow_data)
        return workflow_data

    def list_workflows(self) -> List[str]:
        """Return list of workflow file names (without extension)."""
        if not self.workflows_dir.is_dir():
            return []

        try:
            workflows = set()
            for ext in self.config_loader.get_supported_formats():
                workflows.update([p.stem for p in self.workflows_dir.glob(f"*{ext}")])
            return sorted(list(workflows))
        except OSError as e:
            raise ConfigError(
                f"Failed to list workflows in {self.workflows_dir}: {e}"
            ) from e

    def get_config(self) -> Dict[str, Any]:
        """Returns the currently loaded and merged configuration."""
        return self.config

    def get_timing_config(self) -> dict:
        """
        Get timing configuration settings from the loaded config.

        Returns:
            dict: Dictionary with timing configuration values
        """
        return self.config.get(
            "timing", yaml.safe_load(yaml.dump(self.default_conf["timing"]))
        )

    def is_initialized(self) -> bool:
        """Check if the config directory exists and is properly initialized."""
        return self.config_path.is_dir() and self.workflows_dir.is_dir()
