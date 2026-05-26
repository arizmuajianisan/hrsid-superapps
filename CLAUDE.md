# HRSID (Hirose Identity & Access Management)

Project ini adalah sistem otentikasi dan otorisasi terpusat (SSO/IAM) yang dirancang untuk kebutuhan internal Hirose Electric Indonesia.

Menggunakan monorepo, `hrsid-backend` dan `hrsid-frontend`.

## 1. Tech Stack
*   **Language:** Go (Golang) 1.22+
*   **Database:** PostgreSQL
*   **Driver:** `github.com/jackc/pgx/v5` (via `database/sql` standard interface)
*   **Auth:** JWT (Stateless) + Refresh Token (Database-backed for security)
*   **Migration:** `goose`
*   **Encryption:** `bcrypt`

## 2. Arsitektur Database
*   **Internal PK:** `BIGSERIAL` (untuk efisiensi JOIN dan index).
*   **Public ID:** `UUID` (untuk eksposur API/URL guna mencegah ID Enumeration Attack).
*   **Relasi:** 
    *   `users` (id: bigint, public_id: uuid)
    *   `sessions` (user_id merujuk ke users.id)
    *   `departments` (relasi ke users)

## 3. Struktur Routing
Semua endpoint berada di bawah prefix `/api/v1`.
*   **Public:** `POST /register`, `POST /login`, `POST /refresh`, `POST /logout`, `GET /healthcheck`
*   **Protected (User):** `GET /me`, `GET /my-apps`, `GET /sessions`, `POST /sessions/{id}/revoke`
*   **Protected (Admin):**
    *   Users: `GET /admin/users`, `POST /admin/users/{id}/deactivate`, `POST /admin/users/{id}/reactivate`
    *   Applications: `GET /admin/applications`, `POST /admin/applications`, `PUT /admin/applications/{id}`, `DELETE /admin/applications/{id}`
    *   Departments: `GET /admin/departments`
    *   Audit: `GET /admin/audit-logs`

## 4. Keamanan & Middleware
*   **CORS:** Enabled (hardcoded `localhost:5173` untuk dev — perlu env var sebelum production).
*   **Auth:** Menggunakan `authenticate` middleware (JWT validation).
*   **RBAC:** Middleware `requireAdmin` untuk endpoint administratif.
*   **Logout:** Menggunakan *Soft Block* (`is_blocked = true`) pada `sessions` tabel. Tidak memerlukan JWT yang valid — hanya butuh cookie `hrs_session`.
*   **Refresh Token Rotation:** Setiap call ke `/refresh` akan memblokir session lama dan membuat session baru. Token lama langsung tidak berlaku setelah dipakai sekali.
*   **Cookie:** `Secure: false` saat development HTTP. Set ke `true` sebelum deploy ke production HTTPS.

## 5. Panduan Pengembangan
*   **Menjalankan Migrasi:** `make migrate-up`
*   **Seeding Admin:** `DATABASE_URL=... make seed`
*   **Menjalankan Server:** `go run ./cmd/api/`
*   **Mapping ID:** Gunakan `GetInternalIDByPublicID` di `UserModel` untuk translasi `UUID` (input) ke `BIGINT` (internal/db).

## 6. Current Roadmap
- [x] Migrasi ke `BIGSERIAL` + `UUID` (Public ID)
- [x] Implementasi Middleware RBAC
- [x] Implementasi Seeder Admin menggunakan `pgx/stdlib`
- [x] Security review & bug fixes backend (deactivated user login, double query logout, refresh token rotation)
- [x] Buat frontend menggunakan Nuxt.js (Nuxt 4 + Nuxt UI v4 + Pinia, SPA mode)
- [x] Implementasi Frontend (silent refresh via `useApi` composable)
- [x] Implementasi integrasi aplikasi ke dashboard `my-apps`
- [x] Admin: Manajemen User (create, activate/deactivate)
- [x] Admin: Manajemen Aplikasi CRUD (create, edit, delete, assign department access)
- [x] Admin: Audit Log viewer
- [x] User: Manajemen Sesi (lihat & revoke sesi aktif)
- [ ] Hardening production: `Secure` cookie, CORS env var, input validation on `/register`
