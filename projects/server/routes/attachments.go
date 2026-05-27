package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"issue_tracker/models"
	"issue_tracker/utils"

	"github.com/go-chi/chi/v5"
)

type AttachmentHandler struct {
	DB        *sql.DB
	UploadDir string
}

func RegisterAttachmentRoutes(r chi.Router, h *AttachmentHandler) {
	r.Get("/", h.ListAttachments)
	r.Post("/", h.UploadAttachment)
}

func RegisterAttachmentDownloadRoute(r chi.Router, h *AttachmentHandler) {
	r.Get("/download", h.DownloadAttachment)
	r.Delete("/", h.DeleteAttachment)
}

func AttachmentsForIssue(db *sql.DB, issueID int64) ([]models.Attachment, error) {
	rows, err := db.Query(
		"SELECT * FROM attachments WHERE issue_id = ? ORDER BY created_at DESC, id DESC", issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return models.ScanAttachments(rows)
}

func (h *AttachmentHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
	issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
	if err != nil {
		utils.WriteError(w, utils.NewBadRequest("invalid issue id"))
		return
	}
	if _, err := fetchIssue(h.DB, issueID); err != nil {
		if err == sql.ErrNoRows {
			utils.WriteError(w, utils.NewNotFound("issue not found"))
		} else {
			utils.WriteError(w, utils.NewInternalError("database error"))
		}
		return
	}
	items, err := AttachmentsForIssue(h.DB, issueID)
	if err != nil {
		slog.Error("list attachments", "error", err)
		utils.WriteError(w, utils.NewInternalError("database error"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *AttachmentHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
	if err != nil {
		utils.WriteError(w, utils.NewBadRequest("invalid issue id"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			utils.WriteError(w, utils.NewBadRequest("file too large, max 10MB"))
		} else {
			utils.WriteError(w, utils.NewBadRequest("invalid multipart form"))
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.WriteError(w, utils.NewBadRequest("no file part in multipart form"))
		return
	}
	defer file.Close()

	origName := header.Filename
	if origName == "" {
		origName = "unnamed"
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	data, err := io.ReadAll(file)
	if err != nil {
		utils.WriteError(w, utils.NewInternalError("failed to read file"))
		return
	}
	size := int64(len(data))

	storedName := utils.StoredFileName(origName)
	filePath, err := utils.EnsureInsideUploadDir(h.UploadDir, storedName)
	if err != nil {
		utils.WriteError(w, utils.NewBadRequest("invalid file name"))
		return
	}
	os.MkdirAll(h.UploadDir, 0o755)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		slog.Error("write upload file", "error", err)
		utils.WriteError(w, utils.NewInternalError("failed to save file"))
		return
	}

	var a models.Attachment
	err = h.DB.QueryRow(
		`INSERT INTO attachments (issue_id, original_filename, stored_filename, content_type, size_bytes) VALUES (?, ?, ?, ?, ?) RETURNING *`,
		issueID, origName, storedName, contentType, size,
	).Scan(&a.ID, &a.IssueID, &a.OriginalFilename, &a.StoredFilename, &a.ContentType, &a.SizeBytes, &a.CreatedAt)
	if err != nil {
		slog.Error("insert attachment", "error", err)
		utils.WriteError(w, utils.NewInternalError("failed to save attachment record"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

func (h *AttachmentHandler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteError(w, utils.NewBadRequest("invalid attachment id"))
		return
	}

	a, err := models.ScanAttachment(h.DB.QueryRow("SELECT * FROM attachments WHERE id = ?", id))
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteError(w, utils.NewNotFound("attachment not found"))
		} else {
			utils.WriteError(w, utils.NewInternalError("database error"))
		}
		return
	}

	filePath, err := utils.EnsureInsideUploadDir(h.UploadDir, a.StoredFilename)
	if err != nil {
		utils.WriteError(w, utils.NewInternalError("invalid stored path"))
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error("read attachment file", "path", filePath, "error", err)
		utils.WriteError(w, utils.NewNotFound("file not found on disk"))
		return
	}

	w.Header().Set("Content-Type", detectContentType(a.OriginalFilename))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, a.OriginalFilename))
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
	w.Write(data)
}

func (h *AttachmentHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteError(w, utils.NewBadRequest("invalid attachment id"))
		return
	}

	a, err := models.ScanAttachment(h.DB.QueryRow("SELECT * FROM attachments WHERE id = ?", id))
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteError(w, utils.NewNotFound("attachment not found"))
		} else {
			utils.WriteError(w, utils.NewInternalError("database error"))
		}
		return
	}

	h.DB.Exec("DELETE FROM attachments WHERE id = ?", id)
	if fp, err := utils.EnsureInsideUploadDir(h.UploadDir, a.StoredFilename); err == nil {
		os.Remove(fp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
}

func detectContentType(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".md":
		return "text/markdown; charset=utf-8"
	case ".csv":
		return "text/csv; charset=utf-8"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}
