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

func RegisterLabelRoutes(r chi.Router, db *sql.DB) {
	r.Get("/", ListLabels(db))
	r.Post("/", CreateLabel(db))
}

func RegisterIssueLabelRoutes(r chi.Router, db *sql.DB) {
	r.Post("/", AddIssueLabel(db))
	r.Delete("/", RemoveIssueLabel(db))
}

// LabelsForIssue 被 GetIssue 内部复用。
func LabelsForIssue(ctx context.Context, db *sql.DB, issueID int64) ([]models.Label, error) {
	rows, err := queryContext(ctx, db,
		`SELECT labels.* FROM labels
		 JOIN issue_labels ON labels.id = issue_labels.label_id
		 WHERE issue_labels.issue_id = ? ORDER BY labels.name`, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return models.ScanLabels(rows)
}

// ListLabels godoc
// @Summary      List all labels
// @Description  Returns all available labels, ordered by name.
// @Tags         labels
// @Accept       json
// @Produce      json
// @Success      200  {object}  []models.Label
// @Failure      500  {object}  map[string]string
// @Router       /labels [get]
func ListLabels(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := queryContext(r.Context(), db, "SELECT * FROM labels ORDER BY name")
		if err != nil {
			slog.ErrorContext(r.Context(), "list labels failed", "error", err)
			utils.Error500(w, "database error")
			return
		}
		defer rows.Close()

		labels, err := models.ScanLabels(rows)
		if err != nil {
			slog.ErrorContext(r.Context(), "scan labels failed", "error", err)
			utils.Error500(w, "database error")
			return
		}
		utils.JSON(w, labels)
	}
}

// CreateLabel godoc
// @Summary      Create a new label
// @Description  Creates a new label with a name and color.
// @Tags         labels
// @Accept       json
// @Produce      json
// @Param        body  body  models.CreateLabelRequest  true  "Label data"
// @Success      200  {object}  models.Label
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /labels [post]
func CreateLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input models.CreateLabelRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			utils.Error400(w, "invalid JSON body")
			return
		}
		if strings.TrimSpace(input.Name) == "" {
			utils.Error400(w, "name is required")
			return
		}
		if strings.TrimSpace(input.Color) == "" {
			utils.Error400(w, "color is required")
			return
		}

		var l models.Label
		err := queryRowContext(r.Context(), db,
			"INSERT INTO labels (name, color) VALUES (?, ?) RETURNING *",
			strings.TrimSpace(input.Name), strings.TrimSpace(input.Color),
		).Scan(&l.ID, &l.Name, &l.Color)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				utils.Error400(w, "label name already exists")
				return
			}
			slog.ErrorContext(r.Context(), "create label failed", "error", err)
			utils.Error500(w, "failed to create label")
			return
		}
		utils.JSON(w, l)
	}
}

// AddIssueLabel godoc
// @Summary      Attach a label to an issue
// @Description  Links an existing label to an issue.
// @Tags         labels
// @Accept       json
// @Produce      json
// @Param        issueID  path  integer  true  "Issue ID"
// @Param        labelID  path  integer  true  "Label ID"
// @Success      200  {object}  map[string]bool
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues/{issueID}/labels/{labelID} [post]
func AddIssueLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid issue id")
			return
		}
		labelID, err := strconv.ParseInt(chi.URLParam(r, "labelID"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid label id")
			return
		}

		// 同时验证 issue 和 label 存在
		if _, err := fetchIssue(r.Context(), db, issueID); err != nil {
			if err == sql.ErrNoRows {
				utils.Error404(w, "issue not found")
			} else {
				utils.Error500(w, "database error")
			}
			return
		}
		var exists int
		if err := queryRowContext(r.Context(), db, "SELECT 1 FROM labels WHERE id = ?", labelID).Scan(&exists); err != nil {
			utils.Error404(w, "label not found")
			return
		}

		execContext(r.Context(), db,
			"INSERT OR IGNORE INTO issue_labels (issue_id, label_id) VALUES (?, ?)", issueID, labelID)
		utils.JSONLinked(w)
	}
}

// RemoveIssueLabel godoc
// @Summary      Detach a label from an issue
// @Description  Removes the link between a label and an issue.
// @Tags         labels
// @Accept       json
// @Produce      json
// @Param        issueID  path  integer  true  "Issue ID"
// @Param        labelID  path  integer  true  "Label ID"
// @Success      200  {object}  map[string]bool
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues/{issueID}/labels/{labelID} [delete]
func RemoveIssueLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid issue id")
			return
		}
		labelID, err := strconv.ParseInt(chi.URLParam(r, "labelID"), 10, 64)
		if err != nil {
			utils.Error400(w, "invalid label id")
			return
		}
		res, err := execContext(r.Context(), db,
			"DELETE FROM issue_labels WHERE issue_id = ? AND label_id = ?", issueID, labelID)
		if err != nil {
			slog.ErrorContext(r.Context(), "remove label failed", "error", err)
			utils.Error500(w, "database error")
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			utils.Error404(w, "issue label link not found")
			return
		}
		utils.JSONDeleted(w)
	}
}
