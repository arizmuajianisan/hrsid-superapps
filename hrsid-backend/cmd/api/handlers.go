package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/arizmuajianisan/hrsid-backend/internal/data"
)

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Definisikan struct untuk input
	var input struct {
		Identifier string `json:"identifier"` // Bisa Email atau NIK
		Password   string `json:"password"`
	}

	// 2. Baca JSON dari request
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 3. Cari user di database menggunakan model yang sudah kita buat
	user, err := app.models.Users.GetByEmailOrNIK(input.Identifier)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		default:
			app.logger.Println(err)
			http.Error(w, "Server Error", http.StatusInternalServerError)
		}
		return
	}

	// 4. Verifikasi password
	match, err := user.PasswordMatches(input.Password)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	if !match {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 5. Jika cocok, buat Cookie (30 Hari sesuai rencana)
	// Untuk sekarang kita simpan ID user, nanti kita tingkatkan dengan Token/JWT
	cookie := &http.Cookie{
		Name:     "hrs_session",
		Value:    user.ID,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set ke true jika sudah menggunakan HTTPS/Cloudflare Tunnel
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)

	// 6. Kirim response sukses
	response := map[string]interface{}{
		"status": "success",
		"user":   user,
	}

	err = app.writeJSON(w, http.StatusOK, response, nil)
	if err != nil {
		app.logger.Println(err)
	}
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
	// Buat cookie dengan nama yang sama tapi masa berlaku sudah lewat
	cookie := &http.Cookie{
		Name:     "hrs_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false, // Set true jika HTTPS
	}

	http.SetCookie(w, cookie)

	app.writeJSON(w, http.StatusOK, map[string]string{"message": "logged out successfully"}, nil)
}
