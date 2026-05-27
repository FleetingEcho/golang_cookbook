package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"issue_tracker/models"
	"issue_tracker/utils"

	"github.com/go-chi/chi/v5"
)

// ── 路由注册 ──────────────────────────────────────────────────────────────
//
// 设计模式：每个资源一个 RegisterXxxRoutes 函数，在 app.go 中统一调用。
// 避免在 app.go 中写大量路由，也避免在 handler 文件中声明路由。

func RegisterIssueRoutes(r chi.Router, db *sql.DB) {
	r.Get("/", ListIssues(db))
	r.Post("/", CreateIssue(db))
	r.Get("/{id}", GetIssue(db))
	r.Patch("/{id}", UpdateIssue(db))
	r.Delete("/{id}", DeleteIssue(db))
}

// ── 动态 SQL 构建器 ──────────────────────────────────────────────────────
//
// 工程要点：
// 1. 使用 ? 参数化绑定防止 SQL 注入
// 2. 不用 ORM，直接构建 SQL 可精确控制查询计划
// 3. 比 fmt.Sprintf 安全——参数值不会破坏 SQL 语法

type sqlBuilder struct {
	parts []string
	args  []any
}

func (b *sqlBuilder) Add(part string, args ...any) {
	b.parts = append(b.parts, part)
	b.args = append(b.args, args...)
}

func (b *sqlBuilder) Build() (string, []any) {
	return strings.Join(b.parts, " "), b.args
}

// ── context 超时工具 ─────────────────────────────────────────────────────
//
// 工程要点：
// 1. 每个查询都应传递 context，支持超时和取消
// 2. 全局超时由 chi 的 Timeout 中间件保证，这里不做二次包装
// 3. ⚠️ 不要在返回 *sql.Rows 的函数中 defer cancel()！
//    取消 context 会使 rows.Next() 失效，调用方必须在读完 rows 前保持 context 活跃
// 4. QueryRow 立即 Scan，可以安全用 defer cancel

func queryContext(ctx context.Context, db *sql.DB, query string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}

func queryRowContext(ctx context.Context, db *sql.DB, query string, args ...any) *sql.Row {
	return db.QueryRowContext(ctx, query, args...)
}

func execContext(ctx context.Context, db *sql.DB, query string, args ...any) (sql.Result, error) {
	return db.ExecContext(ctx, query, args...)
}

// ── List Issues ──────────────────────────────────────────────────────────

// ListIssues godoc
// @Summary      List issues with pagination and filtering
// @Description  Returns a paginated list of issues, optionally filtered by status, priority, type, label, or search text.
// @Tags         issues
// @Accept       json
// @Produce      json
// @Param        status     query  string  false  "Filter by status (open, in_progress, closed)"
// @Param        priority   query  string  false  "Filter by priority (low, medium, high)"
// @Param        issueType  query  string  false  "Filter by issue type (bug, feature, task, question)"
// @Param        labelId    query  integer false  "Filter by label ID"
// @Param        search     query  string  false  "Search in title and description"
// @Param        limit      query  integer false  "Max results (default 5, max 100)"
// @Param        offset     query  integer false  "Result offset"
// @Success      200  {object}  models.PaginatedResponse[models.Issue]
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues [get]
func ListIssues(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ── 从 request context 中提取 request_id（已由中间件注入） ──
		// 后续的 slog 调用会自动携带该字段
		logger := slog.Default()

		q := parseIssueQuery(r)

		// ── 计数查询 ──────────────────────────────────────────────────
		// 先查总数再查数据是分页的标准做法（两阶段查询）
		cb := &sqlBuilder{}
		cb.Add("SELECT COUNT(DISTINCT issues.id) FROM issues")
		if q.LabelID != nil {
			cb.Add("JOIN issue_labels ON issues.id = issue_labels.issue_id")
		}
		cb.Add("WHERE 1=1")
		addIssueFilters(cb, q)

		var total int64
		cs, ca := cb.Build()
		// 使用 QueryRowContext 传递请求的 context，支持超时和取消
		if err := queryRowContext(r.Context(), db, cs, ca...).Scan(&total); err != nil {
			logger.ErrorContext(r.Context(), "count issues failed", "error", err)
			utils.Error500(w, "database error")
			return
		}

		// ── 数据查询 ──────────────────────────────────────────────────
		dbb := &sqlBuilder{}
		dbb.Add("SELECT DISTINCT issues.* FROM issues")
		if q.LabelID != nil {
			dbb.Add("JOIN issue_labels ON issues.id = issue_labels.issue_id")
		}
		dbb.Add("WHERE 1=1")
		addIssueFilters(dbb, q)
		dbb.Add("ORDER BY updated_at DESC, id DESC")
		dbb.Add("LIMIT ? OFFSET ?", q.Limit, q.Offset)

		ds, da := dbb.Build()
		rows, err := queryContext(r.Context(), db, ds, da...)
		if err != nil {
			logger.ErrorContext(r.Context(), "list issues failed", "error", err)
			utils.Error500(w, "database error")
			return
		}
		defer rows.Close() // ⚠️ 永远不要忘记 close rows！否则连接池会泄漏

		items, err := models.ScanIssues(rows)
		if err != nil {
			logger.ErrorContext(r.Context(), "scan issues failed", "error", err)
			utils.Error500(w, "database error")
			return
		}

		utils.JSON(w, models.PaginatedResponse[models.Issue]{
			Items: items, Total: total, Limit: q.Limit, Offset: q.Offset,
		})
	}
}

