import json
import os
import unittest
from unittest.mock import MagicMock, patch

from floww.core.workspace import WorkspaceManager, create_backend
from floww.core.workspace.backends import HyprlandBackend, WmctrlBackend


class TestWorkspaceManager(unittest.TestCase):
    def setUp(self):
        if "HYPRLAND_INSTANCE_SIGNATURE" in os.environ:
            del os.environ["HYPRLAND_INSTANCE_SIGNATURE"]
        self._cm_patcher = patch("floww.core.workspace.factory.ConfigManager")
        mock_cm_cls = self._cm_patcher.start()
        self.mock_config = MagicMock()
        self.mock_config.get_general_config.return_value = {
            "workspace_backend": "auto",
            "show_notifications": True,
        }
        mock_cm_cls.return_value = self.mock_config

    def tearDown(self):
        self._cm_patcher.stop()

    def test_auto_selects_wmctrl_when_no_hyprland_and_no_ewmh(self):
        with patch(
            "floww.core.workspace.factory.EwmhBackend.try_create", return_value=None
        ):
            wm = WorkspaceManager()
            self.assertIsInstance(wm._backend, WmctrlBackend)

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    def test_auto_selects_hyprland(self):
        wm = WorkspaceManager()
        self.assertIsInstance(wm._backend, HyprlandBackend)

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.backends.hyprland.run_command")
    def test_switch_hyprland_success(self, mock_run):
        mock_run.return_value = True
        wm = WorkspaceManager()
        result = wm.switch(2)
        self.assertTrue(result)
        mock_run.assert_called_once_with(["hyprctl", "dispatch", "workspace", "2"])

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.backends.hyprland.run_command")
    def test_switch_hyprland_failure(self, mock_run):
        mock_run.return_value = False
        wm = WorkspaceManager()
        self.assertFalse(wm.switch(2))

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.backends.hyprland.subprocess.run")
    def test_get_total_workspaces_hyprland(self, mock_sub_run):
        mock_response = MagicMock()
        mock_response.stdout = json.dumps(
            [{"id": 1, "name": "1"}, {"id": 3, "name": "3"}, {"id": 2, "name": "2"}]
        )
        mock_sub_run.return_value = mock_response

        wm = WorkspaceManager()
        total = wm.get_total_workspaces()
        self.assertEqual(total, 3)
        mock_sub_run.assert_called()

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.backends.hyprland.subprocess.run")
    def test_get_total_workspaces_hyprland_empty(self, mock_sub_run):
        mock_response = MagicMock()
        mock_response.stdout = json.dumps([])
        mock_sub_run.return_value = mock_response

        wm = WorkspaceManager()
        total = wm.get_total_workspaces()
        self.assertEqual(total, 1)

    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    def test_explicit_wmctrl_under_hyprland_env(self):
        wm = WorkspaceManager(backend=create_backend("wmctrl"))
        self.assertIsInstance(wm._backend, WmctrlBackend)


if __name__ == "__main__":
    unittest.main()
