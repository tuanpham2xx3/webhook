# Webhook Auto Deploy System

Hệ thống webhook tự động deploy được viết bằng Go, tích hợp với Discord notifications.

## 🚀 Tính năng

- ✅ Nhận webhook từ GitHub/GitLab
- ✅ Xác thực HMAC SHA-256 signature
- ✅ Kiểm soát IP whitelist
- ✅ Thực thi lệnh deploy tự động
- ✅ Gửi thông báo Discord với embed đẹp
- ✅ Logging chi tiết
- ✅ Health check endpoint
- ✅ Rate limiting middleware
- ✅ Xử lý bất đồng bộ

## 📋 Yêu cầu

- Go 1.21 hoặc cao hơn
- Git (để thực thi lệnh git pull)
- Docker (tùy chọn)

## ⚙️ Cài đặt

### 1. Clone repository

```bash
git clone <your-repo-url>
cd webhook-deploy
```

### 2. Cấu hình biến môi trường

```bash
cp config.env.example config.env
```

Chỉnh sửa file `config.env`:

```env
PORT=8300
WEBHOOK_SECRET=your_very_secure_secret_here
DISCORD_WEBHOOK=https://discord.com/api/webhooks/YOUR_WEBHOOK_URL
ALLOWED_IPS=192.30.252.0/22,185.199.108.0/22
```

### 3. Chạy ứng dụng

#### Với Go native:

```bash
# Tải dependencies
go mod tidy

# Chạy ứng dụng
source config.env && go run main.go
```

#### Với Docker:

```bash
# Build image
docker build -t webhook-deploy .

# Chạy container
docker run -d \
  --name webhook-deploy \
  -p 8300:8300 \
  --env-file config.env \
  webhook-deploy
```

## 🔧 Cấu hình GitHub/GitLab

### GitHub Webhook

1. Vào repository → Settings → Webhooks
2. Thêm webhook mới:
   - **Payload URL**: `http://your-server:8300/deploy`
   - **Content type**: `application/json`
   - **Secret**: (giống với `WEBHOOK_SECRET`)
   - **Events**: Push events
   - **Active**: ✅

### GitLab Webhook

1. Vào project → Settings → Webhooks
2. Thêm webhook:
   - **URL**: `http://your-server:8300/deploy`
   - **Secret Token**: (giống với `WEBHOOK_SECRET`)
   - **Trigger**: Push events
   - **SSL verification**: Enable/Disable tùy setup

## 🎯 API Endpoints

### POST /deploy
Nhận webhook từ Git provider.

**Headers:**
- `Content-Type: application/json`
- `X-Hub-Signature-256: sha256=<signature>` (GitHub)
- `X-GitHub-Signature-256: sha256=<signature>` (GitHub alternative)

**Response:**
```json
{
  "status": "accepted",
  "message": "Deployment initiated"
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "time": "2024-01-15T10:30:00Z"
}
```

## 🔒 Bảo mật

### 1. HMAC Signature Verification
Ứng dụng sử dụng HMAC SHA-256 để xác thực webhook:

```go
// Header: X-Hub-Signature-256: sha256=<hash>
func checkSignature(payload []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedMAC := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
```

### 2. IP Whitelist
Chỉ cho phép IP từ danh sách được cấu hình:

```env
ALLOWED_IPS=192.30.252.0/22,185.199.108.0/22,140.82.112.0/20
```

### 3. GitHub IP Ranges
GitHub sử dụng các IP ranges sau (cập nhật từ GitHub Meta API):
- `192.30.252.0/22`
- `185.199.108.0/22` 
- `140.82.112.0/20`
- `143.55.64.0/20`

## 📱 Discord Notifications

Webhook gửi embed message đẹp với thông tin:

- ✅/❌ Trạng thái deploy
- 📁 Repository name
- 🌿 Branch name
- 📝 Commit message
- 👤 Author
- 🔗 Commit URL
- ⏰ Timestamp

## 🛠️ Cấu hình Đa Ngôn ngữ

Webhook **tự động detect** và support đa ngôn ngữ:

### 🔍 Auto-Detection hỗ trợ:

| Ngôn ngữ | File marker | Deploy commands |
|----------|-------------|-----------------|
| **Go** | `go.mod`, `main.go` | `go mod tidy` → `go build -o app` |
| **Node.js** | `package.json` | `npm ci` → `npm run build` |
| **Python** | `requirements.txt`, `setup.py` | `pip install -r requirements.txt` |
| **PHP** | `composer.json`, `index.php` | `composer install --no-dev` |
| **Java** | `pom.xml`, `build.gradle` | `./mvnw clean package` |
| **.NET** | `*.csproj`, `*.sln` | `dotnet restore` → `dotnet build` |
| **Docker** | `Dockerfile`, `docker-compose.yml` | `docker build` → `docker-compose up` |

### ⚙️ Custom Commands

#### 1. Global Commands (tất cả repo):
```env
DEPLOY_COMMANDS=git pull origin main;npm ci;npm run build;pm2 restart all
```

