package data

import (
	"context"
	"database/sql"
	"time"
)

type Application struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	BaseURL     string  `json:"base_url"`
	IconURL     *string `json:"icon_url"`
	Description string  `json:"description"`
}

type ApplicationModel struct {
	DB *sql.DB
}

// GetByDepartment mengambil daftar aplikasi yang diizinkan untuk departemen tertentu
func (m ApplicationModel) GetByDepartment(deptID int) ([]*Application, error) {
	query := `
		SELECT a.id, a.name, a.slug, a.base_url, a.icon_url, a.description
		FROM applications a
		INNER JOIN department_applications da ON a.id = da.application_id
		WHERE da.department_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, deptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*Application

	for rows.Next() {
		var app Application
		err := rows.Scan(
			&app.ID,
			&app.Name,
			&app.Slug,
			&app.BaseURL,
			&app.IconURL,
			&app.Description,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &app)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}
