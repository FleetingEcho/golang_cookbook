package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"issue_tracker/config"
	_ "issue_tracker/docs"  // registers swagger JSON via init()
	"issue_tracker/middleware"
	"issue_tracker/routes"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// buildApp 构建 HTTP 路由 + 中间件链。
//
// 设计模式：中间件从外到内依次包裹。
// 最先 Use 的中间件包裹在最外层（第一个处理请求，最后一个处理响应）。
//
// 中间件顺序：
// 1. RequestID / Recoverer（最外层：保障每个请求都有 ID，panic 不 crash）
// 2. RealIP / Logger（记录请求信息）
// 3. Timeout / CORS（请求进入业务前拦截）
// 4. RateLimit / Auth（业务安全层）
// 5. Routes（最内层：实际业务处理）
func buildApp(cfg *config.Config, db *sql.DB, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	// ── 全局中间件 ──────────────────────────────────────────────────
	// chi 内置中间件：
	//   RequestID  — 注入 X-Request-Id（如果上游未传递）
	//   RealIP     — 从 X-Forwarded-For / X-Real-IP 解析真实 IP
	//   Recoverer  — 捕获 handler panic，返回 500，防止进程崩溃
	r.Use(chimw.RequestID)
	r.Use(middleware.InjectRequestID) // 我们自定义的 Request ID（也集成 slog）
	r.Use(chimw.RealIP)
	r.Use(structuredLogger)
	r.Use(chimw.Recoverer)

	// ── 条件中间件：超时 ────────────────────────────────────────────
	// chi 的 Timeout 中间件在超时后取消 context，
	// handler 内使用 r.Context() 的 DB 查询会被自动取消。
	r.Use(chimw.Timeout(30 * time.Second))

	// ── 条件中间件：CORS ────────────────────────────────────────────
	// CORS 必须在业务 handler 之前，否则浏览器预检请求（OPTIONS）不会通过。
	r.Use(corsMiddleware)

	// ── 条件中间件：限流 ────────────────────────────────────────────
	// 只有配置了限流才启用，默认不限制。
	if cfg.RateLimitPerSec > 0 {
		slog.Info("rate limiter enabled",
			"rate", cfg.RateLimitPerSec,
			"burst", cfg.RateLimitBurst)
		r.Use(middleware.RateLimit(cfg.RateLimitPerSec, cfg.RateLimitBurst))
	}

	// ── Health + Swagger（无需认证） ───────────────────────────────
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Swagger UI: http://localhost:3001/swagger/index.html
	// 生成命令: swag init -g main.go --output docs/
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// ── API（需要 x-api-key 认证） ─────────────────────────────────
	r.Route("/api", func(api chi.Router) {
		api.Use(middleware.RequireAPIKey(cfg.APIKey))

		api.Route("/issues", func(sub chi.Router) {
			routes.RegisterIssueRoutes(sub, db)
			sub.Route("/{issueID}/comments", func(r chi.Router) {
				routes.RegisterCommentRoutes(r, db)
			})
			sub.Route("/{issueID}/labels/{labelID}", func(r chi.Router) {
				routes.RegisterIssueLabelRoutes(r, db)
			})
			ah := &routes.AttachmentHandler{DB: db, UploadDir: cfg.UploadDir}
			sub.Route("/{issueID}/attachments", func(r chi.Router) {
				routes.RegisterAttachmentRoutes(r, ah)
			})
		})

		api.Route("/comments", func(r chi.Router) {
			routes.RegisterCommentDeleteRoute(r, db)
		})
		api.Route("/labels", func(r chi.Router) {
			routes.RegisterLabelRoutes(r, db)
		})
		api.Route("/attachments/{id}", func(r chi.Router) {
			routes.RegisterAttachmentDownloadRoute(r,
				&routes.AttachmentHandler{DB: db, UploadDir: cfg.UploadDir})
		})
	})

	// ── pprof（调试用，仅 debug 模式） ─────────────────────────────
	// 放在所有 Use + Route 之后，符合 chi 的要求（middleware 先于路由）。
	if cfg.LogLevel == "debug" {
		registerPProfOnChi(r)
	}

	// ── 记录已注册的路由 ───────────────────────────────────────────
	// 启动时打印路由表，方便调试
	walkFunc := func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.Info("route registered", "method", method, "path", route)
		return nil
	}
	if err := chi.Walk(r, walkFunc); err != nil {
		logger.Error("walk routes failed", "error", err)
	}

	return r
}

// registerPProfOnChi 在 chi router 上注册 pprof 端点。
//
// pprof 端点列表：
//   /debug/pprof/          — 首页（go tool pprof ...）
//   /debug/pprof/heap     — 堆内存分配
//   /debug/pprof/goroutine — 所有 goroutine 栈
//   /debug/pprof/profile  — CPU 分析（30 秒采样）
//   /debug/pprof/trace?seconds=5 — 执行追踪
//
// 使用方式：
//   go tool pprof http://127.0.0.1:3001/debug/pprof/heap
//   go tool trace http://127.0.0.1:3001/debug/pprof/trace?seconds=5
//
// 注意：pprof 仅在 debug 日志级别启用，生产环境不暴露。
func registerPProfOnChi(r *chi.Mux) {
	r.HandleFunc("/debug/pprof/*", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	slog.Info("pprof endpoints registered",
		"paths", []string{
			"/debug/pprof/",
			"/debug/pprof/heap",
			"/debug/pprof/goroutine",
			"/debug/pprof/profile",
			"/debug/pprof/trace",
		})
}

// ── CORS 中间件 ───────────────────────────────────────────────────────────
//
// 工程要点：
// 1. CORS 是跨域安全机制，不要用 `Access-Control-Allow-Origin: *`
// 2. 只允许已知的开发来源（localhost）
// 3. OPTIONS 预检请求直接返回 204，不经过业务 handler

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowed := origin == "http://localhost:5173" ||
			origin == "http://localhost:3001" ||
			origin == "http://127.0.0.1:5173" ||
			origin == "http://127.0.0.1:3001"

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers",
				"Content-Type, Accept, Authorization, X-Api-Key")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ── 结构化日志中间件 ──────────────────────────────────────────────────────
//
// 记录每个 HTTP 请求的方法、路径、状态码、耗时和 body 大小。
// 使用 slog.LogAttrs 而非 slog.Info，性能更好（减少内存分配）。

func structuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// NewWrapResponseWriter 包裹 ResponseWriter，捕获状态码和 body 大小
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

		// 执行实际 handler
		next.ServeHTTP(ww, r)

		// 根据状态码选择日志级别
		status := ww.Status()
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.LogAttrs(r.Context(), level, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.Int("bytes", ww.BytesWritten()),
			slog.Duration("duration", time.Since(start)),
			slog.String("remote", r.RemoteAddr),
		)
	})
}
