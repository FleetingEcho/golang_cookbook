package routes

import (
	"context"
	"database/sql"
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

// AttachmentHandler 封装了附件操作所需的依赖。
//
// 设计模式：当 handler 需要多个依赖时，用 struct 替代函数闭包。
// 这样比多层 func(db) func(w,r) 更清晰，也更容易测试。
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

func AttachmentsForIssue(ctx context.Context, db *sql.DB, issueID int64) ([]models.Attachment, error) {
	rows, err := queryContext(ctx, db,
		"SELECT * FROM attachments WHERE issue_id = ? ORDER BY created_at DESC, id DESC", issueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return models.ScanAttachments(rows)
}

// ListAttachments godoc
// @Summary      List attachments for an issue
// @Description  Returns all attachments for a given issue.
// @Tags         attachments
// @Accept       json
// @Produce      json
// @Param        issueID  path  integer  true  "Issue ID"
// @Success      200  {object}  []models.Attachment
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues/{issueID}/attachments [get]
func (h *AttachmentHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
	issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
	if err != nil {
		utils.Error400(w, "invalid issue id")
		return
	}
	if _, err := fetchIssue(r.Context(), h.DB, issueID); err != nil {
		if err == sql.ErrNoRows {
			utils.Error404(w, "issue not found")
		} else {
			utils.Error500(w, "database error")
		}
		return
	}
	items, err := AttachmentsForIssue(r.Context(), h.DB, issueID)
	if err != nil {
		slog.ErrorContext(r.Context(), "list attachments failed", "error", err)
		utils.Error500(w, "database error")
		return
	}
	utils.JSON(w, items)
}

// UploadAttachment godoc
// @Summary      Upload a file attachment
// @Description  Uploads a file as an attachment to an issue (multipart/form-data). Max 10MB.
// @Tags         attachments
// @Accept       multipart/form-data
// @Produce      json
// @Param        issueID  path    integer  true   "Issue ID"
// @Param        file     formData  file    true   "File to upload"
// @Success      200  {object}  models.Attachment
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /issues/{issueID}/attachments [post]
func (h *AttachmentHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	issueID, err := strconv.ParseInt(chi.URLParam(r, "issueID"), 10, 64)
	if err != nil {
		utils.Error400(w, "invalid issue id")
		return
	}

	// ── 请求体大小限制 ──────────────────────────────────────────────
	// http.MaxBytesReader 在读取超过限制时返回错误，而不是缓冲整个请求体。
	// 这防止了恶意大文件耗尽服务器内存。
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			utils.Error400(w, "file too large, max 10MB")
		} else {
			utils.Error400(w, "invalid multipart form")
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.Error400(w, "no file part in multipart form")
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
		slog.ErrorContext(r.Context(), "read upload file failed", "error", err)
		utils.Error500(w, "failed to read file")
		return
	}
	size := int64(len(data))

	// ── 文件名安全 ──────────────────────────────────────────────────
	// 用 UUID + 安全转义，防止路径遍历攻击
	storedName := utils.StoredFileName(origName)
	filePath, err := utils.EnsureInsideUploadDir(h.UploadDir, storedName)
	if err != nil {
		utils.Error400(w, "invalid file name")
		return
	}

	// ── 写入磁盘 ────────────────────────────────────────────────────
	// os.WriteFile 在写入前 truncate，确保不会残留旧数据。
	if err := os.MkdirAll(h.UploadDir, 0o755); err != nil {
		utils.Error500(w, "failed to create upload directory")
		return
	}
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		slog.ErrorContext(r.Context(), "write upload file failed", "path", filePath, "error", err)
		utils.Error500(w, "failed to save file")
		return
	}

	// ── 写入数据库 ──────────────────────────────────────────────────
	var a models.Attachment
	err = queryRowContext(r.Context(), h.DB,
		`INSERT INTO attachments (issue_id, original_filename, stored_filename, content_type, size_bytes)
		 VALUES (?, ?, ?, ?, ?) RETURNING *`,
		issueID, origName, storedName, contentType, size,
	).Scan(&a.ID, &a.IssueID, &a.OriginalFilename, &a.StoredFilename, &a.ContentType, &a.SizeBytes, &a.CreatedAt)
	if err != nil {
		slog.ErrorContext(r.Context(), "insert attachment failed", "error", err)
		utils.Error500(w, "failed to save attachment record")
		return
	}

	utils.JSON(w, a)
}

