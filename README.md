# Blog System

一个基于 **Go + Wails + Vue 3 + MariaDB** 的个人博客系统。项目同时支持：

- **前台博客**：公开访问文章列表、文章详情、分类、归档、最新评论与评论提交。
- **后台管理**：登录后管理文章、分类、标签、评论、用户和站点统计。
- **HTTP API 模式**：作为本地 Web 服务运行，适合浏览器访问和开发调试。
- **Wails 桌面模式**：保留桌面应用能力，前端资源内嵌到 Go 程序中。

当前前端使用 hash 路由，访问路径示例：

- 博客首页：`http://localhost:8080/#/blog`
- 后台登录：`http://localhost:8080/#/login`
- 后台首页：`http://localhost:8080/#/admin`

---

## 技术栈

### 后端

- **Go 1.22**
- **Wails v2**：桌面应用框架与资源嵌入
- **net/http**：HTTP API 与静态文件服务
- **GORM**：ORM、迁移、关联查询
- **MariaDB / MySQL Driver**：数据库存储
- **JWT**：登录认证
- **bcrypt**：密码哈希
- **goldmark**：Markdown 渲染为 HTML

### 前端

- **Vue 3**
- **Vue Router 4**，使用 `createWebHashHistory()`
- **Vite 5**
- **Tailwind CSS 3** + `@tailwindcss/typography`
- **markdown-it**：前端 Markdown 预览
- **highlight.js**：代码高亮
- **lodash-es**

### 数据库

- **MariaDB 10.6+** 推荐
- 字符集：`utf8mb4`
- 引擎：`InnoDB`
- 主要表：`users`、`categories`、`tags`、`posts`、`post_tag_relation`、`comments`、`sessions`

---

## 功能概览

### 前台博客

- 文章列表分页
- 文章详情页
- 按分类查看文章
- 关键词搜索文章
- 按月份归档筛选文章
- 分类列表及文章数量
- 最新评论侧边栏
- 文章评论列表
- 登录用户提交评论
- 访客注册 / 登录入口

### 后台管理

- JWT 登录认证
- 仪表盘统计
- 文章列表、创建、编辑、删除
- Markdown 编辑与 HTML 渲染
- 分类创建、编辑、删除
- 标签创建、删除
- 评论列表、审核状态更新、删除
- 图片上传
- 用户列表
- 管理员强制修改用户密码
- 用户角色调整
- 用户删除

### 权限与安全

- 密码使用 bcrypt 哈希存储。
- JWT 默认有效期 24 小时。
- 后台接口需要 `Authorization: Bearer <token>`。
- `/admin` 前端路由需要登录。
- `subscriber` 用户会被前端拦截，不能进入后台。
- 用户管理接口要求管理员权限，并会重新查询数据库中的当前用户角色。
- 防止删除当前账号。
- 防止删除或降级最后一个管理员。
- 防止删除仍拥有文章的用户。
- 删除用户时会清理会话并将其评论解绑。

---

## 项目结构

```text
blog-system/
├── backend/
│   ├── api.go              # HTTP API、路由、鉴权中间件、静态资源服务
│   ├── app.go              # 程序入口：数据库初始化、迁移、种子数据、Wails 启动
│   ├── auth.go             # 用户注册、登录、JWT、密码哈希
│   ├── database.go         # MariaDB 连接配置、连接池、全局 DB 实例
│   ├── models.go           # GORM 模型：用户、分类、标签、文章、评论等
│   ├── seed_data.go        # 初始分类、标签、示例文章等种子数据
│   ├── service.go          # 博客业务逻辑：文章、分类、标签、评论、统计
│   ├── upload.go           # 图片上传与上传文件访问
│   ├── service_test.go     # 后端服务测试
│   ├── wails.json          # Wails 配置
│   └── uploads/            # 上传文件目录
├── frontend/
│   ├── src/
│   │   ├── assets/         # 静态资源，如头像
│   │   ├── components/
│   │   │   ├── blog/       # 前台博客页面组件
│   │   │   ├── Dashboard.vue
│   │   │   ├── Login.vue
│   │   │   ├── PostEditor.vue
│   │   │   ├── PostList.vue
│   │   │   ├── CategoryList.vue
│   │   │   ├── TagList.vue
│   │   │   ├── CommentList.vue
│   │   │   └── UserList.vue
│   │   ├── router/         # Vue Router 路由配置
│   │   ├── services/       # API 与认证封装
│   │   ├── utils/          # Wails/浏览器环境兼容工具
│   │   ├── App.vue
│   │   ├── main.js
│   │   └── style.css
│   ├── package.json
│   ├── vite.config.js
│   ├── tailwind.config.js
│   └── index.html
├── sql/
│   └── schema.sql          # MariaDB DDL 参考脚本
├── .env.example            # 环境变量示例
├── setup.sh                # 快速初始化脚本
└── README.md
```

