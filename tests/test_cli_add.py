import pytest
from typer.testing import CliRunner

from floww.cli import app
from floww.config import ConfigManager


@pytest.fixture(autouse=True)
def reset_singleton():
    """Reset the ConfigManager singleton between tests."""
    original_instance = ConfigManager.instance
    ConfigManager.instance = None
    yield
    ConfigManager.instance = original_instance


@pytest.fixture(autouse=True)
def set_xdg_config_home(tmp_path, monkeypatch):
    # Redirect config home to a temporary directory
    config_home = tmp_path / "config"
    monkeypatch.setenv("XDG_CONFIG_HOME", str(config_home))
    return config_home


def test_add_new_workflow(set_xdg_config_home):
    """Test adding a new workflow."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Add a new workflow
    result = runner.invoke(
        app, ["add", "myflow"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0
    assert "Created new workflow: myflow (yaml)" in result.stdout

    # Verify file was created
    config_dir = set_xdg_config_home / "floww"
    workflow_file = config_dir / "workflows" / "myflow.yaml"
    assert workflow_file.exists()


def test_add_workflow_with_type(set_xdg_config_home):
    """Test adding a new workflow with a specific file type."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Add a new workflow with JSON format
    result = runner.invoke(
        app,
        ["add", "jsonflow", "--type", "json"],
        env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
    )
    assert result.exit_code == 0
    assert "Created new workflow: jsonflow (json)" in result.stdout

    # Verify file was created
    config_dir = set_xdg_config_home / "floww"
    workflow_file = config_dir / "workflows" / "jsonflow.json"
    assert workflow_file.exists()


def test_add_existing_workflow_same_extension(set_xdg_config_home):
    """Test adding a workflow that already exists with the same extension."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Create a workflow file manually
    config_dir = set_xdg_config_home / "floww"
    workflows_dir = config_dir / "workflows"
    workflows_dir.mkdir(parents=True, exist_ok=True)
    workflow_file = workflows_dir / "existing.yaml"
    workflow_file.write_text("dummy: 1")

    # Try to add a workflow with the same name
    result = runner.invoke(
        app, ["add", "existing"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 1
    assert (
        "Workflow 'existing' already exists with extension(s): .yaml" in result.stdout
    )


def test_add_existing_workflow_different_extension(set_xdg_config_home):
    """Test adding a workflow that already exists with a different extension."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Create a workflow file manually with JSON extension
    config_dir = set_xdg_config_home / "floww"
    workflows_dir = config_dir / "workflows"
    workflows_dir.mkdir(parents=True, exist_ok=True)
    workflow_file = workflows_dir / "crosstype.json"
    workflow_file.write_text('{"dummy": 1}')

    # Try to add a workflow with the same name but YAML extension
    result = runner.invoke(
        app,
        ["add", "crosstype", "--type", "yaml"],
        env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
    )
    assert result.exit_code == 1
    assert (
        "Workflow 'crosstype' already exists with extension(s): .json" in result.stdout
    )


def test_add_existing_workflow_multiple_extensions(set_xdg_config_home):
    """Test adding a workflow that already exists in multiple formats."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Create workflow files manually with different extensions
    config_dir = set_xdg_config_home / "floww"
    workflows_dir = config_dir / "workflows"
    workflows_dir.mkdir(parents=True, exist_ok=True)

    (workflows_dir / "multitype.json").write_text('{"dummy": 1}')
    (workflows_dir / "multitype.yaml").write_text("dummy: 1")

    # Try to add a workflow with the same name but TOML extension
    result = runner.invoke(
        app,
        ["add", "multitype", "--type", "toml"],
        env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
    )
    assert result.exit_code == 1
    assert "Workflow 'multitype' already exists with extension(s):" in result.stdout
    assert ".json" in result.stdout
    assert ".yaml" in result.stdout
