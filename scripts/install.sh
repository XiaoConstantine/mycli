#!/bin/sh
set -e

GITHUB_REPO="XiaoConstantine/mycli"
BINARY_NAME="mycli"
INSTALL_DIR="$HOME/.mycli/bin"


detect_platform() {
    case "$(uname -s)" in
        Linux*)     platform=linux;;
        Darwin*)    platform=darwin;;
        *)          platform="unsupported"
    esac
    echo $platform
}

detect_arch() {
    case "$(uname -m)" in
        x86_64*)    arch=x86_64;;
        arm64*)     arch=arm64;;
        *)          arch="unsupported"
    esac
    echo $arch
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

DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}_${PLATFORM}_${ARCH}.tar.gz"

echo "Downloading ${BINARY_NAME} version ${VERSION} for ${PLATFORM}_${ARCH}..."

mkdir -p "$INSTALL_DIR"

curl -L "$DOWNLOAD_URL" | tar xz -C "$INTALL_DIR"
if [ $? -ne 0 ]; then
    echo "Failed to download or extract the binary. Please check the URL or install manually."
    echo "Download URL: $DOWNLOAD_URL"
    exit 1
fi

chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Update PATH in .zshrc if not already present
if ! grep -q "$INSTALL_DIR" "$HOME/.zshrc"; then
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.zshrc"
    echo "Added $INSTALL_DIR to your PATH in .zshrc. Please restart your terminal or run 'source ~/.zshrc' to apply the changes."
fi

echo "Installation complete. You can now use '${BINARY_NAME}' command."
echo "If the command is not recognized, please restart your terminal or run 'source ~/.zshrc'."
