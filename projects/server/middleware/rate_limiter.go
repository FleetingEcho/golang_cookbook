package middleware

import (
	"net/http"
	"sync"
	"time"
)

// ── Token Bucket 限流器 ────────────────────────────────────────────────────
//
// 工程要点：
// 1. 使用 Token Bucket（令牌桶）算法：每秒填充 rate 个令牌，桶最多 burst 个
// 2. 比简单的计数器更平滑：允许短时突发流量
// 3. 单机限流，多实例需用 Redis 等集中式方案
// 4. 生产环境推荐用 golang.org/x/time/rate 标准库实现

type RateLimiter struct {
	mu       sync.Mutex   // 保护并发访问
	tokens   float64      // 当前令牌数
	lastRefill time.Time  // 上次补充时间
	rate     float64      // 每秒补充速率
	burst    float64      // 桶容量（最大突发）
}

// NewRateLimiter 创建一个令牌桶限流器。
//
// rate: 每秒允许的请求数
// burst: 允许的突发请求数（桶大小）
func NewRateLimiter(rate, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:    float64(burst),
		lastRefill: time.Now(),
		rate:      float64(rate),
		burst:     float64(burst),
	}
}

// Allow 判断当前请求是否允许通过。
//
// 每次调用会先补充令牌（基于时间流逝），然后消耗一个令牌。
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// 计算从上次补充到现在应该增加的令牌数
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.rate
	// 不能超过桶容量
	if rl.tokens > rl.burst {
		rl.tokens = rl.burst
	}
	rl.lastRefill = now

	// 如果还有令牌，消耗一个并放行
	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

// RateLimit 创建一个 HTTP 中间件，对匹配的请求进行限流。
//
// 返回 429 Too Many Requests + JSON 错误体。
func RateLimit(rate, burst int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(rate, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				w.Header().Set("Retry-After", "1")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ── 使用示例（在 app.go 中） ──────────────────────────────────────────────
//
// 如果 cfg.RateLimitPerSec > 0，可以这样启用：
//
//	if cfg.RateLimitPerSec > 0 {
//	    r.Use(middleware.RateLimit(cfg.RateLimitPerSec, cfg.RateLimitBurst))
//	}
