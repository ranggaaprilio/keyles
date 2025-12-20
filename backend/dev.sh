#!/bin/bash

# Development script to run backend with Air (live reload)
# Usage: ./dev.sh

# Check if Air is installed
if ! command -v air &> /dev/null && [ ! -f ~/go/bin/air ]; then
    echo "Air is not installed. Installing..."
    go install github.com/air-verse/air@latest
fi

# Run Air
if command -v air &> /dev/null; then
    air
else
    ~/go/bin/air
fi
