package models

import (
	"database/sql"
	"strings"
)

type Issue struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	IssueType   string  `json:"issueType"`
	Assignee    *string `json:"assignee"`
	CreatedBy   string  `json:"createdBy"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

func ScanIssue(s interface{ Scan(dest ...any) error }) (Issue, error) {
	var i Issue
	err := s.Scan(&i.ID, &i.Title, &i.Description, &i.Status,
		&i.Priority, &i.IssueType, &i.Assignee, &i.CreatedBy, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func ScanIssues(rows *sql.Rows) ([]Issue, error) {
	items := make([]Issue, 0)
	for rows.Next() {
		item, err := ScanIssue(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ValidateIssueFields(title, description, priority, issueType string) string {
	if strings.TrimSpace(title) == "" {
		return "title is required"
	}
	if strings.TrimSpace(description) == "" {
		return "description is required"
	}
	switch priority {
	case "low", "medium", "high":
	default:
		return "priority must be low, medium, or high"
	}
	switch issueType {
	case "bug", "feature", "task", "question":
	default:
		return "issue_type must be bug, feature, task, or question"
	}
	return ""
}

func ValidateStatus(status string) string {
	switch status {
	case "open", "in_progress", "closed":
		return ""
	default:
		return "status must be open, in_progress, or closed"
	}
}

type IssueDetail struct {
	Issue
	Labels      []Label      `json:"labels"`
	Comments    []Comment    `json:"comments"`
	Attachments []Attachment `json:"attachments"`
}
