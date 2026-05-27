package routes

import (
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

func RegisterCommentRoutes(r chi.Router, db *sql.DB) {
	r.Get("/", ListComments(db))
	r.Post("/", CreateComment(db))
}

func RegisterCommentDeleteRoute(r chi.Router, db *sql.DB) {
	r.Delete("/{id}", DeleteComment(db))
}

func CommentsForIssue(db *sql.DB, issueID int64) ([]models.Comment, error) {
	rows, err := db.Query(
		"SELECT * FROM comments WHERE issue_id = ? ORDER BY created_at ASC, id ASC", issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return models.ScanComments(rows)
}

func ListComments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid issue id"))
			return
		}
		if _, err := fetchIssue(db, issueID); err != nil {
			if err == sql.ErrNoRows {
				utils.WriteError(w, utils.NewNotFound("issue not found"))
			} else {
				utils.WriteError(w, utils.NewInternalError("database error"))
			}
			return
		}
		comments, err := CommentsForIssue(db, issueID)
		if err != nil {
			slog.Error("list comments", "error", err)
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	}
}

func CreateComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid issue id"))
			return
		}
		var input models.CreateCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid JSON body"))
			return
		}
		if strings.TrimSpace(input.Author) == "" {
			utils.WriteError(w, utils.NewBadRequest("author is required"))
			return
		}
		if strings.TrimSpace(input.Body) == "" {
			utils.WriteError(w, utils.NewBadRequest("body is required"))
			return
		}
		if _, err := fetchIssue(db, issueID); err != nil {
			if err == sql.ErrNoRows {
				utils.WriteError(w, utils.NewNotFound("issue not found"))
			} else {
				utils.WriteError(w, utils.NewInternalError("database error"))
			}
			return
		}

		var c models.Comment
		err = db.QueryRow(
			`INSERT INTO comments (issue_id, author, body) VALUES (?, ?, ?) RETURNING *`,
			issueID, strings.TrimSpace(input.Author), strings.TrimSpace(input.Body),
		).Scan(&c.ID, &c.IssueID, &c.Author, &c.Body, &c.CreatedAt)
		if err != nil {
			slog.Error("create comment", "error", err)
			utils.WriteError(w, utils.NewInternalError("failed to create comment"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(c)
	}
}

func DeleteComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid comment id"))
			return
		}
		res, err := db.Exec("DELETE FROM comments WHERE id = ?", id)
		if err != nil {
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			utils.WriteError(w, utils.NewNotFound("comment not found"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
	}
}
