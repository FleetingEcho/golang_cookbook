# ── Makefile ────────────────────────────────────────────────────────────────
#
# 工程要点：
# 1. .PHONY 标记非文件目标，防止与同名文件冲突
# 2. 每个目标有明确职责，支持组合调用
# 3. 使用 $(MAKE) 而非直接 make，确保传递 flags

.PHONY: run build test seed clean lint coverage docker

# ── 开发 ────────────────────────────────────────────────────────────────────

run:  ## 启动开发服务器（默认 127.0.0.1:3001）
	go run ./

build: ## 编译到 bin/issue_tracker
	go build -ldflags="-s -w" -o bin/issue_tracker ./

# ── 测试 ────────────────────────────────────────────────────────────────────

test: ## 运行全部测试
	go test -v -count=1 -timeout 30s ./...

test-short: ## 仅运行单元测试（跳过集成测试）
	go test -v -count=1 -short -timeout 10s ./...

coverage: ## 运行测试并生成覆盖率报告
	go test -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "coverage report: coverage.html"

# ── 数据库 ──────────────────────────────────────────────────────────────────

seed: ## 重置数据库并灌入种子数据
	@mkdir -p data uploads
	@rm -f data/issue_tracker.db
	@sqlite3 data/issue_tracker.db < migrations/0001_init.sql
	@sqlite3 data/issue_tracker.db < scripts/seed.sql
	@echo "seeded data/issue_tracker.db"

# ── Docker ──────────────────────────────────────────────────────────────────

docker-build: ## 构建 Docker 镜像
	docker build -t issue-tracker-api:latest .

docker-run: ## 启动 Docker 容器
	docker run -d \
		--name issue-tracker-api \
		-p 3001:3001 \
		-v ./data:/data \
		-v ./uploads:/uploads \
		-e ISSUE_TRACKER_UPLOAD_DIR=/uploads \
		-e DATABASE_URL="sqlite:///data/issue_tracker.db" \
		issue-tracker-api:latest

# ── 代码质量 ────────────────────────────────────────────────────────────────

lint: ## 运行 linter（需安装 golangci-lint）
	golangci-lint run ./...

vet: ## 运行 go vet 静态分析
	go vet ./...

# ── 清理 ────────────────────────────────────────────────────────────────────

clean: ## 清理编译产物和数据库
	rm -rf bin/ coverage.out coverage.html
	rm -f data/issue_tracker.db
	@find uploads -type f ! -name '.gitkeep' -delete 2>/dev/null || true

# ── 帮助 ────────────────────────────────────────────────────────────────────

help: ## 显示所有目标
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
