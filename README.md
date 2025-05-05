# floww

[![GitHub release](https://img.shields.io/github/v/release/dagimg-dot/floww)](https://github.com/dagimg-dot/floww/releases/latest)
[![License](https://img.shields.io/github/license/dagimg-dot/floww)](LICENSE)
[![Downloads](https://img.shields.io/github/downloads/dagimg-dot/floww/total)](https://github.com/dagimg-dot/floww/releases)


```text
  /$$$$$$  /$$
 /$$__  $$| $$
| $$  \__/| $$  /$$$$$$  /$$  /$$  /$$ /$$  /$$  /$$
| $$$$    | $$ /$$__  $$| $$ | $$ | $$| $$ | $$ | $$
| $$_/    | $$| $$  \ $$| $$ | $$ | $$| $$ | $$ | $$
| $$      | $$| $$  | $$| $$ | $$ | $$| $$ | $$ | $$
| $$      | $$|  $$$$$$/|  $$$$$/$$$$/|  $$$$$/$$$$/
|__/      |__/ \______/  \_____/\___/  \_____/\___/
```

**floww** is a command-line utility designed to streamline your workflow setup on Linux desktops. Define your desired workspace layouts and application sets in simple configuration files currently supporting ***yaml***, ***json*** and ***toml***, and let `floww` automate the process of switching workspaces and launching applications.

![showcase](https://res.cloudinary.com/drrfofxv2/image/upload/v1746433774/portfolio/floww-1746433696-1.gif)

## Features

*   **Workflows:** Define your workflows in a configuration file of your choice. (currently supporting ***yaml***, ***json*** and ***toml***)
*   **Workspace Management:** Automatically switch to specified virtual desktops/workspaces.
*   **Application Launching:** Launch binaries, Flatpaks, and Snaps with arguments.
*   **Flexible Timing:** Configure wait times between actions for smoother transitions.
*   **Interactive Mode:** Select workflows easily if you don't specify one.
*   **Validation:** Check your workflow files for correctness before applying them.
*   **Simple CLI:** Manage workflows with intuitive commands (`init`, `list`, `apply`, `add`, `edit`, `remove`, `validate`).

## Prerequisites

Before installing `floww`, ensure you have the following dependencies:

1.  **Workspace Switching Backend:** `floww` needs a tool to interact with your window manager/desktop environment. It prioritizes `ewmhlib` (Python library, which is already packaged in `floww`) and falls back to `wmctrl` (command-line tool). If the default `ewmhlib` fails, `floww` will fall back to `wmctrl`.
    *   **`wmctrl`:** A common command-line tool, often needed as a fallback or if `ewmhlib` encounters issues (especially on Wayland setups where EWMH support might be incomplete).
    *   **Installation (wmctrl):**
        *   Debian/Ubuntu: `sudo apt update && sudo apt install wmctrl`
        *   Fedora: `sudo dnf install wmctrl`
        *   Arch Linux: `sudo pacman -S wmctrl`
2.  **Notification Daemon (Optional but Recommended):** `floww` uses `notify-send` to display completion or error messages.
    *   **Installation (notify-send):**
        *   Debian/Ubuntu: `sudo apt update && sudo apt install libnotify-bin`
        *   Fedora: `sudo dnf install libnotify`
        *   Arch Linux: `sudo pacman -S libnotify`
3.  **Text Editor (for `edit` command):** The `floww edit` command uses the editor specified in your `$EDITOR` environment variable. If not set, it tries common editors like `vim`, `vi`, or `nano`. Set `$EDITOR` for the best experience: `export EDITOR=vim` (add this to your shell configuration, e.g., `.bashrc` or `.zshrc`).

## Installation and Update

1. **Using script**

```bash
curl -fsSL https://raw.githubusercontent.com/dagimg-dot/floww/refs/heads/main/scripts/install.sh | sh
```

2. **Using [eget](https://github.com/zyedidia/eget)**
  
```bash
eget dagimg-dot/floww
```

## Building from Source

It's recommended to install `floww` in a virtual environment. [`uv`](https://github.com/astral-sh/uv) is recommended for this.

1.  **Clone the repository (if you haven't already):**
    ```bash
    git clone https://github.com/dagimg-dot/floww.git 
    cd floww
    ```

2.  **Create and activate a virtual environment (using `uv`):**
    ```bash
    uv venv
    source .venv/bin/activate
    ```

3.  **Install `floww`:**
    ```bash
    make install
    ```

4.  **Verify installation:**
    ```bash
    floww --version
    ```

## Usage

`floww` provides several commands to manage and apply your workflows.

> **Note:** 
> - All options have short and long forms. like `-e` and `--edit`.
> - The commands `apply`, `remove`, `edit` and `validate` have autocompletion for workflow names. Checkout `floww --show-completion` and `floww --install-completion` for more information.

1.  **Initialize Configuration:**
    Run this first to create the necessary configuration directory (`~/.config/floww`) and workflows subdirectory.
    ```bash
    floww init
    ```
    *   Use `floww init --example` to also create a sample `example.yaml` workflow file.
    *   Use `floww init --example --type <type>` to create a sample workflow file of the specified type.

2.  **List Available Workflows:**
    See the names of all workflows defined in `~/.config/floww/workflows/`.
    ```bash
    floww list
    ```

3.  **Add a New Workflow:**
    Create a new, basic workflow file.
    ```bash
    floww add <workflow-name>
    ```
    *   Example: `floww add coding` creates `~/.config/floww/workflows/coding.yaml`.
    *   Use `floww add <workflow-name> --edit` or `-e` to open the new file in your default editor immediately after creation.
    *   Use `floww add <workflow-name> --type <type>` or `-t <type>` to create a new workflow file of the specified type. (currently supporting ***yaml***, ***json*** and ***toml***)

4.  **Edit an Existing Workflow:**
    Open a workflow file in your default editor.
    ```bash
    floww edit <workflow-name>
    ```
    *   If `<workflow-name>` is omitted, you'll get an interactive list to choose from.
    *   Example: `floww edit coding`

5.  **Validate a Workflow:**
    Check if a workflow file has the correct structure and syntax without applying it. Useful for debugging.
    ```bash
    floww validate <workflow-name>
    ```
    *   If `<workflow-name>` is omitted, you'll get an interactive list to choose from.
    *   Example: `floww validate coding`

6.  **Apply a Workflow:**
    Execute the steps defined in a workflow: switch workspaces and launch applications.
    ```bash
    floww apply <workflow-name>
    ```
    *   If `<workflow-name>` is omitted, `floww` will present an interactive list of available workflows to choose from.
    *   Example: `floww apply coding`
    *   Use `floww apply --file <file-path>` to apply a workflow from a file path.
    *   Use `floww apply --append` to start the workflow from the last workspace.

7.  **Remove a Workflow:**
    Delete a workflow file.
    ```bash
    floww remove <workflow-name-1> <workflow-name-2> ...
    ```
    *   If `<workflow-name>` is omitted, you'll get an interactive list to choose from.
    *   By default, it asks for confirmation. Use `--force` or `-f` to skip confirmation.
    *   Example: `floww remove old-project --force`

8.  **Get Help:**
    ```bash
    floww --help
    floww <command> --help # e.g., floww apply --help
    ```

9.  **Control Logging:**
    Use the `--log-level` or `-l` option with any command to set verbosity (DEBUG, INFO, WARNING, ERROR, CRITICAL). Default is WARNING.
    ```bash
    floww --log-level DEBUG apply my-workflow
    ```

## Configuration

`floww` uses a configuration directory typically located at `~/.config/floww` (following the XDG Base Directory Specification).

### Global Configuration (`config.yaml`)

*   **Location:** `~/.config/floww/config.yaml`
*   **Purpose:** Holds global settings for `floww`. Currently, primarily used for timing delays during workflow application.
*   **Format:** YAML
*   **Default Structure (if file is empty or missing):**
    ```yaml
    general:
      show_notifications: true # Whether to show notifications
    timing:
      workspace_switch_wait: 3.0  # Seconds to wait AFTER launching all apps in a workspace before switching to the next, unless overridden by a specific app's 'wait'.
      app_launch_wait: 1.0      # Default seconds to wait AFTER launching an app, IF it's NOT the last app in its workspace list AND doesn't have its own 'wait' value.
      respect_app_wait: true   # If true, use app-specific 'wait' values when present. If false, ignore app 'wait' values and always use 'app_launch_wait' (or 0 for the last app).
    ```
*   You can create this file and customize these values. Invalid values (e.g., negative numbers for waits, non-booleans for `respect_app_wait`) will cause `floww` to revert to the default for that specific setting.

### Workflows (`workflows/`)

*   **Location:** `~/.config/floww/workflows/`
*   **Purpose:** Contains individual workflow definitions.
*   **Format:** Each workflow is a separate file named `<workflow-name>.(yaml|json|toml)`.

#### Workflow File Structure (YAML)

Each workflow configuration file defines the sequence of actions:

## Example Workflow (`~/.config/floww/workflows/web-dev.yaml`)

```yaml
description: "Standard Web Development Setup"

workspaces:
  - target: 0 # Workspace 1 (index 0)
    apps:
      - name: "VS Code (Project)"
        exec: "code"
        type: "binary"
        args: ["~/dev/my-web-project"]
        wait: 2.0 # Give VS Code time to load

      - name: "Project Terminal"
        exec: "gnome-terminal"
        args: ["--working-directory=~/dev/my-web-project"]

  - target: 1 # Workspace 2 (index 1)
    apps:
      - name: "Firefox (Local Dev)"
        exec: "org.mozilla.firefox" # Assuming Firefox installed via Flatpak
        type: "flatpak"
        args: ["http://localhost:3000"]
        wait: 1.5

      - name: "Firefox (Docs)"
        exec: "org.mozilla.firefox"
        type: "flatpak"
        args: ["--new-window", "https://developer.mozilla.org/"]

  - target: 3 # Workspace 4 (index 3)
    apps:
      - name: "Communication (Slack)"
        exec: "slack" # Assuming Slack installed via Snap
        type: "snap"

# After setting up workspaces 0, 1, and 3, switch back to workspace 0.
final_workspace: 0
```

> **Note:** The workflow file can be `yaml`, `json` or `toml`. However, it must be valid for the type it is. 

**Key Fields Explained:**

*   `description` (Optional, String): Human-readable description.
*   `workspaces` (Required, List): A list of workspace objects to process sequentially.
    *   `target` (Required, Integer >= 0): The zero-indexed workspace number to switch to.
    *   `apps` (Required, List): A list of application objects to launch on this workspace.
        *   `name` (Required, String): A name for the application (used in output messages).
        *   `exec` (Required, String): The command, executable path, Flatpak App ID, or Snap name. Tilde (`~`) is expanded to the user's home directory.
        *   `type` (Optional, String): Specifies how to launch the app. Defaults to `binary`.
            *   `binary`: Executes the `exec` value directly (e.g., `/usr/bin/firefox`, `code`, `~/scripts/my_tool`).
            *   `flatpak`: Runs `flatpak run <exec> [args...]`. `exec` should be the Flatpak Application ID (e.g., `org.mozilla.firefox`).
            *   `snap`: Runs `<exec> [args...]`. `exec` should be the snap command name (e.g., `spotify`).
        *   `args` (Optional, List): A list of command-line arguments to pass to the application. Arguments are converted to strings. Tilde (`~`) is expanded within string arguments.
        *   `wait` (Optional, Float >= 0): Seconds to wait after launching *this specific application*. If `respect_app_wait` is `true` in `config.yaml`, this value is used instead of the global `app_launch_wait`. If this is the *last app* in the *last workspace*, this wait occurs *before* switching to the `final_workspace` (if defined). If this is the *last app* in an *intermediate* workspace, this wait occurs *before* switching to the *next* workspace (overriding `workspace_switch_wait`).

*   `final_workspace` (Optional, Integer >= 0): After processing all workspace definitions in the `workspaces` list, switch to this workspace number.

**To apply this example:**

1.  Save the content above as `~/.config/floww/workflows/web-dev.yaml`.
2.  Run: `floww apply web-dev`

## Contributing

Contributions are welcome! Please follow these steps:

1.  **Fork** the repository on GitHub.
2.  **Clone** your fork locally: `git clone <your-fork-url>`
3.  Create a **new branch** for your feature or fix: `git checkout -b feature/my-new-feature` or `git checkout -b fix/bug-description`.
4.  Make your **code changes**.
5.  **Add tests** for your changes in the `tests/` directory.
6.  **Run tests:** Ensure all tests pass using `make test` (this uses `pytest`).
7.  **Lint your code:** Ensure code style consistency using `make lint` (this uses `ruff`).
8.  **Commit** your changes with a clear message: `git commit -m "Add feature: description"`
9.  **Push** your branch to your fork: `git push origin feature/my-new-feature`
10. Create a **Pull Request** on the original repository's `main` branch.

Please also consider opening an issue first to discuss significant changes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Common Errors & Troubleshooting

*   **Error: floww is not initialized. Please run 'floww init' first.**
    *   **Cause:** The configuration directory (`~/.config/floww`) hasn't been created.
    *   **Solution:** Run `floww init`.

*   **Error: Workflow `name` not found...**
    *   **Cause:** The specified workflow file (`<name>.yaml`) doesn't exist in `~/.config/floww/workflows/`.
    *   **Solution:** Check the spelling of the workflow name. Use `floww list` to see available workflows. Ensure the file is in the correct directory.

*   **Error: Validation failed: <schema error message>**
    *   **Cause:** The workflow file has incorrect syntax or structure (e.g., missing required keys like `target` or `apps`, incorrect data types).
    *   **Solution:** Run `floww validate <name>` to get specific details about the schema violation. Compare your workflow file structure against the [Configuration](#configuration) section and the [Example Workflow](#example-workflow). If it's a YAML file, check YAML indentation carefully.

*   **Error launching <App Name>: Command not found...**
    *   **Cause:** The executable specified in the `exec` field cannot be found.
    *   **Solution:**
        *   For `type: binary`: Ensure the command is spelled correctly and is available in your system's `$PATH`, or provide the full path to the executable (e.g., `/usr/local/bin/mytool`, `~/scripts/run_dev.sh`). Ensure the file has execute permissions (`chmod +x <file>`).
        *   For `type: flatpak`: Ensure the Flatpak Application ID is correct and the application is installed (`flatpak list`).
        *   For `type: snap`: Ensure the snap command name is correct and the snap is installed (`snap list`).

*   **Error launching <App Name>: Permission denied...**
    *   **Cause:** You don't have permission to execute the file specified in `exec` (usually for `type: binary`).
    *   **Solution:** Check the file permissions. You might need to add execute permissions: `chmod +x /path/to/executable`.

*   **Workspace switching doesn't work or switches incorrectly:**
    *   **Cause:** Issues with `ewmhlib` or `wmctrl`, or incompatibility with your specific desktop environment/window manager (especially on Wayland).
    *   **Solution:**
        *   Ensure `wmctrl` is installed (see [Prerequisites](#prerequisites)). `floww` should fall back to it if `ewmhlib` fails.
        *   Check if your desktop environment fully supports EWMH (Extended Window Manager Hints), especially if using Wayland. Some Wayland compositors have limited or no support for external workspace control via these methods.
        *   Run `floww apply <name> --log-level DEBUG` and check the logs for specific errors related to `EwmhRoot` or `wmctrl`.
        *   Consult your desktop environment's documentation regarding external workspace control.

*   **Notifications ("Workflow applied successfully", "Workflow completed with errors") don't appear:**
    *   **Cause:** `notify-send` command is not available or the notification daemon isn't running.
    *   **Solution:** Install `libnotify-bin` or equivalent package for your distribution (see [Prerequisites](#prerequisites)). Ensure your desktop environment's notification service is running.
*   **Final workspace is not being applied:**
    *   **Cause:** The last app in the last workspace might be taking too long to launch.
    *   **Solution:** Add a `wait` value to the last app in the last workspace.
    *   **Note:** If the last app in the last workspace has a `wait` value, the final workspace will be applied after the wait time.
    *   **Cause:** You are using `toml` as the workflow file type and final_workspace is at the end of the file.
    *   **Solution:** Bring the `final_workspace` definition up above the worskpaces section
