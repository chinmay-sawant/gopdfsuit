#!/bin/bash
set -e

# Directory context
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

TYPST_VERSION="v0.11.0"
TYPST_DIR="typst-x86_64-unknown-linux-musl"
TYPST_ARCHIVE="typst.tar.xz"
TYPST_URL="https://github.com/typst/typst/releases/download/${TYPST_VERSION}/${TYPST_ARCHIVE}"

# 1. Download and Install Typst if not present
if [ ! -f "$TYPST_DIR/typst" ]; then
    echo "⬇️  Downloading Typst ($TYPST_VERSION)..."
    curl -L -o "$TYPST_ARCHIVE" "$TYPST_URL"
    
    echo "📦 Extracting Typst..."
    tar -xf "$TYPST_ARCHIVE"
    rm "$TYPST_ARCHIVE"
    
    chmod +x "$TYPST_DIR/typst"
    echo "✅ Typst installed locally in $TYPST_DIR"
else
    echo "✅ Typst already installed."
fi

# 2. Generate Benchmark Data
echo "📊 Generating dataset (data.json)..."
go run gen_data.go

echo "🎉 Setup complete! You can now run the benchmarks using:"
echo "   go test -bench=Benchmark -benchmem -v ./internal/pdf"
