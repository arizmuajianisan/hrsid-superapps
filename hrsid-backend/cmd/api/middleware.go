package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/arizmuajianisan/hrsid-backend/internal/auth" // Sesuaikan dengan module name Anda
	"github.com/arizmuajianisan/hrsid-backend/internal/data"
	"github.com/golang-jwt/jwt/v5"
)

// Kita buat tipe data khusus untuk context key agar tidak bentrok
type contextKey string

const userContextKey = contextKey("user")
const sessionIDContextKey = contextKey("session_id")

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Ambil header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		// 2. Format header harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Unauthorized: Invalid Authorization Format", http.StatusUnauthorized)
			return
		}

		accessToken := parts[1]
		claims := &auth.UserClaims{}

		// 3. Parsing dan validasi JWT token
		token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
			// Pastikan metode signing-nya HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(app.config.jwt.secret), nil
		})

		// 4. Jika token expired atau rusak, langsung block di sini
		if err != nil || !token.Valid {
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Unauthorized: Token Expired", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Unauthorized: Invalid Token", http.StatusUnauthorized)
			return
		}

		// 5. Verifikasi session ke DB — menangkap logout & deactivate secara instan
		session, err := app.models.Sessions.GetByID(claims.SessionID)
		if err != nil || session.IsBlocked || session.Expiry.Before(time.Now()) {
			http.Error(w, "Unauthorized: Session is invalid", http.StatusUnauthorized)
			return
		}

		// 6. Ubah claims menjadi struct data.User
		user := &data.User{
			ID:           claims.UserID,
			NIK:          claims.NIK,
			Role:         claims.Role,
			DepartmentID: claims.DepartmentID,
		}

		// 6. Masukkan objek user dan session ID ke Context, lalu teruskan request
		ctx := context.WithValue(r.Context(), userContextKey, user)
		ctx = context.WithValue(ctx, sessionIDContextKey, claims.SessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// log permulaan request
		app.logger.Printf(
			"%s %s %s | %s",
			r.RemoteAddr,
			strings.ToUpper(r.Method),
			r.URL.RequestURI(),
			r.Proto,
		)

		next.ServeHTTP(lrw, r)

		// log selesai dengan status & latency
		duration := time.Since(start)
		app.logger.Printf(
			"%s %s %s | %d | %s",
			r.RemoteAddr,
			strings.ToUpper(r.Method),
			r.URL.RequestURI(),
			lrw.statusCode,
			duration,
		)
	})
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.statusCode == 0 {
		// kalo belum ditulis header, asumsikan 200
		lrw.statusCode = http.StatusOK
	}
	return lrw.ResponseWriter.Write(b)
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

func (app *application) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Ambil user dari context (yang sebelumnya sudah diset oleh middleware authenticate)
		user, ok := r.Context().Value(userContextKey).(*data.User)

		// 2. Jika user tidak ditemukan atau role-nya bukan admin, tolak akses!
		if !ok || user.Role != "admin" {
			http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		// 3. Jika lolos, teruskan request ke handler tujuan
		next.ServeHTTP(w, r)
	})
}
