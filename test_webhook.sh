#!/bin/bash

# Script test webhook nhanh - WEBHOOK TEST SCRIPT
# Usage: ./test_webhook.sh [SECRET] [REPO_NAME]

echo "🚀 WEBHOOK TEST SCRIPT"
echo "====================="

# Default values
WEBHOOK_URL="https://webhook1.iceteadev.site/deploy"
SECRET="${1:-du_an_cua_tuan}"
REPO_NAME="${2:-user/test-repo}"

echo "📡 Webhook URL: $WEBHOOK_URL"
echo "🔑 Secret: $SECRET"
echo "📦 Repository: $REPO_NAME"
echo ""

# Tạo test payload
PAYLOAD=$(cat <<EOF
{
  "ref": "refs/heads/main",
  "repository": {
    "name": "$(echo $REPO_NAME | cut -d'/' -f2)",
    "full_name": "$REPO_NAME",
    "html_url": "https://github.com/$REPO_NAME"
  },
  "pusher": {
    "name": "Test User",
    "email": "test@example.com"
  },
  "head_commit": {
    "id": "1234567890abcdef1234567890abcdef12345678",
    "message": "Test deployment via script",
    "url": "https://github.com/$REPO_NAME/commit/1234567890abcdef1234567890abcdef12345678"
  }
}
EOF
)

echo "📝 Payload được gửi:"
echo "$PAYLOAD" | jq '.' 2>/dev/null || echo "$PAYLOAD"
echo ""

# Tạo signature
echo "🔐 Đang tạo HMAC SHA256 signature..."
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" | cut -d' ' -f2)
echo "✅ Signature: sha256=$SIGNATURE"
echo ""

# Gửi request
echo "📤 Đang gửi webhook request..."
RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=$SIGNATURE" \
  -H "X-GitHub-Event: push" \
  -H "X-GitHub-Delivery: test-delivery-$(date +%s)" \
  -H "User-Agent: GitHub-Hookshot/test-script" \
  -d "$PAYLOAD")

# Parse response
HTTP_BODY=$(echo "$RESPONSE" | sed -n '1,/HTTP_STATUS:/p' | head -n -1)
HTTP_STATUS=$(echo "$RESPONSE" | grep "HTTP_STATUS:" | cut -d':' -f2)

echo "📨 Response Status: $HTTP_STATUS"
echo "📋 Response Body:"
echo "$HTTP_BODY" | jq '.' 2>/dev/null || echo "$HTTP_BODY"
echo ""

# Kiểm tra kết quả
case $HTTP_STATUS in
  200)
    echo "✅ SUCCESS! Webhook đã được xử lý thành công"
    echo "🎉 Deployment đang được thực hiện..."
    ;;
  401)
    echo "❌ ERROR: Invalid signature"
    echo "💡 Kiểm tra lại secret có đúng không:"
    echo "   - Secret trên GitHub webhook settings"
    echo "   - WEBHOOK_SECRET environment variable trên server"
    ;;
  400)
    echo "❌ ERROR: Bad request"
    echo "💡 Có thể do payload format không đúng"
    ;;
  404)
    echo "❌ ERROR: Endpoint not found"
    echo "💡 Kiểm tra URL webhook có đúng không"
    ;;
  500)
    echo "❌ ERROR: Server error"
    echo "💡 Có lỗi xảy ra trên server, kiểm tra logs"
    ;;
  *)
    echo "❌ ERROR: Unexpected status code $HTTP_STATUS"
    echo "💡 Kiểm tra server và network connection"
    ;;
esac

echo ""
echo "📚 Để debug thêm:"
echo "   1. Kiểm tra GitHub webhook deliveries"
echo "   2. Xem server logs"
echo "   3. Kiểm tra Discord notifications"
echo "   4. Đọc HUONG_DAN_SU_DUNG_WEBHOOK.md"
echo ""
echo "🔄 Usage examples:"
echo "   ./test_webhook.sh"
echo "   ./test_webhook.sh \"my_secret_123\""
echo "   ./test_webhook.sh \"my_secret_123\" \"company/my-project\"" 