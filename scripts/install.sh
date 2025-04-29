#!/usr/bin/env bash
# Downloads the latest floww release from https://github.com/dagimg-dot/floww/releases
# and installs it into ~/.local/bin.

set -euo pipefail

APP_NAME="floww"
REPO="dagimg-dot/floww"
INSTALL_BASE_DIR="$HOME/.local"
INSTALL_DIR="$INSTALL_BASE_DIR/$APP_NAME"
BIN_DIR="$INSTALL_BASE_DIR/bin"
LOG_FILE="/tmp/${APP_NAME}-installer.log"
PLATFORM="linux"
ARCH="x86_64"
EXT="tar.gz"

DEPENDENCIES=(curl tar)

logger() {
    local type="$1"
    shift
    local message="$*"
    local timestamp
    timestamp=$(date +"%Y/%m/%d %H:%M:%S")
    local color_reset='\033[0m'
    local color_red='\033[0;31m'
    local color_blue='\033[0;34m'

    # Send logger output to stderr instead of stdout
    if [[ "$type" == "error" ]]; then
        echo -e "${timestamp} -- ${APP_NAME}-Installer [Error]: ${color_red}${message}${color_reset}" >&2
    elif [[ "$type" == "info" ]]; then
        echo -e "${timestamp} -- ${APP_NAME}-Installer [Info]: ${color_blue}${message}${color_reset}" >&2
    else
        echo -e "${timestamp} -- ${APP_NAME}-Installer [Log]: ${message}" >&2
    fi
    # Append to log file separately
    echo -e "${timestamp} -- ${APP_NAME}-Installer [$type]: ${message}" >>"${LOG_FILE}"
}

remove_if_exists() {
    local target="$1"
    if [[ -z "$target" ]]; then
        logger error "No target specified in remove_if_exists function"
        return 1
    fi
    if [[ -e "$target" || -L "$target" ]]; then
        logger info "Removing '$target'..."
        rm -rf "$target"
        logger info "'$target' removed."
    else
        logger info "'$target' does not exist. Skipping removal."
    fi
}

check_dependencies() {
    logger info "Checking required dependencies: ${DEPENDENCIES[*]}"
    local missing_deps=()

    for pkg in "${DEPENDENCIES[@]}"; do
        if ! command -v "$pkg" >/dev/null 2>&1; then
            logger info "'$pkg' not found."
            missing_deps+=("$pkg")
        else
            logger info "'$pkg' is installed."
        fi
    done

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        logger error "Missing required dependencies: ${missing_deps[*]}"
        logger error "Please install them using your distribution's package manager and run this script again."
        logger error "Example commands for common distributions:"
        logger error "  Debian/Ubuntu: sudo apt install ${missing_deps[*]}"
        logger error "  Fedora/RHEL: sudo dnf install ${missing_deps[*]}"
        logger error "  Arch Linux: sudo pacman -S ${missing_deps[*]}"
        exit 1
    else
        logger info "All dependencies are satisfied."
    fi
}

get_latest_version() {
    local release_url="https://api.github.com/repos/${REPO}/releases/latest"
    logger info "Fetching latest version information from $release_url"

    # Capture the curl output separately from logging
    local api_response
    api_response=$(curl -fsS "${release_url}") || {
        logger error "Failed to fetch release information"
        exit 1
    }

    local latest_version
    latest_version=$(echo "$api_response" | grep '"tag_name":' | sed -E 's/.*"tag_name": ?"([^"]+)".*/\1/')

    if [[ -z "$latest_version" ]]; then
        logger error "Could not parse version from GitHub API response"
        exit 1
    fi

    logger info "Latest version found: $latest_version"
    printf "%s" "$latest_version" # Use printf to avoid newline and prevent log contamination
}

get_asset_info() {
    local version="$1"
    local version_no_v="${version#v}"

    ASSET_NAME="${APP_NAME}-${version_no_v}-${PLATFORM}-${ARCH}.${EXT}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${version}/${ASSET_NAME}"

    logger info "Asset Name: $ASSET_NAME"
    logger info "Download URL: $DOWNLOAD_URL"
}

delete_old_version() {
    logger info "Checking for existing installation of $APP_NAME..."

    local binary_to_check
    if [[ -L "$BIN_DIR/$APP_NAME" ]]; then
        binary_to_check=$(readlink -f "$BIN_DIR/$APP_NAME")
        logger info "Found link: $BIN_DIR/$APP_NAME -> $binary_to_check"
    elif [[ -f "$INSTALL_DIR/$APP_NAME" ]]; then
        binary_to_check="$INSTALL_DIR/$APP_NAME"
        logger info "Found binary directly in: $binary_to_check"
    else
        binary_to_check=$(command -v "$APP_NAME") || true
    fi

    remove_if_exists "$INSTALL_DIR"
    remove_if_exists "$BIN_DIR/$APP_NAME"

    logger info "Old version removal complete."
}

