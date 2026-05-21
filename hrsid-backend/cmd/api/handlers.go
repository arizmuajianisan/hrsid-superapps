package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/arizmuajianisan/hrsid-backend/internal/auth"
	"github.com/arizmuajianisan/hrsid-backend/internal/data"
)

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 1. Validasi User berdasarkan Email atau NIK
	user, err := app.models.Users.GetByIdentifier(input.Identifier)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. Cek Password menggunakan golang.org/x/crypto/bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 3. Generate JWT Access Token (Umur Pendek: 15 Menit)
	accessToken, err := auth.GenerateAccessToken(
		user.ID,
		user.NIK,
		user.Role,
		user.DepartmentID,
		app.config.jwt.secret,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. Generate Opaque Refresh Token (Umur Panjang: 30 Hari)
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Hitung Hash dari Refresh Token untuk disimpan di DB
	hash := sha256.Sum256([]byte(refreshToken))
	refreshTokenHash := hash[:]

	// 6. Catat Sesi Baru ke Tabel `sessions`
	session := &data.Session{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		IPAddress:        r.RemoteAddr, // Nanti bisa dipoles untuk menangani Cloudflare IP
		UserAgent:        r.UserAgent(),
		Expiry:           time.Now().Add(30 * 24 * time.Hour), // 30 Hari
	}

	err = app.models.Sessions.Insert(session)
	if err != nil {
		app.logger.Println("Session insert error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 7. Set Refresh Token ke HttpOnly Cookie
	cookie := &http.Cookie{
		Name:     "hrs_session", // Nama cookie tetap sama agar frontend tidak bingung
		Value:    refreshToken,  // Nilainya sekarang adalah string random panjang (bukan ID user lagi)
		Path:     "/",
		Expires:  session.Expiry,
		HttpOnly: true,
		Secure:   false, // Set ke true jika sudah menggunakan HTTPS di produksi
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	// 8. Kirim Access Token ke Frontend via JSON Response
	responseData := map[string]interface{}{
		"access_token": accessToken,
		"user": map[string]interface{}{
			"full_name": user.FullName,
			"role":      user.Role,
		},
	}

	app.writeJSON(w, http.StatusOK, responseData, nil)
}

func (app *application) refreshHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil cookie hrs_session dari request
	cookie, err := r.Cookie("hrs_session")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, "Unauthorized: No refresh token found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 2. Validasi token ke database tabel sessions (mencocokkan hash)
	// Sesuaikan huruf kecil/besar app.models.sessions dengan struct Anda
	session, err := app.models.Sessions.GetByRefreshToken(cookie.Value)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			http.Error(w, "Unauthorized: Invalid or expired session", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 3. Ambil data user pemilik session tersebut untuk mengisi data di JWT baru
	user, err := app.models.Users.GetByID(session.UserID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			http.Error(w, "Unauthorized: User not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. Generate JWT Access Token baru (Umur Pendek: 15 Menit)
	accessToken, err := auth.GenerateAccessToken(
		user.ID,
		user.NIK,
		user.Role,
		user.DepartmentID,
		app.config.jwt.secret,
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Kirim Access Token baru ke Frontend
	responseData := map[string]interface{}{
		"access_token": accessToken,
	}

	app.writeJSON(w, http.StatusOK, responseData, nil)
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Struct untuk menerima input JSON
	var input struct {
		NIK          string `json:"nik"`
		Email        string `json:"email"`
		Password     string `json:"password"`
		FullName     string `json:"full_name"`
		DepartmentID int    `json:"department_id"`
	}

	// 2. Baca input
	err := app.readJSON(w, r, &input)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 3. Copy data ke struct User
	user := &data.User{
		NIK:          input.NIK,
		Email:        input.Email,
		FullName:     input.FullName,
		DepartmentID: input.DepartmentID,
		Role:         "user", // Default role
	}

	// 4. Hash password karyawan tersebut
	err = user.SetPassword(input.Password)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 5. Simpan ke database
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			http.Error(w, "Email already exists", http.StatusConflict)
		case errors.Is(err, data.ErrDuplicateNIK):
			http.Error(w, "NIK already exists", http.StatusConflict)
		default:
			app.logger.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// 6. Berikan response sukses
	err = app.writeJSON(w, http.StatusCreated, map[string]interface{}{"user": user}, nil)
	if err != nil {
		app.logger.Println(err)
	}
}

func (app *application) listMyAppsHandler(w http.ResponseWriter, r *http.Request) {
	// Ambil user dari context (berkat middleware authenticate)
	user := app.contextGetUser(r)

	// Ambil aplikasi berdasarkan department_id si user
	apps, err := app.models.Applications.GetByDepartment(user.DepartmentID)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]interface{}{"applications": apps}, nil)
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("hrs_session")
	if err == nil {
		// 1. Ambil token mentah dari cookie
		tokenString := cookie.Value

		// 2. Lakukan hashing dengan cara yang sama persis seperti saat INSERT & REFRESH
		hash := sha256.Sum256([]byte(tokenString))
		tokenHash := hash[:] // Konversi ke slice []byte

		// 3. Eksekusi query UPDATE dengan konversi parameter yang aman
		query := `UPDATE sessions SET is_blocked = true WHERE refresh_token_hash = $1`
		result, err := app.db.ExecContext(r.Context(), query, hash[:])
		if err != nil {
			app.logger.Println("DB Logout Error:", err)
		} else {
			rows, _ := result.RowsAffected()
			app.logger.Printf("Logout: %d session blocked", rows)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Gunakan app.db atau sesuaikan dengan variabel koneksi DB di struct app Anda
		_, err = app.db.ExecContext(ctx, query, tokenHash)
		if err != nil {
			app.logger.Println("Database logout update error:", err)
		}
	}

	// 4. Hapus cookie di browser (Tetap jalankan ini agar frontend bersih)
	newCookie := &http.Cookie{
		Name:     "hrs_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, newCookie)

	app.writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"}, nil)
}

func (app *application) meHandler(w http.ResponseWriter, r *http.Request) {
	// Mengambil data user yang sudah divalidasi oleh middleware dari context
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Karena di middleware kita hanya set 4 field krusial, kita bisa return ini langsung.
	// Jika nanti butuh data lengkap (seperti Email/Tanggal Join), Anda bisa lakukan query DB di sini menggunakan user.ID.
	responseData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":            user.ID,
			"nik":           user.NIK,
			"role":          user.Role,
			"department_id": user.DepartmentID,
		},
	}

	app.writeJSON(w, http.StatusOK, responseData, nil)
}

func (app *application) deactivateUserHandler(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")

	// 1. Translasi PublicID ke InternalID
	internalID, err := app.models.Users.GetInternalIDByPublicID(publicID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. Gunakan internalID untuk operasi database yang cepat
	err = app.models.Users.Deactivate(internalID) // Pastikan method Deactivate menerima int64
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	app.writeJSON(w, http.StatusOK, map[string]string{"message": "User deactivated"}, nil)
}
