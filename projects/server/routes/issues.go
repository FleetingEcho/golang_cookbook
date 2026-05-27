package routes

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"issue_tracker/middleware"
	"issue_tracker/models"
	"issue_tracker/utils"

	"github.com/go-chi/chi/v5"
)

func RegisterIssueRoutes(r chi.Router, db *sql.DB) {
	r.Get("/", ListIssues(db))
	r.Post("/", CreateIssue(db))
	r.Get("/{id}", GetIssue(db))
	r.Patch("/{id}", UpdateIssue(db))
	r.Delete("/{id}", DeleteIssue(db))
}

// ── SQL Builder ──────────────────────────────────────────────────────────────

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

// ── List ─────────────────────────────────────────────────────────────────────

func ListIssues(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := parseIssueQuery(r)

		// Count
		cb := &sqlBuilder{}
		cb.Add("SELECT COUNT(DISTINCT issues.id) FROM issues")
		if q.LabelID != nil {
			cb.Add("JOIN issue_labels ON issues.id = issue_labels.issue_id")
		}
		cb.Add("WHERE 1=1")
		addIssueFilters(cb, q)

		var total int64
		cs, ca := cb.Build()
		if err := db.QueryRow(cs, ca...).Scan(&total); err != nil {
			slog.Error("count issues", "error", err, "request_id", middleware.GetRequestID(r.Context()))
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}

		// Data
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
		rows, err := db.Query(ds, da...)
		if err != nil {
			slog.Error("list issues", "error", err, "request_id", middleware.GetRequestID(r.Context()))
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		defer rows.Close()

		items, err := models.ScanIssues(rows)
		if err != nil {
			slog.Error("scan issues", "error", err, "request_id", middleware.GetRequestID(r.Context()))
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.PaginatedResponse[models.Issue]{
			Items: items, Total: total, Limit: q.Limit, Offset: q.Offset,
		})
	}
}

func parseIssueQuery(r *http.Request) models.IssueQuery {
	q := models.IssueQuery{Limit: 5}
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
		if l, err := strconv.ParseInt(v, 10, 64); err == nil && l > 0 {
			q.Limit = l
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

// ── Create ───────────────────────────────────────────────────────────────────

func CreateIssue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input models.CreateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid JSON body"))
			return
		}
		if msg := models.ValidateIssueFields(input.Title, input.Description, input.Priority, input.IssueType); msg != "" {
			utils.WriteError(w, utils.NewBadRequest(msg))
			return
		}
		if strings.TrimSpace(input.CreatedBy) == "" {
			utils.WriteError(w, utils.NewBadRequest("created_by is required"))
			return
		}

		tx, err := db.Begin()
		if err != nil {
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		defer tx.Rollback()

		var id int64
		err = tx.QueryRow(
			`INSERT INTO issues (title, description, priority, issue_type, assignee, created_by) VALUES (?, ?, ?, ?, ?, ?) RETURNING id`,
			strings.TrimSpace(input.Title), strings.TrimSpace(input.Description),
			input.Priority, input.IssueType, input.Assignee, strings.TrimSpace(input.CreatedBy),
		).Scan(&id)
		if err != nil {
			slog.Error("create issue", "error", err)
			utils.WriteError(w, utils.NewInternalError("failed to create issue"))
			return
		}

		if err := linkLabels(tx, id, input.LabelIDs); err != nil {
			slog.Error("link labels", "error", err)
			utils.WriteError(w, utils.NewInternalError("failed to link labels"))
			return
		}
		if err := tx.Commit(); err != nil {
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		writeIssueDetail(w, db, id)
	}
}

// ── Get ──────────────────────────────────────────────────────────────────────

func GetIssue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid issue id"))
			return
		}
		writeIssueDetail(w, db, id)
	}
}

// ── Update ───────────────────────────────────────────────────────────────────

func UpdateIssue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid issue id"))
			return
		}

		var input models.UpdateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid JSON body"))
			return
		}

		current, err := fetchIssue(db, id)
		if err != nil {
			if err == sql.ErrNoRows {
				utils.WriteError(w, utils.NewNotFound("issue not found"))
			} else {
				utils.WriteError(w, utils.NewInternalError("database error"))
			}
			return
		}

		title := coalesce(input.Title, current.Title)
		desc := coalesce(input.Description, current.Description)
		status := coalesce(input.Status, current.Status)
		priority := coalesce(input.Priority, current.Priority)
		issueType := coalesce(input.IssueType, current.IssueType)
		var assignee *string
		if input.Assignee != nil {
			assignee = *input.Assignee
		} else {
			assignee = current.Assignee
		}

		if msg := models.ValidateIssueFields(title, desc, priority, issueType); msg != "" {
			utils.WriteError(w, utils.NewBadRequest(msg))
			return
		}
		if msg := models.ValidateStatus(status); msg != "" {
			utils.WriteError(w, utils.NewBadRequest(msg))
			return
		}

		tx, err := db.Begin()
		if err != nil {
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec(
			`UPDATE issues SET title=?, description=?, status=?, priority=?, issue_type=?, assignee=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
			strings.TrimSpace(title), strings.TrimSpace(desc),
			status, priority, issueType, assignee, id,
		)
		if err != nil {
			slog.Error("update issue", "error", err, "issue_id", id)
			utils.WriteError(w, utils.NewInternalError("failed to update issue"))
			return
		}

		if input.LabelIDs != nil {
			tx.Exec("DELETE FROM issue_labels WHERE issue_id = ?", id)
			if err := linkLabels(tx, id, input.LabelIDs); err != nil {
				utils.WriteError(w, utils.NewInternalError("failed to update labels"))
				return
			}
		}

		if err := tx.Commit(); err != nil {
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		writeIssueDetail(w, db, id)
	}
}

// ── Delete ───────────────────────────────────────────────────────────────────

func DeleteIssue(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid issue id"))
			return
		}
		res, err := db.Exec("DELETE FROM issues WHERE id = ?", id)
		if err != nil {
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			utils.WriteError(w, utils.NewNotFound("issue not found"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func fetchIssue(db *sql.DB, id int64) (models.Issue, error) {
	return models.ScanIssue(db.QueryRow("SELECT * FROM issues WHERE id = ?", id))
}

func writeIssueDetail(w http.ResponseWriter, db *sql.DB, id int64) {
	issue, err := fetchIssue(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteError(w, utils.NewNotFound("issue not found"))
		} else {
			utils.WriteError(w, utils.NewInternalError("database error"))
		}
		return
	}

	labels, err := LabelsForIssue(db, id)
	if err != nil {
		slog.Error("fetch labels", "issue_id", id, "error", err)
		utils.WriteError(w, utils.NewInternalError("database error"))
		return
	}
	comments, err := CommentsForIssue(db, id)
	if err != nil {
		slog.Error("fetch comments", "issue_id", id, "error", err)
		utils.WriteError(w, utils.NewInternalError("database error"))
		return
	}
	attachments, err := AttachmentsForIssue(db, id)
	if err != nil {
		slog.Error("fetch attachments", "issue_id", id, "error", err)
		utils.WriteError(w, utils.NewInternalError("database error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.IssueDetail{
		Issue: issue, Labels: labels, Comments: comments, Attachments: attachments,
	})
}

func linkLabels(tx *sql.Tx, issueID int64, ids []int64) error {
	for _, lid := range ids {
		if _, err := tx.Exec("INSERT OR IGNORE INTO issue_labels (issue_id, label_id) VALUES (?, ?)", issueID, lid); err != nil {
			return err
		}
	}
	return nil
}

func coalesce(p *string, fallback string) string {
	if p != nil {
		return *p
	}
	return fallback
}
