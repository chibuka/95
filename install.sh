#!/bin/sh
set -e

# 95 CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/chibuka/95-cli/main/install.sh | sh


REPO="chibuka/95-cli"
BINARY_NAME="95"
INSTALL_DIR="$HOME/.local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Detect OS and architecture
detect_platform() {
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$os" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        *)
            printf "${RED}error: unsupported operating system: %s${NC}\n" "$os"
            exit 1
            ;;
    esac

    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            printf "${RED}error: unsupported architecture: %s${NC}\n" "$arch"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
}

# Get latest release version
get_latest_version() {
    printf "${CYAN}→ fetching latest release...${NC}\n"

    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    fi

    if [ -z "$VERSION" ]; then
        printf "${RED}error: failed to fetch latest version${NC}\n"
        exit 1
    fi

    printf "${GREEN}✓ latest version: %s${NC}\n" "$VERSION"
}

install_binary() {
    download_url="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}-${PLATFORM}"
    tmp_file="/tmp/${BINARY_NAME}"

    printf "${CYAN}→ downloading %s for %s...${NC}\n" "$BINARY_NAME" "$PLATFORM"

    if ! curl -fsSL "$download_url" -o "$tmp_file"; then
        printf "${RED}error: failed to download binary${NC}\n"
        printf "${YELLOW}url: %s${NC}\n" "$download_url"
        exit 1
    fi

    # Make binary executable
    chmod +x "$tmp_file"

    # Create install directory if it doesn't exist
    mkdir -p "$INSTALL_DIR"

    # Move binary to install directory
    mv "$tmp_file" "$INSTALL_DIR/$BINARY_NAME"

    printf "${GREEN}✓ installed to %s/%s${NC}\n" "$INSTALL_DIR" "$BINARY_NAME"
}

# Check if install directory is in PATH
check_path() {
    case ":$PATH:" in
        *":$INSTALL_DIR:"*)
            ;;
        *)
            echo ""
            printf "${YELLOW}warning: %s is not in your PATH${NC}\n" "$INSTALL_DIR"
            echo ""
            echo "add this line to your shell config file (~/.bashrc, ~/.zshrc, etc.):"
            echo ""
            printf "${CYAN}  export PATH=\"\$PATH:%s\"${NC}\n" "$INSTALL_DIR"
            echo ""
            ;;
    esac
}

print_success() {
    echo ""
    printf "${GREEN}installation complete!${NC}\n"
    echo ""
    echo "run your first command:"
    printf "${CYAN}  %s${NC}\n" "$BINARY_NAME"
    echo ""
}

main() {
    echo ""

    printf "${YELLOW}"
    cat << "EOF"
________   ___  ________   _______   ________ ___  ___      ___ _______
|\   ___  \|\  \|\   ___  \|\  ___ \ |\  _____\\  \|\  \    /  /|\  ___ \
\ \  \\ \  \ \  \ \  \\ \  \ \   __/|\ \  \__/\ \  \ \  \  /  / | \   __/|
\ \  \\ \  \ \  \ \  \\ \  \ \  \_|/_\ \   __\\ \  \ \  \/  / / \ \  \_|/__
 \ \  \\ \  \ \  \ \  \\ \  \ \  \_|\ \ \  \_| \ \  \ \    / /   \ \  \_|\ \
  \ \__\\ \__\ \__\ \__\\ \__\ \_______\ \__\   \ \__\ \__/ /     \ \_______\
   \|__| \|__|\|__|\|__| \|__|\|_______|\|__|    \|__|\|__|/       \|_______|
EOF
    printf "${NC}"
    echo ""

    echo "ninefive cli installer"
    echo ""

    detect_platform
    get_latest_version
    install_binary
    check_path
    print_success
}

main
