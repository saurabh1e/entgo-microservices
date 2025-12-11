#!/bin/sh
set -e

echo "🚀 Starting development entrypoint..."

# Ensure gqlgen is installed (this will download its dependencies including urfave/cli/v3)
echo "📦 Ensuring gqlgen is available..."
go install github.com/99designs/gqlgen@v0.17.84

# Run go mod tidy
echo "📦 Running go mod tidy..."
if go mod tidy; then
    echo "✅ Go modules tidied successfully"
else
    echo "⚠️  Warning: go mod tidy had issues (may be okay in workspace mode)"
fi

# Run go generate
echo "🔄 Generating all code with generate.go..."
if go generate generate.go; then
    echo "✅ Code generation completed successfully"
else
    echo "❌ Failed to generate code"
    exit 1
fi

# Start air (or whatever command was passed)
echo "🔥 Starting Air for hot reload..."
exec "$@"