#### 2. Repository-specific Commands:
```env
# Cho repository "user/my-api"
DEPLOY_COMMANDS_USER_MY_API=git pull origin main;go build -o api;sudo systemctl restart my-api

# Cho repository "company/frontend"  
DEPLOY_COMMANDS_COMPANY_FRONTEND=git pull origin main;npm ci;npm run build;sudo systemctl restart nginx
```

#### 3. Working Directory:
```env
# Global working directory
WORK_DIR=/opt/projects

# Repository-specific
WORK_DIR_USER_MY_API=/opt/my-api
WORK_DIR_COMPANY_FRONTEND=/var/www/html
```

### 📝 Ví dụ cấu hình cho từng ngôn ngữ:

**Node.js API:**
```env
DEPLOY_COMMANDS_USER_NODE_API=git pull origin main;npm ci;npm run build;pm2 restart node-api
WORK_DIR_USER_NODE_API=/opt/node-api
```

**Python Flask:**
```env
DEPLOY_COMMANDS_USER_FLASK_APP=git pull origin main;pip install -r requirements.txt;sudo systemctl restart flask-app
WORK_DIR_USER_FLASK_APP=/opt/flask-app
```

**PHP Laravel:**
```env
DEPLOY_COMMANDS_USER_LARAVEL=git pull origin main;composer install --no-dev;php artisan cache:clear;sudo systemctl restart nginx
WORK_DIR_USER_LARAVEL=/var/www/laravel
```

**Docker Compose:**
```env
DEPLOY_COMMANDS_USER_DOCKER_APP=git pull origin main;docker-compose down;docker-compose build;docker-compose up -d
WORK_DIR_USER_DOCKER_APP=/opt/docker-app
```

### 🔄 Command Prioritization

Webhook sẽ chọn deploy commands theo thứ tự ưu tiên:

1. **Repository-specific commands** (`DEPLOY_COMMANDS_REPO_NAME`)
2. **Global custom commands** (`DEPLOY_COMMANDS`) 
3. **Auto-detected commands** (dựa trên project type)

## 📝 Logging

Ứng dụng ghi log chi tiết:

```
2024/01/15 10:30:00 [2024-01-15 10:30:00] POST /deploy from 192.30.252.1
2024/01/15 10:30:00 Received webhook for repository: user/repo, ref: refs/heads/main
2024/01/15 10:30:00 Starting deployment for user/repo
2024/01/15 10:30:01 Executing: git pull origin main
2024/01/15 10:30:02 Command successful: git pull origin main
2024/01/15 10:30:02 Sending Discord notification...
2024/01/15 10:30:03 Discord notification sent successfully
```

## 🐳 Docker Compose

Tạo file `docker-compose.yml`:

```yaml
version: '3.8'
services:
  webhook-deploy:
    build: .
    ports:
      - "8300:8300"
    environment:
      - PORT=8300
      - WEBHOOK_SECRET=your_secret
      - DISCORD_WEBHOOK=your_discord_url
      - ALLOWED_IPS=
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock # Nếu cần deploy Docker
```

## 🔄 Systemd Service

Tạo service để chạy tự động:

```ini
[Unit]
Description=Webhook Deploy Service
After=network.target

[Service]
Type=simple
User=webhook
ExecStart=/opt/webhook-deploy/webhook-deploy
WorkingDirectory=/opt/webhook-deploy
Restart=always
RestartSec=10
Environment=PORT=8300
Environment=WEBHOOK_SECRET=your_secret
Environment=DISCORD_WEBHOOK=your_discord_url

[Install]
WantedBy=multi-user.target
```

## 🚨 Troubleshooting

### 1. Signature verification failed
- Kiểm tra `WEBHOOK_SECRET` khớp với GitHub/GitLab
- Kiểm tra header `X-Hub-Signature-256`

### 2. Discord notification không gửi được
- Kiểm tra `DISCORD_WEBHOOK` URL
- Kiểm tra network connectivity

### 3. Deploy commands thất bại
- Kiểm tra quyền thực thi
- Kiểm tra working directory
- Xem log chi tiết

### 4. Auto-detection không đúng ngôn ngữ
- Kiểm tra file marker tồn tại (package.json, go.mod, etc.)
- Thiết lập custom commands với `DEPLOY_COMMANDS_REPO_NAME`
- Kiểm tra `WORK_DIR` đúng đường dẫn

### 5. Multiple project types detected
- Webhook sẽ chạy commands cho tất cả types detected
- Sử dụng custom commands để control chính xác
- Ví dụ: Project có cả Dockerfile và package.json

## 📞 Hỗ trợ

Nếu gặp vấn đề, vui lòng:
1. Kiểm tra logs
2. Kiểm tra cấu hình
3. Test với `/health` endpoint
4. Tạo issue trên repository

## 📄 License

MIT License - Xem file LICENSE để biết thêm chi tiết. 