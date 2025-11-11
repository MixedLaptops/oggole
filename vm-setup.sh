#!/bin/bash
set -e

# One-time VM setup script for Oggole
# Run this ONCE when setting up a new VM

echo "🔧 Setting up Oggole on VM..."

# Configuration
POSTGRES_DIR="$HOME/oggole/postgres"
APP_DIR="$HOME/oggole"

# Check if running on VM (not locally)
if [ ! -d "/var/lib" ]; then
    echo "⚠️  Warning: This doesn't look like a VM. Are you sure you want to continue? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 1
    fi
fi

# Create PostgreSQL data directory with proper permissions
echo "📁 Creating PostgreSQL data directory at $POSTGRES_DIR..."
mkdir -p "$POSTGRES_DIR"

# Create app directory
echo "📁 Creating app directory at $APP_DIR..."
mkdir -p "$APP_DIR"
mkdir -p "$APP_DIR/nginx"
mkdir -p "$APP_DIR/nginx/ssl"

echo ""
echo "✅ VM setup complete!"
echo ""
echo "📋 Next steps:"
echo ""
echo "1. (Optional) Set PostgreSQL password:"
echo "   export POSTGRES_PASSWORD=your_secure_password"
echo "   (If not set, defaults to 'oggole')"
echo ""
echo "2. Upload configuration files:"
echo "   On your LOCAL machine:"
echo "     scp docker-compose.prod.yml user@your-vm:$APP_DIR/"
echo "     scp deploy.sh user@your-vm:$APP_DIR/"
echo "     scp nginx/nginx.conf user@your-vm:$APP_DIR/nginx/"
echo "     ssh user@your-vm 'chmod +x ~/oggole/deploy.sh'"
echo ""
echo "3. (Optional) Setup GitHub Container Registry authentication:"
echo "   If using private images, run:"
echo "     echo \$GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin"
echo ""
echo "4. Deploy the application:"
echo "   cd $APP_DIR"
echo "   ./deploy.sh"
echo ""
echo "   NOTE: PostgreSQL will initialize automatically on first run!"
echo ""
echo "5. Initialize database schema:"
echo "   docker exec oggole-app /app/oggole init-db"
echo "   docker exec oggole-app /app/oggole seed-data"
echo ""
echo "📍 Important paths:"
echo "   - PostgreSQL data: ~/oggole/postgres/"
echo "   - App files: ~/oggole/"
echo "   - Nginx config: ~/oggole/nginx/nginx.conf"
