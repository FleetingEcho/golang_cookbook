package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"issue_tracker/config"
	"issue_tracker/middleware"
	"issue_tracker/routes"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func buildApp(cfg *config.Config, db *sql.DB) http.Handler {
	r := chi.NewRouter()

	// 全局中间件
	r.Use(chimw.RequestID)
	r.Use(middleware.InjectRequestID)
	r.Use(chimw.RealIP)
	r.Use(structuredLogger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	// Health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API
	r.Route("/api", func(api chi.Router) {
		api.Use(middleware.RequireAPIKey(cfg.APIKey))

		api.Route("/issues", func(sub chi.Router) {
			routes.RegisterIssueRoutes(sub, db)
			sub.Route("/{issueID}/comments", func(r chi.Router) { routes.RegisterCommentRoutes(r, db) })
			sub.Route("/{issueID}/labels/{labelID}", func(r chi.Router) { routes.RegisterIssueLabelRoutes(r, db) })
			ah := &routes.AttachmentHandler{DB: db, UploadDir: cfg.UploadDir}
			sub.Route("/{issueID}/attachments", func(r chi.Router) { routes.RegisterAttachmentRoutes(r, ah) })
		})

		api.Route("/comments", func(r chi.Router) { routes.RegisterCommentDeleteRoute(r, db) })
		api.Route("/labels", func(r chi.Router) { routes.RegisterLabelRoutes(r, db) })
		api.Route("/attachments/{id}", func(r chi.Router) {
			routes.RegisterAttachmentDownloadRoute(r, &routes.AttachmentHandler{DB: db, UploadDir: cfg.UploadDir})
		})
	})

	return r
}

// ── CORS ─────────────────────────────────────────────────────────────────────

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := origin == "http://localhost:5173" || origin == "http://localhost:3001" ||
			origin == "http://127.0.0.1:5173" || origin == "http://127.0.0.1:3001"
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, X-Api-Key")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── 结构化日志 ─────────────────────────────────────────────────────────────────

func structuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		level := slog.LevelInfo
		s := ww.Status()
		if s >= 500 {
			level = slog.LevelError
		} else if s >= 400 {
			level = slog.LevelWarn
		}
		slog.LogAttrs(r.Context(), level, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", s),
			slog.Int("bytes", ww.BytesWritten()),
			slog.Duration("duration", time.Since(start)),
			slog.String("request_id", middleware.GetRequestID(r.Context())),
		)
	})
}
