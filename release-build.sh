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
    ["linux_amd64"]="linux amd64"
    ["linux_arm64"]="linux arm64"
    ["windows"]="windows amd64"
    ["darwin_amd64"]="darwin amd64"
    ["darwin_arm64"]="darwin arm64"
)

BINARY_NAME="unmineable-solXEN"

echo "Building version: $VERSION"

# Create build directory
mkdir -p "$BUILD_DIR"

# Function to build with appropriate flags
build() {
    local OS=$1
    local ARCH=$2
    local OUTPUT=$3

    if [ "$OS" = "windows" ]; then
        # Windows doesn't support full static linking
        GOOS=$OS GOARCH=$ARCH go build -o "$OUTPUT"
    elif [ "$OS" = "darwin" ]; then
        # macOS: use partial static linking
        GOOS=$OS GOARCH=$ARCH go build -ldflags="-extldflags=-static-libgcc" -o "$OUTPUT"
    else
        # Linux and others: use full static linking
        CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build -ldflags='-w -s' -tags netgo -installsuffix netgo -o "$OUTPUT"
    fi
}

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
    
    build $OS $ARCH "$BUILD_DIR/${BINARY_NAME}"
    
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