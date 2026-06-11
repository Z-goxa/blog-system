# blog-system 部署到公网：Vercel 前端 + VPS 后端

本项目不是纯静态博客：前台文章、后台管理、评论、用户系统都依赖 Go API 和 MariaDB。

推荐部署结构：

```text
用户浏览器
  ↓
Vercel：Vue 前端静态资源
  ↓ /api 代理
VPS：Go API 服务
  ↓
VPS：MariaDB
```

## 1. 域名规划

建议准备两个域名：

- 博客前台：`https://blog.example.com` → Vercel
- API 后端：`https://api.example.com` → VPS

如果暂时没有自定义域名，也可以先用：

- Vercel 默认域名：`https://xxx.vercel.app`
- VPS 公网 IP 或临时域名

## 2. Vercel 部署前端

### 2.1 导入项目

在 Vercel 新建项目时选择本仓库，并设置：

- Framework Preset：`Vite`
- Root Directory：`frontend`
- Build Command：`npm run build`
- Output Directory：`dist`

项目已准备好：

- `frontend/vercel.json`
- `frontend/.env.production.example`

### 2.2 配置 API 代理

打开：

```text
frontend/vercel.json
```

把：

```json
"destination": "https://api.example.com/api/:path*"
```

改成你的真实后端域名，例如：

```json
"destination": "https://api.yourdomain.com/api/:path*"
```

这样前端仍然请求 `/api`，由 Vercel 转发到后端。

这种方式的优点：

- 前端代码不用改；
- 浏览器看到的是同源 `/api`；
- 后端可以少处理一部分 CORS 问题。

## 3. 后端 VPS 部署

### 3.1 服务器准备

VPS 上需要：

- Linux
- Go
- MariaDB
- Nginx 或 Caddy

建议目录：

```text
/opt/blog-system
  ├── backend
  ├── .env
  └── blog-system-new
```

### 3.2 构建后端

本地或服务器上执行：

```bash
cd /opt/blog-system/backend
go build -o blog-system-new .
```

### 3.3 配置生产 `.env`

参考项目根目录 `.env.example`，在服务器上创建：

```bash
/opt/blog-system/.env
```

至少需要配置：

```env
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=blog_user
DB_PASSWORD=请换成强密码
DB_NAME=blog_db
JWT_SECRET=请换成32字节以上随机字符串
```

不要把生产 `.env` 提交到 Git。

## 4. 数据库迁移

### 4.1 本地导出

```bash
mysqldump --default-character-set=utf8mb4 -u root -p blog_db > blog_db.sql
```

### 4.2 服务器导入

```bash
mysql --default-character-set=utf8mb4 -u blog_user -p blog_db < blog_db.sql
```

导入后建议登录后台立刻修改默认管理员密码。

## 5. systemd 常驻后端服务

在 VPS 创建：

```bash
sudo nano /etc/systemd/system/blog-system.service
```

示例内容：

```ini
[Unit]
Description=blog-system Go API
After=network.target mariadb.service

[Service]
Type=simple
WorkingDirectory=/opt/blog-system/backend
EnvironmentFile=/opt/blog-system/.env
ExecStart=/opt/blog-system/backend/blog-system-new -api
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now blog-system
sudo systemctl status blog-system
```

## 6. Nginx 反向代理 API

示例：

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

HTTPS 可以用 Certbot 或 Caddy 自动申请。

## 7. 本地构建说明

现在 `frontend/vite.config.js` 默认输出到：

```text
frontend/dist
```

适合 Vercel。

如果本地开发时仍想把前端静态资源打进 Go 后端目录，执行：

```bash
cd frontend
BUILD_TARGET=backend npm run build
```

输出目录会变成：

```text
backend/frontend/dist
```

## 8. 上线后检查

前端：

```text
https://blog.yourdomain.com/#/blog
```

API：

```text
https://api.yourdomain.com/api/public/posts
```

Vercel 代理：

```text
https://blog.yourdomain.com/api/public/posts
```

如果这三个都正常，说明前后端链路打通。

## 9. 注意事项

- 后台入口已经从前台导航隐藏，但 `/admin` 路由仍存在。
- 生产环境必须使用强密码和强 `JWT_SECRET`。
- 建议定期备份 MariaDB。
- API 域名建议只开放 80/443，不要直接暴露数据库端口。
- 如果不用 Vercel `/api` 代理，而是设置 `VITE_API_URL=https://api.xxx/api`，后端需要配置 CORS。
