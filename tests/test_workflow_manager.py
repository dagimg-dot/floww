import pytest
from unittest.mock import patch, MagicMock

from floww.workflow_manager import WorkflowManager


@pytest.fixture
def workflow_manager():
    return WorkflowManager()


def test_apply_empty_workflow(workflow_manager):
    workflow_data = {"workspaces": []}

    success = workflow_manager.apply(workflow_data)
    assert success is True


def test_apply_workflow_with_description(workflow_manager):
    workflow_data = {"description": "Test workflow", "workspaces": []}

    with patch("floww.workflow_manager.typer.echo") as mock_echo:
        success = workflow_manager.apply(workflow_data)
        assert success is True
        mock_echo.assert_any_call("Workflow description: Test workflow")


def test_apply_workflow_switch_workspace_failure(workflow_manager):
    workflow_data = {"workspaces": [{"target": 1, "apps": []}]}

    # Mock the workspace manager to fail switching
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=False)

    success = workflow_manager.apply(workflow_data)
    assert success is False
    workflow_manager.workspace_mgr.switch.assert_called_once_with(1)


def test_apply_workflow_launch_apps(workflow_manager):
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1"},
                    {"name": "App2", "exec": "app2", "args": ["--flag"]},
                ],
            }
        ]
    }

    # Mock successful workspace switch
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    # Mock successful app launches
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    with patch("floww.workflow_manager.typer.echo") as mock_echo:
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify workspace was switched
        workflow_manager.workspace_mgr.switch.assert_called_once_with(1)

        # Verify apps were launched
        assert workflow_manager.app_launcher.launch_app.call_count == 2
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App1", "exec": "app1"}
        )
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App2", "exec": "app2", "args": ["--flag"]}
        )

        # Verify user feedback
        mock_echo.assert_any_call("Switching to workspace 1...")
        mock_echo.assert_any_call("Launching App1...")
        mock_echo.assert_any_call("Launching App2...")
        mock_echo.assert_any_call("Workflow applied successfully")


def test_apply_workflow_app_launch_failure(workflow_manager):
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1"},
                    {"name": "App2", "exec": "app2"},
                ],
            }
        ]
    }

    # Mock successful workspace switch
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    # Mock app launch failure
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=False)

    with patch("floww.workflow_manager.typer.echo") as mock_echo:
        success = workflow_manager.apply(workflow_data)
        # Overall workflow should still succeed even if app launch fails
        assert success is True

        # Verify error messages
        mock_echo.assert_any_call("Failed to launch App1")
        mock_echo.assert_any_call("Failed to launch App2")


def test_apply_workflow_multiple_workspaces(workflow_manager):
    workflow_data = {
        "workspaces": [
            {"target": 1, "apps": [{"name": "App1", "exec": "app1"}]},
            {"target": 2, "apps": [{"name": "App2", "exec": "app2"}]},
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    success = workflow_manager.apply(workflow_data)
    assert success is True

    # Verify workspace switches
    assert workflow_manager.workspace_mgr.switch.call_count == 2
    workflow_manager.workspace_mgr.switch.assert_any_call(1)
    workflow_manager.workspace_mgr.switch.assert_any_call(2)

    # Verify app launches
    assert workflow_manager.app_launcher.launch_app.call_count == 2
    workflow_manager.app_launcher.launch_app.assert_any_call(
        {"name": "App1", "exec": "app1"}
    )
    workflow_manager.app_launcher.launch_app.assert_any_call(
        {"name": "App2", "exec": "app2"}
    )
