package models

type CreateIssueRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	IssueType   string  `json:"issueType"`
	Assignee    *string `json:"assignee"`
	CreatedBy   string  `json:"createdBy"`
	LabelIDs    []int64 `json:"labelIds"`
}

type UpdateIssueRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Status      *string  `json:"status"`
	Priority    *string  `json:"priority"`
	IssueType   *string  `json:"issueType"`
	Assignee    **string `json:"assignee"`
	LabelIDs    []int64  `json:"labelIds"`
}

type IssueQuery struct {
	Status    *string
	Priority  *string
	IssueType *string
	LabelID   *int64
	Search    *string
	Limit     int64
	Offset    int64
}

type CreateCommentRequest struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

type CreateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
