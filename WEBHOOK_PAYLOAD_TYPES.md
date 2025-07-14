# 🎯 Webhook Payload Types - Complete Guide

Webhook này hỗ trợ **3 loại payload** khác nhau từ GitHub, mỗi loại có mục đích và cách sử dụng riêng:

## 📋 Tổng quan các loại Payload

| Payload Type | Khi nào trigger | Source | Ưu điểm | Nhược điểm |
|--------------|-----------------|--------|---------|------------|
| **Push Events** | Push code | GitHub tự động | Đơn giản, tự động | Chỉ deploy code |
| **Package Events** | Push Docker image | GitHub Container Registry | Tự động khi push image | Ít thông tin custom |
| **Workflow Payload** | GitHub Actions | Custom workflow | Linh hoạt, nhiều thông tin | Cần setup workflow |

---

## 1️⃣ GitHub Push Events (Chuẩn)

### 🎯 **Khi nào sử dụng:**
- Deploy source code sau khi push
- Trigger build trên server
- Restart services

### 📝 **Payload Structure:**
```json
{
  "ref": "refs/heads/main",
  "repository": {
    "name": "my-repo",
    "full_name": "owner/my-repo",
    "html_url": "https://github.com/owner/my-repo"
  },
  "pusher": {
    "name": "Developer Name",
    "email": "dev@example.com"
  },
  "head_commit": {
    "id": "abc123...",
    "message": "Fix bug in authentication",
    "url": "https://github.com/owner/my-repo/commit/abc123"
  }
}
```

### ⚙️ **Configuration:**
```bash
# Trong config.env
DEPLOY_COMMANDS_OWNER_MY_REPO=git pull origin main;npm ci;npm run build;pm2 restart app
WORK_DIR_OWNER_MY_REPO=/opt/my-repo
```

### 🎨 **Discord Notification:**
```
✅ Deployment Successful - Code Deployment
Repository: owner/my-repo
├─ Branch: main
├─ Commit: abc123 (Fix bug in authentication)
└─ Author: Developer Name
```

---

## 2️⃣ GitHub Package Events (Chuẩn)

### 🎯 **Khi nào sử dụng:**
- Deploy Docker image sau khi push lên GitHub Container Registry
- Automatic container updates

### 📝 **Payload Structure:**
```json
{
  "action": "published",
  "package": {
    "name": "my-docker-app",
    "package_version": "1.2.3",
    "registry": {
      "name": "container",
      "type": "docker",
      "url": "https://ghcr.io"
    }
  },
  "repository": {
    "name": "my-repo",
    "full_name": "owner/my-repo"
  }
}
```

### ⚙️ **Configuration:**
```bash
# GitHub Webhook settings - chọn "Package" events
# Deployment commands vẫn dùng config.env như bình thường
DEPLOY_COMMANDS_OWNER_MY_REPO=docker pull ghcr.io/owner/my-repo:latest;docker stop my-app;docker run -d my-app
```

### 🎨 **Discord Notification:**
```
✅ Deployment Successful - Package Deployment  
Repository: owner/my-repo
├─ Package: my-docker-app
├─ Version: 1.2.3
└─ Registry: docker
```

---

## 3️⃣ Custom Workflow Payload (GitHub Actions)

### 🎯 **Khi nào sử dụng:**
- Muốn control chi tiết deployment process
- Multi-environment deployment (production/staging)
- Custom Docker tags và metadata
- Combine build + deploy trong 1 workflow

### 📝 **Payload Structure:**
```json
{
  "ref": "refs/heads/main",
  "repository": { /* standard repo info */ },
  "pusher": { /* standard pusher info */ },
  "head_commit": { /* standard commit info */ },
  
  "docker": {
    "registry": "ghcr.io",
    "image_name": "tuanpham2xx3/mrs_address_be",
    "latest_tag": "latest", 
    "versioned_tag": "main-fdc36db",
    "latest_image": "ghcr.io/tuanpham2xx3/mrs_address_be:latest",
    "versioned_image": "ghcr.io/tuanpham2xx3/mrs_address_be:main-fdc36db",
    "pull_command": "docker pull ghcr.io/tuanpham2xx3/mrs_address_be:latest"
  },
  
  "deployment": {
    "environment": "production", // hoặc "staging"
    "branch": "main",
    "commit": "fdc36db...",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

### ⚙️ **Setup Workflow:**
Sử dụng file `examples/github-actions-workflow.yml`:

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy via Webhook

on:
  push:
    branches: [ main, develop ]

jobs:
  build-and-deploy:
    steps:
      - name: Build Docker Image
        # ... build steps ...
      
      - name: Send Custom Webhook
        run: |
          # Create custom payload với Docker info
          # Generate signature 
          # Send to webhook endpoint
```

### ⚙️ **GitHub Secrets Required:**
```
WEBHOOK_URL=https://your-webhook-domain.com
WEBHOOK_SECRET=du_an_cua_tuan
```

### 🤖 **Auto Commands:**
Webhook tự động generate commands từ payload:
```bash
# Production
docker pull ghcr.io/tuanpham2xx3/mrs_address_be:latest
docker stop tuanpham2xx3/mrs_address_be || true  
docker rm tuanpham2xx3/mrs_address_be || true
docker run -d --name tuanpham2xx3/mrs_address_be -p 8100:8100 ghcr.io/tuanpham2xx3/mrs_address_be:latest

# Staging  
docker run -d --name tuanpham2xx3/mrs_address_be-staging -p 8101:8100 ...
```

### 🎨 **Discord Notification:**
```
✅ Deployment Successful - Workflow Deployment
Repository: tuanpham2xx3/mrs_address_be
├─ Environment: production
├─ Branch: main
├─ Commit: fdc36db
├─ Docker Image: ghcr.io/tuanpham2xx3/mrs_address_be:latest
├─ Registry: ghcr.io
└─ Tags: latest: latest, versioned: main-fdc36db
```

---

## 🔧 Workflow Detection Logic

Webhook tự động detect payload type:

```go
// Detection priority:
1. payload.Docker.ImageName != "" && payload.Deployment.Environment != ""
   → Custom Workflow Payload

2. eventType == "package" && payload.Action == "published"  
   → GitHub Package Events

3. eventType == "push" || payload.Ref != ""
   → GitHub Push Events
```

## 🎯 Recommendations

### 🥇 **For Simple Projects:**
- **GitHub Push Events** - Deploy code changes automatically

### 🥈 **For Docker Projects:**
- **GitHub Package Events** - Auto deploy when image is updated  

### 🥉 **For Complex Projects:**
- **Custom Workflow Payload** - Full control over build + deployment process

## 📊 Feature Comparison

| Feature | Push Events | Package Events | Workflow Payload |
|---------|-------------|----------------|------------------|
| **Auto trigger** | ✅ Push code | ✅ Push image | ❌ Manual setup |
| **Docker info** | ❌ | ⚠️ Basic | ✅ Detailed |
| **Environment control** | ❌ | ❌ | ✅ |
| **Custom commands** | ✅ Config file | ✅ Config file | ✅ Auto + Config |
| **Setup complexity** | 🟢 Easy | 🟡 Medium | 🔴 Advanced |

## 🚀 Getting Started

1. **Basic setup:** Start với Push Events
2. **Docker ready:** Add Package Events  
3. **Production ready:** Implement Workflow Payload với multi-environment

Mỗi loại payload đều được hỗ trợ đầy đủ trong webhook này! 🎉 