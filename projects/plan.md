# Go 后端重写计划

## 概述

将现有的 Rust/Axum issue tracker 后端用 Go 重写，保持完全一致的 API 接口（确保前端零改动），放入 `server_go/` 目录。

---

## 技术栈

| 层         | 选择                                | 理由                                                                 |
|------------|-------------------------------------|----------------------------------------------------------------------|
| **语言**   | Go 1.23+                            | 最新的 range-over-func、泛型稳定、`net/http` 方法路由模式             |
| **路由**   | `go-chi/chi/v5`                     | 标准 `http.Handler` 接口、中间件链式、不侵入业务代码                  |
| **数据库** | `modernc.org/sqlite` (pure Go)      | 纯 Go SQLite，无 CGO 依赖，交叉编译友好                               |
| **UUID**   | `google/uuid`                       | V4 UUID                                                              |
| **日志**   | `log/slog` (标准库)                 | Go 1.21 内置的结构化日志                                              |
| **测试**   | `testing` + `httptest` (标准库)     | 集成测试用内存 SQLite + 测试 HTTP 服务器                              |

**为什么不用 Gin？** Chi 使用标准 `http.Handler` 接口——中间件和 handler 与标准库完全互通，更符合"最新的 Go 语法"的方向。Go 1.22 的 `net/http` 方法路由已经是产品级，Chi 在此基础上仅添加了中间件链和路径参数提取。

---

## 目录结构

```
server_go/
├── cmd/
│   └── server/
│       └── main.go              # 入口：加载配置、连接 DB、启动 HTTP 服务
├── internal/
│   ├── config/
│   │   └── config.go            # 环境变量 → Config 结构体
│   ├── db/
│   │   └── sqlite.go            # 连接池、自动运行 migration
│   ├── models/
│   │   └── models.go            # Issue / Comment / Label / Attachment 结构体
│   ├── dto/
│   │   ├── request.go           # 请求 DTO（含 validate 标签）
│   │   └── response.go          # 响应 DTO（含泛型 PaginatedResponse[T]）
│   ├── handlers/
│   │   ├── issues.go            # issue CRUD + 查询/过滤/分页
│   │   ├── comments.go          # comment 增删查
│   │   ├── labels.go            # label 增删查 + 关联/取消关联 issue
│   │   └── attachments.go       # 附件上传/下载/删除
│   ├── middleware/
│   │   ├── auth.go              # x-api-key 校验
│   │   └── request_id.go        # X-Request-Id 注入
│   ├── storage/
│   │   └── file.go              # 安全存储文件名、路径校验
│   └── router/
│       └── router.go            # 路由注册 + 中间件 + CORS
├── migrations/
│   └── 0001_init.sql            # (复用原 SQLite DDL)
├── data/                        # SQLite 数据库文件
├── storage/uploads/             # 上传文件存储
├── go.mod
└── Makefile
```

---

## 核心设计决策

### 1. 泛型精确用法

只在 **一处** 使用泛型——`PaginatedResponse[T]`，因为这是 Go 泛型最清晰的适用场景：

```go
type PaginatedResponse[T any] struct {
    Items  []T  `json:"items"`
    Total  int64 `json:"total"`
    Limit  int64 `json:"limit"`
    Offset int64 `json:"offset"`
}
```

DTO 转换保持手写（从 model 结构体转 response），不在泛型上过度工程——代码可读性优先于"硬用泛型"。

### 2. CamelCase JSON

所有请求/响应字段使用 `json:"camelCase"` 标签，与前端 TypeScript 类型完全对齐。

### 3. 错误处理

统一错误类型 `AppError`，带 HTTP status code + 错误消息，通过 `http.Handler` 中间件收集并返回 `{"error": "..."}` JSON 格式。

### 4. 文件上传安全

复用原 Rust 逻辑：
- UUID 前缀文件名避免冲突/路径遍历
- 严格禁止 `../` 和路径分隔符
- 存储路径在 `uploadDir` 内部

### 5. 与 Rust 版的 API 差异说明

| 特性             | Rust 版                          | Go 版                                  |
|------------------|----------------------------------|----------------------------------------|
| 运行时           | Tokio (async)                    | Go 原生 goroutine                      |
| DB 迁移          | 启动时自动运行 raw SQL           | 同上，`db.Exec`                      |
| 日志             | tracing / tracing-subscriber     | slog 结构化日志                        |
| 请求 ID          | tower-http 中间件                | 手写中间件 + slog 字段                 |
| 分页             | QueryBuilder 动态 SQL            | 手写字符串拼接（参数化绑定防注入）     |
| 最大上传         | 10MB，axum 中间件层              | 10MB，`http.MaxBytesReader`            |
| CORS             | tower-http                       | chi 内置 CORS 中间件 / 手写           |
| API Key 校验     | 层中间件                         | Chi 路由组中间件                       |

---

## 实现步骤

### Step 1: 脚手架

- `go mod init`，初始化 `server_go/` 目录结构
- 创建 `config.go`（从环境变量读取配置）
- 创建 `sqlite.go`（连接池 + 自动执行 DDL）
- 创建 `models.go`（4 个结构体）
- 创建 `dto/request.go` + `dto/response.go`（含 `PaginatedResponse[T]`）
- 创建 `error.go`（统一 AppError）

### Step 2: 中间件

- `middleware/auth.go` — x-api-key 校验
- `middleware/request_id.go` — 注入 X-Request-Id，写入 slog

### Step 3: 路由 + Handler

- `router/router.go` — 注册所有路由 + CORS + 中间件链
- `handlers/issues.go` — list/create/get/update/delete + 分页 + 过滤 + 搜索
- `handlers/comments.go` — list/create/delete
- `handlers/labels.go` — list/create/attach/detach
- `handlers/attachments.go` — list/upload/download/delete
- `storage/file.go` — 文件名校验 + 路径安全

### Step 4: main.go + 测试

- `cmd/server/main.go` — 入口
- 集成测试：内存 SQLite + `httptest`，覆盖全部 CRUD 端点

### Step 5: Makefile + 文档

- `Makefile` 含 `run` / `build` / `test` / `seed` 目标
- 简要 README（可选）

---

## 风险与注意

1. **动态 SQL 构建** — Rust 的 `QueryBuilder` 很优雅，Go 没有对应库。手写条件拼接+参数化绑定需要小心 SQL 注入（用 `?` 占位符，不拼接用户输入）。
2. **Multipart 文件上传** — Go 的 `r.MultipartForm` API 偏底层，需要小心处理 10MB 限制和 field 名称 `"file"`。
3. **测试 DB 隔离** — 每个集成测试用例需要独立事务 + 回滚，或独立内存 DB 实例，避免测试间污染。
4. **Time 类型** — SQLite 没有原生 datetime，原项目用 TEXT `CURRENT_TIMESTAMP` 存储。Go 侧用 `time.Time` 序列化为 RFC3339 字符串。

---

## 工作量估计

- **文件数**: ~15 个 Go 源文件
- **预计代码行**: ~1200–1500 行
- **与前端兼容性**: 100%（接口路径、JSON 字段名、状态码、错误格式完全一致）
