package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/arizmuajianisan/hrsid-backend/internal/data"
)

// Kita buat tipe data khusus untuk context key agar tidak bentrok
type contextKey string

const userContextKey = contextKey("user")

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Ambil cookie dari request
		cookie, err := r.Cookie("hrs_session")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(w, "Unauthorized: No session found", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// 2. Cari user berdasarkan ID yang ada di cookie
		user, err := app.models.Users.GetByID(cookie.Value)
		if err != nil {
			if errors.Is(err, data.ErrRecordNotFound) {
				http.Error(w, "Unauthorized: Invalid session", http.StatusUnauthorized)
				return
			}
			app.logger.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 3. Masukkan data user ke dalam Context
		ctx := context.WithValue(r.Context(), userContextKey, user)

		// 4. Lanjutkan ke handler berikutnya dengan context baru
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Kita bungkus ResponseWriter agar bisa menangkap status code
		// Karena standard http.ResponseWriter tidak punya cara ambil status code setelah ditulis
		app.logger.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())

		next.ServeHTTP(w, r)

		app.logger.Printf("completed in %s", time.Since(start))
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.logger.Printf("PANIC: %s", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Izinkan origin tertentu (nanti ganti dengan domain frontend Anda)
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // WAJIB untuk Cookies

		// Jika request adalah OPTIONS (pre-flight), langsung balas 200
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
