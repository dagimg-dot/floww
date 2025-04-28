# floww

CLI tool to manage Linux workspaces and launch applications based on YAML workflows.

## Installation

Install with uv in a venv:

```bash
uv install .
```

## Usage

```bash
floww init          # scaffold config and workflows directories
floww list          # list available workflows
floww apply <name>  # apply given workflow (or interactive chooser)
floww --version     # show version
```

Run `floww --help` or `floww <command> --help` for more details.

## Configuration

- Global config file: `~/.config/floww/config.yaml`
- Workflows directory: `~/.config/floww/workflows/`

Define your workflows in individual `.yaml` files under the workflows directory.

## License

This project is licensed under the MIT License. See `LICENSE` for details.
