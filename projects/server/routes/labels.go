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

func RegisterLabelRoutes(r chi.Router, db *sql.DB) {
	r.Get("/", ListLabels(db))
	r.Post("/", CreateLabel(db))
}

func RegisterIssueLabelRoutes(r chi.Router, db *sql.DB) {
	r.Post("/", AddIssueLabel(db))
	r.Delete("/", RemoveIssueLabel(db))
}

func LabelsForIssue(db *sql.DB, issueID int64) ([]models.Label, error) {
	rows, err := db.Query(
		`SELECT labels.* FROM labels
		 JOIN issue_labels ON labels.id = issue_labels.label_id
		 WHERE issue_labels.issue_id = ? ORDER BY labels.name`, issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return models.ScanLabels(rows)
}

func ListLabels(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM labels ORDER BY name")
		if err != nil {
			slog.Error("list labels", "error", err)
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		defer rows.Close()
		labels, err := models.ScanLabels(rows)
		if err != nil {
			slog.Error("scan labels", "error", err)
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(labels)
	}
}

func CreateLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input models.CreateLabelRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid JSON body"))
			return
		}
		if strings.TrimSpace(input.Name) == "" {
			utils.WriteError(w, utils.NewBadRequest("name is required"))
			return
		}
		if strings.TrimSpace(input.Color) == "" {
			utils.WriteError(w, utils.NewBadRequest("color is required"))
			return
		}
		var l models.Label
		err := db.QueryRow("INSERT INTO labels (name, color) VALUES (?, ?) RETURNING *",
			strings.TrimSpace(input.Name), strings.TrimSpace(input.Color),
		).Scan(&l.ID, &l.Name, &l.Color)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				utils.WriteError(w, utils.NewBadRequest("label name already exists"))
				return
			}
			slog.Error("create label", "error", err)
			utils.WriteError(w, utils.NewInternalError("failed to create label"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(l)
	}
}

func AddIssueLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid issue id"))
			return
		}
		labelID, err := strconv.ParseInt(chi.URLParam(r, "labelID"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid label id"))
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
		var exists int
		if err := db.QueryRow("SELECT 1 FROM labels WHERE id = ?", labelID).Scan(&exists); err != nil {
			utils.WriteError(w, utils.NewNotFound("label not found"))
			return
		}
		db.Exec("INSERT OR IGNORE INTO issue_labels (issue_id, label_id) VALUES (?, ?)", issueID, labelID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"linked": true})
	}
}

func RemoveIssueLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid issue id"))
			return
		}
		labelID, err := strconv.ParseInt(chi.URLParam(r, "labelID"), 10, 64)
		if err != nil {
			utils.WriteError(w, utils.NewBadRequest("invalid label id"))
			return
		}
		res, err := db.Exec("DELETE FROM issue_labels WHERE issue_id = ? AND label_id = ?", issueID, labelID)
		if err != nil {
			utils.WriteError(w, utils.NewInternalError("database error"))
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			utils.WriteError(w, utils.NewNotFound("issue label link not found"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
	}
}
