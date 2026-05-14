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

```bash
🚀 make run-dev
2026/05/14 07:31:46 database connection pool established
2026/05/14 07:31:46 goose: no migrations to run. current version: 20260513190621
2026/05/14 07:31:46 database migrations applied successfully
2026/05/14 07:31:46 starting server on :4000
```

### Healthcheck

```bash
get /api/v1/healthcheck
```

Responses:

```bash
> GET /api/v1/healthcheck HTTP/1.1
> Host: localhost:4000
> Cookie: hrs_session=81cd0f87-69a6-41c1-a93f-44baef34182d
> User-Agent: insomnia/12.5.0
> Accept: */*

* Mark bundle as not supporting multiuse

< HTTP/1.1 200 OK
< Access-Control-Allow-Credentials: true
< Access-Control-Allow-Headers: Content-Type, Authorization
< Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
< Access-Control-Allow-Origin: http://localhost:5173
< Date: Thu, 14 May 2026 00:43:02 GMT
< Content-Length: 2
< Content-Type: text/plain; charset=utf-8
```

### Register

```bash
POST /api/v1/register
Body:
{
    "nik": "2024001",
    "email": "developer@hrs-id.com",
    "password": "rahasia-banget",
    "full_name": "Arz Lab Developer",
    "department_id": 1
}
```

Responses:

```bash
{
	"user": {
		"id": "81cd0f87-69a6-41c1-a93f-44baef34182d",
		"nik": "2024001",
		"email": "developer@hrs-id.com",
		"full_name": "Arz Lab Developer",
		"role": "user",
		"department_id": 1,
		"is_active": false,
		"created_at": "2026-05-14T02:00:56.845503+07:00"
	}
}
```

### Login

```bash
POST /api/v1/login
Body:
{
    "email": "developer@hrs-id.com",
    "password": "rahasia-banget"
}
```

Responses:

```bash
{
	"status": "success",
	"user": {
		"id": "81cd0f87-69a6-41c1-a93f-44baef34182d",
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

It will generate cookie for 30 days.

### My Apps

This will return of apps that available for the users.

```bash
GET /api/v1/my-apps
```

Responses:

```bash
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

### Logout

```bash
POST /api/v1/logout
```

Responses:

```bash
{
	"message": "logged out successfully"
}
```
