import json
import yaml
from pathlib import Path
import toml

from .errors import ConfigLoadError


class ConfigLoader:
    """A utility class for loading and saving configuration files in various formats."""

    def __init__(self):
        self._loaders = {
            ".toml": self._load_toml,
            ".yaml": self._load_yaml,
            ".yml": self._load_yaml,
            ".json": self._load_json,
        }

        self._savers = {
            ".toml": self._save_toml,
            ".yaml": self._save_yaml,
            ".yml": self._save_yaml,
            ".json": self._save_json,
        }

    def load(self, file_path: str | Path) -> dict:
        """
        Load configuration from a file based on its extension.

        Args:
            file_path: Path to the configuration file

        Returns:
            Dictionary containing the loaded configuration

        Raises:
            ConfigLoadError: If the file format is unsupported or the file cannot be loaded
        """
        path = Path(file_path)

        if not path.exists():
            raise ConfigLoadError(f"Configuration file not found: {file_path}")

        extension = path.suffix.lower()
        if extension not in self._loaders:
            raise ConfigLoadError(f"Unsupported configuration format: {extension}")

        try:
            return self._loaders[extension](path)
        except Exception as e:
            raise ConfigLoadError(f"Failed to load {extension} file: {e}")

    def save(self, data: dict, file_path: str | Path) -> None:
        """
        Save configuration data to a file based on its extension.

        Args:
            data: Dictionary containing the configuration data
            file_path: Path to the configuration file

        Raises:
            ConfigLoadError: If the file format is unsupported or the file cannot be saved
        """
        path = Path(file_path)

        extension = path.suffix.lower()
        if extension not in self._savers:
            raise ConfigLoadError(
                f"Unsupported configuration format for saving: {extension}"
            )

        try:
            self._savers[extension](data, path)
        except Exception as e:
            raise ConfigLoadError(f"Failed to save {extension} file: {e}")

    def _load_toml(self, file_path: Path) -> dict:
        """Load configuration from a TOML file."""
        try:
            return toml.load(file_path)
        except Exception as e:
            raise ConfigLoadError(f"Error parsing TOML file {file_path}: {e}")

    def _save_toml(self, data: dict, file_path: Path) -> None:
        """Save configuration to a TOML file."""
        try:
            with open(file_path, "w") as f:
                toml.dump(data, f)
        except Exception as e:
            raise ConfigLoadError(f"Error saving TOML file {file_path}: {e}")

    def _load_yaml(self, file_path: Path) -> dict:
        """Load configuration from a YAML file."""
        try:
            with open(file_path, "r") as f:
                data = yaml.safe_load(f)
                return data if data is not None else {}
        except Exception as e:
            raise ConfigLoadError(f"Error parsing YAML file {file_path}: {e}")

    def _save_yaml(self, data: dict, file_path: Path) -> None:
        """Save configuration to a YAML file."""
        try:
            with open(file_path, "w") as f:
                yaml.dump(data, f, default_flow_style=False, sort_keys=False)
        except Exception as e:
            raise ConfigLoadError(f"Error saving YAML file {file_path}: {e}")

    def _load_json(self, file_path: Path) -> dict:
        """Load configuration from a JSON file."""
        try:
            with open(file_path, "r") as f:
                return json.load(f)
        except Exception as e:
            raise ConfigLoadError(f"Error parsing JSON file {file_path}: {e}")

    def _save_json(self, data: dict, file_path: Path) -> None:
        """Save configuration to a JSON file."""
        try:
            with open(file_path, "w") as f:
                json.dump(data, f, indent=2)
        except Exception as e:
            raise ConfigLoadError(f"Error saving JSON file {file_path}: {e}")

    def get_supported_formats(self) -> list[str]:
        """Return a list of supported file extensions."""
        return list(self._loaders.keys())

    def is_supported_format(self, file_path: str | Path) -> bool:
        """Check if the given file has a supported format."""
        extension = Path(file_path).suffix.lower()
        return extension in self._loaders
