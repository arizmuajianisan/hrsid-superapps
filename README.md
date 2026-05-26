# Hirose Indonesia Superapps

Centralized SSO/IAM platform for internal applications at Hirose Electric Indonesia.

## Structure

```
hrsid-superapps/
├── hrsid-backend/   # Go REST API (port 4000)
└── hrsid-frontend/  # Nuxt 4 SPA (port 5173)
```

## Features

- SSO login with JWT + refresh token rotation
- Role-based access control (user / admin)
- App portal — shows apps available per department
- Session management — view and revoke active devices
- Admin panel:
  - User management (create, activate/deactivate)
  - Application management (CRUD + department access)
  - Audit log viewer

## Quick Start

**Backend:**
```bash
cd hrsid-backend
go run ./cmd/api
```

**Frontend:**
```bash
cd hrsid-frontend
npm run dev
```
