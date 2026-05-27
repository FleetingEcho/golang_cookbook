package utils

import (
	"encoding/json"
	"net/http"
)

// ── 统一 JSON 响应 ──────────────────────────────────────────────────────────
//
// 工程要点：
// 1. 所有响应统一走这些函数，确保 JSON 格式一致
// 2. 错误响应永远返回 {"error": "..."} 结构
// 3. 成功响应直接序列化数据对象
// 4. Content-Type 在中心设置，不会遗漏

// JSON 写成功响应（HTTP 200）
func JSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

// JSONStatus 写指定状态码的成功响应
func JSONStatus(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// JSONDeleted 写删除成功的标准响应 {"deleted": true}
//
// 这是 REST API 的常见约定：DELETE 操作返回 200 + 确认体，
// 而不是 204 No Content，因为前端需要读到响应体来更新状态。
func JSONDeleted(w http.ResponseWriter) {
	JSON(w, map[string]bool{"deleted": true})
}

// JSONLinked 写关联操作成功的标准响应 {"linked": true}
func JSONLinked(w http.ResponseWriter) {
	JSON(w, map[string]bool{"linked": true})
}

// WriteError 将 AppError 写为 JSON 错误响应
func WriteError(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Message,
	})
}

// ── HTTP 错误速查 ───────────────────────────────────────────────────────────

// Error400 快速返回 400 Bad Request
func Error400(w http.ResponseWriter, msg string) {
	WriteError(w, NewBadRequest(msg))
}

// Error404 快速返回 404 Not Found
func Error404(w http.ResponseWriter, msg string) {
	WriteError(w, NewNotFound(msg))
}

// Error500 快速返回 500 Internal Server Error
func Error500(w http.ResponseWriter, msg string) {
	WriteError(w, NewInternalError(msg))
}
