# Backend Service

Just run using:

```bash
go run ./cmd/api
```

Or,

```bash
make run-dev
```

It will also automatically run the migrations at start.

## Routes

After run, the apps will serve at port 4000.

```
2026/05/14 07:31:46 database connection pool established
2026/05/14 07:31:46 database migrations applied successfully
2026/05/14 07:31:46 starting server on :4000
```

### Public Routes

---

#### Healthcheck

```
GET /api/v1/healthcheck
```

Response: `200 OK`

---

#### Register

```
POST /api/v1/register
```

Body:
```json
{
    "nik": "2024001",
    "email": "developer@hrs-id.com",
    "password": "rahasia-banget",
    "full_name": "Arz Lab Developer",
    "department_id": 1
}
```

Response `201 Created`:
```json
{
    "user": {
        "id": "1",
        "nik": "2024001",
        "email": "developer@hrs-id.com",
        "full_name": "Arz Lab Developer",
        "role": "user",
        "department_id": 1,
        "is_active": true,
        "created_at": "2026-05-14T02:00:56.845503+07:00"
    }
}
```

---

#### Login

```
POST /api/v1/login
```

Body:
```json
{
    "identifier": "developer@hrs-id.com",
    "password": "rahasia-banget"
}
```

`identifier` accepts either email or NIK.

Response `200 OK` — sets `hrs_session` HttpOnly cookie (30 days):
```json
{
    "access_token": "<jwt>",
    "user": {
        "full_name": "Arz Lab Developer",
        "role": "user"
    }
}
```

---

#### Refresh

```
POST /api/v1/refresh
```

Requires `hrs_session` cookie. Issues a new access token **and rotates the refresh token** — the old cookie is invalidated and a new one is set.

Response `200 OK`:
```json
{
    "access_token": "<new_jwt>"
}
```

---

#### Logout

```
POST /api/v1/logout
```

Blocks the current session in the database and clears the `hrs_session` cookie. Works even if the access token has already expired.

Response `200 OK`:
```json
{
    "message": "logged out successfully"
}
```

---

### Protected Routes

All protected routes require `Authorization: Bearer <access_token>` header.

---

#### Me

```
GET /api/v1/me
```

Response `200 OK`:
```json
{
    "user": {
        "id": "1",
        "nik": "2024001",
        "role": "user",
        "department_id": 1
    }
}
```

---

#### My Apps

Returns apps accessible to the user's department.

```
GET /api/v1/my-apps
```

Response `200 OK`:
```json
{
    "applications": [
        {
            "id": 1,
            "name": "Hi-DSign",
            "slug": "hi-dsign",
            "base_url": "https://hi-dsign.hrs-id.com",
            "icon_url": null,
            "description": "Electronic Drawing Approval System"
        }
    ]
}
```

---

### Admin Routes

Requires a valid access token with `role: admin`.

---

#### Deactivate User

Deactivates a user and immediately blocks all their active sessions.

```
POST /api/v1/admin/users/{public_id}/deactivate
```

Response `200 OK`:
```json
{
    "message": "User deactivated"
}
```