---

## 环境要求

- Go **1.22+**
- Node.js **18+**（建议 20+）
- npm
- MariaDB **10.6+** 或兼容 MySQL 服务
- 可选：Wails CLI（需要桌面构建/开发时使用）

安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

---

## 配置说明

复制环境变量模板：

```bash
cp .env.example .env
```

按实际情况修改 `.env`：

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=blog_db

DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=300
DB_CONN_MAX_IDLE_TIME=120

JWT_SECRET=change-me-please-use-openssl-rand-hex-32
```

建议生成强随机 JWT 密钥：

```bash
openssl rand -hex 32
```

> 如果没有设置 `JWT_SECRET`，程序会在启动时生成一次性临时密钥；重启后旧 token 会全部失效。

---

## 数据库初始化

创建数据库：

```bash
mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS blog_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

项目启动时会通过 GORM 自动迁移表结构，并重建已知外键。`sql/schema.sql` 主要作为 DDL 参考或手动初始化使用。

如果要手动导入 DDL：

```bash
mysql -u root -p blog_db < sql/schema.sql
```

> 注意：当前 `schema.sql` 尾部的示例插入语句仅作参考，实际默认管理员和种子数据由 Go 程序初始化更可靠。

---

## 启动方式

### 1. 安装依赖

后端：

```bash
cd backend
go mod tidy
```

前端：

```bash
cd frontend
npm install
```

### 2. 构建前端资源

后端 HTTP 静态文件服务读取 `frontend/dist`：

```bash
cd frontend
npm run build
```

### 3. 启动 HTTP API / Web 服务

```bash
cd backend
go run . -api
```

默认端口为 `8080`，可通过环境变量覆盖：

```bash
API_PORT=8081 go run . -api
```

访问：

- `http://localhost:8080/#/blog`
- `http://localhost:8080/#/login`
- `http://localhost:8080/#/admin`

首次空库启动时会创建默认管理员：

```text
用户名：admin
密码：admin123
```

首次登录后请尽快修改密码。

### 4. 构建二进制

```bash
cd backend
go build -o blog-system-new .
./blog-system-new -api
```

### 5. Wails 桌面模式

```bash
cd backend
wails dev
```

或生产构建：

```bash
cd backend
wails build
```

> 代码中会先启动 HTTP API 服务，再进入 Wails 应用流程。当前更常用的是 `-api` 模式作为本地 Web 服务运行。

---

## 路由说明

### 前端路由

项目使用 hash 路由：

| 路径 | 页面 |
| --- | --- |
| `#/blog` | 博客首页 |
| `#/blog/about` | 关于页 |
| `#/blog/cat/:cat` | 分类文章列表 |
| `#/blog/:slug` | 文章详情 |
| `#/login` | 登录页 |
| `#/admin` | 后台仪表盘 |
| `#/admin/posts` | 文章管理 |
| `#/admin/editor` | 文章编辑器 |
| `#/admin/categories` | 分类管理 |
| `#/admin/tags` | 标签管理 |
| `#/admin/comments` | 评论管理 |
| `#/admin/users` | 用户管理 |

