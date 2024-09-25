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

BUILD_DIR="unmineable-solXEN"

# Clean up previous build artifacts
echo "Cleaning up previous build artifacts..."
rm -rf "$BUILD_DIR"
rm -f unmineable-solXEN-*.zip unmineable-solXEN-*.tar.gz

# Define build configurations
declare -A OS_ARCH=(
    ["linux"]="amd64"
    ["windows"]="amd64"
    ["darwin_amd64"]="darwin amd64"
    ["darwin_arm64"]="darwin arm64"
)

BINARY_NAME="unmineable-solXEN"

echo "Building version: $VERSION"

# Create build directory
mkdir -p "$BUILD_DIR"

# Build for each OS and architecture
for OS_ARCH_KEY in "${!OS_ARCH[@]}"; do
    IFS=' ' read -r OS ARCH <<< "${OS_ARCH[$OS_ARCH_KEY]}"
    
    # Handle cases where ARCH is empty (for linux and windows)
    if [ -z "$ARCH" ]; then
        ARCH=$OS
        OS=${OS_ARCH_KEY}
    fi
    
    echo "Building for $OS ($ARCH)..."
    
    if [ "$OS" == "windows" ]; then
        BINARY_NAME="unmineable-solXEN.exe"
    else
        BINARY_NAME="unmineable-solXEN"
    fi
    
    GOOS=$OS GOARCH=$ARCH go build -o "$BUILD_DIR/${BINARY_NAME}"
    
    if [ $? -ne 0 ]; then
        echo "Build failed for $OS $ARCH"
        exit 1
    fi
    
    if [ "$OS" == "windows" ]; then
        ARCHIVE_NAME="unmineable-solXEN-${VERSION}-${OS}-${ARCH}.zip"
        (cd "$BUILD_DIR/.." && zip -r "$ARCHIVE_NAME" "unmineable-solXEN")
    else
        ARCHIVE_NAME="unmineable-solXEN-${VERSION}-${OS}-${ARCH}.tar.gz"
        tar -czvf "$ARCHIVE_NAME" -C "$BUILD_DIR/.." "unmineable-solXEN"
    fi
    
    echo "Archive created: $ARCHIVE_NAME"
    
    # Clean up binary
    rm "$BUILD_DIR/${BINARY_NAME}"
done

echo "Build complete for all platforms."