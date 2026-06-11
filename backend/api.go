package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// APIServer HTTP API 服务（供浏览器开发模式使用）
type APIServer struct {
	blog   *BlogService
	auth   *AuthService
	upload *UploadService
}

func NewAPIServer() *APIServer {
	return &APIServer{
		blog:   NewBlogService(),
		auth:   NewAuthService(),
		upload: NewUploadService(),
	}
}

func getAPIPort() string {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

// StartAPIServer 启动 HTTP API 服务
func StartAPIServer() {
	server := NewAPIServer()

	// 最关键的修改：不使用 ServeMux，直接使用简单的处理器
	// 这样完全避免了 Go http 包的自动重定向功能
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// 公开博客 API：供博客首页/详情页从 MariaDB 读取真实数据
			if r.URL.Path == "/api/public/auth/register" && r.Method == http.MethodPost {
				server.handleRegister(w, r)
				return
			} else if r.URL.Path == "/api/public/auth/login" && r.Method == http.MethodPost {
				server.handleLogin(w, r)
				return
			} else if r.URL.Path == "/api/public/posts" && r.Method == http.MethodGet {
				server.handleGetPublishedPosts(w, r)
				return
			} else if r.URL.Path == "/api/public/categories" && r.Method == http.MethodGet {
				server.handleGetPublicCategories(w, r)
				return
			} else if r.URL.Path == "/api/public/archives" && r.Method == http.MethodGet {
				server.handleGetPublicArchives(w, r)
				return
			} else if r.URL.Path == "/api/public/comments/recent" && r.Method == http.MethodGet {
				server.handleGetRecentComments(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/public/posts/") && strings.HasSuffix(r.URL.Path, "/comments") && r.Method == http.MethodGet {
				server.handleGetPostComments(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/public/posts/") && strings.HasSuffix(r.URL.Path, "/comments") && r.Method == http.MethodPost {
				server.handleCreatePostComment(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/public/posts/") && r.Method == http.MethodGet {
				server.handleGetPostBySlug(w, r)
				return
			}

			// 后台 API 路由处理 - 手动匹配方法和路径
			if r.URL.Path == "/api/auth/login" && r.Method == http.MethodPost {
				server.handleLogin(w, r)
				return
			} else if r.URL.Path == "/api/stats" && r.Method == http.MethodGet {
				server.withAuth(server.handleGetStats)(w, r)
				return
			} else if r.URL.Path == "/api/users" && r.Method == http.MethodGet {
				server.withAdminAuth(server.handleGetUsers)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/users/") && strings.HasSuffix(r.URL.Path, "/password") && r.Method == http.MethodPut {
				server.withAdminAuth(server.handleForceUpdateUserPassword)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/users/") && strings.HasSuffix(r.URL.Path, "/role") && r.Method == http.MethodPut {
				server.withAdminAuth(server.handleUpdateUserRole)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/users/") && r.Method == http.MethodDelete {
				server.withAdminAuth(server.handleDeleteUser)(w, r)
				return
			} else if r.URL.Path == "/api/posts" && r.Method == http.MethodGet {
				server.withAuth(server.handleGetAllPosts)(w, r)
				return
			} else if r.URL.Path == "/api/posts" && r.Method == http.MethodPost {
				server.withAuth(server.handleCreatePost)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/posts/") && r.Method == http.MethodGet {
				server.withAuth(server.handleGetPost)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/posts/") && r.Method == http.MethodPut {
				server.withAuth(server.handleUpdatePost)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/posts/") && r.Method == http.MethodDelete {
				server.withAuth(server.handleDeletePost)(w, r)
				return
			} else if r.URL.Path == "/api/categories" && r.Method == http.MethodGet {
				server.withAuth(server.handleGetCategories)(w, r)
				return
			} else if r.URL.Path == "/api/categories" && r.Method == http.MethodPost {
				server.withAuth(server.handleCreateCategory)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/categories/") && r.Method == http.MethodPut {
				server.withAuth(server.handleUpdateCategory)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/categories/") && r.Method == http.MethodDelete {
				server.withAuth(server.handleDeleteCategory)(w, r)
				return
			} else if r.URL.Path == "/api/tags" && r.Method == http.MethodGet {
				server.withAuth(server.handleGetTags)(w, r)
				return
			} else if r.URL.Path == "/api/tags" && r.Method == http.MethodPost {
				server.withAuth(server.handleCreateTag)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/tags/") && r.Method == http.MethodDelete {
				server.withAuth(server.handleDeleteTag)(w, r)
				return
			} else if r.URL.Path == "/api/comments" && r.Method == http.MethodGet {
				server.withAuth(server.handleGetAllComments)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/comments/") && strings.HasSuffix(r.URL.Path, "/status") && r.Method == http.MethodPut {
				server.withAuth(server.handleUpdateCommentStatus)(w, r)
				return
			} else if strings.HasPrefix(r.URL.Path, "/api/comments/") && r.Method == http.MethodDelete {
				server.withAuth(server.handleDeleteComment)(w, r)
				return
			} else if r.URL.Path == "/api/upload" && r.Method == http.MethodPost {
				server.withAuth(server.handleUpload)(w, r)
				return
			}
			// 未匹配的API路径返回404
			writeError(w, http.StatusNotFound, "API路径未找到")
			return
		}

		if strings.HasPrefix(r.URL.Path, "/uploads/") {
			// 上传文件路由处理
			mux := http.NewServeMux()
			mux.Handle("/uploads/", &UploadHandler{UploadDir: "uploads"})
			mux.ServeHTTP(w, r)
			return
		}

		// 根路径直接返回 index.html
		if r.URL.Path == "/" {
			filePath := "./frontend/dist/index.html"
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("❌ 无法读取 index.html: %v", err)
				w.WriteHeader(404)
				w.Write([]byte("404 - Not Found"))
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Write(content)
			return
		}

		// 处理其他静态文件
		// 使用 filepath.Join 正确构建路径
		filePath := filepath.Join("./frontend/dist", strings.TrimPrefix(r.URL.Path, "/"))
		if _, err := os.Stat(filePath); err == nil {
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("❌ 无法读取文件: %v", err)
				w.WriteHeader(500)
				w.Write([]byte("500 - Internal Server Error"))
				return
			}
			// 开发/本地部署阶段禁用缓存，避免浏览器继续加载旧构建资源。
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			// 设置内容类型
			if strings.HasSuffix(filePath, ".css") {
				w.Header().Set("Content-Type", "text/css")
			} else if strings.HasSuffix(filePath, ".js") {
				w.Header().Set("Content-Type", "application/javascript")
			} else if strings.HasSuffix(filePath, ".json") {
				w.Header().Set("Content-Type", "application/json")
			} else if strings.HasSuffix(filePath, ".png") {
				w.Header().Set("Content-Type", "image/png")
			} else if strings.HasSuffix(filePath, ".jpg") || strings.HasSuffix(filePath, ".jpeg") {
				w.Header().Set("Content-Type", "image/jpeg")
			} else if strings.HasSuffix(filePath, ".gif") {
				w.Header().Set("Content-Type", "image/gif")
			} else if strings.HasSuffix(filePath, ".svg") {
				w.Header().Set("Content-Type", "image/svg+xml")
			} else if strings.HasSuffix(filePath, ".ico") {
				w.Header().Set("Content-Type", "image/x-icon")
			}
			w.Write(content)
			return
		}

		// 处理 SPA 路由（返回 index.html）
		spafilePath := "./frontend/dist/index.html"
		spacontent, spaerr := os.ReadFile(spafilePath)
		if spaerr != nil {
			log.Printf("❌ SPA 路由: 无法读取 index.html: %v", spaerr)
			w.WriteHeader(404)
			w.Write([]byte("404 - Not Found"))
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Write(spacontent)
	})

	addr := ":" + getAPIPort()
	log.Printf("🌐 HTTP API 服务已启动: http://localhost%s/api", addr)
	log.Printf("🎨 前端静态文件服务: http://localhost%s", addr)
	log.Printf("✅ 服务已准备完毕，访问: http://localhost%s", addr)

	// 直接启动服务器
	if err := http.ListenAndServe(addr, server.corsMiddleware(handler)); err != nil {
		log.Fatalf("❌ API 服务启动失败: %v", err)
	}
}

func (s *APIServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type apiHandler func(http.ResponseWriter, *http.Request)

func (s *APIServer) withAuth(handler apiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := claimsFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		var user User
		if err := s.blog.db.Select("id", "role", "status").First(&user, claims.UserID).Error; err != nil {
			writeError(w, http.StatusUnauthorized, "用户不存在")
			return
		}
		if user.Status != "active" {
			writeError(w, http.StatusForbidden, "账号不可用")
			return
		}
		if user.Role == "subscriber" {
			writeError(w, http.StatusForbidden, "访客没有后台访问权限")
			return
		}
		handler(w, r)
	}
}

func (s *APIServer) withAdminAuth(handler apiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := claimsFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		var user User
		if err := s.blog.db.Select("id", "role", "status").First(&user, claims.UserID).Error; err != nil {
			writeError(w, http.StatusUnauthorized, "用户不存在")
			return
		}
		if user.Status != "active" {
			writeError(w, http.StatusForbidden, "账号不可用")
			return
		}
		if user.Role != "admin" {
			writeError(w, http.StatusForbidden, "需要管理员权限")
			return
		}
		handler(w, r)
	}
}

