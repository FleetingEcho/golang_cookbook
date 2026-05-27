package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"issue_tracker/config"
	"issue_tracker/db"
	"issue_tracker/models"

	"github.com/go-chi/chi/v5"
)

// ── 测试工具 ───────────────────────────────────────────────────────────────
//
// 工程要点：
// 1. 每个测试使用独立的内存 SQLite 数据库，互不干扰
// 2. t.Helper() 标记辅助函数，报错时指向调用行而非辅助函数内部
// 3. t.TempDir() 自动生成临时目录，测试结束后自动清理

// setupTestDB 创建测试用的数据库和路由。
//
// 返回 (router, cleanupFunc)。
// 使用 cleanupFunc 确保数据库连接被关闭。
func setupTestDB(t *testing.T) (chi.Router, func()) {
	t.Helper()

	cfg := &config.Config{
		BindAddr:    "127.0.0.1:0",
		DatabaseURL: "sqlite://file::memory:?cache=shared",
		UploadDir:   t.TempDir(),
		APIKey:      "test-key",
	}

	database, err := db.Connect(cfg.DatabaseURL, 1, 1)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	r := chi.NewRouter()
	r.Route("/api", func(api chi.Router) {
		// 测试环境自动注入 API Key
		api.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("x-api-key", "test-key")
				next.ServeHTTP(w, r)
			})
		})

		api.Route("/issues", func(sub chi.Router) {
			RegisterIssueRoutes(sub, database)
			sub.Route("/{issueID}/comments", func(r chi.Router) {
				RegisterCommentRoutes(r, database)
			})
			sub.Route("/{issueID}/labels/{labelID}", func(r chi.Router) {
				RegisterIssueLabelRoutes(r, database)
			})
		})
		api.Route("/comments", func(r chi.Router) {
			RegisterCommentDeleteRoute(r, database)
		})
		api.Route("/labels", func(r chi.Router) {
			RegisterLabelRoutes(r, database)
		})
	})

	return r, func() { database.Close() }
}

