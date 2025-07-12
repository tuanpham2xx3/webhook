# Yêu cầu xây dựng Webhook API để auto deploy

## 1. Yêu cầu cơ bản

- Endpoint HTTP nhận request từ hệ thống bên ngoài (ví dụ: GitHub, GitLab)
- Nhận và xử lý payload dạng JSON
- Thực thi lệnh trên server (ví dụ: git pull, build, restart app...)
- Trả về mã trạng thái HTTP phù hợp (200 OK, 400/500 khi lỗi)

## 2. Yêu cầu về bảo mật

- **Xác thực webhook:**
  - Kiểm tra secret/token gửi kèm header hoặc payload
  - Kiểm tra IP nguồn (whitelist IP GitHub/GitLab nếu cần)
  - Hoặc xác thực chữ ký HMAC (signature)
- Chỉ cho phép phương thức POST
- Giới hạn tần suất (rate limit) nếu cần
- Ghi log các request đến webhook (thời gian, nội dung, IP...)

## 3. Yêu cầu về xử lý và quản lý

- Kiểm soát lỗi khi thực thi script (catch error, báo lỗi rõ ràng)
- (Tùy chọn) Thực hiện bất đồng bộ nếu thao tác lâu
- Trả về kết quả thành công/thất bại rõ ràng
- Có thể tích hợp gửi thông báo (Slack, Email, Discord...) sau deploy

## 4. Checklist tiêu chuẩn

-

## 5. Ví dụ kiểm tra secret trong Python Flask

```python
from flask import Flask, request, abort
app = Flask(__name__)
SECRET = 'your_secret_here'
@app.route('/deploy', methods=['POST'])
def deploy():
    token = request.headers.get('X-Hub-Signature-256')
    if not token or not check_signature(request.data, token, SECRET):
        abort(403)
    # Triển khai logic deploy ở đây
    return 'OK', 200
```

(Hàm check\_signature dùng hmac theo chuẩn của GitHub/GitLab)