download_archive() {
    local tmp_file="/tmp/$ASSET_NAME"
    logger info "Downloading $APP_NAME ($ASSET_NAME)..."
    remove_if_exists "$tmp_file"

    if curl --progress-bar -fSL -o "$tmp_file" "${DOWNLOAD_URL}"; then
        logger info "Download finished successfully: $tmp_file"
    else
        logger error "Download failed! Check URL, network, and permissions."
        logger error "URL: ${DOWNLOAD_URL}"
        remove_if_exists "$tmp_file"
        exit 1
    fi
    echo "$tmp_file"
}

install_app() {
    local archive_path="$1"
    local actual_binary_path="$INSTALL_DIR/$APP_NAME"

    logger info "Installing $APP_NAME from $archive_path..."

    mkdir -p "$INSTALL_DIR" "$BIN_DIR"

    logger info "Extracting archive to $INSTALL_DIR..."
    if tar -xzf "$archive_path" -C "$INSTALL_DIR"; then
        logger info "Archive extracted successfully."
    else
        logger error "Failed to extract archive '$archive_path'."
        # Clean up potentially partially extracted files
        remove_if_exists "$INSTALL_DIR"
        exit 1
    fi

    if [[ ! -f "$actual_binary_path" ]]; then
        logger error "Binary '$APP_NAME' not found in the expected location after extraction: $actual_binary_path"
        logger error "Please check the archive structure. Expected the binary to be at the root of the archive or directly named '$APP_NAME'."
        local found_binary
        found_binary=$(find "$INSTALL_DIR" -type f -executable -name "$APP_NAME" | head -n 1)
        if [[ -n "$found_binary" ]]; then
            logger info "Found potential binary at: $found_binary"
            logger error "Installer needs adjustment for this archive structure."
            remove_if_exists "$INSTALL_DIR"
            exit 1
        else
            logger error "Could not locate the '$APP_NAME' binary within the extracted files."
            remove_if_exists "$INSTALL_DIR"
            exit 1
        fi
    fi

    local link_path="$BIN_DIR/$APP_NAME"
    logger info "Creating symbolic link: $link_path -> $actual_binary_path"
    ln -sf "$actual_binary_path" "$link_path"

    logger info "$APP_NAME installed successfully to $actual_binary_path"
    logger info "Symlink created at $link_path"

    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        logger error "Warning: '$HOME/.local/bin' is not found in your PATH."
        logger error "You may need to add it to your shell configuration (e.g., ~/.bashrc, ~/.zshrc)"
        logger error "Example: export PATH=\"\$HOME/.local/bin:\$PATH\""
        logger error "Please open a new terminal or source your configuration file after adding it."
    else
        logger info "'$HOME/.local/bin' found in PATH."
    fi

    logger info "Run '$APP_NAME --help' to get started."
    logger info "Installation logs saved in: ${LOG_FILE}"
}

check_if_installed() {
    local installed_version=""
    if command -v "$APP_NAME" >/dev/null 2>&1; then
        logger info "'$APP_NAME' command found in PATH."
        # Capture version output separately from logging
        local version_output
        version_output=$($APP_NAME --version 2>/dev/null) || {
            logger error "Failed to get version information"
            printf ""
            return
        }
        installed_version=$(echo "$version_output" | awk '{print $3}')
    fi

    if [[ -n "$installed_version" ]]; then
        logger info "Detected installed version: v$installed_version"
        printf "%s" "$installed_version"
    else
        logger info "'$APP_NAME' not found or version command failed."
        printf ""
    fi
}

update_app() {
    local installed_version_no_v="$1"
    local latest_version="$2"
    local latest_version_no_v="${latest_version#v}"

    logger info "Checking for updates..."
    logger info "Installed version: v$installed_version_no_v"
    logger info "Latest available version: $latest_version"

    if [[ "$installed_version_no_v" != "$latest_version_no_v" ]]; then
        logger info "Newer version ($latest_version) available. Updating..."

        get_asset_info "$latest_version"
        local downloaded_archive
        downloaded_archive=$(download_archive)

        delete_old_version

        install_app "$downloaded_archive"
        logger info "$APP_NAME updated successfully to $latest_version."
    else
        logger info "You already have the latest version (v$installed_version_no_v) installed."
        exit 0
    fi
}

main() {
    echo "" >"$LOG_FILE"
    logger info "Starting $APP_NAME installation script..."

    check_dependencies

    local latest_version
    latest_version=$(get_latest_version)

    local installed_version
    installed_version=$(check_if_installed)

    if [[ -n "$installed_version" ]]; then
        logger info "$APP_NAME v$installed_version is currently installed."
        update_app "$installed_version" "$latest_version"
    else
        logger info "$APP_NAME is not currently installed."
        logger info "Proceeding with fresh installation of version $latest_version."

        get_asset_info "$latest_version"

        local downloaded_archive
        downloaded_archive=$(download_archive)

        install_app "$downloaded_archive"
    fi

    # Clean up temporary files after successful installation
    remove_if_exists "$downloaded_archive"

    logger info "$APP_NAME installation/update process finished."
}

main "$@"
