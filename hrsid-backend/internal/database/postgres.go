package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/arizmuajianisan/hrsid-backend/migrations"
	_ "github.com/jackc/pgx/v5/stdlib" // Standard library driver untuk pgx
	"github.com/pressly/goose/v3"
)

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Set durasi maksimal koneksi idle
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(15 * time.Minute)

	// Cek koneksi dengan context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	// Beritahu goose bahwa kita menggunakan embed FS
	goose.SetBaseFS(migrations.FS)

	// Set dialect ke postgres
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// Jalankan migrasi 'Up'
	// "." berarti mencari file sql di root folder embed FS tadi
	if err := goose.Up(db, "."); err != nil {
		return err
	}

	return nil
}
