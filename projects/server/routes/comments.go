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

func RegisterCommentRoutes(r chi.Router, db *sql.DB) {
	r.Get("/", ListComments(db))
	r.Post("/", CreateComment(db))
}

func RegisterCommentDeleteRoute(r chi.Router, db *sql.DB) {
	r.Delete("/{id}", DeleteComment(db))
}

// CommentsForIssue 被 GetIssue 内部复用。
// 接收 context 以支持超时链路。
func CommentsForIssue(ctx context.Context, db *sql.DB, issueID int64) ([]models.Comment, error) {
	rows, err := queryContext(ctx, db,
		"SELECT * FROM comments WHERE issue_id = ? ORDER BY created_at ASC, id ASC", issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return models.ScanComments(rows)
}

// ListComments godoc
// @Summary      List comments for an issue
// @Description  Returns all comments for a given issue, ordered by creation date.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        issueID  path  integer  true  "Issue ID"
// @Success      200  {object}  []models.Comment
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues/{issueID}/comments [get]
func ListComments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid issue id")
			return
		}
		if _, err := fetchIssue(r.Context(), db, issueID); err != nil {
			if err == sql.ErrNoRows {
				utils.Error404(w, "issue not found")
			} else {
				slog.ErrorContext(r.Context(), "fetch issue failed", "error", err)
				utils.Error500(w, "database error")
			}
			return
		}
		comments, err := CommentsForIssue(r.Context(), db, issueID)
		if err != nil {
			slog.ErrorContext(r.Context(), "list comments failed", "error", err)
			utils.Error500(w, "database error")
			return
		}
		utils.JSON(w, comments)
	}
}

// CreateComment godoc
// @Summary      Add a comment to an issue
// @Description  Creates a new comment on the specified issue.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        issueID  path  integer                    true  "Issue ID"
// @Param        body     body  models.CreateCommentRequest  true  "Comment data"
// @Success      200  {object}  models.Comment
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues/{issueID}/comments [post]
func CreateComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid issue id")
			return
		}

		var input models.CreateCommentRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			utils.Error400(w, "invalid JSON body")
			return
		}
		if strings.TrimSpace(input.Author) == "" {
			utils.Error400(w, "author is required")
			return
		}
		if strings.TrimSpace(input.Body) == "" {
			utils.Error400(w, "body is required")
			return
		}
		if _, err := fetchIssue(r.Context(), db, issueID); err != nil {
			if err == sql.ErrNoRows {
				utils.Error404(w, "issue not found")
			} else {
				utils.Error500(w, "database error")
			}
			return
		}

		var c models.Comment
		err = queryRowContext(r.Context(), db,
			`INSERT INTO comments (issue_id, author, body) VALUES (?, ?, ?) RETURNING *`,
			issueID, strings.TrimSpace(input.Author), strings.TrimSpace(input.Body),
		).Scan(&c.ID, &c.IssueID, &c.Author, &c.Body, &c.CreatedAt)
		if err != nil {
			slog.ErrorContext(r.Context(), "create comment failed", "error", err)
			utils.Error500(w, "failed to create comment")
			return
		}

		utils.JSON(w, c)
	}
}

// DeleteComment godoc
// @Summary      Delete a comment
// @Description  Deletes a comment by its ID.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        id   path  integer  true  "Comment ID"
// @Success      200  {object}  map[string]bool
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /comments/{id} [delete]
func DeleteComment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid comment id")
			return
		}
		res, err := execContext(r.Context(), db, "DELETE FROM comments WHERE id = ?", id)
		if err != nil {
			slog.ErrorContext(r.Context(), "delete comment failed", "error", err)
			utils.Error500(w, "database error")
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			utils.Error404(w, "comment not found")
			return
		}
		utils.JSONDeleted(w)
	}
}