### 后端公开 API

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/public/auth/register` | 注册访客用户 |
| `POST` | `/api/public/auth/login` | 前台登录 |
| `GET` | `/api/public/posts` | 获取已发布文章列表，支持分页/分类/搜索/归档 |
| `GET` | `/api/public/posts/{slug}` | 根据 slug 获取文章详情 |
| `GET` | `/api/public/posts/{slug}/comments` | 获取文章评论 |
| `POST` | `/api/public/posts/{slug}/comments` | 提交文章评论，需要登录 |
| `GET` | `/api/public/categories` | 获取公开分类及文章数量 |
| `GET` | `/api/public/archives` | 获取归档月份列表 |
| `GET` | `/api/public/comments/recent` | 获取最新评论 |

`/api/public/posts` 常用查询参数：

- `page`：页码，默认 `1`
- `pageSize`：每页数量，默认 `10`
- `category`：分类 slug
- `search`：关键词
- `archive`：归档月份，格式 `YYYY-MM`

### 后台 API

以下接口除登录外均需要 JWT：

```http
Authorization: Bearer <token>
```

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/auth/login` | 后台登录 |
| `GET` | `/api/stats` | 获取统计数据 |
| `GET` | `/api/posts` | 获取所有文章 |
| `POST` | `/api/posts` | 创建文章 |
| `GET` | `/api/posts/{id}` | 获取单篇文章 |
| `PUT` | `/api/posts/{id}` | 更新文章 |
| `DELETE` | `/api/posts/{id}` | 删除文章 |
| `GET` | `/api/categories` | 获取分类 |
| `POST` | `/api/categories` | 创建分类 |
| `PUT` | `/api/categories/{id}` | 更新分类 |
| `DELETE` | `/api/categories/{id}` | 删除分类 |
| `GET` | `/api/tags` | 获取标签 |
| `POST` | `/api/tags` | 创建标签 |
| `DELETE` | `/api/tags/{id}` | 删除标签 |
| `GET` | `/api/comments` | 获取所有评论 |
| `PUT` | `/api/comments/{id}/status` | 更新评论状态 |
| `DELETE` | `/api/comments/{id}` | 删除评论 |
| `POST` | `/api/upload` | 上传图片 |
| `GET` | `/api/users` | 获取用户列表，管理员限定 |
| `PUT` | `/api/users/{id}/password` | 强制修改用户密码，管理员限定 |
| `PUT` | `/api/users/{id}/role` | 修改用户角色，管理员限定 |
| `DELETE` | `/api/users/{id}` | 删除用户，管理员限定 |

上传后的文件通过以下路径访问：

```text
/uploads/<filename>
```

---

## 数据模型摘要

- `User`：用户、角色、状态、头像、简介、最后登录时间。
- `Category`：分类、slug、父分类、排序、扩展元数据。
- `Tag`：标签、slug、扩展元数据。
- `Post`：文章标题、slug、摘要、Markdown、HTML、状态、分类、标签、发布时间、浏览量。
- `PostTagRelation`：文章与标签多对多关联。
- `Comment`：文章评论、登录用户/访客信息、父评论、审核状态。

文章状态：

- `draft`
- `pending`
- `published`
- `private`
- `trash`

评论状态：

- `pending`
- `approved`
- `spam`
- `trash`

用户角色：

- `admin`
- `editor`
- `author`
- `subscriber`

---

## 常用开发命令

前端开发服务器：

```bash
cd frontend
npm run dev
```

前端构建：

```bash
cd frontend
npm run build
```

后端运行：

```bash
cd backend
go run . -api
```

后端构建：

```bash
cd backend
go build -o blog-system-new .
```

运行测试：

```bash
cd backend
go test ./...
```

快速初始化脚本：

```bash
./setup.sh
```

---

## 部署提示

1. 准备 MariaDB，并创建 `blog_db`。
2. 配置 `.env`，尤其是 `DB_PASSWORD` 和 `JWT_SECRET`。
3. 执行 `frontend npm run build` 生成 `frontend/dist`。
4. 在 `backend` 目录执行 `go build -o blog-system-new .`。
5. 运行 `./blog-system-new -api`。
6. 使用 Nginx/Caddy 反向代理到 `localhost:8080`（如需公网访问）。
7. 首次登录默认管理员后，立即修改默认密码。

---

## 注意事项

- 当前项目主要面向个人博客和本地/小型部署。
- `JWT_SECRET` 未固定会导致每次重启后登录状态失效。
- 上传目录为 `backend/uploads`，部署时需要保证进程有写入权限，并做好备份。
- 文章正文会保存 Markdown，同时缓存渲染后的 HTML。
- 后端 Markdown 渲染开启了 `html.WithUnsafe()`，如果开放给不可信作者，需要补充 HTML 清洗策略。
- `sql/schema.sql` 是结构参考；实际运行时以 GORM 自动迁移和代码中的外键重建逻辑为准。

---

## License

MIT License

---

## 作者

**小龙** - 由景龙召唤的全栈开发助手 🐉