// testRequest 发送测试 HTTP 请求。
//
// 使用 httptest.NewRequest + httptest.NewRecorder，
// 无需实际启动 HTTP 服务器，测试速度更快。
func testRequest(router chi.Router, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("x-api-key", "test-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

// ── 表驱动测试 ────────────────────────────────────────────────────────────
//
// 工程要点：
// 1. 表驱动测试是 Go 的标准测试模式
// 2. 每个测试用例是一个 struct，包含 name、input、expected
// 3. 添加/修改测试用例只需加一行 struct 条目
// 4. t.Run 使用子测试，单个用例失败不影响其他用例

// TestIssueCRUD 用表驱动方式测试 issue 的增删改查全流程。
func TestIssueCRUD(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()

	// ── 先创建 ─────────────────────────────────────────────────────
	w := testRequest(router, "POST", "/api/issues",
		`{"title":"Test Issue","description":"Test Desc","priority":"high","issueType":"bug","createdBy":"tester"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("create: got %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var detail models.IssueDetail
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// ── 表驱动测试 GET ─────────────────────────────────────────────
	getCases := []struct {
		name       string
		id         int64
		wantStatus int
		wantTitle  string
	}{
		{name: "existing issue", id: detail.ID, wantStatus: 200, wantTitle: "Test Issue"},
		{name: "non-existent issue", id: 99999, wantStatus: 404, wantTitle: ""},
		{name: "invalid id (zero)", id: 0, wantStatus: 404, wantTitle: ""},
	}

	for _, tc := range getCases {
		t.Run("GET/"+tc.name, func(t *testing.T) {
			w := testRequest(router, "GET", "/api/issues/"+strconv.FormatInt(tc.id, 10), "")
			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
			if tc.wantStatus == 200 {
				var d models.IssueDetail
				json.Unmarshal(w.Body.Bytes(), &d)
				if d.Title != tc.wantTitle {
					t.Errorf("title: got %q, want %q", d.Title, tc.wantTitle)
				}
			}
		})
	}

	// ── 表驱动测试 UPDATE ──────────────────────────────────────────
	updateCases := []struct {
		name       string
		id         int64
		body       string
		wantStatus int
		wantTitle  string
	}{
		{
			name:       "update title only",
			id:         detail.ID,
			body:       `{"title":"Updated Title"}`,
			wantStatus: 200,
			wantTitle:  "Updated Title",
		},
		{
			name:       "update status only",
			id:         detail.ID,
			body:       `{"status":"in_progress"}`,
			wantStatus: 200,
			wantTitle:  "Updated Title",
		},
		{
			name:       "invalid status",
			id:         detail.ID,
			body:       `{"status":"bad_status"}`,
			wantStatus: 400,
			wantTitle:  "",
		},
		{
			name:       "non-existent issue",
			id:         99999,
			body:       `{"title":"Nope"}`,
			wantStatus: 404,
			wantTitle:  "",
		},
	}

	for _, tc := range updateCases {
		t.Run("PATCH/"+tc.name, func(t *testing.T) {
			w := testRequest(router, "PATCH", "/api/issues/"+strconv.FormatInt(tc.id, 10), tc.body)
			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d: %s", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == 200 {
				var d models.IssueDetail
				json.Unmarshal(w.Body.Bytes(), &d)
				if d.Title != tc.wantTitle {
					t.Errorf("title: got %q, want %q", d.Title, tc.wantTitle)
				}
			}
		})
	}

	// ── 表驱动测试 DELETE ──────────────────────────────────────────
	deleteCases := []struct {
		name       string
		id         int64
		wantStatus int
	}{
		{name: "existing issue", id: detail.ID, wantStatus: 200},
		{name: "already deleted", id: detail.ID, wantStatus: 404},
		{name: "non-existent", id: 99999, wantStatus: 404},
	}

	for _, tc := range deleteCases {
		t.Run("DELETE/"+tc.name, func(t *testing.T) {
			w := testRequest(router, "DELETE", "/api/issues/"+strconv.FormatInt(tc.id, 10), "")
			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tc.wantStatus)
			}
		})
	}
}

// ── 创建 issue 校验 ───────────────────────────────────────────────────────
//
// 使用表驱动测试验证各种无效输入。

func TestCreateIssueValidation(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()

	cases := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "empty title", body: `{"title":"","description":"d","priority":"high","issueType":"bug","createdBy":"t"}`, wantStatus: 400},
		{name: "bad priority", body: `{"title":"t","description":"d","priority":"urgent","issueType":"bug","createdBy":"t"}`, wantStatus: 400},
		{name: "bad issue type", body: `{"title":"t","description":"d","priority":"high","issueType":"invalid","createdBy":"t"}`, wantStatus: 400},
		{name: "empty created_by", body: `{"title":"t","description":"d","priority":"high","issueType":"bug","createdBy":""}`, wantStatus: 400},
		{name: "valid", body: `{"title":"t","description":"d","priority":"high","issueType":"bug","createdBy":"t"}`, wantStatus: 200},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := testRequest(router, "POST", "/api/issues", tc.body)
			if w.Code != tc.wantStatus {
				t.Errorf("status: got %d, want %d: %s", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}
}

// ── 分页测试 ──────────────────────────────────────────────────────────────

func TestPagination(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()

	// 先创建 5 条 issue
	for i := 0; i < 5; i++ {
		testRequest(router, "POST", "/api/issues",
			`{"title":"T`+strconv.Itoa(i)+`","description":"d","priority":"medium","issueType":"task","createdBy":"t"}`)
	}

	tests := []struct {
		name        string
		query       string
		wantItems   int
		wantTotal   int64
		wantLimit   int64
		wantOffset  int64
	}{
		{name: "page 1, size 2", query: "?limit=2&offset=0", wantItems: 2, wantTotal: 5, wantLimit: 2, wantOffset: 0},
		{name: "page 2, size 2", query: "?limit=2&offset=2", wantItems: 2, wantTotal: 5, wantLimit: 2, wantOffset: 2},
		{name: "page 3, size 2", query: "?limit=2&offset=4", wantItems: 1, wantTotal: 5, wantLimit: 2, wantOffset: 4},
		{name: "default pagination", query: "", wantItems: 5, wantTotal: 5, wantLimit: 25, wantOffset: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := testRequest(router, "GET", "/api/issues"+tc.query, "")
			if w.Code != 200 {
				t.Fatalf("status: got %d", w.Code)
			}
			var r models.PaginatedResponse[models.Issue]
			json.Unmarshal(w.Body.Bytes(), &r)
			if len(r.Items) != tc.wantItems {
				t.Errorf("items: got %d, want %d", len(r.Items), tc.wantItems)
			}
			if r.Total != tc.wantTotal {
				t.Errorf("total: got %d, want %d", r.Total, tc.wantTotal)
			}
			if r.Limit != tc.wantLimit {
				t.Errorf("limit: got %d, want %d", r.Limit, tc.wantLimit)
			}
			if r.Offset != tc.wantOffset {
				t.Errorf("offset: got %d, want %d", r.Offset, tc.wantOffset)
			}
		})
	}
}

// ── Label 流程测试 ─────────────────────────────────────────────────────────

func TestLabelsFlow(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()

	// 先创建一个 label
	w := testRequest(router, "POST", "/api/labels", `{"name":"bug","color":"#ff0000"}`)
	if w.Code != 200 {
		t.Fatalf("create label: %d: %s", w.Code, w.Body.String())
	}
	var label models.Label
	json.Unmarshal(w.Body.Bytes(), &label)

	// 创建 issue 时关联 label
	w = testRequest(router, "POST", "/api/issues",
		`{"title":"Buggy","description":"Found a bug","priority":"high","issueType":"bug","createdBy":"t","labelIds":[`+
			strconv.FormatInt(label.ID, 10)+`]}`)
	if w.Code != 200 {
		t.Fatalf("create issue with label: %d: %s", w.Code, w.Body.String())
	}

	var detail models.IssueDetail
	json.Unmarshal(w.Body.Bytes(), &detail)
	if len(detail.Labels) != 1 || detail.Labels[0].Name != "bug" {
		t.Fatalf("labels: got %+v, want [bug]", detail.Labels)
	}
}

// ── Comment 全流程测试 ────────────────────────────────────────────────────

func TestCommentsFlow(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()

	// 创建 issue
	w := testRequest(router, "POST", "/api/issues",
		`{"title":"T","description":"d","priority":"low","issueType":"task","createdBy":"t"}`)
	var detail models.IssueDetail
	json.Unmarshal(w.Body.Bytes(), &detail)
	issueID := detail.ID

	// 创建 comment
	w = testRequest(router, "POST", "/api/issues/"+strconv.FormatInt(issueID, 10)+"/comments",
		`{"author":"Teng","body":"ok"}`)
	if w.Code != 200 {
		t.Fatalf("create comment: %d: %s", w.Code, w.Body.String())
	}

	var c models.Comment
	json.Unmarshal(w.Body.Bytes(), &c)
	if c.Author != "Teng" {
		t.Fatalf("author: got %q, want 'Teng'", c.Author)
	}

	// 删除 comment
	w = testRequest(router, "DELETE", "/api/comments/"+strconv.FormatInt(c.ID, 10), "")
	if w.Code != 200 {
		t.Fatalf("delete: %d", w.Code)
	}

	// 确认删除后 404
	w = testRequest(router, "DELETE", "/api/comments/"+strconv.FormatInt(c.ID, 10), "")
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