func claimsFromRequest(r *http.Request) (*Claims, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return nil, fmt.Errorf("未授权，请先登录")
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	claims, err := ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("Token 无效或已过期")
	}
	return claims, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseID(r *http.Request) (uint, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		// 手动路由模式下没有 ServeMux path variables，从路径末尾向前找第一个数字段。
		// 支持 /api/posts/8、/api/tags/23、/api/comments/1/status 等路径。
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		for i := len(parts) - 1; i >= 0; i-- {
			if _, err := strconv.ParseUint(parts[i], 10, 32); err == nil {
				idStr = parts[i]
				break
			}
		}
	}
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("无效的 ID")
	}
	return uint(id), nil
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	resp, err := s.auth.Login(req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	resp, err := s.auth.Register(req.Username, req.Email, req.Password, req.DisplayName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

type userDTO struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	DisplayName string     `json:"display_name"`
	AvatarURL   string     `json:"avatar_url"`
	Bio         string     `json:"bio"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func toUserDTO(user User) userDTO {
	return userDTO{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Bio:         user.Bio,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func (s *APIServer) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	if err := s.blog.db.Order("created_at DESC").Find(&users).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "查询用户失败: "+err.Error())
		return
	}
	dtos := make([]userDTO, 0, len(users))
	for _, user := range users {
		dtos = append(dtos, toUserDTO(user))
	}
	writeJSON(w, http.StatusOK, dtos)
}

type updateUserRoleRequest struct {
	Role string `json:"role"`
}

func (s *APIServer) handleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var req updateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	role := strings.TrimSpace(req.Role)
	allowed := map[string]bool{"admin": true, "editor": true, "author": true, "subscriber": true}
	if !allowed[role] {
		writeError(w, http.StatusBadRequest, "无效角色")
		return
	}
	var existing User
	if err := s.blog.db.First(&existing, id).Error; err != nil {
		writeError(w, http.StatusNotFound, "用户不存在")
		return
	}
	if existing.Role == "admin" && role != "admin" {
		var adminCount int64
		if err := s.blog.db.Model(&User{}).Where("role = ?", "admin").Count(&adminCount).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "检查管理员数量失败: "+err.Error())
			return
		}
		if adminCount <= 1 {
			writeError(w, http.StatusBadRequest, "不能降级最后一个管理员")
			return
		}
	}
	if err := s.blog.db.Model(&User{}).Where("id = ?", id).Update("role", role).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "更新角色失败: "+err.Error())
		return
	}
	var user User
	if err := s.blog.db.First(&user, id).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "读取用户失败: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toUserDTO(user))
}

type updateUserPasswordRequest struct {
	Password string `json:"password"`
}

func (s *APIServer) handleForceUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var req updateUserPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "密码长度至少 6 位")
		return
	}
	hash, err := hashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "密码加密失败: "+err.Error())
		return
	}
	if err := s.blog.db.Model(&User{}).Where("id = ?", id).Update("password_hash", hash).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "修改密码失败: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *APIServer) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	claims, err := claimsFromRequest(r)
	if err == nil && claims.UserID == id {
		writeError(w, http.StatusBadRequest, "不能删除当前登录的管理员账号")
		return
	}
	var user User
	if err := s.blog.db.First(&user, id).Error; err != nil {
		writeError(w, http.StatusNotFound, "用户不存在")
		return
	}
	if user.Role == "admin" {
		var adminCount int64
		if err := s.blog.db.Model(&User{}).Where("role = ?", "admin").Count(&adminCount).Error; err != nil {
			writeError(w, http.StatusInternalServerError, "检查管理员数量失败: "+err.Error())
			return
		}
		if adminCount <= 1 {
			writeError(w, http.StatusBadRequest, "不能删除最后一个管理员")
			return
		}
	}
	var postCount int64
	if err := s.blog.db.Model(&Post{}).Where("user_id = ?", id).Count(&postCount).Error; err != nil {
		writeError(w, http.StatusInternalServerError, "检查用户文章失败: "+err.Error())
		return
	}
	if postCount > 0 {
		writeError(w, http.StatusBadRequest, "该用户名下仍有文章，不能直接删除")
		return
	}
	if err := s.blog.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Comment{}).Where("user_id = ?", id).Update("user_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM sessions WHERE user_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&User{}, id).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "删除用户失败: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *APIServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.blog.GetStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *APIServer) handleGetAllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := s.blog.GetAllPosts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, posts)
}

func (s *APIServer) handleGetPublishedPosts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	categorySlug := strings.TrimSpace(r.URL.Query().Get("category"))
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	archive := strings.TrimSpace(r.URL.Query().Get("archive"))
	posts, total, err := s.blog.GetPublishedPostsByCategory(page, pageSize, categorySlug, search, archive)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"posts": posts, "total": total})
}

func (s *APIServer) handleGetPublicCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.blog.GetPublishedCategories()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

func (s *APIServer) handleGetPublicArchives(w http.ResponseWriter, r *http.Request) {
	archives, err := s.blog.GetPublishedArchives()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, archives)
}

func (s *APIServer) handleGetRecentComments(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 20 {
		limit = 5
	}
	comments, err := s.blog.GetLatestApprovedComments(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

func (s *APIServer) handleGetPostBySlug(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/api/public/posts/")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "slug 不能为空")
		return
	}
	post, err := s.blog.GetPostBySlug(slug)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, post)
}

func publicPostCommentSlug(path string) string {
	slug := strings.TrimPrefix(path, "/api/public/posts/")
	slug = strings.TrimSuffix(slug, "/comments")
	return strings.Trim(slug, "/")
}

func (s *APIServer) handleGetPostComments(w http.ResponseWriter, r *http.Request) {
	slug := publicPostCommentSlug(r.URL.Path)
	comments, err := s.blog.GetCommentsByPostSlug(slug)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

type commentRequest struct {
	Content string `json:"content"`
}

func (s *APIServer) handleCreatePostComment(w http.ResponseWriter, r *http.Request) {
	claims, err := claimsFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req commentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	slug := publicPostCommentSlug(r.URL.Path)
	comment, err := s.blog.CreateComment(slug, claims.UserID, req.Content)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, comment)
}

func (s *APIServer) handleGetPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	post, err := s.blog.GetPostByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "文章不存在")
		return
	}
	writeJSON(w, http.StatusOK, post)
}

type postRequest struct {
	Title      string   `json:"title"`
	Slug       string   `json:"slug"`
	Content    string   `json:"content"`
	CategoryID uint     `json:"category_id"`
	Tags       []string `json:"tags"`
	Status     string   `json:"status"`
}

func (s *APIServer) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	var req postRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	id, err := s.blog.CreatePost(req.Title, req.Slug, req.Content, req.CategoryID, req.Tags, req.Status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]uint{"id": id})
}

func (s *APIServer) handleUpdatePost(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var req postRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.blog.UpdatePost(id, req.Title, req.Slug, req.Content, req.CategoryID, req.Tags, req.Status); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "更新成功"})
}

func (s *APIServer) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.blog.DeletePost(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

func (s *APIServer) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.blog.GetCategories()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, categories)
}

type categoryRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

func (s *APIServer) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.blog.CreateCategory(req.Name, req.Slug, req.Description); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "创建成功"})
}

func (s *APIServer) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var req categoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.blog.UpdateCategory(id, req.Name, req.Slug, req.Description); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "更新成功"})
}

func (s *APIServer) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.blog.DeleteCategory(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

func (s *APIServer) handleGetTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.blog.GetTags()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tags)
}

type tagRequest struct {
	Name string `json:"name"`
}

func (s *APIServer) handleCreateTag(w http.ResponseWriter, r *http.Request) {
	var req tagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.blog.CreateTag(req.Name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "创建成功"})
}

func (s *APIServer) handleDeleteTag(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.blog.DeleteTag(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

func (s *APIServer) handleGetAllComments(w http.ResponseWriter, r *http.Request) {
	comments, err := s.blog.GetAllComments()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

type commentStatusRequest struct {
	Status string `json:"status"`
}

func (s *APIServer) handleUpdateCommentStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var req commentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.blog.UpdateCommentStatus(id, req.Status); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "更新成功"})
}

func (s *APIServer) handleDeleteComment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.blog.DeleteComment(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

func (s *APIServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "文件过大或格式错误")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "未找到上传文件")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取文件失败")
		return
	}

	result, err := s.upload.UploadImage(header.Filename, data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
