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

	// Manajemen sesi milik user yang sedang login
	listMySessions := http.HandlerFunc(app.listMySessionsHandler)
	mux.Handle("GET /api/v1/sessions", app.authenticate(listMySessions))

	revokeMySession := http.HandlerFunc(app.revokeMySessionHandler)
	mux.Handle("POST /api/v1/sessions/{id}/revoke", app.authenticate(revokeMySession))

	// --- ADMIN ONLY ROUTES ---
	listUsersHandler := http.HandlerFunc(app.listUsersHandler)
	mux.Handle("GET /api/v1/admin/users", app.authenticate(app.requireAdmin(listUsersHandler)))

	// Perhatikan penggunaan penulisan parameter {id} sesuai standar bawaan Go modern
	deactivateHandler := http.HandlerFunc(app.deactivateUserHandler)
	mux.Handle("POST /api/v1/admin/users/{id}/deactivate", app.authenticate(app.requireAdmin(deactivateHandler)))

	reactivateHandler := http.HandlerFunc(app.reactivateUserHandler)
	mux.Handle("POST /api/v1/admin/users/{id}/reactivate", app.authenticate(app.requireAdmin(reactivateHandler)))

	listAuditLogs := http.HandlerFunc(app.listAuditLogsHandler)
	mux.Handle("GET /api/v1/admin/audit-logs", app.authenticate(app.requireAdmin(listAuditLogs)))

	listAppsAdmin := http.HandlerFunc(app.listApplicationsAdminHandler)
	mux.Handle("GET /api/v1/admin/applications", app.authenticate(app.requireAdmin(listAppsAdmin)))

	createApp := http.HandlerFunc(app.createApplicationHandler)
	mux.Handle("POST /api/v1/admin/applications", app.authenticate(app.requireAdmin(createApp)))

	updateApp := http.HandlerFunc(app.updateApplicationHandler)
	mux.Handle("PUT /api/v1/admin/applications/{id}", app.authenticate(app.requireAdmin(updateApp)))

	deleteApp := http.HandlerFunc(app.deleteApplicationHandler)
	mux.Handle("DELETE /api/v1/admin/applications/{id}", app.authenticate(app.requireAdmin(deleteApp)))

	listDepts := http.HandlerFunc(app.listDepartmentsHandler)
	mux.Handle("GET /api/v1/admin/departments", app.authenticate(app.requireAdmin(listDepts)))

	return app.recoverPanic(app.enableCORS(app.logRequest(mux)))
}
