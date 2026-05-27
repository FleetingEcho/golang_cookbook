package models

import "database/sql"

type Attachment struct {
	ID               int64  `json:"id"`
	IssueID          int64  `json:"issueId"`
	OriginalFilename string `json:"originalFilename"`
	StoredFilename   string `json:"storedFilename"`
	ContentType      string `json:"contentType"`
	SizeBytes        int64  `json:"sizeBytes"`
	CreatedAt        string `json:"createdAt"`
}

func ScanAttachment(s interface{ Scan(dest ...any) error }) (Attachment, error) {
	var a Attachment
	err := s.Scan(&a.ID, &a.IssueID, &a.OriginalFilename,
		&a.StoredFilename, &a.ContentType, &a.SizeBytes, &a.CreatedAt)
	return a, err
}

func ScanAttachments(rows *sql.Rows) ([]Attachment, error) {
	items := make([]Attachment, 0)
	for rows.Next() {
		item, err := ScanAttachment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
