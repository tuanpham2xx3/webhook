# 🚀 Setup MRS Address BE - Webhook Configuration

## 📋 Environment Variables Required

### Webhook Server Configuration
```bash
# Basic webhook settings
export PORT=8300
export WEBHOOK_SECRET="your_webhook_secret_here"
export DISCORD_WEBHOOK="https://discord.com/api/webhooks/your_webhook_url"

# Global working directory (optional)
export WORK_DIR=/opt/projects
```

### Project-Specific Configuration

#### ⚠️ **IMPORTANT**: Repository Name Transformation
Webhook transforms repository names:
- `owner/mrs_address_be` → `OWNER_MRS_ADDRESS_BE`
- Replace `/` with `_` and convert to uppercase
- Replace `-` with `_`

#### For `company/mrs_address_be` repository:
```bash
# Deploy commands cho production
export DEPLOY_COMMANDS_COMPANY_MRS_ADDRESS_BE="git pull origin main;go mod tidy;go build -o mrs-address-be;sudo systemctl restart mrs-address-be"

# Working directory cho project này
export WORK_DIR_COMPANY_MRS_ADDRESS_BE="/opt/mrs-address-be"
```

#### For other repositories, follow this pattern:
```bash
# Example: owner/my-project → OWNER_MY_PROJECT
export DEPLOY_COMMANDS_OWNER_MY_PROJECT="command1;command2;command3"
export WORK_DIR_OWNER_MY_PROJECT="/path/to/project"
```

## 🔧 Deployment Commands Examples

### Go API Server
```bash
export DEPLOY_COMMANDS_COMPANY_MRS_ADDRESS_BE="git pull origin main;go mod tidy;go test ./...;go build -o mrs-address-be;sudo systemctl restart mrs-address-be"
```

### Go with Docker
```bash
export DEPLOY_COMMANDS_COMPANY_MRS_ADDRESS_BE="git pull origin main;go mod download;docker build -t mrs-address-be .;docker stop mrs-address-be || true;docker run -d --name mrs-address-be -p 8080:8080 mrs-address-be"
```

### Go with Docker Compose
```bash
export DEPLOY_COMMANDS_COMPANY_MRS_ADDRESS_BE="git pull origin main;docker-compose down;docker-compose build;docker-compose up -d"
```

## 🐳 Docker Setup

### Dockerfile optimization for Go:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o mrs-address-be

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mrs-address-be .
EXPOSE 8080
CMD ["./mrs-address-be"]
```

## 📊 GitHub Actions Configuration

### Required GitHub Secrets:
```
WEBHOOK_SECRET - Same as webhook server
WEBHOOK_URL - Your webhook endpoint (e.g., https://webhook.yourdomain.com/deploy)
```

### Repository Settings:
1. Go to Settings → Secrets and variables → Actions
2. Add the required secrets above
3. Configure webhook in repository settings

## 🔍 Testing & Debugging

### Test webhook locally:
```bash
# Start webhook server
go run main.go

# Test signature generation
go run test_debug.go
```

### Check webhook logs:
```bash
# View webhook logs
sudo journalctl -f -u your-webhook-service

# Check application logs
sudo journalctl -f -u mrs-address-be
```

## 🚨 Common Issues & Solutions

### 1. **Invalid Signature Error**
```
2025/07/13 10:58:13 Invalid signature from 172.17.0.1:55186
```

**Solutions:**
- ✅ Check `WEBHOOK_SECRET` matches in both GitHub and webhook server
- ✅ Verify no extra whitespace in secret
- ✅ Ensure payload format is compact (no extra spaces)

### 2. **No Deploy Commands Found**
```
No deployment commands configured for owner/repo
```

**Solutions:**
- ✅ Use correct environment variable format: `DEPLOY_COMMANDS_OWNER_REPO`
- ✅ Transform repository name correctly (uppercase, replace `/` and `-` with `_`)

### 3. **Permission Denied**
```
Permission denied when executing commands
```

**Solutions:**
- ✅ Check user permissions for git pull, build, and service restart
- ✅ Configure sudo permissions for systemctl commands
- ✅ Verify working directory permissions

### 4. **Service Not Starting**
```
Failed to restart service
```

**Solutions:**
- ✅ Check systemd service file exists and is enabled
- ✅ Verify binary path in service file
- ✅ Check application logs for startup errors

## 📝 Service File Example

Create `/etc/systemd/system/mrs-address-be.service`:
```ini
[Unit]
Description=MRS Address BE API Server
After=network.target

[Service]
Type=simple
User=deploy
WorkingDirectory=/opt/mrs-address-be
ExecStart=/opt/mrs-address-be/mrs-address-be
Restart=always
RestartSec=5
Environment=PORT=8080
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

Enable and start service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable mrs-address-be
sudo systemctl start mrs-address-be
```

## 🌐 Nginx Configuration

Add to nginx config:
```nginx
location /api/mrs-address/ {
    proxy_pass http://localhost:8080/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## 🎯 Success Checklist

- [ ] Webhook server running on port 8300
- [ ] Environment variables set correctly
- [ ] GitHub webhook configured with correct secret
- [ ] Deploy commands tested manually
- [ ] Service file created and enabled
- [ ] Nginx configuration updated
- [ ] Firewall rules configured
- [ ] Discord notifications working

---

## 📞 Quick Reference

**Repository**: `company/mrs_address_be`
**Env Variable**: `DEPLOY_COMMANDS_COMPANY_MRS_ADDRESS_BE`
**Working Dir**: `WORK_DIR_COMPANY_MRS_ADDRESS_BE`
**Service Name**: `mrs-address-be` 