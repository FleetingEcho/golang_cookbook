package models

import "database/sql"

type Label struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func ScanLabel(s interface{ Scan(dest ...any) error }) (Label, error) {
	var l Label
	err := s.Scan(&l.ID, &l.Name, &l.Color)
	return l, err
}

func ScanLabels(rows *sql.Rows) ([]Label, error) {
	items := make([]Label, 0)
	for rows.Next() {
		item, err := ScanLabel(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
