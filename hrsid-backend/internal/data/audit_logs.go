package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID              string          `json:"id"`
	PublicID        string          `json:"public_id"`
	Action          string          `json:"action"`
	ActorUserID     *string         `json:"actor_user_id,omitempty"`
	ActorIdentifier *string         `json:"actor_identifier,omitempty"`
	ActorFullName   *string         `json:"actor_full_name,omitempty"`
	ActorNIK        *string         `json:"actor_nik,omitempty"`
	TargetUserID    *string         `json:"target_user_id,omitempty"`
	TargetFullName  *string         `json:"target_full_name,omitempty"`
	TargetNIK       *string         `json:"target_nik,omitempty"`
	IPAddress       string          `json:"ip_address"`
	UserAgent       string          `json:"user_agent"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

type AuditLogModel struct {
	DB *sql.DB
}

// Insert records a new audit log entry. Best-effort: callers should typically
// log errors but not fail the parent request when this fails.
func (m AuditLogModel) Insert(action string, actorUserID, actorIdentifier *string, targetUserID *string, ip, userAgent string, metadata map[string]any) error {
	var metaBytes []byte
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metaBytes = b
	}

	query := `
		INSERT INTO audit_logs (action, actor_user_id, actor_identifier, target_user_id, ip_address, user_agent, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query,
		action,
		actorUserID,
		actorIdentifier,
		targetUserID,
		ip,
		userAgent,
		metaBytes,
	)
	return err
}

func (m AuditLogModel) GetAll(limit int) ([]*AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	query := `
		SELECT
			a.public_id, a.action,
			a.actor_user_id, a.actor_identifier,
			actor.full_name, actor.nik,
			a.target_user_id,
			target.full_name, target.nik,
			COALESCE(a.ip_address, ''), COALESCE(a.user_agent, ''),
			a.metadata, a.created_at
		FROM audit_logs a
		LEFT JOIN users actor  ON actor.id  = a.actor_user_id
		LEFT JOIN users target ON target.id = a.target_user_id
		ORDER BY a.created_at DESC
		LIMIT $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []*AuditLog{}
	for rows.Next() {
		var a AuditLog
		var meta sql.NullString
		err := rows.Scan(
			&a.PublicID, &a.Action,
			&a.ActorUserID, &a.ActorIdentifier,
			&a.ActorFullName, &a.ActorNIK,
			&a.TargetUserID,
			&a.TargetFullName, &a.TargetNIK,
			&a.IPAddress, &a.UserAgent,
			&meta, &a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if meta.Valid {
			a.Metadata = json.RawMessage(meta.String)
		}
		logs = append(logs, &a)
	}
	return logs, rows.Err()
}
