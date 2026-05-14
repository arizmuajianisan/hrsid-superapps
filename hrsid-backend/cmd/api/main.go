package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arizmuajianisan/hrsid-backend/internal/data"
	"github.com/arizmuajianisan/hrsid-backend/internal/database"
	"github.com/joho/godotenv"
)

// Struktur ini menampung semua dependensi aplikasi
type application struct {
	config config
	db     *sql.DB
	logger *log.Logger
	models data.Models // Tambahkan ini
}

type config struct {
	port int
	dsn  string
}

func main() {
	// 1. Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Ambil konfigurasi dari flag atau environment variable
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.dsn, "db-dsn", os.Getenv("DB_DSN"), "PostgreSQL DSN")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// 3. Inisialisasi koneksi database
	db, err := database.OpenDB(cfg.dsn)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Printf("database connection pool established")

	// 4. JALANKAN MIGRASI OTOMATIS
	err = database.Migrate(db)
	if err != nil {
		logger.Fatal("migration failed:", err)
	}
	logger.Printf("database migrations applied successfully")

	// 5. Inisialisasi aplikasi dan jalankan server
	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
		models: data.NewModels(db), // Inisialisasi model
	}

	// 5. Konfigurasi HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(), // Kita akan buat fungsi routes() nanti
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting server on %s", srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}
