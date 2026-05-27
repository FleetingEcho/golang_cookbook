package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// ── Request ID 注入中间件 ──────────────────────────────────────────────────
//
// 工程要点：
// 1. 每个请求分配唯一 ID，贯穿整个请求生命周期
// 2. 写入响应头 X-Request-Id，方便客户端关联请求和日志
// 3. 存入 context，可在 handler 中通过 GetRequestID 取出
// 4. 将 request_id 加入 slog 的 Logger，让日志自动携带该字段

type contextKey string

const RequestIDKey contextKey = "request_id"

// InjectRequestID 给每个请求注入 X-Request-Id。
func InjectRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果上游已经传递了请求 ID（如网关），优先复用
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}

		w.Header().Set("X-Request-Id", id)

		// 将 request_id 存入 context
		ctx := context.WithValue(r.Context(), RequestIDKey, id)

		// 创建一个携带 request_id 的新 logger
		// 后续 handler 用 slog.Default() 时会自动附带该字段
		logger := slog.Default().With("request_id", id)

		// 将 logger 存入 context
		ctx = context.WithValue(ctx, loggerKey{}, logger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loggerKey 是用于在 context 中存储/读取 logger 的 key 类型。
// 使用自定义类型而非 string，避免与其它包的 context key 冲突。
type loggerKey struct{}

// GetLogger 从 context 中取出携带 request_id 的 logger。
//
// 使用方式：
//
//	logger := middleware.GetLogger(r.Context())
//	logger.Info("doing something")
func GetLogger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// GetRequestID 从 context 中取出请求 ID。
func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDKey).(string)
	return id
}
