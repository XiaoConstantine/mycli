#!/bin/zsh
set -e

GITHUB_REPO="XiaoConstantine/mycli"
BINARY_NAME="mycli"

detect_platform() {
    case "$(uname -s)" in
        Linux*)     echo "linux";;
        Darwin*)    echo "darwin";;
        *)          echo "unsupported";;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64*)    echo "x86_64";;
        arm64*)     echo "arm64";;
        *)          echo "unsupported";;
    esac
}

PLATFORM=$(detect_platform)
ARCH=$(detect_arch)

if [ "$PLATFORM" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
    echo "Unsupported platform or architecture. Please install manually."
    exit 1
fi

# Fetch the latest release version
VERSION=$(curl -s https://api.github.com/repos/${GITHUB_REPO}/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$VERSION" ]; then
    echo "Failed to fetch the latest version. Please check your internet connection or install manually."
    exit 1
fi
# Remove 'v' prefix from VERSION if present
VERSION=${VERSION#v}
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/v${VERSION}/${BINARY_NAME}_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"

echo "Downloading ${BINARY_NAME} version ${VERSION} for ${PLATFORM}_${ARCH}..."

# Create installation directory
INSTALL_DIR="${HOME}/.mycli/bin"
mkdir -p "$INSTALL_DIR"

# Download the file first
TEMP_FILE=$(mktemp)
if ! curl -L "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
    echo "Failed to download the binary. Please check the URL or install manually."
    rm -f "$TEMP_FILE"
    exit 1
fi

# Check if the downloaded file is empty
if [ ! -s "$TEMP_FILE" ]; then
    echo "Downloaded file is empty. Installation failed."
    rm -f "$TEMP_FILE"
    exit 1
fi

# Print the file type for debugging
file "$TEMP_FILE"

# Try to extract
if ! tar xzf "$TEMP_FILE" -C "$INSTALL_DIR"; then
    echo "Failed to extract the archive. The downloaded file may not be a valid tar.gz archive."
    rm -f "$TEMP_FILE"
    exit 1
fi

rm -f "$TEMP_FILE"

# Ensure the binary exists and is executable
if [ ! -x "$INSTALL_DIR/${BINARY_NAME}" ]; then
    echo "Binary not found or not executable. Installation failed."
    exit 1
fi

# Add to PATH if not already there
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo '\n# mycli installation' >> ~/.zshrc
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> ~/.zshrc
    echo "Updated ~/.zshrc with the new PATH. Please restart your terminal or run 'source ~/.zshrc' to apply changes."
else
    echo "$INSTALL_DIR is already in your PATH."
fi

echo "Installation complete. The ${BINARY_NAME} binary is located at: $INSTALL_DIR/${BINARY_NAME}"
echo "You can now use '${BINARY_NAME}' command. If it's not recognized, please restart your terminal or run 'source ~/.zshrc'."
