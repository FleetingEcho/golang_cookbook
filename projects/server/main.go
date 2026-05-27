// ── Swagger 元数据 ─────────────────────────────────────────────────────────
//
// 使用 swaggo 自动生成 OpenAPI 文档。
// 生成命令： swag init -g main.go --output docs/
// 访问地址： http://localhost:3001/swagger/index.html

// @title           Issue Tracker API
// @version         1.0.0
// @description     A production-style Go REST API for tracking issues, comments, labels, and file attachments.
// @host            localhost:3001
// @BasePath        /api
// @securityDefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        x-api-key
// @description                 API key for authentication. Default: dev-secret

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"issue_tracker/config"
	"issue_tracker/db"
)

func main() {
	// ── 1. Fail Fast 配置 ───────────────────────────────────────────
	cfg := config.MustLoad()

	// ── 2. 初始化结构化日志 ─────────────────────────────────────────
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     parseLogLevel(cfg.LogLevel),
		AddSource: true,
	})))

	slog.Info("starting server",
		"version", runtime.Version(),
		"bind", cfg.BindAddr,
		"upload_dir", cfg.UploadDir,
		"max_upload_mb", cfg.MaxUploadBytes/(1<<20),
	)

	// ── 3. 连接数据库 ───────────────────────────────────────────────
	database, err := db.Connect(cfg.DatabaseURL, cfg.DBMaxOpenConns, cfg.DBMaxIdleConns)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// ── 4. 创建上传目录 ─────────────────────────────────────────────
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		slog.Error("create upload directory failed", "path", cfg.UploadDir, "error", err)
		os.Exit(1)
	}

	// ── 5. 构建 HTTP 处理器 ─────────────────────────────────────────
	handler := buildApp(cfg, database, slog.Default())

	// ── 6. HTTP 服务器配置 ──────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.BindAddr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// ── 7. 优雅关闭 ─────────────────────────────────────────────────
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh

		slog.Info("shutting down", "signal", sig.String())

		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("shutdown timed out, forcing close", "error", err)
			if closeErr := srv.Close(); closeErr != nil {
				slog.Error("force close failed", "error", closeErr)
			}
		}
		database.Close()
		slog.Info("server stopped")
	}()

	// ── 8. 启动 ─────────────────────────────────────────────────────
	slog.Info("listening", "addr", cfg.BindAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

// parseLogLevel 将字符串日志级别转为 slog.Level。
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
