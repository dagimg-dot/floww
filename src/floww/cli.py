import typer
import questionary
import logging
from typing import Optional

from . import __version__ as VERSION
from .config import ConfigManager, WorkflowSchemaError
from .workflow_manager import WorkflowManager

FLOWW_ART = r"""

  /$$$$$$  /$$                                      
 /$$__  $$| $$                                      
| $$  \__/| $$  /$$$$$$  /$$  /$$  /$$ /$$  /$$  /$$
| $$$$    | $$ /$$__  $$| $$ | $$ | $$| $$ | $$ | $$
| $$_/    | $$| $$  \ $$| $$ | $$ | $$| $$ | $$ | $$
| $$      | $$| $$  | $$| $$ | $$ | $$| $$ | $$ | $$
| $$      | $$|  $$$$$$/|  $$$$$/$$$$/|  $$$$$/$$$$/
|__/      |__/ \______/  \_____/\___/  \_____/\___/ 
"""

app = typer.Typer(
    help="floww - your workflow automations in one place",
    context_settings={"help_option_names": ["-h", "--help"]},
)

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


@app.command()
def init(create_example: bool = typer.Option(False, "--example", "-e", help="Create an example workflow")):
    """Initialize configuration directory and workflows folder."""
    cfg = ConfigManager()
    cfg.init(create_example=create_example)
    typer.echo(f"Initialized config at {cfg.config_path}")


@app.command(name="list")
def list_workflows():
    """List available workflows."""
    cfg = ConfigManager()
    names = cfg.list_workflow_names()
    if not names:
        typer.echo("No workflows found")
    else:
        typer.echo("Available workflows:")
        for name in names:
            typer.echo(f"  - {name}")


@app.command()
def apply(name: Optional[str] = typer.Argument(None, help="Workflow name to apply")):
    """Apply the named workflow (or interactive chooser if no name provided)."""
    cfg = ConfigManager()
    workflow_mgr = WorkflowManager()

    # Get the workflow name [argument or interactive chooser]
    workflow_name = name
    if not workflow_name:
        # Interactive chooser
        available = cfg.list_workflow_names()
        if not available:
            typer.echo("No workflows found")
            raise typer.Exit(1)

        workflow_name = questionary.select(
            "Select a workflow to apply:",
            choices=available,
            use_arrow_keys=True,
            use_shortcuts=True,
            qmark="❯",
        ).ask()

        if not workflow_name:
            raise typer.Exit()

    try:
        typer.echo(f"Loading workflow: {workflow_name}")
        workflow_data = cfg.load_workflow(workflow_name)

        success = workflow_mgr.apply(workflow_data)
        if not success:
            typer.echo("Failed to apply workflow completely")
            raise typer.Exit(1)

    except FileNotFoundError as e:
        typer.echo(f"Error: {e}")
        raise typer.Exit(1)
    except WorkflowSchemaError as e:
        typer.echo(f"Invalid workflow format: {e}")
        raise typer.Exit(1)
    except Exception as e:
        typer.echo(f"Error applying workflow: {e}")
        logger.exception("Unexpected error")
        raise typer.Exit(1)


@app.callback(invoke_without_command=True)
def customer_option_callback(
    ctx: typer.Context,
    version: bool = typer.Option(False, "--version", "-v", help="Show version"),
):
    """Show version or help and exit."""
    if version:
        typer.echo(f"floww version {VERSION}")
        raise typer.Exit()
    if ctx.invoked_subcommand is None:
        typer.echo(FLOWW_ART)
        typer.echo(ctx.get_help())
        raise typer.Exit()
