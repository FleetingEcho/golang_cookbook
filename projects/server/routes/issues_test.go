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

func setupTestDB(t *testing.T) (chi.Router, func()) {
	t.Helper()
	cfg := &config.Config{
		BindAddr:    "127.0.0.1:0",
		DatabaseURL: "sqlite://file::memory:?cache=shared",
		UploadDir:   t.TempDir(),
		APIKey:      "test-key",
	}
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	r := chi.NewRouter()
	r.Route("/api", func(api chi.Router) {
		api.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				r.Header.Set("x-api-key", "test-key")
				next.ServeHTTP(w, r)
			})
		})
		api.Route("/issues", func(sub chi.Router) {
			RegisterIssueRoutes(sub, database)
			sub.Route("/{issueID}/comments", func(r chi.Router) { RegisterCommentRoutes(r, database) })
			sub.Route("/{issueID}/labels/{labelID}", func(r chi.Router) { RegisterIssueLabelRoutes(r, database) })
		})
		api.Route("/comments", func(r chi.Router) { RegisterCommentDeleteRoute(r, database) })
		api.Route("/labels", func(r chi.Router) { RegisterLabelRoutes(r, database) })
	})
	return r, func() { database.Close() }
}

func req(router chi.Router, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("x-api-key", "test-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

func TestCreateAndListIssues(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()
	for i := 0; i < 3; i++ {
		w := req(router, "POST", "/api/issues",
			`{"title":"T`+strconv.Itoa(i)+`","description":"D`+strconv.Itoa(i)+`","priority":"high","issueType":"bug","createdBy":"tester"}`)
		if w.Code != 200 {
			t.Fatalf("create %d: %d: %s", i, w.Code, w.Body.String())
		}
	}
	w := req(router, "GET", "/api/issues?limit=10&offset=0", "")
	if w.Code != 200 {
		t.Fatalf("list: %d", w.Code)
	}
	var r models.PaginatedResponse[models.Issue]
	json.Unmarshal(w.Body.Bytes(), &r)
	if len(r.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(r.Items))
	}
	if r.Total != 3 {
		t.Fatalf("expected total 3, got %d", r.Total)
	}
}

func TestCreateIssueValidation(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()
	w := req(router, "POST", "/api/issues", `{"title":"","description":"d","priority":"high","issueType":"bug","createdBy":"t"}`)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetIssue(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()
	w := req(router, "POST", "/api/issues", `{"title":"My Issue","description":"My Desc","priority":"high","issueType":"bug","createdBy":"alice"}`)
	var detail models.IssueDetail
	json.Unmarshal(w.Body.Bytes(), &detail)
	if detail.Title != "My Issue" || detail.CreatedBy != "alice" {
		t.Fatalf("unexpected: %+v", detail)
	}
}

func TestPagination(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()
	for i := 0; i < 5; i++ {
		req(router, "POST", "/api/issues", `{"title":"T`+strconv.Itoa(i)+`","description":"d","priority":"medium","issueType":"task","createdBy":"t"}`)
	}
	w := req(router, "GET", "/api/issues?limit=2&offset=1", "")
	if w.Code != 200 {
		t.Fatalf("list: %d", w.Code)
	}
	var r models.PaginatedResponse[models.Issue]
	json.Unmarshal(w.Body.Bytes(), &r)
	if len(r.Items) != 2 || r.Total != 5 || r.Limit != 2 || r.Offset != 1 {
		t.Fatalf("pagination mismatch: %+v", r)
	}
}

func TestLabelsFlow(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()
	w := req(router, "POST", "/api/labels", `{"name":"bug","color":"#ff0000"}`)
	var label models.Label
	json.Unmarshal(w.Body.Bytes(), &label)
	w = req(router, "POST", "/api/issues", `{"title":"B","description":"b","priority":"high","issueType":"bug","createdBy":"t","labelIds":[`+strconv.FormatInt(label.ID, 10)+`]}`)
	var detail models.IssueDetail
	json.Unmarshal(w.Body.Bytes(), &detail)
	if len(detail.Labels) != 1 || detail.Labels[0].Name != "bug" {
		t.Fatalf("labels: %+v", detail.Labels)
	}
}

func TestCommentsFlow(t *testing.T) {
	router, cleanup := setupTestDB(t)
	defer cleanup()
	w := req(router, "POST", "/api/issues", `{"title":"T","description":"d","priority":"low","issueType":"task","createdBy":"t"}`)
	var detail models.IssueDetail
	json.Unmarshal(w.Body.Bytes(), &detail)
	iid := detail.ID

	w = req(router, "POST", "/api/issues/"+strconv.FormatInt(iid, 10)+"/comments", `{"author":"Teng","body":"ok"}`)
	var c models.Comment
	json.Unmarshal(w.Body.Bytes(), &c)
	if c.Author != "Teng" {
		t.Fatalf("author mismatch: %+v", c)
	}

	w = req(router, "DELETE", "/api/comments/"+strconv.FormatInt(c.ID, 10), "")
	if w.Code != 200 {
		t.Fatalf("delete: %d", w.Code)
	}
	w = req(router, "DELETE", "/api/comments/"+strconv.FormatInt(c.ID, 10), "")
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
