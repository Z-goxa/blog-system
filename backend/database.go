package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 数据库配置
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig 返回默认配置（从环境变量读取）
func DefaultConfig() *Config {
	// 尝试在多个路径加载 .env 文件
	envPaths := []string{".env", "../.env", "backend/.env"}
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("✅ 已加载环境变量文件: %s", path)
			break
		}
	}

	// 从环境变量读取配置，如果不存在则使用默认值
	host := getEnv("DB_HOST", "localhost")
	port := getEnvAsInt("DB_PORT", 3306)
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "blog_db")
	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetime := getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)
	connMaxIdleTime := getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 120)

	return &Config{
		Host:            host,
		Port:            port,
		User:            user,
		Password:        password,
		DBName:          dbName,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: time.Duration(connMaxLifetime) * time.Second,
		ConnMaxIdleTime: time.Duration(connMaxIdleTime) * time.Second,
	}
}

// getEnv 读取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 读取环境变量并转换为 int
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}

var dbInstance *gorm.DB

// InitDB 初始化数据库连接（带连接池）
func InitDB(cfg *Config) (*gorm.DB, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 构建 DSN（MariaDB 兼容 MySQL 格式）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	log.Printf("🔌 正在连接数据库: %s@%s:%d/%s", cfg.User, cfg.Host, cfg.Port, cfg.DBName)

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		PrepareStmt: true,
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取通用数据库对象 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取 sql.DB 失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	log.Println("✅ 数据库连接成功")
	
	dbInstance = db
	
	return db, nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return dbInstance
}

// SetDB 设置数据库实例
func SetDB(db *gorm.DB) {
	dbInstance = db
}
