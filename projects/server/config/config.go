package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config 是应用的全局配置。
//
// 设计原则：
// 1. 所有配置来自环境变量，不引入配置文件（12-Factor App）
// 2. 优先使用带默认值的 safe getter，而不是直接 os.Getenv
// 3. 配置验证集中在 MustLoad 中，fail fast
type Config struct {
	BindAddr    string        // 监听地址，如 "127.0.0.1:3001"
	DatabaseURL string        // SQLite 路径，如 "sqlite://./data/db.sqlite"
	UploadDir   string        // 上传文件存储目录
	APIKey      string        // API 鉴权密钥

	// 以下为可选的进阶配置，从环境变量读取
	MaxUploadBytes   int64         // 单文件最大字节数，默认 10MB
	ReadTimeout      time.Duration // HTTP 读超时
	WriteTimeout     time.Duration // HTTP 写超时
	IdleTimeout      time.Duration // HTTP 空闲超时
	ShutdownTimeout  time.Duration // 优雅关闭等待时间
	DBMaxOpenConns   int           // 数据库最大连接数
	DBMaxIdleConns   int           // 数据库最大空闲连接数
	RateLimitPerSec  int           // 每秒最多请求数（0=不限制）
	RateLimitBurst   int           // 突发允许的请求数
	LogLevel         string        // 日志级别：debug, info, warn, error
}

// getEnv 是带默认值的环境变量读取器。
// 这是 Go 中常见的工具函数模式：避免在每个字段处写 if/else。
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getEnvInt 将环境变量解析为 int，失败时返回默认值。
//
// 使用场景：数据库连接数、超时秒数等整型配置。
// 注意：环境变量全是字符串，解析错误不能 panic，要静默 fallback。
func getEnvInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

// Load 读取环境变量并返回配置。
//
// 设计模式：将所有配置读取集中在一处，不散落在各模块中。
// 这样新开发者看这一个文件就知道"这个应用需要配什么"。
func Load() *Config {
	cwd, _ := os.Getwd()

	// ── 核心配置 ──────────────────────────────────────────────────────

	bind := getEnv("ISSUE_TRACKER_BIND_ADDR", "127.0.0.1:3001")

	dbURL := getEnv("DATABASE_URL",
		"sqlite://"+filepath.Join(cwd, "data", "issue_tracker.db"))

	uploadDir := getEnv("ISSUE_TRACKER_UPLOAD_DIR",
		filepath.Join(cwd, "uploads"))

	apiKey := getEnv("ISSUE_TRACKER_API_KEY", "dev-secret")

	// ── 进阶配置 ──────────────────────────────────────────────────────

	return &Config{
		BindAddr:         bind,
		DatabaseURL:       dbURL,
		UploadDir:         uploadDir,
		APIKey:            apiKey,
		MaxUploadBytes:    int64(getEnvInt("ISSUE_TRACKER_MAX_UPLOAD_BYTES", 10<<20)),   // 10 MB
		ReadTimeout:       time.Duration(getEnvInt("ISSUE_TRACKER_READ_TIMEOUT_SEC", 10)) * time.Second,
		WriteTimeout:      time.Duration(getEnvInt("ISSUE_TRACKER_WRITE_TIMEOUT_SEC", 30)) * time.Second,
		IdleTimeout:       time.Duration(getEnvInt("ISSUE_TRACKER_IDLE_TIMEOUT_SEC", 60)) * time.Second,
		ShutdownTimeout:   time.Duration(getEnvInt("ISSUE_TRACKER_SHUTDOWN_TIMEOUT_SEC", 15)) * time.Second,
		DBMaxOpenConns:    getEnvInt("ISSUE_TRACKER_DB_MAX_OPEN_CONNS", 1),
		DBMaxIdleConns:    getEnvInt("ISSUE_TRACKER_DB_MAX_IDLE_CONNS", 1),
		RateLimitPerSec:   getEnvInt("ISSUE_TRACKER_RATE_LIMIT_PER_SEC", 0),
		RateLimitBurst:    getEnvInt("ISSUE_TRACKER_RATE_LIMIT_BURST", 10),
		LogLevel:          getEnv("ISSUE_TRACKER_LOG_LEVEL", "debug"),
	}
}

// MustLoad 加载配置并在验证失败时 panic。
//
// 设计原则：Fail Fast — 配置错误应该在进程启动时暴露，而不是在运行时。
// 如果绑定地址无效、端口被占用等，立即崩溃比运行到一半再挂更易调试。
func MustLoad() *Config {
	cfg := Load()

	// ── 配置验证 ──────────────────────────────────────────────────────

	// 验证监听地址格式
	if _, err := net.ResolveTCPAddr("tcp", cfg.BindAddr); err != nil {
		panic(fmt.Sprintf("ISSUE_TRACKER_BIND_ADDR invalid: %v", err))
	}

	// 验证上传限制（不能为负，不能超过 100MB 硬限制）
	if cfg.MaxUploadBytes <= 0 || cfg.MaxUploadBytes > 100<<20 {
		panic("ISSUE_TRACKER_MAX_UPLOAD_BYTES must be between 1 and 104857600")
	}

	// 验证日志级别
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		panic("ISSUE_TRACKER_LOG_LEVEL must be debug, info, warn, or error")
	}

	return cfg
}
