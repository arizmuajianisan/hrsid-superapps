package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrDuplicateNIK   = errors.New("duplicate NIK")
	ErrRecordNotFound = errors.New("record not found")
)

type User struct {
	ID           string    `json:"id"`
	NIK          string    `json:"nik"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	PasswordHash []byte    `json:"-"` // "-" agar tidak muncul di JSON
	Role         string    `json:"role"`
	DepartmentID int       `json:"department_id"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserModel struct {
	DB *sql.DB
}

// Password management
func (u *User) SetPassword(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return nil
}

func (u *User) PasswordMatches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func (m UserModel) GetByEmailOrNIK(identifier string) (*User, error) {
	query := `
		SELECT id, nik, email, password_hash, full_name, role, department_id, is_active, created_at
		FROM users
		WHERE (email = $1 OR nik = $1) AND is_active = true`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, identifier).Scan(
		&user.ID,
		&user.NIK,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.DepartmentID,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (nik, email, password_hash, full_name, role, department_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	args := []interface{}{
		user.NIK,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.DepartmentID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Eksekusi query dan ambil ID serta CreatedAt yang dibuat oleh DB
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		switch {
		// Cek jika ada pelanggaran constraint UNIQUE (NIK atau Email sudah ada)
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_nik_key"`:
			return ErrDuplicateNIK
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetByID(id string) (*User, error) {
	query := `
		SELECT id, nik, email, full_name, role, department_id, is_active, created_at
		FROM users
		WHERE id = $1 AND is_active = true`

	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.NIK, &user.Email, &user.FullName, &user.Role, &user.DepartmentID, &user.IsActive, &user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}
