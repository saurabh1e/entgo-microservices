#!/usr/bin/env bash
#
# sync-grpc-templates.sh
# Synchronize the fixed serviceTemplate from core to all other microservices
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SOURCE_FILE="$PROJECT_ROOT/core/cmd/generate-grpc/main.go"
MICROSERVICES=(
    "attendance"
    "fleet"
    "route"
    "notification"
    "reporting"
    "yard"
    "auth"
)

echo "🔧 Synchronizing gRPC template logic from core to all microservices..."
echo ""

# Extract the serviceTemplate section from core
TEMP_TEMPLATE=$(mktemp)
awk '/^const serviceTemplate = `/,/^`$/' "$SOURCE_FILE" > "$TEMP_TEMPLATE"

if [ ! -s "$TEMP_TEMPLATE" ]; then
    echo "❌ Failed to extract serviceTemplate from core"
    rm -f "$TEMP_TEMPLATE"
    exit 1
fi

echo "✅ Extracted template from core service"
echo ""

for service in "${MICROSERVICES[@]}"; do
    TARGET_FILE="$PROJECT_ROOT/$service/cmd/generate-grpc/main.go"

    if [ ! -f "$TARGET_FILE" ]; then
        echo "⚠️  Skipping $service - file not found"
        continue
    fi

    echo "📝 Updating $service..."

    # Create a backup
    cp "$TARGET_FILE" "$TARGET_FILE.backup"

    # Replace the serviceTemplate section
    awk '
    /^const serviceTemplate = `/ {
        skip=1
        while ((getline line < "'"$TEMP_TEMPLATE"'") > 0) {
            print line
        }
        close("'"$TEMP_TEMPLATE"'")
    }
    skip && /^`$/ {
        skip=0
        next
    }
    !skip { print }
    ' "$TARGET_FILE" > "$TARGET_FILE.tmp"

    mv "$TARGET_FILE.tmp" "$TARGET_FILE"

    echo "  ✅ Updated $service"

    # Regenerate gRPC code
    echo "  🔄 Regenerating gRPC code for $service..."
    cd "$PROJECT_ROOT/$service"
    rm -f grpc/*_service.go grpc/service_registry.go
    if go run cmd/generate-grpc/main.go 2>&1 | tail -5; then
        echo "  ✅ Successfully regenerated gRPC code for $service"
    else
        echo "  ❌ Failed to regenerate gRPC code for $service"
        # Restore backup
        mv "$TARGET_FILE.backup" "$TARGET_FILE"
    fi

    echo ""
done

rm -f "$TEMP_TEMPLATE"

echo "✅ Template synchronization complete!"
echo ""
echo "Testing builds..."
cd "$PROJECT_ROOT"
for service in core attendance fleet route; do
    echo "Building $service/grpc..."
    if cd "$PROJECT_ROOT/$service" && go build ./grpc/... 2>&1 | head -10; then
        echo "  ✅ $service builds successfully"
    else
        echo "  ❌ $service has build errors"
    fi
done

