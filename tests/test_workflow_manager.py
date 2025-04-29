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
        mock_echo.assert_any_call("Workflow: Test workflow")


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

    with patch("floww.workflow_manager.typer.secho") as mock_secho:
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

        # Verify success message with color
        mock_secho.assert_any_call("✓ Workflow applied successfully", fg="green")


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

    with patch("floww.workflow_manager.typer.secho") as mock_secho:
        success = workflow_manager.apply(workflow_data)
        assert success is False

        # Verify error messages with correct prefix
        mock_secho.assert_any_call("    ✗ Failed to launch App1", fg="red")
        mock_secho.assert_any_call("    ✗ Failed to launch App2", fg="red")
        mock_secho.assert_any_call("⚠ Workflow completed with errors", fg="yellow")


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


def test_apply_workflow_with_wait(workflow_manager):
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1", "wait": 0.5},
                    {
                        "name": "App2",
                        "exec": "app2",
                        "wait": "invalid",
                    },  # Test invalid wait
                    {"name": "App3", "exec": "app3"},  # No wait after last app
                ],
            }
        ]
    }

    # Mock successful workspace switch and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Patch time.sleep and typer.echo/secho
    with (
        patch("floww.workflow_manager.time.sleep") as mock_sleep,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True  # Should succeed even with wait issues

        # Verify workspace switch
        workflow_manager.workspace_mgr.switch.assert_called_once_with(1)

        # Verify app launches
        assert workflow_manager.app_launcher.launch_app.call_count == 3
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App1", "exec": "app1", "wait": 0.5}
        )
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App2", "exec": "app2", "wait": "invalid"}
        )
        workflow_manager.app_launcher.launch_app.assert_any_call(
            {"name": "App3", "exec": "app3"}
        )

        # Verify time.sleep was called correctly:
        # 1. After App1 (0.5s from app-specific wait)
        # 2. After App2 (1s from default app_launch_wait)
        # 3. After workspace completion (3s from workspace_switch_wait)
        assert mock_sleep.call_count == 1  # Only App1's wait should be applied
        mock_sleep.assert_called_with(0.5)


def test_workflow_respects_workspace_switch_wait(workflow_manager):
    """Test that workflow manager respects workspace switch wait time."""
    workflow_data = {
        "workspaces": [
            {"target": 1, "apps": [{"name": "App1", "exec": "app1"}]},
            {"target": 2, "apps": [{"name": "App2", "exec": "app2"}]},
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 1.5,
        "app_launch_wait": 0,  # No wait after app launches
        "respect_app_wait": True,
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.workflow_manager.time.sleep") as mock_sleep,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify workspace switches happened
        assert workflow_manager.workspace_mgr.switch.call_count == 2

        # Verify time.sleep was called after first workspace only
        assert mock_sleep.call_count == 1
        mock_sleep.assert_called_with(1.5)


def test_workflow_respects_app_launch_wait(workflow_manager):
    """Test that workflow manager respects app launch wait time."""
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1"},
                    {"name": "App2", "exec": "app2"},
                ],
            },
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 0,  # No wait after workspace switch
        "app_launch_wait": 0.7,  # Wait after each app launch
        "respect_app_wait": True,
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.workflow_manager.time.sleep") as mock_sleep,
        patch("floww.workflow_manager.typer.echo") as mock_echo,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called after first app launch, not after last app
        assert mock_sleep.call_count == 1
        mock_sleep.assert_called_once_with(0.7)

        # Verify waiting message was displayed
        mock_echo.assert_any_call("    ... Waiting 0.7s before next action...")


def test_workflow_respects_app_specific_wait(workflow_manager):
    """Test that workflow manager respects app-specific wait time."""
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1", "wait": 1.2},  # App-specific wait
                    {"name": "App2", "exec": "app2"},  # No app-specific wait
                ],
            },
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 0,  # No wait after workspace switch
        "app_launch_wait": 0.5,  # Default wait after app launch
        "respect_app_wait": True,  # Respect app-specific wait
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.workflow_manager.time.sleep") as mock_sleep,
        patch("floww.workflow_manager.typer.echo") as mock_echo,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called with app-specific wait for App1
        # and default wait for App2
        assert mock_sleep.call_count == 1  # Only for App1, App2 is last app
        mock_sleep.assert_called_once_with(1.2)  # App-specific wait time

        # Verify waiting message was displayed with app-specific wait time
        mock_echo.assert_any_call("    ... Waiting 1.2s before next action...")


def test_workflow_disables_app_specific_wait(workflow_manager):
    """Test that workflow manager can ignore app-specific wait time."""
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {
                        "name": "App1",
                        "exec": "app1",
                        "wait": 1.2,
                    },  # App-specific wait (should be ignored)
                    {"name": "App2", "exec": "app2"},  # No app-specific wait
                ],
            },
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 0,  # No wait after workspace switch
        "app_launch_wait": 0.5,  # Default wait after app launch
        "respect_app_wait": False,  # IGNORE app-specific wait
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.workflow_manager.time.sleep") as mock_sleep,
        patch("floww.workflow_manager.typer.echo") as mock_echo,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called with config wait for App1, not app-specific
        assert mock_sleep.call_count == 1  # Only for App1, App2 is last app
        mock_sleep.assert_called_once_with(0.5)  # Config wait time

        # Verify waiting message was displayed with config wait time
        mock_echo.assert_any_call("    ... Waiting 0.5s before next action...")


def test_workflow_respects_last_app_wait(workflow_manager):
    """Test that workflow manager respects wait time from last app in a workspace."""
    workflow_data = {
        "workspaces": [
            {
                "target": 1,
                "apps": [
                    {"name": "App1", "exec": "app1"},
                    {"name": "App2", "exec": "app2", "wait": 2.5},  # Last app with wait
                ],
            },
            {"target": 2, "apps": [{"name": "App3", "exec": "app3"}]},
        ]
    }

    # Mock successful workspace switches and app launches
    workflow_manager.workspace_mgr.switch = MagicMock(return_value=True)
    workflow_manager.app_launcher.launch_app = MagicMock(return_value=True)

    # Mock the config manager to return custom timing settings
    mock_timing_config = {
        "workspace_switch_wait": 1.0,  # Wait after workspace switch
        "app_launch_wait": 0.5,  # Default wait after app launch
        "respect_app_wait": True,  # Respect app-specific wait
    }
    workflow_manager.config_mgr.get_timing_config = MagicMock(
        return_value=mock_timing_config
    )

    with (
        patch("floww.workflow_manager.time.sleep") as mock_sleep,
    ):
        success = workflow_manager.apply(workflow_data)
        assert success is True

        # Verify time.sleep was called for:
        # 1. After App1 (0.5s from app_launch_wait)
        # 2. After workspace 1 (2.5s from App2's wait)
        # 3. After workspace 2 (2.5s from App2's wait)
        assert mock_sleep.call_count == 3
        mock_sleep.assert_any_call(0.5)  # App1 launch wait
        mock_sleep.assert_any_call(2.5)  # App2's wait (used for both workspaces)
