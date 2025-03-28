#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Print debug information
echo "Current directory: $(pwd)"
echo "Checking for required files..."
ls -la levelEditor/*.go

echo "Searching for asset directory..."
find mazeGame -type d -name "assets" -print

echo "Finding PNG files..."
find mazeGame/assets -name "*.png" | head -5

# Build and run the level editor
echo "Building level editor..."
go build -o levelEditor/levelEditor levelEditor/*.go

echo "Running level editor..."
./levelEditor/levelEditor 