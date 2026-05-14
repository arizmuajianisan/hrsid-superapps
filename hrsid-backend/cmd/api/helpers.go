package main

import (
	"encoding/json"
	"net/http"

	"github.com/arizmuajianisan/hrsid-backend/internal/data"
)

// writeJSON adalah helper untuk mengirim response JSON
func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// readJSON adalah helper untuk membaca body request JSON ke sebuah struct
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576 // Limit 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // Error jika user kirim field yang tidak ada di struct

	err := dec.Decode(dst)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		return nil
	}
	return user
}
