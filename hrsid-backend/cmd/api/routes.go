package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// --- PUBLIC ROUTES ---
	mux.HandleFunc("POST /api/v1/register", app.registerUserHandler)
	mux.HandleFunc("POST /api/v1/login", app.loginHandler)
	mux.HandleFunc("POST /api/v1/refresh", app.refreshHandler)
	mux.HandleFunc("GET /api/v1/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// --- PROTECTED ROUTES (User Biasa & Admin) ---
	// Endpoint /me yang baru saja Anda buat
	meHandler := http.HandlerFunc(app.meHandler)
	mux.Handle("GET /api/v1/me", app.authenticate(meHandler))

	myAppsHandler := http.HandlerFunc(app.listMyAppsHandler)
	mux.Handle("GET /api/v1/my-apps", app.authenticate(myAppsHandler))

	mux.HandleFunc("POST /api/v1/logout", app.logoutHandler)

	// --- ADMIN ONLY ROUTES ---
	// Perhatikan penggunaan penulisan parameter {id} sesuai standar bawaan Go modern
	deactivateHandler := http.HandlerFunc(app.deactivateUserHandler)
	mux.Handle("POST /api/v1/admin/users/{id}/deactivate", app.authenticate(app.requireAdmin(deactivateHandler)))

	return app.recoverPanic(app.enableCORS(app.logRequest(mux)))
}
