import unittest
from unittest.mock import patch, MagicMock
import os
import json
from floww.core.workspace import WorkspaceManager
from floww.core.errors import WorkspaceError


class TestWorkspaceManager(unittest.TestCase):
    def setUp(self):
        # Clear environment before each test
        if "HYPRLAND_INSTANCE_SIGNATURE" in os.environ:
            del os.environ["HYPRLAND_INSTANCE_SIGNATURE"]

    @patch("floww.core.workspace.EWMHLIB_AVAILABLE", False)
    @patch("floww.core.workspace.run_command")
    def test_init_no_hyprland_no_ewmh(self, mock_run):
        wm = WorkspaceManager()
        self.assertFalse(wm.is_hyprland)
        self.assertFalse(wm.use_ewmh)
        self.assertEqual(wm.wmctrl_cmd, "wmctrl")

    @patch("floww.core.workspace.EWMHLIB_AVAILABLE", False)
    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    def test_init_hyprland(self):
        wm = WorkspaceManager()
        self.assertTrue(wm.is_hyprland)
        self.assertFalse(wm.use_ewmh)
        self.assertEqual(wm.hyprctl_cmd, "hyprctl")
        self.assertIsNone(wm.wmctrl_cmd)

    @patch("floww.core.workspace.EWMHLIB_AVAILABLE", False)
    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.run_command")
    def test_switch_hyprland_success(self, mock_run):
        mock_run.return_value = True
        wm = WorkspaceManager()
        result = wm.switch(2)
        self.assertTrue(result)
        mock_run.assert_called_with(["hyprctl", "dispatch", "workspace", "2"])

    @patch("floww.core.workspace.EWMHLIB_AVAILABLE", False)
    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("floww.core.workspace.run_command")
    def test_switch_hyprland_failure(self, mock_run):
        mock_run.return_value = False
        wm = WorkspaceManager()
        with self.assertRaises(WorkspaceError):
            wm.switch(2)

    @patch("floww.core.workspace.EWMHLIB_AVAILABLE", False)
    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("subprocess.run")
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

    @patch("floww.core.workspace.EWMHLIB_AVAILABLE", False)
    @patch.dict(os.environ, {"HYPRLAND_INSTANCE_SIGNATURE": "test_sig"})
    @patch("subprocess.run")
    def test_get_total_workspaces_hyprland_empty(self, mock_sub_run):
        mock_response = MagicMock()
        mock_response.stdout = json.dumps([])
        mock_sub_run.return_value = mock_response

        wm = WorkspaceManager()
        total = wm.get_total_workspaces()
        self.assertEqual(total, 1)


if __name__ == "__main__":
    unittest.main()
