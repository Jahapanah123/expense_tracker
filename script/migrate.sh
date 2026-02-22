#!/bin/bash
set -e

ACTION="${1:-up}"

# Added 'drop' to validation
if [[ "$ACTION" != "up" && "$ACTION" != "down" && "$ACTION" != "status" && "$ACTION" != "drop" ]]; then
    echo "Usage: $0 [up|down|status|drop]"
    exit 1
fi

# Run migrate with -verbose to see errors
migrate -path ./migrations -database "postgres://jahapanah:123456@localhost:5432/expense_tracker?sslmode=disable" -verbose "$ACTION"