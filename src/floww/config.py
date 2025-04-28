import os
import yaml
from pathlib import Path


class ConfigManager:
    """Handles loading and initializing configuration and workflows."""

    def __init__(self, config_path: Path = None):
        # Allow override, else respect XDG_CONFIG_HOME or default to ~/.config
        if config_path:
            self.config_path = config_path
        else:
            base = Path(os.environ.get("XDG_CONFIG_HOME", Path.home() / ".config"))
            self.config_path = base / "floww"
        self.config_file = self.config_path / "config.yaml"
        self.workflows_dir = self.config_path / "workflows"

    def init(self):
        """Create config directory, default config.yaml, and workflows directory."""
        self.config_path.mkdir(parents=True, exist_ok=True)
        if not self.config_file.exists():
            default_content = {"workspaces": {}}
            with open(self.config_file, "w") as f:
                yaml.dump(default_content, f)
        self.workflows_dir.mkdir(exist_ok=True)

    def list_workflows(self) -> list[str]:
        """Return list of workflow file names (without extension)."""
        return [p.stem for p in self.workflows_dir.glob("*.yaml")]

    def load_workflow(self, name: str) -> dict:
        """Load a workflow by name."""
        path = self.workflows_dir / f"{name}.yaml"
        with open(path) as f:
            return yaml.safe_load(f)
