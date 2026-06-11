package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// jwtSecret 从环境变量 JWT_SECRET 读取；未设置时生成一次性随机密钥并警告。
var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
)

func getJWTSecret() []byte {
	jwtSecretOnce.Do(func() {
		if s := os.Getenv("JWT_SECRET"); s != "" {
			jwtSecret = []byte(s)
			return
		}
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			log.Fatalf("❌ 生成临时 JWT 密钥失败: %v", err)
		}
		jwtSecret = []byte(hex.EncodeToString(buf))
		log.Println("⚠️  未设置 JWT_SECRET，已生成临时密钥；重启后现有 token 会全部失效。请在 .env 中配置 JWT_SECRET=。")
	})
	return jwtSecret
}

// AuthService 认证服务
type AuthService struct {
	db *gorm.DB
}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{
		db: GetDB(),
	}
}

// Claims JWT 声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Register 注册访客用户
func (s *AuthService) Register(username, email, password, displayName string) (*LoginResponse, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	displayName = strings.TrimSpace(displayName)
	if username == "" || email == "" || password == "" {
		return nil, fmt.Errorf("用户名、邮箱和密码不能为空")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("密码长度至少 6 位")
	}
	if displayName == "" {
		displayName = username
	}

	var count int64
	if err := s.db.Model(&User{}).Where("username = ? OR email = ?", username, email).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查用户失败: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("用户名或邮箱已存在")
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}
	user := User{
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Role:         "subscriber",
		DisplayName:  displayName,
		Status:       "active",
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("注册失败: %w", err)
	}
	token, err := generateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}
	return &LoginResponse{Token: token, User: user}, nil
}

// Login 用户登录
func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
	var user User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 验证密码
	if err := verifyPassword(user.PasswordHash, password); err != nil {
		return nil, fmt.Errorf("密码错误")
	}

	// 生成 JWT Token
	token, err := generateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("生成 token 失败: %w", err)
	}

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// generateToken 生成 JWT Token
func generateToken(user *User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证 Token
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// hashPassword 对密码进行哈希处理
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// verifyPassword 验证密码
func verifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
