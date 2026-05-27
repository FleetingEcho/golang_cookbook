package utils

import "net/http"

// ── 统一错误类型 ──────────────────────────────────────────────────────────
//
// AppError 是应用层错误，包含 HTTP 状态码和用户可读的消息。
// 实现 error 接口，可被标准库的 errors.Is/As 匹配。

type AppError struct {
	Status  int    // HTTP 状态码
	Message string // 用户可见的错误描述
}

func (e *AppError) Error() string { return e.Message }

// ── 错误工厂函数 ──────────────────────────────────────────────────────────

func NewBadRequest(msg string) *AppError {
	return &AppError{Status: http.StatusBadRequest, Message: msg}
}

func NewNotFound(msg string) *AppError {
	return &AppError{Status: http.StatusNotFound, Message: msg}
}

func NewUnauthorized() *AppError {
	return &AppError{Status: http.StatusUnauthorized, Message: "missing or invalid x-api-key"}
}

func NewInternalError(msg string) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Message: msg}
}
