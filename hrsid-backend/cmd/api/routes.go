package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/register", app.registerUserHandler)
	mux.HandleFunc("POST /api/v1/login", app.loginHandler)
	mux.HandleFunc("POST /api/v1/logout", app.logoutHandler)
	mux.HandleFunc("GET /api/v1/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	myAppsHandler := http.HandlerFunc(app.listMyAppsHandler)
	mux.Handle("GET /api/v1/my-apps", app.authenticate(myAppsHandler))

	return app.recoverPanic(app.enableCORS(app.logRequest(mux)))
}
