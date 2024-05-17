#!/bin/bash

# Project directory (replace with your project's directory)
PROJECT_DIR=$(pwd)

# Output directory for binaries
OUTPUT_DIR="${PROJECT_DIR}/bin"

# List of target OS and architecture combinations
TARGETS=(
    "windows/386"
    "windows/amd64"
    "linux/386"
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

# Ensure output directory exists
mkdir -p "${OUTPUT_DIR}"

# Build for each target
for TARGET in "${TARGETS[@]}"; do
    OS=$(echo "${TARGET}" | cut -d'/' -f1)
    ARCH=$(echo "${TARGET}" | cut -d'/' -f2)
    OUTPUT_NAME="${OUTPUT_DIR}/confman-${OS}-${ARCH}"

    if [ "$OS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi

    echo "Building for ${OS}/${ARCH}..."
    env GOOS="${OS}" GOARCH="${ARCH}" go build -o "${OUTPUT_NAME}" ./cmd/confman 

    if [ $? -ne 0 ]; then
        echo "Failed to build for ${OS}/${ARCH}"
        exit 1
    fi
done

echo "Build completed. Binaries are located in the ${OUTPUT_DIR} directory."

