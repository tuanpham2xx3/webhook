@echo off
REM Webhook Auto Deploy Startup Script for Windows

echo 🚀 Starting Webhook Auto Deploy System...

REM Kiểm tra file cấu hình
if not exist "config.env" (
    echo ⚠️  File config.env không tồn tại. Tạo từ template...
    copy config.env.example config.env
    echo ✅ File config.env đã được tạo. Vui lòng chỉnh sửa cấu hình trước khi chạy lại.
    pause
    exit /b 1
)

REM Load environment variables từ file config.env
echo 📋 Loading configuration...
for /f "eol=# tokens=*" %%a in (config.env) do set %%a

REM Kiểm tra Go installation
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ Go chưa được cài đặt. Vui lòng cài đặt Go 1.21 trở lên.
    pause
    exit /b 1
)

echo 📦 Installing dependencies...
go mod tidy

echo 🔨 Building application...
go build -o webhook-deploy.exe .

if %PORT%=="" set PORT=8300

echo 🎯 Starting webhook server on port %PORT%...
echo 🔗 Health check: http://localhost:%PORT%/health
echo 📥 Webhook endpoint: http://localhost:%PORT%/deploy
echo.
echo 📱 Discord webhook configured
echo.
echo Press Ctrl+C to stop...
echo.

REM Chạy ứng dụng
webhook-deploy.exe 