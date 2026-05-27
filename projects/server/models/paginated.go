package models

// PaginatedResponse 泛型分页包装 —— 本项目唯一使用泛型的地方
type PaginatedResponse[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}