// DownloadAttachment godoc
// @Summary      Download an attachment
// @Description  Downloads the original file by attachment ID.
// @Tags         attachments
// @Produce      octet-stream
// @Param        id   path  integer  true  "Attachment ID"
// @Success      200  {file}  binary
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /attachments/{id}/download [get]
func (h *AttachmentHandler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error400(w, "invalid attachment id")
		return
	}

	a, err := models.ScanAttachment(
		queryRowContext(r.Context(), h.DB, "SELECT * FROM attachments WHERE id = ?", id))
	if err != nil {
		if err == sql.ErrNoRows {
			utils.Error404(w, "attachment not found")
		} else {
			slog.ErrorContext(r.Context(), "fetch attachment failed", "error", err)
			utils.Error500(w, "database error")
		}
		return
	}

	filePath, err := utils.EnsureInsideUploadDir(h.UploadDir, a.StoredFilename)
	if err != nil {
		slog.ErrorContext(r.Context(), "invalid stored path", "stored", a.StoredFilename)
		utils.Error500(w, "invalid stored path")
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.ErrorContext(r.Context(), "read attachment file failed", "path", filePath, "error", err)
		utils.Error404(w, "file not found on disk")
		return
	}

	// ── 设置正确的 Content-Type 和 Content-Disposition ──────────────
	// 这确保浏览器能正确识别文件类型并触发下载行为。
	w.Header().Set("Content-Type", detectContentType(a.OriginalFilename))
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, a.OriginalFilename))
	w.Header().Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
	w.Write(data)
}

// DeleteAttachment godoc
// @Summary      Delete an attachment
// @Description  Deletes an attachment record and its file from disk.
// @Tags         attachments
// @Accept       json
// @Produce      json
// @Param        id   path  integer  true  "Attachment ID"
// @Success      200  {object}  map[string]bool
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /attachments/{id} [delete]
func (h *AttachmentHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.Error400(w, "invalid attachment id")
		return
	}

	a, err := models.ScanAttachment(
		queryRowContext(r.Context(), h.DB, "SELECT * FROM attachments WHERE id = ?", id))
	if err != nil {
		if err == sql.ErrNoRows {
			utils.Error404(w, "attachment not found")
		} else {
			slog.ErrorContext(r.Context(), "fetch attachment failed", "error", err)
			utils.Error500(w, "database error")
		}
		return
	}

	// ── 先删数据库记录，再删文件 ──────────────────────────────────
	// 即使文件删除失败，数据库记录也已清理，不会产生"幽灵记录"。
	_, err = execContext(r.Context(), h.DB, "DELETE FROM attachments WHERE id = ?", id)
	if err != nil {
		slog.ErrorContext(r.Context(), "delete attachment record failed", "error", err)
		utils.Error500(w, "database error")
		return
	}

	// 文件删除失败不返回错误——文件可能已被手动清理
	if fp, err := utils.EnsureInsideUploadDir(h.UploadDir, a.StoredFilename); err == nil {
		if err := os.Remove(fp); err != nil {
			slog.WarnContext(r.Context(), "failed to remove attachment file",
				"path", fp, "error", err)
		}
	}

	utils.JSONDeleted(w)
}

// detectContentType 根据扩展名推断 Content-Type。
//
// 工程要点：
// 1. 不依赖第三方库（如 mime_guess），减少依赖
// 2. 只用扩展名判断，不读取文件头（读文件头需要多一次 IO）
// 3. 无法识别的扩展名返回 application/octet-stream
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
