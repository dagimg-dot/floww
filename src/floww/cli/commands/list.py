import typer

from floww.cli.helpers import check_initialized, print_error, logger
from floww import ConfigManager, ConfigError


def list_workflows():
    """List available workflows."""
    check_initialized()
    try:
        cfg = ConfigManager()
        names = cfg.list_workflow_names()
        if not names:
            typer.echo("No workflows found")
        else:
            typer.echo("Available workflows:")
            for name in names:
                typer.echo(f"  - {name}")
    except ConfigError as e:
        print_error(f"Failed to list workflows: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred while listing workflows.")
        logger.exception(f"Unexpected list error: {e}")
        raise typer.Exit(1)
