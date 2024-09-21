#!/bin/bash

set -e

# Default version
VERSION="0.0.0"

# Parse command line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -v|--version) VERSION="$2"; shift ;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

BUILD_DIR="xoon"

# Clean up previous build artifacts
echo "Cleaning up previous build artifacts..."
rm -rf "$BUILD_DIR"
rm -f xoon-*.zip xoon-*.tar.gz

# Define build configurations
declare -A OS_ARCH=(
    ["linux"]="amd64"
    ["windows"]="amd64"
)

BINARY_NAME="xoon"

echo "Building version: $VERSION"

# Create build directory
mkdir -p "$BUILD_DIR"

# Build for each OS and architecture
for OS in "${!OS_ARCH[@]}"; do
    ARCH=${OS_ARCH[$OS]}
    echo "Building for $OS ($ARCH)..."
    
    if [ "$OS" == "windows" ]; then
        BINARY_NAME="xoon.exe"
    else
        BINARY_NAME="xoon"
    fi
    
    GOOS=$OS GOARCH=$ARCH go build -o "$BUILD_DIR/${BINARY_NAME}"
    
    if [ $? -ne 0 ]; then
        echo "Build failed for $OS"
        exit 1
    fi
    
    if [ "$OS" == "windows" ]; then
        ARCHIVE_NAME="xoon-${VERSION}-${OS}-${ARCH}.zip"
        (cd "$BUILD_DIR" && zip "../$ARCHIVE_NAME" "${BINARY_NAME}")
    else
        ARCHIVE_NAME="xoon-${VERSION}-${OS}-${ARCH}.tar.gz"
        tar -czvf "$ARCHIVE_NAME" -C "$BUILD_DIR" "${BINARY_NAME}"
    fi
    
    echo "Archive created: $ARCHIVE_NAME"
    
    # Clean up binary
    rm "$BUILD_DIR/${BINARY_NAME}"
done

echo "Build complete for all platforms."