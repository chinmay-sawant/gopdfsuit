#!/bin/bash

# k6 Installation Script for Ubuntu WSL2
# This script installs k6 load testing tool

set -e

echo "========================================"
echo "k6 Installation Script for Ubuntu WSL2"
echo "========================================"

# Check if k6 is already installed
if command -v k6 &> /dev/null; then
    echo "k6 is already installed!"
    k6 version
    exit 0
fi

echo "Installing k6..."

# Method 1: Using apt (Recommended)
echo "Adding k6 GPG key and repository..."

# Install required packages
sudo apt-get update
sudo apt-get install -y gnupg software-properties-common

# Add k6 GPG key
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69

# Add k6 repository
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list

# Update and install k6
sudo apt-get update
sudo apt-get install -y k6

# Verify installation
echo ""
echo "========================================"
echo "k6 Installation Complete!"
echo "========================================"
k6 version

echo ""
echo "Quick test commands:"
echo "  k6 run test/smoke_test.js"
echo "  k6 run test/load_test.js"
echo "  k6 run --vus 10 --duration 30s test/smoke_test.js"
