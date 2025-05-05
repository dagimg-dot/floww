from enum import Enum

DEFAULT_CONFIG = {
    "general": {"show_notifications": True},
    "timing": {
        "workspace_switch_wait": 3,  # Seconds to wait AFTER apps in a workspace before switching
        "app_launch_wait": 1,  # Default seconds to wait after launching each app (if not last)
        "respect_app_wait": True,  # Whether to use app-specific 'wait' values
    },
}

# Sample workflow content for different formats
SAMPLE_WORKFLOW_CONTENT = {
    "description": "An example workflow.",
    "workspaces": [
        {
            "target": 1,
            "apps": [
                {"name": "Terminal", "exec": "gnome-terminal"},
            ],
        },
        {
            "target": 2,
            "apps": [
                {
                    "name": "Browser",
                    "exec": "firefox",
                    "args": ["https://github.com/dagimg-dot/floww"],
                },
            ],
        },
    ],
}


class FileType(str, Enum):
    YAML = "yaml"
    JSON = "json"
    TOML = "toml"
