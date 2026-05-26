package data

import (
	"context"
	"database/sql"
	"time"
)

type Department struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type DepartmentModel struct {
	DB *sql.DB
}

func (m DepartmentModel) GetAll() ([]*Department, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, `SELECT id, name FROM departments ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var depts []*Department
	for rows.Next() {
		var d Department
		if err := rows.Scan(&d.ID, &d.Name); err != nil {
			return nil, err
		}
		depts = append(depts, &d)
	}
	return depts, rows.Err()
}
