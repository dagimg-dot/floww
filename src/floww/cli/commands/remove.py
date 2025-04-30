import typer
from typing import Optional, List
import questionary

from floww.cli.helpers import (
    check_initialized,
    print_error,
    logger,
    get_workflow_name,
)
from floww import ConfigManager

cfg = ConfigManager()


def remove(
    names: Optional[List[str]] = typer.Argument(
        None,
        autocompletion=lambda: cfg.list_workflow_names(),
        help="Name(s) of the workflow(s) to remove",
    ),
    force: bool = typer.Option(False, "--force", "-f", help="Skip confirmation prompt"),
):
    """Remove a workflow file."""
    check_initialized()

    names = names or []

    if not names:
        selected = get_workflow_name(None, "remove", cfg)
        names = [selected]

    # Collect all files to remove
    all_files = []
    for workflow_name in names:
        found = []
        for ext in cfg.config_loader.get_supported_formats():
            path = cfg.workflows_dir / f"{workflow_name}{ext}"
            if path.is_file():
                found.append(path)
        if not found:
            print_error(f"Workflow '{workflow_name}' not found")
            raise typer.Exit(1)
        all_files.extend(found)

    if not force:
        if len(names) > 1:
            prompt = f"Are you sure you want to remove workflows {', '.join(names)}?"
        else:
            prompt = f"Are you sure you want to remove workflow '{names[0]}'?"
        confirm = questionary.confirm(prompt, default=False).ask()
        if not confirm:
            typer.echo("Operation cancelled")
            raise typer.Exit(0)

    try:
        for workflow_file in all_files:
            workflow_file.unlink()
            typer.echo(f"Removed workflow: {workflow_file.name}")
    except OSError as e:
        print_error(f"Failed to remove workflow file: {e}")
        raise typer.Exit(1)
    except Exception as e:
        print_error("An unexpected error occurred while removing the workflow.")
        logger.exception(f"Unexpected error removing workflow '{names[0]}': {e}")
        raise typer.Exit(1)
