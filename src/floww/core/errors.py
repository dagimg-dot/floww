class FlowwError(Exception):
    """Base exception class for floww application errors."""

    pass


class ConfigError(FlowwError):
    """Errors related to configuration loading or validation."""

    pass


class ConfigLoadError(ConfigError):
    """Errors related to loading configuration files."""

    pass


class WorkflowNotFoundError(ConfigError):
    """Specific error for when a workflow file is not found."""

    pass


class WorkflowSchemaError(ConfigError):
    """Errors related to invalid workflow file structure or content."""

    pass


class WorkspaceError(FlowwError):
    """Errors related to workspace management."""

    pass


class AppLaunchError(FlowwError):
    """Errors related to launching applications."""

    pass
