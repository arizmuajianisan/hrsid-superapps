package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

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

// clientIP mengekstrak IP asli klien. Prioritas: X-Forwarded-For pertama,
// lalu X-Real-IP, lalu r.RemoteAddr (di-strip portnya).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Format: "client, proxy1, proxy2" — ambil yang paling kiri
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// audit menulis entri audit log secara best-effort: kegagalan hanya di-log,
// tidak menggagalkan request induk.
func (app *application) audit(action string, actorUserID, actorIdentifier, targetUserID *string, r *http.Request, metadata map[string]any) {
	err := app.models.AuditLogs.Insert(
		action,
		actorUserID,
		actorIdentifier,
		targetUserID,
		clientIP(r),
		r.UserAgent(),
		metadata,
	)
	if err != nil {
		app.logger.Printf("audit insert failed (action=%s): %v", action, err)
	}
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		return nil
	}
	return user
}
