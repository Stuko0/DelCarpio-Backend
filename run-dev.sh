#!/usr/bin/env bash
# Development launcher for Del Carpio backend.
# Usage: bash run-dev.sh
set -a
source "$(dirname "$0")/.env"
set +a

cd "$(dirname "$0")"
echo "Starting Del Carpio backend..."
exec go run ./cmd/serve/
