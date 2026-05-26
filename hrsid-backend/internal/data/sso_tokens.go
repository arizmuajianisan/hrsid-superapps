package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SSOLaunchToken struct {
	ID            int64
	Token         string
	UserID        string
	ApplicationID int
	ExpiresAt     time.Time
	UsedAt        *time.Time

	// Populated via JOIN on ValidateAndConsume
	AppSSOSecret string
	UserNIK      string
	UserEmail    string
	UserFullName string
	UserRole     string
	UserDept     string
	UserDeptID   int
}

type SSOTokenModel struct {
	DB *sql.DB
}

func (m SSOTokenModel) Insert(userID string, appID int, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO sso_launch_tokens (token, user_id, application_id, expires_at)
		VALUES ($1, $2, $3, $4)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, token, userID, appID, expiresAt)
	return err
}

// ValidateAndConsume looks up the token, marks it as used, and returns user + app info.
// Returns ErrRecordNotFound if not found, expired, or already used.
func (m SSOTokenModel) ValidateAndConsume(token string) (*SSOLaunchToken, error) {
	query := `
		SELECT
			t.id, t.user_id, t.application_id, t.expires_at, t.used_at,
			a.sso_secret,
			u.nik, u.email, u.full_name, u.role,
			d.name, u.department_id
		FROM sso_launch_tokens t
		INNER JOIN applications a ON t.application_id = a.id
		INNER JOIN users u ON t.user_id = u.id
		INNER JOIN departments d ON u.department_id = d.id
		WHERE t.token = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var ott SSOLaunchToken
	err := m.DB.QueryRowContext(ctx, query, token).Scan(
		&ott.ID,
		&ott.UserID,
		&ott.ApplicationID,
		&ott.ExpiresAt,
		&ott.UsedAt,
		&ott.AppSSOSecret,
		&ott.UserNIK,
		&ott.UserEmail,
		&ott.UserFullName,
		&ott.UserRole,
		&ott.UserDept,
		&ott.UserDeptID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	if ott.UsedAt != nil || time.Now().After(ott.ExpiresAt) {
		return nil, ErrRecordNotFound
	}

	// Mark as used immediately (single-use)
	_, err = m.DB.ExecContext(ctx, `UPDATE sso_launch_tokens SET used_at = NOW() WHERE id = $1`, ott.ID)
	if err != nil {
		return nil, err
	}

	return &ott, nil
}
