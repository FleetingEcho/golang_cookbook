package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Connect 打开 SQLite 连接池并执行 migration。
//
// 工程要点：
// 1. 使用 database/sql 标准接口，不绑定 ORM
// 2. SQLite 限制 MaxOpenConns=1，因为 SQLite 写操作是串行化的
// 3. 优先使用文件路径而非 :memory: 以保留调试数据
func Connect(databaseURL string, maxOpenConns, maxIdleConns int) (*sql.DB, error) {
	// ── 解析连接串 ────────────────────────────────────────────────────
	// 支持格式：sqlite:///absolute/path 或 sqlite://relative/path
	path := strings.TrimPrefix(databaseURL, "sqlite://")

	// 确保父目录存在（如果是文件路径而非 :memory:）
	if path != ":memory:" {
		dir := filepath.Dir(path)
		if dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create db dir: %w", err)
			}
		}
	}

	// ── 打开连接池 ────────────────────────────────────────────────────
	// sql.Open 不会真正连接，只验证 DSN 格式。实际连接在第一次 Query 时建立。
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// ── 连接池配置 ────────────────────────────────────────────────────
	// SQLite 特点：写操作是串行的（只有一个写事务可以执行）。
	// 设置 MaxOpenConns=1 避免多个连接互相阻塞。
	// 如果是 PostgreSQL 或 MySQL，通常设置为 10-50。
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(30 * time.Minute) // 定期回收连接，避免陈旧连接
	db.SetConnMaxIdleTime(5 * time.Minute)  // 空闲连接超过此时间则关闭

	// ── 验证连接可达 ──────────────────────────────────────────────────
	// Ping 实际创建连接，验证 DSN 和权限。
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	// ── 执行 migration ────────────────────────────────────────────────
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	slog.Info("database connected",
		"driver", "sqlite",
		"path", path,
		"max_open", maxOpenConns,
	)
	return db, nil
}

// runMigrations 按顺序执行 DDL 文件。
//
// 工程要点：
// 1. 在生产中应使用 golang-migrate 或 goose 等迁移工具
// 2. 这里用简单的 Exec 是为了零依赖启动
// 3. 多个迁移文件按文件名排序执行（0001 → 0002 → ...）
// 4. 使用 `IF NOT EXISTS` 让迁移幂等
func runMigrations(db *sql.DB) error {
	// ── 创建迁移追踪表（记录哪些迁移已执行） ──────────────────────────
	// 这模拟了 golang-migrate 等工具的 _migrations 表。
	// 目的是让迁移幂等：已运行的迁移不会重复执行。
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			filename TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// ── 读取并执行迁移文件 ────────────────────────────────────────────
	// 在生产中会查询 _migrations 表确定哪些迁移尚未执行。
	// 这里简化实现：启动时执行完整 DDL（幂等）。
	data, err := os.ReadFile("migrations/0001_init.sql")
	if err != nil {
		// 尝试从上层目录读取（适配不同工作目录）
		data, err = os.ReadFile("../migrations/0001_init.sql")
		if err != nil {
			return fmt.Errorf("read migration file: %w", err)
		}
	}

	if _, err := db.Exec(string(data)); err != nil {
		return fmt.Errorf("exec migration: %w", err)
	}

	// 记录迁移状态（仅记录一次，INSERT OR IGNORE 幂等）
	_, _ = db.Exec(
		"INSERT OR IGNORE INTO _migrations (filename) VALUES (?)",
		"0001_init.sql",
	)

	slog.Info("migrations applied", "files", []string{"0001_init.sql"})
	return nil
}
