#!/bin/bash

# Webhook Auto Deploy Startup Script

set -e  # Exit on any error

echo "Starting Webhook Auto Deploy System..."

# Kiểm tra file cấu hình
if [ ! -f "config.env" ]; then
    echo "File config.env khong ton tai. Tao tu template..."
    cp config.env.example config.env
    echo "File config.env da duoc tao. Vui long chinh sua cau hinh truoc khi chay lai."
    exit 1
fi

# Load only environment variables needed by Go app
echo "Loading configuration..."

# Danh sach cac bien can export (sua tuy app ban)
export $(grep -E '^(PORT|WEBHOOK_SECRET|DISCORD_WEBHOOK|WORK_DIR)=' config.env | grep -v '^#' | xargs)

# Kiem tra Go installation
if ! command -v go &> /dev/null; then
    echo "Go chua duoc cai dat. Vui long cai dat Go 1.21 tro len."
    exit 1
fi

echo "Installing dependencies..."
go mod tidy

echo "Building application..."
go build -o webhook-deploy .

echo "Starting webhook server on port ${PORT:-8300}..."
echo "Health check: http://localhost:${PORT:-8300}/health"
echo "Webhook endpoint: http://localhost:${PORT:-8300}/deploy"
echo ""
echo "Discord webhook configured: $(echo $DISCORD_WEBHOOK | sed 's/webhooks\/[0-9]*\/[a-zA-Z0-9_-]*/webhooks\/***\/***/')"
echo ""
echo "Press Ctrl+C to stop..."
echo ""

# Chay ung dung
./webhook-deploy
