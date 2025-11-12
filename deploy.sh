#!/bin/bash
set -e

# VM deployment script for Oggole
# Run this on the VM to deploy/update the application

echo "🚀 Deploying Oggole..."

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"

# Check if docker-compose file exists
if [ ! -f "$COMPOSE_FILE" ]; then
    echo "❌ Error: $COMPOSE_FILE not found"
    exit 1
fi

# Pull latest image from GitHub Container Registry
echo "📦 Pulling latest image from GitHub Packages..."
docker compose -f "$COMPOSE_FILE" pull oggole

# Restart services with new image
echo "🔄 Restarting services..."
docker compose -f "$COMPOSE_FILE" up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to start..."
sleep 5

# Show status
echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Service Status:"
docker compose -f "$COMPOSE_FILE" ps

echo ""
echo "📝 Recent logs:"
docker compose -f "$COMPOSE_FILE" logs --tail=20 oggole

echo ""
echo "🌐 Application should be available at:"
echo "   - http://localhost:80"
echo "   - http://your-vm-ip:80"
