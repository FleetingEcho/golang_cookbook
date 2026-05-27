# ── 多阶段构建 Dockerfile ─────────────────────────────────────────────────
#
# 工程要点：
# 1. 多阶段构建：构建阶段用完整 Go 镜像，运行阶段用 scratch（最小镜像）
# 2. 最终镜像仅包含编译后的二进制，无 shell、无包管理器
# 3. 最终镜像大小约 5-15MB（vs 包含 OS 的 ~1GB）
# 4. 安全：scratch 镜像的攻击面最小

# ── Stage 1: 构建 ─────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 先复制依赖文件（利用 Docker 层缓存）
# 只要 go.mod/go.sum 不变，这一步就不会重跑
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# 说明：
# CGO_ENABLED=0  — 禁用 CGO，静态编译，确保在 scratch 中可运行
# -ldflags="-s -w" — 去掉调试信息，减小二进制体积约 30%

# ── Stage 2: 运行 ─────────────────────────────────────────────────────────
FROM scratch

# 复制编译好的二进制
COPY --from=builder /app/server /server

# 复制静态资源（migration 文件等）
COPY --from=builder /app/migrations /migrations

# 创建数据目录（scratch 中没有 mkdir，但空目录可映射 volume）
# 运行时可挂载：docker run -v ./data:/data ...

EXPOSE 3001

# 设置默认环境变量
ENV ISSUE_TRACKER_BIND_ADDR=0.0.0.0:3001
ENV ISSUE_TRACKER_UPLOAD_DIR=/uploads
ENV ISSUE_TRACKER_LOG_LEVEL=info

ENTRYPOINT ["/server"]
