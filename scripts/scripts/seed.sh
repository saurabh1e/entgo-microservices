#!/bin/bash
# Seed admin user, role, and permissions in Docker postgres

set -e

# Find auth container
AUTH_CONTAINER=$(docker ps --format '{{.Names}}' | grep "entgo_auth" | head -1)

if [ -z "$AUTH_CONTAINER" ]; then
    echo "❌ Auth container not running. Start it first: docker-compose up -d"
    exit 1
fi

echo "🌱 Seeding database in Docker..."
echo ""

# Run seed command in container
docker exec -it ${AUTH_CONTAINER} go run ./cmd/seed/main.go

echo ""
echo "✅ Done! Login with: admin / admin123"
echo "⚠️  Change password immediately!"

