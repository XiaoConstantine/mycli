#!/bin/sh
set -e

RELEASES_URL="https://github.com/XiaoConstantine/mycli/releases"
BINARY_NAME="mycli"

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

VERSION=$(curl -sL ${RELEASES_URL}/latest | sed -n 's/.*tag\/\(.*\)".*/\1/p')
DOWNLOAD_URL="${RELEASES_URL}/download/${VERSION}/${BINARY_NAME}_${PLATFORM}_${ARCH}.tar.gz"

curl -L "$DOWNLOAD_URL" | tar xz -C /tmp
sudo mv /tmp/${BINARY_NAME} /usr/local/bin/
echo "Installation complete. You can now use '${BINARY_NAME}' command."
