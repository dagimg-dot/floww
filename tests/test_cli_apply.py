import pytest
from typer.testing import CliRunner
from unittest.mock import patch, MagicMock, ANY

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


@pytest.fixture
def setup_workflow_file(set_xdg_config_home):
    """Set up a workflow file for testing."""
    config_dir = set_xdg_config_home / "floww"
    workflows_dir = config_dir / "workflows"
    workflows_dir.mkdir(parents=True, exist_ok=True)

    # Create a valid workflow file
    example_workflow = workflows_dir / "example.yaml"
    example_workflow.write_text("""
description: "Example workflow"
workspaces:
  - target: 1
    apps:
      - name: "Terminal"
        exec: "gnome-terminal"
""")

    return example_workflow


@pytest.fixture
def setup_external_file(tmp_path):
    """Set up an external workflow file for testing."""
    external_file = tmp_path / "external_workflow.yaml"
    external_file.write_text("""
description: "External workflow"
workspaces:
  - target: 2
    apps:
      - name: "Browser"
        exec: "firefox"
""")

    return external_file


def test_apply_workflow_by_name(set_xdg_config_home, setup_workflow_file):
    """Test applying a workflow by name with fully mocked dependencies."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Set up patches to prevent real application execution
    with patch("floww.cli.commands.apply.WorkflowManager") as wm_mock:
        # Configure mocks
        wm_instance = MagicMock()
        wm_mock.return_value = wm_instance
        wm_instance.apply.return_value = True

        # Set up patched logger to capture output
        with patch("floww.cli.commands.apply.logger") as logger_mock:
            # Apply the workflow by name
            result = runner.invoke(
                app,
                ["apply", "example"],
                env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
            )

            assert result.exit_code == 0

            # Verify logger calls
            logger_mock.info.assert_any_call("Loading workflow: example")
            logger_mock.info.assert_any_call("Applying workflow: example")

            # Verify workflow manager apply was called
            wm_instance.apply.assert_called_once()


def test_apply_workflow_from_file(set_xdg_config_home, setup_external_file):
    """Test applying a workflow directly from a file with fully mocked dependencies."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Set up patches to prevent real application execution
    with patch("floww.cli.commands.apply.WorkflowManager") as wm_mock:
        # Configure mocks
        wm_instance = MagicMock()
        wm_mock.return_value = wm_instance
        wm_instance.apply.return_value = True

        # Set up patched logger to capture output
        with patch("floww.cli.commands.apply.logger") as logger_mock:
            # Apply the workflow from file
            result = runner.invoke(
                app,
                ["apply", "--file", str(setup_external_file)],
                env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
            )

            assert result.exit_code == 0

            # Verify logger calls
            logger_mock.info.assert_any_call(
                f"Loading workflow from file: {setup_external_file}"
            )
            logger_mock.info.assert_any_call(
                f"Applying workflow: {setup_external_file}"
            )

            # Verify workflow manager apply was called
            wm_instance.apply.assert_called_once()


def test_apply_nonexistent_file(set_xdg_config_home, tmp_path):
    """Test applying a workflow from a nonexistent file."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Define a nonexistent file path
    nonexistent_file = tmp_path / "nonexistent.yaml"

    # Set up patches to prevent real application execution
    with patch("floww.cli.commands.apply.WorkflowManager") as wm_mock:
        # Configure mocks
        wm_instance = MagicMock()
        wm_mock.return_value = wm_instance

        # Try to apply a workflow from a nonexistent file
        result = runner.invoke(
            app,
            ["apply", "--file", str(nonexistent_file)],
            env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
        )

        assert result.exit_code == 1
        assert "Workflow file not found" in result.stdout

        # Verify workflow manager apply was NOT called
        wm_instance.apply.assert_not_called()


def test_apply_workflow_failure(set_xdg_config_home, setup_workflow_file):
    """Test handling when workflow application fails."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Set up patches to prevent real application execution
    with (
        patch("floww.cli.commands.apply.WorkflowManager") as wm_mock,
        patch("floww.cli.commands.apply.print_error") as print_error_mock,
    ):
        # Configure mocks to simulate workspace error
        wm_instance = MagicMock()
        wm_mock.return_value = wm_instance

        # Use WorkspaceError to match the exception handling in apply.py
        from floww import WorkspaceError

        wm_instance.apply.side_effect = WorkspaceError("Workflow application failed")

        # Apply the workflow
        result = runner.invoke(
            app, ["apply", "example"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
        )

        # Now it should exit with code 1 due to the WorkspaceError
        assert result.exit_code == 1

        # Verify error was printed
        print_error_mock.assert_called_once()

        # Verify workflow manager apply was called (and raised an exception)
        wm_instance.apply.assert_called_once()


def test_apply_workflow_append_flag(set_xdg_config_home, setup_workflow_file):
    """Test applying a workflow with the --append flag."""
    runner = CliRunner()

    # Initialize config
    result = runner.invoke(
        app, ["init"], env={"XDG_CONFIG_HOME": str(set_xdg_config_home)}
    )
    assert result.exit_code == 0

    # Set up patches to prevent real application execution
    with patch("floww.cli.commands.apply.WorkflowManager") as wm_mock:
        # Configure mocks
        wm_instance = MagicMock()
        wm_mock.return_value = wm_instance
        wm_instance.apply.return_value = True  # Simulate successful application

        # Apply the workflow with the --append flag
        result = runner.invoke(
            app,
            ["apply", "example", "--append"],
            env={"XDG_CONFIG_HOME": str(set_xdg_config_home)},
        )

        # Assert the command exited successfully
        assert result.exit_code == 0

        # Assert that the WorkflowManager's apply method was called
        # with the append argument set to True.
        # We use ANY for the first argument (workflow_data) as we are
        # primarily interested in the value of the 'append' flag here.
        wm_instance.apply.assert_called_once_with(ANY, True)
