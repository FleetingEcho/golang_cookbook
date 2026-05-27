package utils

import (
	"encoding/json"
	"net/http"
)

type AppError struct {
	Status  int
	Message string
}

func (e *AppError) Error() string { return e.Message }

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

func WriteError(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Message})
}
