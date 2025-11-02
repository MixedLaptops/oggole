#!/bin/bash
set -e

# One-time VM setup script for Oggole
# Run this ONCE when setting up a new VM

echo "🔧 Setting up Oggole on VM..."

# Configuration
DATA_DIR="$HOME/oggole/data"
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

# Create data directory with proper permissions
echo "📁 Creating data directory at $DATA_DIR..."
mkdir -p "$DATA_DIR"

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
echo "1. Initialize and upload database:"
echo "   On your LOCAL machine:"
echo "     cd src"
echo "     go run init_db.go"
echo "     go run seed_data.go"
echo "     scp whoknows.db user@your-vm:~/oggole/data/oggole.db"
echo ""
echo "2. Upload configuration files:"
echo "   On your LOCAL machine:"
echo "     scp docker-compose.prod.yml user@your-vm:$APP_DIR/"
echo "     scp deploy.sh user@your-vm:$APP_DIR/"
echo "     scp nginx/nginx.conf user@your-vm:$APP_DIR/nginx/"
echo "     chmod +x $APP_DIR/deploy.sh"
echo ""
echo "3. (Optional) Setup GitHub Container Registry authentication:"
echo "   If using private images, run:"
echo "     echo \$GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin"
echo ""
echo "4. Deploy the application:"
echo "   cd $APP_DIR"
echo "   ./deploy.sh"
echo ""
echo "📍 Important paths:"
echo "   - Database: ~/oggole/data/oggole.db"
echo "   - App files: ~/oggole/"
echo "   - Nginx config: ~/oggole/nginx/nginx.conf"
