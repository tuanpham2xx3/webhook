# HƯỚNG DẪN SỬ DỤNG WEBHOOK - TRÁNH LỖI INVALID SIGNATURE

## 🚨 PHẦN QUAN TRỌNG - ĐỌC TRƯỚC KHI SỬ DỤNG

### Nguyên nhân chính gây lỗi "Invalid signature":
1. **Thiếu header signature** - Không gửi `X-Hub-Signature-256` hoặc `X-GitHub-Signature-256`
2. **Sai format signature** - Thiếu prefix `sha256=` hoặc sai encoding
3. **Secret không khớp** - Secret trên GitHub khác với secret trên server
4. **Payload bị thay đổi** - Body request không giống với khi tạo signature

---

## 📋 THIẾT LẬP GITHUB WEBHOOK

### 1. Cấu hình Webhook trên GitHub

Đi tới repository > Settings > Webhooks > Add webhook:

```
Payload URL: https://webhook1.iceteadev.site/deploy
Content type: application/json
Secret: [SECRET_CỦA_BẠN] (phải giống với WEBHOOK_SECRET trên server)
Which events: Just the push event
Active: ✅ (check)
```

### 2. Lấy Secret từ Server

**🔑 QUAN TRỌNG**: Secret phải giống nhau giữa GitHub và server!

```bash
# Kiểm tra secret hiện tại trên server
echo $WEBHOOK_SECRET
```

Nếu chưa có, thêm vào file `.env` hoặc environment:
```env
WEBHOOK_SECRET=du_an_cua_tuan
```

---

## 🔧 CẤU HÌNH DEPLOYMENT CHO DỰ ÁN

### Format Environment Variables

```env
# Commands để deploy
DEPLOY_COMMANDS_OWNER_REPO_NAME=command1;command2;command3

# Working directory
WORK_DIR_OWNER_REPO_NAME=/path/to/project
```

### Quy tắc đặt tên:
- Repository `company/my-api` → `COMPANY_MY_API`
- Repository `user/frontend-app` → `USER_FRONTEND_APP`
- Thay `/` và `-` bằng `_`, chuyển thành chữ hoa

### Ví dụ cấu hình:

#### Go API
```env
DEPLOY_COMMANDS_COMPANY_GO_API=git pull origin main;go mod tidy;go test ./...;go build -o api-server;sudo systemctl restart go-api
WORK_DIR_COMPANY_GO_API=/opt/go-api
```

#### React Frontend
```env
DEPLOY_COMMANDS_USER_FRONTEND=git pull origin main;npm ci;npm run build;rsync -av --delete build/ /var/www/html/;sudo systemctl restart nginx
WORK_DIR_USER_FRONTEND=/opt/frontend
```

#### Python Django
```env
DEPLOY_COMMANDS_COMPANY_DJANGO_APP=git pull origin main;pip install -r requirements.txt;python manage.py collectstatic --noinput;python manage.py migrate;sudo systemctl restart django-app
WORK_DIR_COMPANY_DJANGO_APP=/opt/django-app
```

---

## 🧪 KIỂM TRA WEBHOOK

### 1. Sử dụng Test Script có sẵn

```bash
cd test/
go run webhook_sender.go https://webhook1.iceteadev.site/deploy du_an_cua_tuan
```

### 2. Kiểm tra bằng curl

```bash
#!/bin/bash

# Thông tin cấu hình
WEBHOOK_URL="https://webhook1.iceteadev.site/deploy"
SECRET="du_an_cua_tuan"  # Secret thực của bạn

# Tạo payload test
PAYLOAD='{
  "ref": "refs/heads/main",
  "repository": {
    "name": "test-repo",
    "full_name": "user/test-repo",
    "html_url": "https://github.com/user/test-repo"
  },
  "pusher": {
    "name": "Test User",
    "email": "test@example.com"
  },
  "head_commit": {
    "id": "1234567890abcdef1234567890abcdef12345678",
    "message": "Test deployment",
    "url": "https://github.com/user/test-repo/commit/1234567890abcdef1234567890abcdef12345678"
  }
}'

# Tạo signature HMAC SHA256
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$SECRET" | cut -d' ' -f2)

# Gửi request
curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=$SIGNATURE" \
  -H "X-GitHub-Event: push" \
  -H "X-GitHub-Delivery: test-delivery-$(date +%s)" \
  -H "User-Agent: GitHub-Hookshot/test" \
  -d "$PAYLOAD"
```

### 3. Kiểm tra với PowerShell (Windows)

```powershell
# Chạy script có sẵn
.\test_webhook.ps1

# Hoặc với custom secret và repo
.\test_webhook.ps1 -Secret "du_an_cua_tuan" -RepoName "your-username/your-repo"
```

### 4. Kiểm tra với Python