func parseIssueQuery(r *http.Request) models.IssueQuery {
	q := models.IssueQuery{Limit: 25}
	if v := r.URL.Query().Get("status"); v != "" {
		q.Status = &v
	}
	if v := r.URL.Query().Get("priority"); v != "" {
		q.Priority = &v
	}
	if v := r.URL.Query().Get("issueType"); v != "" {
		q.IssueType = &v
	}
	if v := r.URL.Query().Get("labelId"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			q.LabelID = &id
		}
	}
	if v := r.URL.Query().Get("search"); v != "" {
		q.Search = &v
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if l, err := strconv.ParseInt(v, 10, 64); err == nil && l > 0 && l <= 100 {
			q.Limit = l // 上限 100 防止恶意大分页
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if o, err := strconv.ParseInt(v, 10, 64); err == nil && o >= 0 {
			q.Offset = o
		}
	}
	return q
}

func addIssueFilters(b *sqlBuilder, q models.IssueQuery) {
	if q.Status != nil {
		b.Add("AND issues.status = ?", *q.Status)
	}
	if q.Priority != nil {
		b.Add("AND issues.priority = ?", *q.Priority)
	}
	if q.IssueType != nil {
		b.Add("AND issues.issue_type = ?", *q.IssueType)
	}
	if q.LabelID != nil {
		b.Add("AND issue_labels.label_id = ?", *q.LabelID)
	}
	if q.Search != nil {
		p := "%" + *q.Search + "%"
		b.Add("AND (issues.title LIKE ? OR issues.description LIKE ?)", p, p)
	}
}

// CreateIssue godoc
// @Summary      Create a new issue
// @Description  Creates a new issue with optional label associations.
// @Tags         issues
// @Accept       json
// @Produce      json
// @Param        body  body  models.CreateIssueRequest  true  "Issue data"
// @Success      200   {object}  models.IssueDetail
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /issues [post]
func CreateIssue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ── 请求体解码 ──────────────────────────────────────────────
		// 注意：json.NewDecoder 默认使用 io.EOF 表示空 body
		// 使用 Decode 而不是 Unmarshal，支持流式解析
		var input models.CreateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			utils.Error400(w, "invalid JSON body")
			return
		}

		// ── 业务逻辑校验（非 typescript 的类型校验） ────────────────
		// 注意：这层校验是必需的——即使前端做了校验，恶意请求仍可绕过
		if msg := models.ValidateIssueFields(input.Title, input.Description, input.Priority, input.IssueType); msg != "" {
			utils.Error400(w, msg)
			return
		}
		if strings.TrimSpace(input.CreatedBy) == "" {
			utils.Error400(w, "created_by is required")
			return
		}

		// ── 事务 ────────────────────────────────────────────────────
		// 事务确保两个操作（插入 issue + 关联 label）一起成功或一起失败。
		// 不使用事务：可能出现 issue 插入成功但 label 关联失败的中间状态。
		tx, err := db.Begin()
		if err != nil {
			slog.ErrorContext(r.Context(), "begin tx failed", "error", err)
			utils.Error500(w, "database error")
			return
		}
		// ⚠️ 事务必须 Rollback 或 Commit。
		// 如果提前 return 忘记调用，连接会被事务锁住直到超时。
		// defer tx.Rollback() 在正常 Commit 后调用是安全的（第二次调用是 no-op）。
		defer tx.Rollback()

		var id int64
		// RETURNING 是 SQLite 3.35+ 的特性，避免先 INSERT 再 SELECT
		err = tx.QueryRow(
			`INSERT INTO issues (title, description, priority, issue_type, assignee, created_by)
			 VALUES (?, ?, ?, ?, ?, ?) RETURNING id`,
			strings.TrimSpace(input.Title), strings.TrimSpace(input.Description),
			input.Priority, input.IssueType, input.Assignee, strings.TrimSpace(input.CreatedBy),
		).Scan(&id)
		if err != nil {
			slog.ErrorContext(r.Context(), "create issue failed", "error", err)
			utils.Error500(w, "failed to create issue")
			return
		}

		if err := linkLabelsTx(tx, id, input.LabelIDs); err != nil {
			slog.ErrorContext(r.Context(), "link labels failed", "error", err)
			utils.Error500(w, "failed to link labels")
			return
		}

		// ── 提交事务 ────────────────────────────────────────────────
		if err := tx.Commit(); err != nil {
			slog.ErrorContext(r.Context(), "commit tx failed", "error", err)
			utils.Error500(w, "database error")
			return
		}

		writeIssueDetail(w, r.Context(), db, id)
	}
}

