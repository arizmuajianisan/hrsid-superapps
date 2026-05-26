package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrDuplicateSlug = errors.New("duplicate slug")

type Application struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	BaseURL       string  `json:"base_url"`
	IconURL       *string `json:"icon_url"`
	Description   string  `json:"description"`
	DepartmentIDs []int   `json:"department_ids"`
}

type ApplicationModel struct {
	DB *sql.DB
}

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
		err := rows.Scan(&app.ID, &app.Name, &app.Slug, &app.BaseURL, &app.IconURL, &app.Description)
		if err != nil {
			return nil, err
		}
		apps = append(apps, &app)
	}
	return apps, rows.Err()
}

func (m ApplicationModel) GetAll() ([]*Application, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, `SELECT id, name, slug, base_url, icon_url, description FROM applications ORDER BY name`)
	if err != nil {
		return nil, err
	}

	apps := []*Application{}
	appMap := map[int]*Application{}
	for rows.Next() {
		var a Application
		a.DepartmentIDs = []int{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Slug, &a.BaseURL, &a.IconURL, &a.Description); err != nil {
			rows.Close()
			return nil, err
		}
		apps = append(apps, &a)
		appMap[a.ID] = &a
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	deptRows, err := m.DB.QueryContext(ctx, `SELECT application_id, department_id FROM department_applications`)
	if err != nil {
		return nil, err
	}
	defer deptRows.Close()

	for deptRows.Next() {
		var appID, deptID int
		if err := deptRows.Scan(&appID, &deptID); err != nil {
			return nil, err
		}
		if a, ok := appMap[appID]; ok {
			a.DepartmentIDs = append(a.DepartmentIDs, deptID)
		}
	}
	return apps, deptRows.Err()
}

func (m ApplicationModel) Create(app *Application) error {
	query := `INSERT INTO applications (name, slug, base_url, icon_url, description) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, app.Name, app.Slug, app.BaseURL, app.IconURL, app.Description).Scan(&app.ID)
	if err != nil {
		if strings.Contains(err.Error(), "slug") {
			return ErrDuplicateSlug
		}
		return err
	}
	return nil
}

func (m ApplicationModel) Update(app *Application) error {
	query := `UPDATE applications SET name=$1, slug=$2, base_url=$3, icon_url=$4, description=$5 WHERE id=$6`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := m.DB.ExecContext(ctx, query, app.Name, app.Slug, app.BaseURL, app.IconURL, app.Description, app.ID)
	if err != nil {
		if strings.Contains(err.Error(), "slug") {
			return ErrDuplicateSlug
		}
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m ApplicationModel) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := m.DB.ExecContext(ctx, `DELETE FROM applications WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m ApplicationModel) SetDepartments(appID int, deptIDs []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `DELETE FROM department_applications WHERE application_id=$1`, appID); err != nil {
		return err
	}
	for _, deptID := range deptIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO department_applications (department_id, application_id) VALUES ($1, $2)`, deptID, appID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
