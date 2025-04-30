import pytest
from typer.testing import CliRunner
from floww.cli import app
from floww import ConfigManager


@pytest.fixture(autouse=True)
def reset_singleton():
    """Reset the ConfigManager singleton between tests."""
    original = ConfigManager.instance
    ConfigManager.instance = None
    yield
    ConfigManager.instance = original


@pytest.fixture(autouse=True)
def set_xdg_config_home(tmp_path, monkeypatch):
    # Redirect XDG_CONFIG_HOME to a temp dir
    config_home = tmp_path / "config"
    monkeypatch.setenv("XDG_CONFIG_HOME", str(config_home))
    return config_home


def test_remove_single_workflow(set_xdg_config_home):
    runner = CliRunner()
    # Initialize config
    result = runner.invoke(app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)})
    assert result.exit_code == 0

    # Create one workflow file
    cfg = ConfigManager()
    workflows_dir = cfg.workflows_dir
    file_path = workflows_dir / "demo.yaml"
    workflows_dir.mkdir(parents=True, exist_ok=True)
    file_path.write_text("dummy: 1")
    assert file_path.exists()

    # Force-remove the workflow
    result = runner.invoke(
        app,
        ["remove", "demo", "--force"],
        env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
    )
    assert result.exit_code == 0
    assert f"Removed workflow: {file_path.name}" in result.stdout
    assert not file_path.exists()


def test_remove_multiple_workflows(set_xdg_config_home):
    runner = CliRunner()
    # Initialize config
    result = runner.invoke(app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)})
    assert result.exit_code == 0

    # Create two workflow files
    cfg = ConfigManager()
    workflows_dir = cfg.workflows_dir
    workflows_dir.mkdir(parents=True, exist_ok=True)
    file1 = workflows_dir / "foo.yaml"
    file2 = workflows_dir / "bar.yaml"
    file1.write_text("dummy: foo")
    file2.write_text("dummy: bar")
    assert file1.exists() and file2.exists()

    # Force-remove both
    result = runner.invoke(
        app,
        ["remove", "foo", "bar", "--force"],
        env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
    )
    assert result.exit_code == 0
    out = result.stdout
    assert "Removed workflow: foo.yaml" in out
    assert "Removed workflow: bar.yaml" in out
    assert not file1.exists()
    assert not file2.exists()


def test_remove_nonexistent_workflow(set_xdg_config_home):
    runner = CliRunner()
    # Initialize config
    result = runner.invoke(app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)})
    assert result.exit_code == 0

    # Attempt to remove a workflow that doesn't exist
    result = runner.invoke(
        app,
        ["remove", "nope", "--force"],
        env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
    )
    assert result.exit_code == 1
    assert "Workflow 'nope' not found" in result.stdout 