// GetIssue godoc
// @Summary      Get issue detail
// @Description  Returns a single issue with its labels, comments, and attachments.
// @Tags         issues
// @Accept       json
// @Produce      json
// @Param        id   path  integer  true  "Issue ID"
// @Success      200  {object}  models.IssueDetail
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues/{id} [get]
func GetIssue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid issue id")
			return
		}
		writeIssueDetail(w, r.Context(), db, id)
	}
}

// UpdateIssue godoc
// @Summary      Update an issue (partial update)
// @Description  Updates specific fields of an issue. Only provided fields are changed.
// @Tags         issues
// @Accept       json
// @Produce      json
// @Param        id    path  integer               true  "Issue ID"
// @Param        body  body  models.UpdateIssueRequest  true  "Fields to update"
// @Success      200   {object}  models.IssueDetail
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /issues/{id} [patch]
func UpdateIssue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid issue id")
			return
		}

		var input models.UpdateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			utils.Error400(w, "invalid JSON body")
			return
		}

		// ── 先读当前值（partial update 需要现有值做 fallback） ────
		current, err := fetchIssue(r.Context(), db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.Error404(w, "issue not found")
			} else {
				slog.ErrorContext(r.Context(), "fetch issue for update failed", "error", err)
				utils.Error500(w, "database error")
			}
			return
		}

		// ── Partial Update 模式 ─────────────────────────────────────
		// 只有前端传了的字段才更新，没传的保持原值。
		// 这是 REST API 的惯用模式，区别于 PUT（全量替换）。
		title := coalesce(input.Title, current.Title)
		desc := coalesce(input.Description, current.Description)
		status := coalesce(input.Status, current.Status)
		priority := coalesce(input.Priority, current.Priority)
		issueType := coalesce(input.IssueType, current.IssueType)

		var assignee *string
		if input.Assignee != nil {
			assignee = *input.Assignee // 允许设为 null（置空 assignee）
		} else {
			assignee = current.Assignee
		}

		if msg := models.ValidateIssueFields(title, desc, priority, issueType); msg != "" {
			utils.Error400(w, msg)
			return
		}
		if msg := models.ValidateStatus(status); msg != "" {
			utils.Error400(w, msg)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			utils.Error500(w, "database error")
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec(
			`UPDATE issues SET title=?, description=?, status=?, priority=?, issue_type=?, assignee=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
			strings.TrimSpace(title), strings.TrimSpace(desc),
			status, priority, issueType, assignee, id,
		)
		if err != nil {
			slog.ErrorContext(r.Context(), "update issue failed", "issue_id", id, "error", err)
			utils.Error500(w, "failed to update issue")
			return
		}

		// ── 全量替换 label 关联 ──────────────────────────────────────
		// 前端传了 labelIds 就全量替换，没传就不动。
		// 简单粗暴：删掉旧的关联，插入新的。
		if input.LabelIDs != nil {
			tx.Exec("DELETE FROM issue_labels WHERE issue_id = ?", id)
			if err := linkLabelsTx(tx, id, input.LabelIDs); err != nil {
				slog.ErrorContext(r.Context(), "update labels failed", "error", err)
				utils.Error500(w, "failed to update labels")
				return
			}
		}

		if err := tx.Commit(); err != nil {
			utils.Error500(w, "database error")
			return
		}
		writeIssueDetail(w, r.Context(), db, id)
	}
}

// DeleteIssue godoc
// @Summary      Delete an issue
// @Description  Deletes an issue and all its associated comments, labels, and attachments.
// @Tags         issues
// @Accept       json
// @Produce      json
// @Param        id   path  integer  true  "Issue ID"
// @Success      200  {object}  map[string]bool
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues/{id} [delete]
func DeleteIssue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid issue id")
			return
		}

		// ── 检查影响行数 ────────────────────────────────────────────
		// RowsAffected 返回实际删除的行数。
		// 如果为 0，说明该记录不存在或已被删除。
		res, err := execContext(r.Context(), db, "DELETE FROM issues WHERE id = ?", id)
		if err != nil {
			slog.ErrorContext(r.Context(), "delete issue failed", "error", err)
			utils.Error500(w, "database error")
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			utils.Error404(w, "issue not found")
			return
		}

		utils.JSONDeleted(w)
	}
}

// ── 内部 helpers ─────────────────────────────────────────────────────────

// fetchIssue 从数据库读取单个 issue。
//
// 使用 QueryRowContext 而非 QueryRow，支持超时和取消。
// 注意：sql.ErrNoRows 是"没有找到"的标准错误，需要调用方自行判断。
func fetchIssue(ctx context.Context, db *sql.DB, id int64) (models.Issue, error) {
	row := queryRowContext(ctx, db, "SELECT * FROM issues WHERE id = ?", id)
	return models.ScanIssue(row)
}

// writeIssueDetail 查询 issue 及其关联数据，写入 JSON 响应。
//
// 这是一个 N+1 查询模式：1 次查 issue + 1 次查 labels + 1 次查 comments + 1 次查 attachments。
// 对于本项目的规模（单用户、少量数据），这种模式简单清晰。
// 对于大规模系统，应使用 JOIN 或 DataLoader 批量加载。
func writeIssueDetail(w http.ResponseWriter, ctx context.Context, db *sql.DB, id int64) {
	issue, err := fetchIssue(ctx, db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.Error404(w, "issue not found")
		} else {
			slog.ErrorContext(ctx, "fetch issue failed", "issue_id", id, "error", err)
			utils.Error500(w, "database error")
		}
		return
	}

	// ── 加载关联数据（顺序执行） ────────────────────────────────────
	// 顺序加载三个关联数据集。对于 SQLite（单连接串行），并发没有收益。
	// 对于 PostgreSQL 等多连接数据库，可改用并发 + errgroup 模式批量加载。
	labels, err := LabelsForIssue(ctx, db, id)
	if err != nil {
		slog.ErrorContext(ctx, "fetch labels failed", "issue_id", id, "error", err)
		utils.Error500(w, "database error")
		return
	}

	comments, err := CommentsForIssue(ctx, db, id)
	if err != nil {
		slog.ErrorContext(ctx, "fetch comments failed", "issue_id", id, "error", err)
		utils.Error500(w, "database error")
		return
	}

	attachments, err := AttachmentsForIssue(ctx, db, id)
	if err != nil {
		slog.ErrorContext(ctx, "fetch attachments failed", "issue_id", id, "error", err)
		utils.Error500(w, "database error")
		return
	}

	utils.JSON(w, models.IssueDetail{
		Issue: issue, Labels: labels, Comments: comments, Attachments: attachments,
	})
}

// linkLabelsTx 在事务中为 issue 关联 label。
//
// INSERT OR IGNORE 确保重复关联不会报错。
func linkLabelsTx(tx *sql.Tx, issueID int64, ids []int64) error {
	for _, lid := range ids {
		if _, err := tx.Exec(
			"INSERT OR IGNORE INTO issue_labels (issue_id, label_id) VALUES (?, ?)",
			issueID, lid,
		); err != nil {
			return err
		}
	}
	return nil
}

// coalesce 是 "if not nil then val else fallback" 的简写。
// 名字来自 SQL 的 COALESCE 函数。
// 这是 Go 中处理 optional pointer 参数的常见工具函数。
func coalesce(p *string, fallback string) string {
	if p != nil {
		return *p
	}
	return fallback
}
