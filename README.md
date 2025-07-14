# Webhook API Documentation

## Overview
This webhook service handles automated deployments for multiple types of projects (Go, Node.js, Python, PHP, Java, .NET, Docker) through GitHub webhooks. The service is deployed at `https://webhook1.iceteadev.site/` via Cloudflare Tunnel.

## Authentication
All webhook requests must include:
- GitHub webhook signature (`X-Hub-Signature-256` header)

## Endpoints

### Main Webhook Endpoint
```
POST https://webhook1.iceteadev.site/deploy
```

### Health Check Endpoint
```
GET https://webhook1.iceteadev.site/health
```

### Headers
- `Content-Type: application/json`
- `X-Hub-Signature-256: sha256=HASH` (HMAC SHA256 signature)
- `X-GitHub-Event: push` (or other GitHub event types)

### Request Body
GitHub webhook payload in JSON format. The service primarily responds to `push` events.

### Response Codes
- `200 OK`: Webhook processed successfully
- `400 Bad Request`: Invalid payload or missing headers
- `401 Unauthorized`: Invalid signature
- `500 Internal Server Error`: Deployment error

## Project Configuration

Projects are configured through environment variables following this pattern:
```env
DEPLOY_COMMANDS_OWNER_REPO_NAME=command1;command2;command3
WORK_DIR_OWNER_REPO_NAME=/path/to/working/directory
```

**Note**: For Docker workflows (GitHub Actions with container registry), working directories are not needed. The webhook will automatically use Docker commands from the payload.

### Example Configuration

For a repository `company/go-api`:
```env
DEPLOY_COMMANDS_COMPANY_GO_API=git pull origin main;go mod tidy;go test ./...;go build -o api-server;sudo systemctl restart go-api
WORK_DIR_COMPANY_GO_API=/opt/go-api
```

## Setting Up GitHub Webhooks

1. Go to your GitHub repository
2. Navigate to Settings > Webhooks
3. Click "Add webhook"
4. Configure the webhook:
   - Payload URL: `https://webhook1.iceteadev.site/deploy`
   - Content type: `application/json`
   - Secret: Your configured webhook secret
   - Events: Select "Just the push event"
   - Active: Check this box

## Supported Project Types

The webhook supports various project types including:
- Go projects (API, Microservices)
- Node.js (React, Express, Next.js)
- Python (Django, Flask, FastAPI)
- PHP (Laravel, WordPress)
- Java (Spring Boot, Gradle)
- .NET Core
- Docker/Docker Compose
- Full-stack and Monorepo projects

## Security Considerations

- Only GitHub IP ranges are allowed
- HMAC SHA256 signature verification
- Cloudflare Tunnel for secure connectivity
- Environment-based configuration

## Troubleshooting

If deployments fail, check:
1. GitHub webhook delivery logs
2. Webhook server logs
3. Project-specific deployment logs
4. Correct environment variable configuration
5. Working directory permissions

## Example GitHub Actions Integration

```yaml
name: Deploy via Webhook

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger deployment webhook
        uses: distributhor/workflow-webhook@v2
        with:
          url: https://webhook1.iceteadev.site/deploy
          secret: ${{ secrets.WEBHOOK_SECRET }}
```

## Rate Limiting

- Webhook requests are processed sequentially
- Multiple deployments for the same repository are queued
- Consider implementing cooldown periods between deployments

## Support

For issues or questions:
1. Check server logs
2. Verify GitHub webhook delivery logs
3. Ensure correct environment configuration
4. Check project-specific deployment requirements 

## Detailed GitHub Webhook Payload

### Required Headers
```http
Content-Type: application/json
X-Hub-Signature-256: sha256=HASH
X-GitHub-Event: push
X-GitHub-Delivery: unique-delivery-id
User-Agent: GitHub-Hookshot/*
```

### Example Payload
```json
{
  "ref": "refs/heads/main",
  "repository": {
    "name": "project-name",
    "full_name": "owner/project-name",
    "html_url": "https://github.com/owner/project-name",
    "clone_url": "https://github.com/owner/project-name.git",
    "default_branch": "main"
  },
  "pusher": {
    "name": "username",
    "email": "user@example.com"
  },
  "head_commit": {
    "id": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
    "message": "Fix logging issue",
    "timestamp": "2024-01-15T10:00:00Z",
    "url": "https://github.com/owner/project-name/commit/6dcb09b5b57875f334f61aebed695e2e4193db5e",
    "author": {
      "name": "User Name",
      "email": "user@example.com"
    }
  },
  "commits": [
    {
      "id": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
      "message": "Fix logging issue",
      "timestamp": "2024-01-15T10:00:00Z",
      "url": "https://github.com/owner/project-name/commit/6dcb09b5b57875f334f61aebed695e2e4193db5e",
      "author": {
        "name": "User Name",
        "email": "user@example.com"
      },
      "added": ["new-file.txt"],
      "removed": ["old-file.txt"],
      "modified": ["modified-file.txt"]
    }
  ]
}
```

### Important Fields Explanation

1. **Headers:**
   - `X-Hub-Signature-256`: HMAC SHA256 signature of the payload using your webhook secret
   - `X-GitHub-Event`: Type of event (e.g., "push", "pull_request")
   - `X-GitHub-Delivery`: Unique identifier for the delivery
   - `User-Agent`: Always starts with "GitHub-Hookshot/"

2. **Payload Fields:**
   - `ref`: Branch or tag that was pushed to (e.g., "refs/heads/main")
   - `repository`: Information about the repository
     - `name`: Repository name
     - `full_name`: Owner and repository name
     - `html_url`: GitHub URL of the repository
     - `clone_url`: Git clone URL
   - `pusher`: Information about who pushed the changes
   - `head_commit`: Details about the latest commit
   - `commits`: Array of all commits in this push
     - `added`: New files added
     - `removed`: Files removed
     - `modified`: Files modified

### Testing Webhook Locally

You can use the provided test script:

```bash
go run test/webhook_test.go https://webhook1.iceteadev.site/deploy your_webhook_secret
```

Or use curl:

```bash
# Generate HMAC SHA256 signature first
WEBHOOK_SECRET="your_secret"
PAYLOAD='{"ref":"refs/heads/main","repository":{"name":"test-repo"}}'
SIGNATURE=$(echo -n "$PAYLOAD" | openssl dgst -sha256 -hmac "$WEBHOOK_SECRET" | cut -d' ' -f2)

# Send test request
curl -X POST https://webhook1.iceteadev.site/deploy \
  -H "Content-Type: application/json" \
  -H "X-Hub-Signature-256: sha256=$SIGNATURE" \
  -H "X-GitHub-Event: push" \
  -H "X-GitHub-Delivery: test-delivery-id" \
  -H "User-Agent: GitHub-Hookshot/test" \
  -d "$PAYLOAD"
```

### Common Response Codes

1. **Success Response:**
```json
{
  "status": "accepted",
  "message": "Deployment initiated"
}
```

2. **Error Responses:**
```json
// 400 Bad Request
{
  "error": "Invalid payload format"
}

// 401 Unauthorized
{
  "error": "Invalid signature"
}

// 403 Forbidden
{
  "error": "IP not in allowed range"
}

// 500 Internal Server Error
{
  "error": "Deployment failed",
  "details": "Error executing deployment commands"
}
``` 