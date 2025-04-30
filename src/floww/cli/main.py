import typer

from floww import __version__ as VERSION
from .helpers import LogLevel, setup_logging, FLOWW_ART

from .commands.init import init
from .commands.list import list_workflows
from .commands.add import add
from .commands.edit import edit
from .commands.remove import remove
from .commands.validate import validate
from .commands.apply import apply


app = typer.Typer(
    help="floww - your workflow automations in one place",
    context_settings={"help_option_names": ["-h", "--help"]},
)

app.command(
    name="init", help="Initialize configuration directory and workflows folder."
)(init)


app.command(name="list", help="List available workflows.")(list_workflows)
app.command(name="add", help="Create a new workflow file with basic structure.")(add)
app.command(name="edit", help="Open a workflow file in the default editor.")(edit)
app.command(name="remove", help="Remove a workflow file.")(remove)
app.command(name="validate", help="Validate a workflow's schema without applying it.")(
    validate
)
app.command(name="apply", help="Apply the named workflow.")(apply)


@app.callback(invoke_without_command=True)
def main_callback(
    ctx: typer.Context,
    log_level: str = LogLevel,
    version: bool = typer.Option(
        False, "--version", "-v", help="Show version", is_eager=True
    ),
):
    """Main entry point, sets up logging and handles version/help."""
    setup_logging(log_level)

    if version:
        typer.echo(f"floww version {VERSION}")
        raise typer.Exit()
    if ctx.invoked_subcommand is None:
        typer.echo(FLOWW_ART)
        typer.echo(ctx.get_help())
        raise typer.Exit()
