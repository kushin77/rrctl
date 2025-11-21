#!/usr/bin/env bash
# rrctl installer script
# Downloads and installs the latest rrctl binary for your platform

set -e

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
REPO_URL="https://github.com/kushin77/elevatedIQ/raw/main/rrctl-opensource"
VERSION="${VERSION:-latest}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${GREEN}==>${NC} $1"
}

warn() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

error() {
    echo -e "${RED}Error:${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        MINGW*|MSYS*|CYGWIN*) os="windows" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        aarch64|arm64)  arch="arm64" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac
    
    # Windows uses .exe extension
    local ext=""
    if [ "$os" = "windows" ]; then
        ext=".exe"
    fi
    
    # For now, only amd64 builds for Windows and Linux
    if [ "$os" = "windows" ] || [ "$os" = "linux" ]; then
        arch="amd64"
    fi
    
    echo "rrctl-${os}-${arch}${ext}"
}

# Download binary
download_binary() {
    local binary="$1"
    local url="${REPO_URL}/${binary}"
    local temp_file="/tmp/${binary}"
    
    info "Downloading rrctl for your platform..."
    info "URL: ${url}"
    
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "${temp_file}" "${url}" || error "Failed to download binary"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "${temp_file}" "${url}" || error "Failed to download binary"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
    
    echo "${temp_file}"
}

# Install binary
install_binary() {
    local temp_file="$1"
    local install_path="${INSTALL_DIR}/rrctl"
    
    # Create install directory if it doesn't exist
    mkdir -p "${INSTALL_DIR}"
    
    # Move binary to install location
    mv "${temp_file}" "${install_path}"
    chmod +x "${install_path}"
    
    info "Installed rrctl to ${install_path}"
}

# Check if install directory is in PATH
check_path() {
    if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
        warn "Installation directory ${INSTALL_DIR} is not in your PATH"
        info "Add it to your PATH by adding this line to your shell profile:"
        echo ""
        echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
        echo ""
    fi
}

main() {
    info "rrctl installer"
    echo ""
    
    # Detect platform
    local binary
    binary=$(detect_platform)
    info "Detected platform: ${binary}"
    
    # Download binary
    local temp_file
    temp_file=$(download_binary "${binary}")
    
    # Install binary
    install_binary "${temp_file}"
    
    # Check PATH
    check_path
    
    # Verify installation
    if command -v rrctl >/dev/null 2>&1; then
        info "Installation successful!"
        echo ""
        rrctl version
    else
        info "Installation complete!"
        warn "rrctl command not found. Make sure ${INSTALL_DIR} is in your PATH."
    fi
    
    echo ""
    info "Get started:"
    echo "    rrctl --help"
    echo "    rrctl repo-defrag --path /path/to/repo"
    echo "    rrctl repo-autofix --path /path/to/repo --dry-run"
}

main "$@"
