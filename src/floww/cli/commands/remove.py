import typer
from typing import Optional
import questionary

from floww.cli.helpers import (
    check_initialized,
    print_error,
    logger,
    get_workflow_name,
)
from floww import ConfigManager


def remove(
    name: Optional[str] = typer.Argument(None, help="Name of the workflow to remove"),
    force: bool = typer.Option(False, "--force", "-f", help="Skip confirmation prompt"),
):
    """Remove a workflow file."""
    check_initialized()
    cfg = ConfigManager()

    workflow_name = get_workflow_name(name, "remove", cfg)

    workflow_files = []
    for ext in cfg.config_loader.get_supported_formats():
        path = cfg.workflows_dir / f"{workflow_name}{ext}"
        if path.is_file():
            workflow_files.append(path)

    if not workflow_files:
        print_error(f"Workflow '{workflow_name}' not found")
        raise typer.Exit(1)

    if not force:
        confirm = questionary.confirm(
            f"Are you sure you want to remove workflow '{workflow_name}'?",
            default=False,
        ).ask()

        if not confirm:
            typer.echo("Operation cancelled")
            raise typer.Exit(0)

    try:
        for workflow_file in workflow_files:
            workflow_file.unlink()
            typer.echo(f"Removed workflow: {workflow_file.name}")
    except OSError as e:
        print_error(f"Failed to remove workflow file: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred while removing the workflow.")
        logger.exception(f"Unexpected error removing workflow '{workflow_name}': {e}")
        raise typer.Exit(1)
