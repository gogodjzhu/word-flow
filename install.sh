#!/bin/bash
set -e

REPO="gogodjzhu/word-flow"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TEMP_DIR=$(mktemp -d)
SUFFIX="tar.gz"

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

detect_os() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "darwin";;
        MINGW*|CYGWIN*|MSYS*) echo "windows";;
        *)          echo "unsupported";;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64";;
        aarch64|arm64) echo "arm64";;
        *)             echo "amd64";;
    esac
}

echo "Fetching latest release info..."
LATEST_JSON=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest")
TAG=$(echo "$LATEST_JSON" | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4)
VERSION=${TAG#v}

if [ -z "$TAG" ]; then
    echo "Error: Could not fetch latest release" >&2
    exit 1
fi

echo "Latest version: $VERSION"

OS=$(detect_os)
ARCH=$(detect_arch)

case "$OS" in
    windows) SUFFIX="zip"; FILENAME="wordflow-${OS}-${ARCH}.zip";;
    linux)   SUFFIX="tar.gz"; FILENAME="wordflow-${OS}-${ARCH}.tar.gz";;
    darwin)  SUFFIX="tar.gz"; FILENAME="wordflow-${OS}-${ARCH}.tar.gz";;
    *) echo "Error: Unsupported operating system: $OS" >&2; exit 1;;
esac
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${FILENAME}"

echo "Downloading ${FILENAME}..."
cd "$TEMP_DIR"
curl -fSL -o "$FILENAME" "$DOWNLOAD_URL"

echo "Verifying checksum..."
SHA256SUMS_URL="https://github.com/${REPO}/releases/download/${TAG}/SHA256SUMS"
curl -fSL -o "SHA256SUMS" "$SHA256SUMS_URL"
sha256sum -c SHA256SUMS --status 2>/dev/null || shasum -a 256 -c SHA256SUMS --status 2>/dev/null || {
    echo "Warning: Checksum verification failed, continuing anyway..." >&2
}

echo "Extracting..."
if [ "$SUFFIX" = "zip" ]; then
    unzip -o "$FILENAME"
    rm -f "$FILENAME"
    EXTRACTED_BIN="wordflow.exe"
else
    tar -xzf "$FILENAME"
    rm "$FILENAME"
    EXTRACTED_BIN="wordflow"
fi

echo "Installing to ${INSTALL_DIR}..."
if [ -d "$INSTALL_DIR" ] && [ ! -w "$INSTALL_DIR" ]; then
    echo "Error: $INSTALL_DIR is not writable, try running with sudo" >&2
    exit 1
fi

mv "$EXTRACTED_BIN" "$INSTALL_DIR/wordflow"
chmod +x "$INSTALL_DIR/wordflow"

echo ""
echo "Installation complete! wordflow installed to ${INSTALL_DIR}/wordflow"
echo "Run 'wordflow --version' to verify."