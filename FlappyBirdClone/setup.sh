#!/bin/bash

# setup_go_deps.sh: Script to install Go dependencies for Flappy Bird Clone.
# Project directory: ~/GolandProjects/FlappyBirdClone
# Purpose: Install and verify Go modules (Ebiten and utilities).

# Exit on any error to prevent partial setup.
set -e

# Print a header for clarity.
echo "=== Setting up Go Dependencies for Flappy Bird Clone ==="

# Step 1: Check for Go installation.
echo "Checking for Go installation..."
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go (version 1.18 or later) from https://go.dev/dl/"
    exit 1
fi
GO_VERSION=$(go version | grep -o 'go[0-9]\.[0-9]\{1,\}')
if [[ ! "$GO_VERSION" =~ go1\.[1][8-9] && ! "$GO_VERSION" =~ go1\.[2-9][0-9] ]]; then
    echo "Error: Go version 1.18 or later is required. Found: $GO_VERSION"
    exit 1
fi
echo "Go is installed: $GO_VERSION"


# Step 2: Install Go dependencies.
echo "Installing Go dependencies..."
# Run go mod tidy to download and clean up dependencies.
go mod tidy
# Explicitly fetch Ebiten and utilities to ensure they’re included.
go get github.com/hajimehoshi/ebiten/v2@v2.8.7
go get github.com/hajimehoshi/ebiten/v2/ebitenutil
# Ensure image/png is available for PNG decoding.
go get image/png
echo "Dependencies installed."

# Step 4: Verify setup.
echo "Verifying Go dependencies..."
# Check go.mod for Ebiten dependency.
if grep -q "github.com/hajimehoshi/ebiten/v2" go.mod; then
    echo "Ebiten dependency confirmed."
else
    echo "Error: Ebiten not found in go.mod."
    exit 1
fi
# Check for main.go to ensure project is intact.
if [ ! -f "main.go" ]; then
    echo "Error: main.go not found in $PROJECT_DIR."
    exit 1
fi
echo "Setup verification complete: Dependencies and main.go present."

# Step 5: Instructions to proceed.
echo "=== Go Dependency Setup Complete ==="
echo "Next steps:"
echo "  1. Ensure image assets (bird.png, pipe.png, ground.png, background.png) are in $PROJECT_DIR."
echo "  2. Run the game with:"
echo "     go run main.go"
echo "See README.md for full project setup and CODE_DOCUMENTATION.md for code details."
