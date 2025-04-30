import pytest
from typer.testing import CliRunner

from floww.cli import app
from floww import ConfigManager


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


def test_validate_valid_workflow(set_xdg_config_home):
    """Test validating a valid workflow."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Create a valid workflow file
    config_dir = set_xdg_config_home / "floww"
    workflows_dir = config_dir / "workflows"
    workflows_dir.mkdir(parents=True, exist_ok=True)
    example_workflow = workflows_dir / "example.yaml"
    example_workflow.write_text("""
description: "Example workflow"
workspaces:
  - target: 1
    apps:
      - name: "Terminal"
        exec: "gnome-terminal"
""")

    # Print directory contents for debugging
    print(f"Contents of {workflows_dir}:")
    for path in workflows_dir.rglob("*"):
        print(f"  {path.relative_to(workflows_dir)}")

    # Validate the example workflow
    result = runner.invoke(
        app, ["validate", "example"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    print(f"Validation result: {result.stdout}")
    assert result.exit_code == 0
    assert "✓ Workflow is valid" in result.stdout


def test_validate_nonexistent_workflow(set_xdg_config_home):
    """Test validating a workflow that doesn't exist."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Try to validate a non-existent workflow
    result = runner.invoke(
        app,
        ["validate", "nonexistent"],
        env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
    )
    assert result.exit_code == 1
    assert "not found" in result.stdout


def test_validate_invalid_workflow(set_xdg_config_home):
    """Test validating an invalid workflow."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Create an invalid workflow file (missing apps key)
    config_dir = set_xdg_config_home / "floww"
    workflows_dir = config_dir / "workflows"
    workflows_dir.mkdir(parents=True, exist_ok=True)
    invalid_workflow = workflows_dir / "invalid.yaml"
    invalid_workflow.write_text("""
description: "Invalid workflow"
workspaces:
  - target: 1
""")

    # Print directory contents for debugging
    print(f"Contents of {workflows_dir}:")
    for path in workflows_dir.rglob("*"):
        print(f"  {path.relative_to(workflows_dir)}")
        with open(path, "r") as f:
            print(f"   Content: {f.read()}")

    # Try to validate the invalid workflow
    result = runner.invoke(
        app, ["validate", "invalid"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    print(f"Validation error result: {result.stdout}")
    assert result.exit_code == 1
    # The specific error message will depend on your validator implementation
    assert "Validation failed" in result.stdout
