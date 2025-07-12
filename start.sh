#!/bin/bash

# Webhook Auto Deploy Startup Script

set -e  # Exit on any error

echo "🚀 Starting Webhook Auto Deploy System..."

# Kiểm tra file cấu hình
if [ ! -f "config.env" ]; then
    echo "⚠️  File config.env không tồn tại. Tạo từ template..."
    cp config.env.example config.env
    echo "✅ File config.env đã được tạo. Vui lòng chỉnh sửa cấu hình trước khi chạy lại."
    exit 1
fi

# Load environment variables
echo "📋 Loading configuration..."
export $(cat config.env | grep -v '^#' | xargs)

# Kiểm tra Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go chưa được cài đặt. Vui lòng cài đặt Go 1.21 trở lên."
    exit 1
fi

echo "📦 Installing dependencies..."
go mod tidy

echo "🔨 Building application..."
go build -o webhook-deploy .

echo "🎯 Starting webhook server on port ${PORT:-8300}..."
echo "🔗 Health check: http://localhost:${PORT:-8300}/health"
echo "📥 Webhook endpoint: http://localhost:${PORT:-8300}/deploy"
echo ""
echo "📱 Discord webhook configured: $(echo $DISCORD_WEBHOOK | sed 's/webhooks\/[0-9]*\/[a-zA-Z0-9_-]*/webhooks\/***\/***/')"
echo ""
echo "Press Ctrl+C to stop..."
echo ""

# Chạy ứng dụng
./webhook-deploy 