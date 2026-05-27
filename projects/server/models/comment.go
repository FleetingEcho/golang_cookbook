package models

import "database/sql"

type Comment struct {
	ID        int64  `json:"id"`
	IssueID   int64  `json:"issueId"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

func ScanComment(s interface{ Scan(dest ...any) error }) (Comment, error) {
	var c Comment
	err := s.Scan(&c.ID, &c.IssueID, &c.Author, &c.Body, &c.CreatedAt)
	return c, err
}

func ScanComments(rows *sql.Rows) ([]Comment, error) {
	items := make([]Comment, 0)
	for rows.Next() {
		item, err := ScanComment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
