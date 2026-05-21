package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/arizmuajianisan/hrsid-backend/internal/database"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type config struct {
	dsn string
}

type userSeed struct {
	nik      string
	email    string
	fullName string
	role     string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var cfg config
	flag.StringVar(&cfg.dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable not set.")
	}

	db, err := database.OpenDB(cfg.dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Printf("database connection pool established")

	password := "rahasia-banget"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	users := []userSeed{
		{
			nik:      "99999999",
			email:    "admin@hrs-id.com",
			fullName: "Super Admin",
			role:     "admin",
		},
		{
			nik:      "00000001",
			email:    "user-testing-1@hrs-id.com",
			fullName: "UserTesting1",
			role:     "user",
		},
		{
			nik:      "00000002",
			email:    "user-testing-2@hrs-id.com",
			fullName: "UserTesting2",
			role:     "user",
		},
	}

	query := `INSERT INTO users (nik, email, password_hash, full_name, role, department_id)
		VALUES ($1, $2, $3, $4, $5, (SELECT id FROM departments LIMIT 1))`

	for _, u := range users {
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", u.email).Scan(&exists)
		if err != nil {
			log.Fatal(err)
		}

		if exists {
			fmt.Printf("User %s NIK: %s sudah ada, skip.\n", u.email, u.nik)
			continue
		}

		_, err = db.Exec(query, u.nik, u.email, hash, u.fullName, u.role)
		if err != nil {
			log.Fatalf("Gagal insert user %s: %v", u.email, err)
		}
		fmt.Printf("User %s (%s) NIK: %s berhasil di-seed!\n", u.fullName, u.email, u.nik)
	}

	fmt.Printf("All users use password: %s", password)
}