```python
import hashlib
import hmac
import json
import requests

def create_webhook_signature(payload, secret):
    """Tạo signature cho webhook"""
    signature = hmac.new(
        secret.encode('utf-8'),
        payload.encode('utf-8'),
        hashlib.sha256
    )
    return f"sha256={signature.hexdigest()}"

# Cấu hình
webhook_url = "https://webhook1.iceteadev.site/deploy"
secret = "du_an_cua_tuan"

# Payload test
payload = {
    "ref": "refs/heads/main",
    "repository": {
        "name": "test-repo",
        "full_name": "user/test-repo",
        "html_url": "https://github.com/user/test-repo"
    },
    "pusher": {
        "name": "Test User",
        "email": "test@example.com"
    },
    "head_commit": {
        "id": "1234567890abcdef1234567890abcdef12345678",
        "message": "Test deployment",
        "url": "https://github.com/user/test-repo/commit/1234567890abcdef"
    }
}

# Chuyển thành JSON
payload_json = json.dumps(payload, separators=(',', ':'))

# Tạo signature
signature = create_webhook_signature(payload_json, secret)

# Headers
headers = {
    "Content-Type": "application/json",
    "X-Hub-Signature-256": signature,
    "X-GitHub-Event": "push",
    "X-GitHub-Delivery": "test-delivery-123",
    "User-Agent": "GitHub-Hookshot/test"
}

# Gửi request
response = requests.post(webhook_url, data=payload_json, headers=headers)

print(f"Status Code: {response.status_code}")
print(f"Response: {response.text}")
```

---

## 🔍 DEBUG LỖI SIGNATURE

### 1. Kiểm tra Headers
Webhook phải có đầy đủ headers:
```http
Content-Type: application/json
X-Hub-Signature-256: sha256=HASH_VALUE
X-GitHub-Event: push
X-GitHub-Delivery: unique-delivery-id
User-Agent: GitHub-Hookshot/*
```

### 2. Kiểm tra Secret
```bash
# Trên server
echo "Secret server: $WEBHOOK_SECRET"

# Trên GitHub: Settings > Webhooks > Edit > Secret
```

### 3. Kiểm tra Format Signature
- ✅ Đúng: `sha256=a1b2c3d4...`
- ❌ Sai: `a1b2c3d4...` (thiếu prefix)
- ❌ Sai: `SHA256=a1b2c3d4...` (sai case)

### 4. Tools Debug Online
- GitHub Webhook Deliveries (xem log gửi/nhận)
- RequestBin.com (test endpoint)
- ngrok (test local)

---

## 📚 CÁC LOẠI DỰ ÁN ĐƯỢC HỖ TRỢ

### Auto-Detection
Webhook tự động nhận diện loại project:

| File Marker | Project Type | Auto Commands |
|-------------|--------------|---------------|
| `go.mod`, `main.go` | Go | `go mod tidy`, `go build` |
| `package.json` | Node.js | `npm ci`, `npm run build` |
| `requirements.txt` | Python | `pip install -r requirements.txt` |
| `composer.json` | PHP | `composer install` |
| `pom.xml` | Java | `./mvnw clean package` |
| `Dockerfile` | Docker | `docker build`, `docker-compose` |

### Custom Commands
Nếu cần custom, dùng environment variables:
```env
# Override auto-detection
DEPLOY_COMMANDS_USER_MYPROJECT=git pull;custom command 1;custom command 2

# Custom working directory
WORK_DIR_USER_MYPROJECT=/custom/path
```

---

## ⚡ TROUBLESHOOTING

### Lỗi thường gặp:

#### 1. "Invalid signature"
```
✅ Kiểm tra secret GitHub = server
✅ Kiểm tra format signature có "sha256="
✅ Kiểm tra payload không bị thay đổi
✅ Test với script có sẵn
```

#### 2. "Unauthorized"
```
✅ Kiểm tra IP whitelist (nếu có)
✅ Kiểm tra headers đầy đủ
✅ Kiểm tra User-Agent
```

#### 3. "Deployment failed"
```
✅ Kiểm tra working directory tồn tại
✅ Kiểm tra permissions (git pull, systemctl)
✅ Kiểm tra commands syntax
✅ Xem logs Discord notification
```

#### 4. Commands không chạy
```
✅ Kiểm tra format environment variable
✅ Kiểm tra tên repository đúng format
✅ Kiểm tra semicolon (;) phân cách commands
```

---

## 📞 HỖ TRỢ

### Logs để check:
1. **Webhook server logs** - xem request đến
2. **GitHub webhook delivery logs** - xem response status
3. **Discord notifications** - xem kết quả deployment
4. **System logs** - `systemctl status service-name`

### Contacts:
- Check GitHub Issues
- Review server logs tại `/var/log/webhook/`
- Discord notifications để biết trạng thái deploy

---

## 🏆 BEST PRACTICES

### 1. Security
- Sử dụng secret mạnh (random, >= 32 characters)
- Không commit secret vào code
- Sử dụng HTTPS cho webhook URL
- Giới hạn IP nếu có thể

### 2. Reliability  
- Test webhook trước khi deploy production
- Backup trước khi auto deploy
- Sử dụng staging environment
- Monitor deployment logs

### 3. Performance
- Giữ commands deploy ngắn gọn
- Sử dụng systemctl restart thay vì kill process
- Cache dependencies khi có thể (npm ci vs npm install)

---

## 📝 TEMPLATE CẤU HÌNH NHANH

Copy và thay đổi theo dự án của bạn:

```env
# Basic webhook config
PORT=8300
WEBHOOK_SECRET=du_an_cua_tuan
DISCORD_WEBHOOK=your_discord_webhook_url

# Project deployment - Thay đổi OWNER và REPO_NAME
DEPLOY_COMMANDS_OWNER_REPO_NAME=git pull origin main;npm ci;npm run build;pm2 restart app
WORK_DIR_OWNER_REPO_NAME=/opt/your-project

# Hoặc dùng auto-detection (không cần config commands)
WORK_DIR_OWNER_REPO_NAME=/opt/your-project
```

**🎯 Bước tiếp theo**: Test webhook với script có sẵn trước khi sử dụng thực tế! 