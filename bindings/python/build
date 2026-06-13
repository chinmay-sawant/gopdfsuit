#!/bin/bash

# Build script for gopdfsuit Python bindings shared library
# This builds the CGO shared library for the current platform

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/pypdfsuit/lib"

# Ensure output directory exists
mkdir -p "$OUTPUT_DIR"

# Detect OS and set output filename
OS=$(uname -s)
case "$OS" in
    Linux)
        OUTPUT_FILE="libgopdfsuit.so"
        ;;
    Darwin)
        OUTPUT_FILE="libgopdfsuit.dylib"
        ;;
    MINGW*|CYGWIN*|MSYS*)
        OUTPUT_FILE="gopdfsuit.dll"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

echo "Building shared library for $OS..."
echo "Output: $OUTPUT_DIR/$OUTPUT_FILE"

cd "$PROJECT_ROOT"

# Build the shared library
CGO_ENABLED=1 go build \
    -buildmode=c-shared \
    -o "$OUTPUT_DIR/$OUTPUT_FILE" \
    ./bindings/python/cgo/

# Keep generated headers in sync with exports for developers and CI diffs.
rm -f "$OUTPUT_DIR/gopdfsuit.h"

echo "Build complete: $OUTPUT_DIR/$OUTPUT_FILE"
echo ""
echo "Library info:"
file "$OUTPUT_DIR/$OUTPUT_FILE"
