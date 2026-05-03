import os
from pathlib import Path

from .core import ConfigManager
from .core import ConfigLoader
from .core import WorkflowManager
from .core import WorkspaceManager
from .core import AppLauncher
from .core import (
    ConfigError,
    ConfigLoadError,
    WorkflowNotFoundError,
    WorkflowSchemaError,
    WorkspaceError,
    AppLaunchError,
)

__version__ = "0.3.2"


def _iter_dotenv_search_dirs() -> list[Path]:
    """Directories to check for a `.env` file (cwd chain, then package / repo parents)."""
    seen: set[Path] = set()
    ordered: list[Path] = []
    cwd = Path.cwd().resolve()
    pkg = Path(__file__).resolve().parent
    for d in (
        cwd,
        *list(cwd.parents)[:8],
        pkg,
        *list(pkg.parents)[:8],
    ):
        if d not in seen:
            seen.add(d)
            ordered.append(d)
    return ordered


def _first_dotenv_path() -> Path | None:
    for root in _iter_dotenv_search_dirs():
        candidate = root / ".env"
        if candidate.is_file():
            return candidate
    return None


def _read_env_value_from_dotenv(key: str = "ENV") -> str:
    path = _first_dotenv_path()
    if path is None:
        return ""
    try:
        text = path.read_text(encoding="utf-8")
    except OSError:
        return ""
    for raw_line in text.splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            continue
        name, _, value = line.partition("=")
        if name.strip() != key:
            continue
        v = value.strip()
        if len(v) >= 2 and v[0] == v[-1] and v[0] in "\"'":
            v = v[1:-1]
        return v.strip()
    return ""


def version_display() -> str:
    """User-visible version (e.g. for ``--version``). Optional suffix from env or ``.env``."""
    extra = os.environ.get("FLOWW_VERSION_SUFFIX", "").strip()
    if extra:
        return __version__ + extra

    env_label = os.environ.get("ENV", "").strip()
    if not env_label:
        env_label = _read_env_value_from_dotenv("ENV").strip()
    if env_label:
        suffix = env_label.lstrip("@")
        return f"{__version__}@{suffix}"

    if os.environ.get("FLOWW_DEV", "").lower() in ("1", "true", "yes", "on"):
        return __version__ + "@dev"
    return __version__


__all__ = [
    "version_display",
    "__version__",
    "ConfigManager",
    "ConfigLoader",
    "WorkflowManager",
    "WorkspaceManager",
    "AppLauncher",
    "ConfigError",
    "ConfigLoadError",
    "WorkflowNotFoundError",
    "WorkflowSchemaError",
    "WorkspaceError",
    "AppLaunchError",
]
