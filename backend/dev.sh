#!/usr/bin/env bash
set -euo pipefail

# Development script to run backend with Air (live reload)
# Usage: ./dev.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

AIR_BIN=""
if command -v air >/dev/null 2>&1; then
	AIR_BIN="$(command -v air)"
elif [ -x "$HOME/go/bin/air" ]; then
	AIR_BIN="$HOME/go/bin/air"
else
	echo "Air is not installed. Install it with: go install github.com/air-verse/air@v1.52.3" >&2
	exit 1
fi

exec "$AIR_BIN"
