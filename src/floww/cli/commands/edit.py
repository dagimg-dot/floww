import typer
from typing import Optional

from floww.cli.helpers import (
    check_initialized,
    print_error,
    logger,
    get_workflow_name,
    open_in_editor,
)
from floww import ConfigManager

cfg = ConfigManager()


def edit(
    name: Optional[str] = typer.Argument(
        None,
        autocompletion=lambda: cfg.list_workflow_names(),
        help="Name of the workflow to edit",
    ),
):
    """Open a workflow file in the default editor."""
    check_initialized()

    workflow_name = get_workflow_name(name, "edit", cfg)

    workflow_files = []
    for ext in cfg.config_loader.get_supported_formats():
        path = cfg.workflows_dir / f"{workflow_name}{ext}"
        if path.is_file():
            workflow_files.append(path)

    if not workflow_files:
        print_error(f"Workflow '{workflow_name}' not found")
        raise typer.Exit(1)

    workflow_file = workflow_files[0]

    try:
        typer.echo(f"Opening workflow '{workflow_name}' in editor...")
        open_in_editor(workflow_file)
    except Exception as e:
        print_error("An unexpected error occurred while opening the editor.")
        logger.exception(f"Unexpected error editing workflow '{workflow_name}': {e}")
        raise typer.Exit(1)
