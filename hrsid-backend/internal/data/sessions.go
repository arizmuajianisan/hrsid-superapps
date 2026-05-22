package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"
)

type Session struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	RefreshTokenHash []byte    `json:"-"` // Sembunyikan dari JSON response
	IPAddress        string    `json:"ip_address"`
	UserAgent        string    `json:"user_agent"`
	IsBlocked        bool      `json:"is_blocked"`
	Expiry           time.Time `json:"expiry"`
	CreatedAt        time.Time `json:"created_at"`
}

type SessionModel struct {
	DB *sql.DB
}

// Insert digunakan saat user sukses login untuk mencatat session baru
func (m SessionModel) Insert(session *Session) error {
	query := `
		INSERT INTO sessions (user_id, refresh_token_hash, ip_address, user_agent, expiry)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, is_blocked`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query,
		session.UserID,
		session.RefreshTokenHash,
		session.IPAddress,
		session.UserAgent,
		session.Expiry,
	).Scan(&session.ID, &session.CreatedAt, &session.IsBlocked)
}

func (m SessionModel) GetByID(id string) (*Session, error) {
	query := `SELECT id, user_id, is_blocked, expiry FROM sessions WHERE id = $1`

	var session Session
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.UserID, &session.IsBlocked, &session.Expiry,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (m SessionModel) BlockByID(id string) error {
	query := `UPDATE sessions SET is_blocked = true WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, id)
	return err
}

// ListByUserID mengambil semua session milik user tertentu (untuk halaman "Sesi Saya")
func (m SessionModel) ListByUserID(userID string) ([]*Session, error) {
	query := `
		SELECT id, user_id, ip_address, user_agent, is_blocked, expiry, created_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 100`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []*Session{}
	for rows.Next() {
		var s Session
		err := rows.Scan(
			&s.ID, &s.UserID, &s.IPAddress, &s.UserAgent,
			&s.IsBlocked, &s.Expiry, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, &s)
	}
	return sessions, rows.Err()
}

// BlockByIDAndUserID memblokir session hanya jika benar-benar milik userID tsb
// (mencegah user lain memblokir session orang lain). Mengembalikan ErrRecordNotFound
// jika tidak ada baris yang cocok.
func (m SessionModel) BlockByIDAndUserID(id, userID string) error {
	query := `UPDATE sessions SET is_blocked = true WHERE id = $1 AND user_id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := m.DB.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// GetByRefreshToken mencari session yang aktif berdasarkan token yang dikirim user
func (m SessionModel) GetByRefreshToken(token string) (*Session, error) {
	// Kita hitung SHA256 hash dari token yang masuk untuk dicocokkan dengan DB
	hash := sha256.Sum256([]byte(token))
	tokenHash := hash[:]

	query := `
		SELECT id, user_id, refresh_token_hash, ip_address, user_agent, is_blocked, expiry, created_at
		FROM sessions
		WHERE refresh_token_hash = $1 AND is_blocked = false AND expiry > NOW()`

	var session Session
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.IPAddress,
		&session.UserAgent,
		&session.IsBlocked,
		&session.Expiry,
		&session.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound // Session tidak valid / expired / di-block
		}
		return nil, err
	}

	return &session, nil
}
