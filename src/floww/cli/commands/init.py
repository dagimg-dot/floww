import typer

from floww.cli.helpers import print_error, logger
from floww import ConfigManager, ConfigError
from floww.utils import FileType


def init(
    create_example: bool = typer.Option(
        False, "--example", "-e", help="Create an example workflow"
    ),
    file_type: FileType = typer.Option(
        FileType.YAML, "--type", "-t", help="File format for the example workflow"
    ),
):
    """Initialize configuration directory and workflows folder."""
    try:
        cfg = ConfigManager()
        cfg.init(create_example=create_example, file_type=file_type.value)
        typer.echo(f"Initialized config at {cfg.config_path}")
    except ConfigError as e:
        print_error(f"Initialization failed: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred during initialization.")
        logger.exception(f"Unexpected init error: {e}")
        raise typer.Exit(1)
