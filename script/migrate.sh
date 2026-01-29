#!/bin/bash
set -e

# Go to project root
cd "$(dirname "$0")/.."

# Default to 'up' if no argument
ACTION="${1:-up}"

# Validate input
if [[ "$ACTION" != "up" && "$ACTION" != "down" && "$ACTION" != "status" ]]; then
    echo "Usage: $0 [up|down|status]"
    exit 1
fi

# Run migrate
migrate -path ./migrations -database "postgres://jahapanah:123456@localhost:5432/expense_tracker?sslmode=disable" "$ACTION"