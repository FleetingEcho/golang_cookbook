# Go 大型项目实践指南

> 从几千行到几十万行，Go 项目的组织方式、常见陷阱和工程实践。
> 适用场景：团队 5 人以上、代码量 10 万行以上、多人协作的中大型后端项目。

---

## 目录

- [一、项目结构](#一项目结构)
- [二、包设计原则](#二包设计原则)
- [三、依赖管理](#三依赖管理)
- [四、错误处理策略](#四错误处理策略)
- [五、配置管理](#五配置管理)
- [六、测试金字塔](#六测试金字塔)
- [七、接口什么时候抽](#七接口什么时候抽)
- [八、并发模式](#八并发模式)
- [九、数据库层](#九数据库层)
- [十、日志、监控、可观测性](#十日志监控可观测性)
- [十一、HTTP 层](#十一http-层)
- [十二、值得用的第三方包](#十二值得用的第三方包)
- [十三、不要用的包](#十三不要用的包)
- [十四、常见的坑](#十四常见的坑)
- [十五、渐进式重构](#十五渐进式重构)

---

## 一、项目结构

### 不要这样按层分包

```
pkg/
├── models/        # 所有 struct
├── handlers/      # 所有 HTTP handler
├── services/      # 所有业务逻辑
├── repositories/  # 所有数据库操作
```

问题：每个目录几十上百个文件，查找困难，修改一个功能要跨 4 个目录。

### 要这样按功能分包

```
internal/
├── user/
│   ├── handler.go      # HTTP 层
│   ├── service.go      # 业务逻辑
│   ├── repository.go   # 数据库操作
│   ├── model.go        # user 相关的 struct
│   └── user_test.go    # 测试
├── order/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   ├── model.go
│   └── order_test.go
├── product/
│   └── ...
└── pkg/                # 真正跨多个模块复用的工具
    ├── config/
    ├── middleware/
    ├── pagination/
    └── testutil/
```

原则：**一个功能的所有代码在一个包里**。改 user 就只改 `internal/user/`，不需要跳 4 个目录。

### cmd/ 放入口

```
cmd/
├── server/main.go       # API 服务入口
├── worker/main.go       # 后台 worker 入口
└── migrate/main.go      # 数据库迁移 CLI
```

每个入口文件应该非常短——加载配置、初始化依赖、启动服务。所有逻辑在 `internal/` 里。

---

## 二、包设计原则

### 2.1 依赖方向：从外向内

```
cmd/                  → 依赖 internal/
internal/user/handler → 依赖 internal/user/service
internal/user/service → 依赖 internal/user/repository
                       → 依赖 internal/pkg/
```

不允许反向依赖。`internal/user/repository` 不能引用 `internal/order/service`。

### 2.2 循环依赖是头号敌人

```
services/user  →  services/order
services/order →  services/user  // ❌ 编译错误
```

解法：**接口在 consumer 侧定义**，依赖倒置。

```go
// services/user/service.go
type OrderChecker interface {
    HasActiveOrders(ctx context.Context, userID string) bool
}

type Service struct {
    orders OrderChecker
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
    if s.orders.HasActiveOrders(ctx, id) {
        return ErrUserHasOrders
    }
    // ...
}
```

`services/order` 实现 `OrderChecker`，不需要 `services/user` 知道自己被实现了。

### 2.3 包大小控制

| 指标 | 警戒线 | 说明 |
|------|--------|------|
| 单个文件行数 | < 500 行 | 超过考虑拆分 |
| 单个包文件数 | < 15 个 | 超过考虑按子功能分包 |
| 单个包公开函数 | < 50 个 | 超过考虑 API 面是否太大 |
| import 路径深度 | < 5 层 | `internal/a/b/c/d/e` → 需要重构 |

### 2.4 internal/ vs pkg/

```
internal/     — 私有代码，外部项目不能导入（Go 编译器强制）
pkg/          — 可公开复用的工具库
```

大多数代码放 `internal/`。只有当你有明确的"这个工具其他项目也会用"时才放 `pkg/`。

---

## 三、依赖管理

### 3.1 依赖注入（不用全局变量）

```go
// ❌ 全局变量——测试互相污染，没法并行
var db *sql.DB

func Handler(w http.ResponseWriter, r *http.Request) {
    rows, _ := db.Query(...)
}

// ✅ 依赖注入——测试传 mock
type Handler struct {
    db    *sql.DB
    cache *redis.Client
    log   *slog.Logger
}

func New(db *sql.DB, cache *redis.Client, log *slog.Logger) *Handler {
    return &Handler{db: db, cache: cache, log: log}
}
```

### 3.2 手动 DI vs Wire vs Fx

| 方式 | 适合场景 | 说明 |
|------|---------|------|
| 手动 DI | < 20 个依赖 | 在 main.go 里 new 所有对象，按顺序传 |
| wire | > 20 个依赖 | Google 的编译期 DI，生成代码 |
| fx | 需要生命周期管理 | Uber 的运行时 DI，带启动/停止钩子 |

推荐：**手动 DI 优先**。Go 的哲学是显式优于隐式。

```go
// cmd/server/main.go
func main() {
    cfg := config.Load()
    db := database.Connect(cfg.DatabaseURL)
    cache := redis.New(cfg.RedisAddr)
    logger := slog.New(...)

    userRepo := user.NewRepository(db)
    userSvc := user.NewService(userRepo, cache, logger)
    userHandler := user.NewHandler(userSvc)

    orderRepo := order.NewRepository(db)
    orderSvc := order.NewService(orderRepo, cache)
    orderHandler := order.NewHandler(orderSvc)

    router := buildRouter(userHandler, orderHandler)
    server.ListenAndServe(cfg.BindAddr, router)
}
```

20 行 main.go，所有依赖一目了然。不需要 wire 的玄学代码生成。

### 3.3 vendor 还是 module cache？

```
go mod vendor    # 把依赖拷到项目里
```

推荐：**CI/CD 环境用 vendor**。确保构建可重现，不依赖上游仓库可用性。
**开发环境**用 module cache（默认），不用 vendor。

---

## 四、错误处理策略

### 4.1 Wrap 错误，不要裸传

```go
// ❌ 裸传——调用方不知道错误来源
func GetUser(id string) (*User, error) {
    rows, err := db.Query(...)
    if err != nil {
        return nil, err
    }
}

// ✅ Wrap——带上上下文
func GetUser(ctx context.Context, id string) (*User, error) {
    rows, err := db.QueryContext(ctx, "SELECT ...", id)
    if err != nil {
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
}

// 日志处用 %+v 打印完整栈
if err != nil {
    log.Errorf("get user failed: %+v", err) // 打印完整 wrap 链
}
```

### 4.2 定义业务错误类型

```go
// internal/pkg/errors/errors.go
var (
    ErrNotFound      = errors.New("resource not found")
    ErrConflict      = errors.New("resource already exists")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrValidation    = errors.New("validation failed")
    ErrRateLimited   = errors.New("rate limit exceeded")
)

// 带业务上下文的错误
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}
```

### 4.3 HTTP 层统一处理错误

```go
// internal/pkg/httputil/errors.go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

// 中间件收集 panic + 业务错误
func ErrorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if rec := recover(); rec != nil {
                // log stack trace
                writeJSON(w, http.StatusInternalServerError, APIError{
                    Code:    "INTERNAL_ERROR",
                    Message: "an internal error occurred",
                })
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### 4.4 不要滥用 panic

Go 没有 try/catch，panic 不是异常机制。

```go
// ❌ 用 panic 传错误
func MustGetUser(id string) *User {
    u, err := GetUser(id)
    if err != nil {
        panic(err)
    }
    return u
}

// ✅ 错误是返回值
func GetUser(id string) (*User, error)
```

只有一种情况用 panic：**配置错误、端口被占用、数据库连不上**——Fail Fast，启动时就崩溃。

---

## 五、配置管理

### 5.1 12-Factor App 配置

所有配置来自环境变量，不写配置文件。

```go
// ❌ config.yaml——容易泄露到 git、环境间不同步
// ✅ 环境变量——同一份二进制部署到任何环境

type Config struct {
    DatabaseURL string `env:"DATABASE_URL,required"`
    RedisAddr   string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
    MaxWorkers  int    `env:"MAX_WORKERS" envDefault:"10"`
}
```

### 5.2 配置验证——启动时 Fail Fast

```go
func MustLoad() *Config {
    cfg := Load()
    if cfg.DatabaseURL == "" {
        panic("DATABASE_URL is required")
    }
    if _, err := net.ResolveTCPAddr("tcp", cfg.BindAddr); err != nil {
        panic(fmt.Sprintf("invalid BIND_ADDR: %v", err))
    }
    if cfg.MaxWorkers <= 0 {
        cfg.MaxWorkers = 10
    }
    return cfg
}
```

带着错误配置启动比运行到一半再炸好 100 倍。

### 5.3 推荐的配置库

| 库 | 说明 |
|----|------|
| `github.com/caarlos0/env/v11` | 最简，只从环境变量读 |
| `github.com/kelseyhightower/envconfig` | 经典，支持 required/default |
| `github.com/spf13/viper` | 全能，支持 env/file/etcd |

---

## 六、测试金字塔

### 6.1 比例

```
         ╱  E2E  ╲          1 : 10 : 100
        ╱ 集成测试 ╲
       ╱  单元测试  ╲
```

- **单元测试**（最多）：测单个函数/struct，mock 外部依赖
- **集成测试**（适量）：真实 DB / Testcontainers，测 repo+DB 交互
- **E2E 测试**（最少）：整个系统起来测核心用户路径

### 6.2 单元测试模式

```go
// internal/user/service_test.go
func TestCreateUser(t *testing.T) {
    t.Parallel() // 并行跑

    tests := []struct {
        name    string
        input   CreateUserInput
        mockFn  func(*MockRepo)
        wantErr bool
    }{
        {
            name:  "valid user",
            input: CreateUserInput{Email: "test@example.com"},
            mockFn: func(m *MockRepo) {
                m.On("FindByEmail", mock.Anything, "test@example.com").
                    Return(nil, ErrNotFound)
                m.On("Create", mock.Anything, mock.Anything).
                    Return(&User{ID: "1"}, nil)
            },
            wantErr: false,
        },
        {
            name:  "duplicate email",
            input: CreateUserInput{Email: "existing@example.com"},
            mockFn: func(m *MockRepo) {
                m.On("FindByEmail", mock.Anything, "existing@example.com").
                    Return(&User{ID: "1"}, nil)
            },
            wantErr: true,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            mockRepo := new(MockRepo)
            tc.mockFn(mockRepo)
            svc := NewService(mockRepo, nil, nil)

            err := svc.CreateUser(context.Background(), tc.input)
            if (err != nil) != tc.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tc.wantErr)
            }
        })
    }
}
```

### 6.3 集成测试用 Testcontainers

```go
// internal/user/repository_test.go
func TestUserRepository(t *testing.T) {
    // 启动 PostgreSQL 容器（Ory/dockertest 或 Testcontainers）
    postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "postgres:16-alpine",
            Env:   map[string]string{"POSTGRES_PASSWORD": "test"},
        },
        Started: true,
    })

    db := connectToPostgres(postgres)
    repo := NewRepository(db)

    // 测试真实数据库操作
    user, err := repo.Create(ctx, &User{Email: "test@example.com"})
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### 6.4 mock 生成

```go
// 用 mockery 从接口自动生成 mock
// 安装：go install github.com/vektra/mockery/v2@latest
// 使用：//go:generate mockery --name UserRepository --output ./mocks

// 然后运行 go generate ./...
```

---

## 七、接口什么时候抽

### 7.1 原则：Consumer 侧定义

```go
// ❌ Producer 侧定义（永远只有一个实现）
type UserRepository interface {  // 这个接口多余
    FindByID(ctx, id) (*User, error)
}

type userRepo struct { db *sql.DB }
func (r *userRepo) FindByID(ctx, id) (*User, error) { ... }

// ✅ Consumer 侧定义（只在需要 mock 时抽）
type userService struct {
    users interface {
        FindByID(ctx, id) (*User, error)
    }
}
```

### 7.2 什么时候需要接口

| 场景 | 需要接口吗 | 原因 |
|------|-----------|------|
| 只有一个实现，测试用真实 DB | ❌ | 直接在测试里连真实 DB |
| 多个实现（MySQL/PostgreSQL/mock） | ✅ | 需要切换实现 |
| 测试要用 mock（不想连 DB） | ✅ | mock 需要接口 |
| 回调/插件模式 | ✅ | 需要用户注入实现 |
| 中间件模式 | ✅ | 装饰器模式需要接口 |

### 7.3 接口大小

Go 的标准库接口通常很小：

```go
// 1 个方法
type io.Reader interface { Read(p []byte) (n int, err error) }

// 2 个方法
type io.ReadWriteCloser interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
}
```

你的接口也应该这样。如果接口超过 3 个方法，考虑拆成多个小接口。

---

## 八、并发模式

### 8.1 errgroup 编排并发任务

```go
import "golang.org/x/sync/errgroup"

func loadIssueDetail(ctx context.Context, id int64) (*IssueDetail, error) {
    g, ctx := errgroup.WithContext(ctx)

    var issue *Issue
    g.Go(func() error {
        issue, err = repo.GetIssue(ctx, id)
        return err
    })

    var labels []Label
    g.Go(func() error {
        labels, err = repo.GetLabels(ctx, id)
        return err
    })

    var comments []Comment
    g.Go(func() error {
        comments, err = repo.GetComments(ctx, id)
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, err
    }

    return &IssueDetail{
        Issue:    issue,
        Labels:   labels,
        Comments: comments,
    }, nil
}
```

比手写 channel 简洁得多，错误会自动传播。

### 8.2 singleflight 防缓存击穿

```go
import "golang.org/x/sync/singleflight"

var sf singleflight.Group

func GetUser(ctx context.Context, id string) (*User, error) {
    // 先从缓存读
    if u, hit := cache.Get(ctx, "user:"+id); hit {
        return u, nil
    }

    // 并发请求只查一次 DB
    v, err, shared := sf.Do("user:"+id, func() (any, error) {
        u, err := repo.GetUser(ctx, id)
        if err != nil {
            return nil, err
        }
        cache.Set(ctx, "user:"+id, u, 5*time.Minute)
        return u, nil
    })

    return v.(*User), err
}
```

### 8.3 worker pool 模式

```go
func ProcessOrders(ctx context.Context, orders []Order) error {
    workers := 10
    ch := make(chan Order, len(orders))

    // 启动 worker
    g, ctx := errgroup.WithContext(ctx)
    for i := 0; i < workers; i++ {
        g.Go(func() error {
            for order := range ch {
                if err := processOrder(ctx, order); err != nil {
                    return err
                }
            }
            return nil
        })
    }

    // 投递任务
    for _, order := range orders {
        select {
        case ch <- order:
        case <-ctx.Done():
            close(ch)
            return ctx.Err()
        }
    }
    close(ch)

    return g.Wait()
}
```

### 8.4 限流

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Limit(100), 200) // 100 req/s, burst 200

func handler(w http.ResponseWriter, r *http.Request) {
    if !limiter.Allow() {
        http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    // ...
}
```

---

## 九、数据库层

### 9.1 database/sql 还是 ORM？

| 方式 | 适合 | 说明 |
|------|------|------|
| `database/sql` | 复杂查询、性能敏感 | 手写 SQL，完全控制 |
| `sqlx` | 中等复杂度 | 比标准库少写 40% 模板代码 |
| `sqlc` | 类型安全 | 从 SQL 生成 Go 代码 |
| `ent` | 关系复杂 | Facebook 的 ORM，类型安全 |
| `gorm` | 快速开发 | 国内流行，但性能一般 |

推荐：**sqlc + database/sql**。SQL 是资产不是负债，手写 SQL 让你知道数据库在干什么。

```sql
-- sqlc 从 SQL 生成 Go 代码
-- query.sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *;
```

```go
// 生成的代码，类型安全
user, err := queries.GetUser(ctx, "user-123")
users, err := queries.ListUsers(ctx, ListUsersParams{Limit: 20, Offset: 0})
```

### 9.2 事务边界要清晰

```go
func CreateOrder(ctx context.Context, input CreateOrderInput) error {
    return db.WithTx(ctx, func(tx *sql.Tx) error {
        // 事务内所有操作
        order, err := tx.ExecContext(ctx, "INSERT INTO orders ...")
        for _, item := range input.Items {
            tx.ExecContext(ctx, "INSERT INTO order_items ...")
        }
        tx.ExecContext(ctx, "UPDATE inventory SET qty = qty - 1 WHERE id = ?", item.ID)
        return nil
    })
}
```

原则：**事务不要跨越 service 边界**。一个事务只在一个 service 方法内完成。

### 9.3 连接池配置

```go
// PostgreSQL
db.SetMaxOpenConns(25)      // 最大连接数
db.SetMaxIdleConns(5)       // 最大空闲连接
db.SetConnMaxLifetime(30 * time.Minute)  // 连接最大存活时间
db.SetConnMaxIdleTime(5 * time.Minute)   // 空闲超时

// SQLite（特殊）
db.SetMaxOpenConns(1)  // SQLite 写串行，多连接无意义
```

### 9.4 N+1 查询

```go
// ❌ N+1
issues, _ := repo.ListIssues(ctx)     // 1 次
for _, issue := range issues {
    comments, _ := repo.GetComments(ctx, issue.ID) // N 次
}

// ✅ JOIN 或批量查询
issues, _ := repo.ListIssuesWithComments(ctx) // 1 次
```

---

## 十、日志、监控、可观测性

### 10.1 结构化日志

```go
// ❌ 非结构化
log.Printf("user %s logged in from %s", id, ip)

// ✅ 结构化——可被日志系统解析
slog.Info("user logged in",
    "user_id", id,
    "ip", ip,
    "duration_ms", elapsed.Milliseconds(),
    "request_id", reqID,
)
```

### 10.2 日志级别策略

| 级别 | 环境 | 内容 |
|------|------|------|
| DEBUG | 开发 | 函数入口/出口、SQL、请求体 |
| INFO | 生产 | 用户操作、状态变更、任务完成 |
| WARN | 生产 | 请求慢、重试、降级 |
| ERROR | 生产 + 告警 | DB 断连、第三方超时、业务异常 |

### 10.3 每个请求一个 request_id

```go
// 中间件注入
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := r.Header.Get("X-Request-Id")
        if id == "" {
            id = uuid.NewString()
        }
        w.Header().Set("X-Request-Id", id)
        ctx := context.WithValue(r.Context(), "request_id", id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 10.4 Metrics

```
Prometheus 4 个黄金指标：
- 请求量（QPS）
- 延迟（P50 / P95 / P99）
- 错误率（5xx / 4xx）
- 饱和度（CPU / 内存 / 连接池）

Histogram 示例：
request_duration_seconds{method, path, status}  — 请求耗时
requests_total{method, path, status}             — 请求计数
db_query_duration_seconds{query}                 — 数据库查询耗时
```

### 10.5 推荐的日志库

| 库 | 性能 | 特点 |
|----|------|------|
| `log/slog` | 中等 | Go 标准库，够用 |
| `rs/zerolog` | 极快 | 零分配 JSON 日志 |
| `uber-go/zap` | 极快 | 性能最好的结构化日志 |

大型项目（QPS > 1000）推荐 zerolog 或 zap。标准库 slog 适合中小项目。

---

## 十一、HTTP 层

### 11.1 中间件顺序

```
RequestID → Logger → Recoverer → Timeout → RateLimit → Auth → Router
```

外层到内层：保障性中间件（ID/日志/恢复）→ 安全中间件（限流/认证）→ 路由

### 11.2 Handler 应该有多厚？

```go
// ❌ 厚 handler——500 行 handler 做所有事
func CreateUser(w http.ResponseWriter, r *http.Request) {
    // 解析请求体
    // 校验字段
    // 查数据库
    // 发邮件
    // 写日志
    // 返回响应
}

// ✅ 薄 handler——只做 HTTP 层的事
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        httputil.WriteError(w, ErrInvalidJSON)
        return
    }

    user, err := h.svc.CreateUser(r.Context(), req.ToInput())
    if err != nil {
        httputil.WriteError(w, err)
        return
    }

    httputil.WriteJSON(w, http.StatusCreated, user.ToResponse())
}
```

### 11.3 统一响应格式

```json
// 成功
{ "data": { "id": "123", "name": "Alice" } }

// 列表
{ "data": [...], "pagination": { "total": 100, "page": 1, "size": 20 } }

// 错误
{ "error": { "code": "VALIDATION_ERROR", "message": "email is required", "details": [...] } }
```

### 11.4 路由设计

```go
// 稳定的 API 版本
/v1/users
/v1/orders

// 不要在 URL 里用动词
// ❌ /v1/createUser
// ✅ /v1/users POST

// 嵌套不要超过 2 层
// ✅ /v1/users/{id}/orders
// ❌ /v1/users/{id}/orders/{oid}/items/{iid}
```

---

## 十二、值得用的第三方包

### 核心基础设施

| 包 | 用途 | 推荐指数 |
|----|------|:--------:|
| `github.com/jackc/pgx/v5` | PostgreSQL 驱动 | ⭐⭐⭐⭐⭐ |
| `github.com/redis/go-redis/v9` | Redis 客户端 | ⭐⭐⭐⭐⭐ |
| `github.com/rs/zerolog` | 高性能日志 | ⭐⭐⭐⭐⭐ |
| `golang.org/x/sync/errgroup` | 并发编排 | ⭐⭐⭐⭐⭐ |
| `golang.org/x/sync/singleflight` | 防缓存击穿 | ⭐⭐⭐⭐ |
| `golang.org/x/time/rate` | 限流器 | ⭐⭐⭐⭐⭐ |
| `github.com/prometheus/client_golang` | Metrics | ⭐⭐⭐⭐⭐ |

### API 层

| 包 | 用途 | 推荐指数 |
|----|------|:--------:|
| `github.com/go-chi/chi/v5` | HTTP 路由 | ⭐⭐⭐⭐⭐ |
| `github.com/go-playground/validator/v10` | 字段校验 | ⭐⭐⭐⭐ |
| `github.com/swaggo/http-swagger` | Swagger UI | ⭐⭐⭐⭐ |
| `github.com/golang-jwt/jwt/v5` | JWT | ⭐⭐⭐⭐⭐ |
| `github.com/go-chi/cors` | CORS | ⭐⭐⭐⭐ |
| `github.com/go-chi/jwtauth/v5` | JWT 中间件 | ⭐⭐⭐⭐ |

### 数据库

| 包 | 用途 | 推荐指数 |
|----|------|:--------:|
| `github.com/jmoiron/sqlx` | SQL 增强 | ⭐⭐⭐⭐ |
| `github.com/sqlc-dev/sqlc` | SQL → Go 代码生成 | ⭐⭐⭐⭐⭐ |
| `github.com/golang-migrate/migrate/v4` | 数据库迁移 | ⭐⭐⭐⭐⭐ |
| `entgo.io/ent` | ORM | ⭐⭐⭐⭐ |
| `github.com/jackc/pgx/v5/pgxpool` | 连接池 | ⭐⭐⭐⭐⭐ |

### 测试

| 包 | 用途 | 推荐指数 |
|----|------|:--------:|
| `github.com/stretchr/testify` | 断言 + mock | ⭐⭐⭐⭐⭐ |
| `github.com/vektra/mockery/v2` | 自动生成 mock | ⭐⭐⭐⭐⭐ |
| `github.com/ory/dockertest/v3` | 容器化集成测试 | ⭐⭐⭐⭐⭐ |
| `github.com/testcontainers/testcontainers-go` | 容器化管理 | ⭐⭐⭐⭐ |
| `github.com/google/go-cmp` | 深度比较 | ⭐⭐⭐⭐ |

### 工具

| 包 | 用途 | 推荐指数 |
|----|------|:--------:|
| `github.com/caarlos0/env/v11` | 环境变量解析 | ⭐⭐⭐⭐⭐ |
| `github.com/spf13/viper` | 配置管理 | ⭐⭐⭐⭐ |
| `github.com/spf13/cobra` | CLI 框架 | ⭐⭐⭐⭐⭐ |
| `github.com/google/uuid` | UUID | ⭐⭐⭐⭐⭐ |
| `github.com/samber/lo` | 泛型工具集合 | ⭐⭐⭐⭐ |

---

## 十三、不要用的包

| 包 | 不推荐理由 |
|----|-----------|
| `github.com/gin-gonic/gin` | 不兼容 `http.Handler`，生态锁定 |
| `github.com/labstack/echo/v4` | 同上 |
| `github.com/kataras/iris` | 太重，生态小 |
| `github.com/gogf/gf` | Go 版 Spring，过度设计 |
| `github.com/astaxie/beego` | 已基本停止维护 |
| `github.com/julienschmidt/httprouter` | chi 已经替代了它 |
| `github.com/gorilla/mux` | 已归档（不再维护） |
| `github.com/go-sql-driver/mysql` | 性能不如 `pgx`（MySQL 本身的问题） |
| `github.com/lib/pq` | 性能不如 `pgx`，已归档 |

---

## 十四、常见的坑

### 14.1 接口接收者用指针还是值？

```go
type User struct {
    Name string
}

// 方法不修改结构体 → 值接收者
func (u User) FullName() string { return u.Name }

// 方法修改结构体 → 指针接收者
func (u *User) SetName(name string) { u.Name = name }

// 一致性原则：如果一个方法用指针，所有方法都用指针
```

### 14.2 Slice 和 Map 的并发安全

```go
// ❌ map 并发读写会 fatal error: concurrent map writes
m := make(map[string]int)
go func() { m["a"] = 1 }()
go func() { m["b"] = 2 }()

// ✅ 用 sync.Map 或 mutex
var mu sync.RWMutex
mu.Lock()
m["a"] = 1
mu.Unlock()

// ✅ 或用 concurrent-map
import "github.com/orcaman/concurrent-map/v2"
```

### 14.3 指针接收者的方法集

```go
type Speaker interface {
    Speak() string
}

type Person struct{ Name string }
func (p *Person) Speak() string { return "Hi, I'm " + p.Name }

var s Speaker
s = Person{"Alice"}  // ❌ *Person 实现了 Speaker，Person 没有
s = &Person{"Alice"} // ✅
```

### 14.4 Slice 底层数组共享

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]  // b = [2, 3]，但底层指向 a 的同一数组
b[0] = 99    // a[1] 也变成 99！

// 要真正复制
b := append([]int{}, a[1:3]...) // 独立副本
```

### 14.5 循环变量捕获（Go < 1.22）

```go
// Go 1.22 以下
for _, v := range items {
    go func() {
        fmt.Println(v) // 所有 goroutine 打印最后一个 v
    }()
}

// 修复
for _, v := range items {
    v := v
    go func() {
        fmt.Println(v)
    }()
}

// Go 1.22+ 已修复，不需要 v := v
```

### 14.6 defer 在循环中的开销

```go
// ❌ defer 在循环中——defer 在函数返回时才执行，阻塞连接释放
for _, item := range items {
    mu.Lock()
    defer mu.Unlock() // 直到函数结束才解锁！
}

// ✅ 用匿名函数包裹
for _, item := range items {
    func() {
        mu.Lock()
        defer mu.Unlock()
        // ...
    }()
}
```

### 14.7 大 map/slice 不释放内存

```go
var m map[int][]byte
for i := 0; i < 100000; i++ {
    m[i] = make([]byte, 1024) // 100MB
}
delete(m, 0)  // map 不会缩小——底层哈希表依然占着 100MB

// 解法：定期重建 map，或用 sync.Map
m = make(map[int][]byte) // 重建，旧 map 被 GC
```

---

## 十五、渐进式重构

大型项目最怕的不是"代码写得不好"，而是"为了改成好的架构停掉所有功能开发三个月"。

### 15.1 绞杀者模式（Strangler Fig）

```go
// 旧系统：users 在老路径
// 新系统：users 在新路径
// 过渡期：两个共存，逐步迁移

router.Route("/v1", oldRouter)  // 旧 API，逐步废弃
router.Route("/v2", newRouter)  // 新 API，新功能只加在这
```

### 15.2 Feature Flag

```go
// 用环境变量或配置中心控制功能开关
if cfg.FeatureNewCheckout {
    return newCheckout(ctx, req)
}
return oldCheckout(ctx, req)
```

### 15.3 并行运行 + 对比

```go
// 新老实现同时跑，对比结果
func GetUser(ctx context.Context, id string) (*User, error) {
    oldUser, oldErr := oldImpl.GetUser(ctx, id) // 旧路径

    newUser, newErr := newImpl.GetUser(ctx, id) // 新路径
    if newErr == nil {
        go compareAndAlert(oldUser, newUser) // 异步对比，不一致就告警
    }

    return oldUser, oldErr // 先走旧路径，稳定后再切
}
```

### 15.4 数据迁移

```
不要一次性 migration：
1. 先加新字段（允许 null，默认值兼容旧代码）
2. 旧代码双写新字段（新旧同时写）
3. 后台任务回填历史数据
4. 切换读取到新字段
5. 删除旧字段
```

---

## 最后的话

> 大型 Go 项目死于**过度抽象**，而不是不够抽象。

- 不要一开始就设计一个完美的架构。架构是长出来的，不是设计出来的。
- 重复三次才抽公共代码。前两次的重复让你真正理解共同点在哪。
- 不确定要不要用泛型？先不用。手写三遍再考虑泛型化。
- 不确定要不要接口？先不抽。需要 mock 测试的时候自然知道在哪抽。
- 每个抽象层次都有成本。一个间接调用可能永远不会被替换，那它就是多余的。

Go 的设计哲学是**简单直白**。读代码的人能在 5 秒内理解你在做什么，比"用了 6 种设计模式"重要得多。
